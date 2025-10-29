/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/record"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
	"github.com/kubernetes-tenants/tenant-operator/internal/apply"
	"github.com/kubernetes-tenants/tenant-operator/internal/graph"
	"github.com/kubernetes-tenants/tenant-operator/internal/metrics"
	"github.com/kubernetes-tenants/tenant-operator/internal/readiness"
	"github.com/kubernetes-tenants/tenant-operator/internal/template"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	// Annotation key for tracking Once creation policy
	AnnotationCreatedOnce = "kubernetes-tenants.org/created-once"

	// Finalizer for tenant cleanup
	TenantFinalizer = "tenant.operator.kubernetes-tenants.org/finalizer"
)

// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenants/finalizers,verbs=update
// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenanttemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenantregistries,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts;services;configmaps;secrets;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments;statefulsets;daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs;cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile applies all resources for a tenant
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()

	// Fetch Tenant
	tenant := &tenantsv1.Tenant{}
	if err := r.Get(ctx, req.NamespacedName, tenant); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Tenant")
		metrics.TenantReconcileDuration.WithLabelValues("error").Observe(time.Since(startTime).Seconds())
		return ctrl.Result{}, err
	}

	// Handle deletion with finalizer
	if !tenant.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(tenant, TenantFinalizer) {
			// Perform cleanup with deletion policies
			if err := r.cleanupTenantResources(ctx, tenant); err != nil {
				logger.Error(err, "Failed to cleanup tenant resources")
				metrics.TenantReconcileDuration.WithLabelValues("error").Observe(time.Since(startTime).Seconds())
				return ctrl.Result{}, err
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(tenant, TenantFinalizer)
			if err := r.Update(ctx, tenant); err != nil {
				logger.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}

			logger.Info("Tenant cleanup completed", "tenant", tenant.Name)
			metrics.TenantReconcileDuration.WithLabelValues("success").Observe(time.Since(startTime).Seconds())
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(tenant, TenantFinalizer) {
		controllerutil.AddFinalizer(tenant, TenantFinalizer)
		if err := r.Update(ctx, tenant); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		logger.Info("Finalizer added to Tenant", "tenant", tenant.Name)
		// Requeue to continue with reconciliation
		return ctrl.Result{Requeue: true}, nil
	}

	// Build template variables from annotations
	vars, err := r.buildTemplateVariablesFromAnnotations(tenant)
	if err != nil {
		logger.Error(err, "Failed to build template variables")
		r.updateStatus(ctx, tenant, 0, 0, 0, false, "VariablesBuildError", err.Error())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Collect all resources from Tenant.Spec (already rendered by Registry controller)
	allResources := r.collectResourcesFromTenant(tenant)

	// Build dependency graph
	depGraph, err := graph.BuildGraph(allResources)
	if err != nil {
		logger.Error(err, "Failed to build dependency graph")
		r.updateStatus(ctx, tenant, 0, 0, 0, false, "DependencyError", err.Error())
		return ctrl.Result{}, err
	}

	// Get sorted resources
	sortedNodes, err := depGraph.TopologicalSort()
	if err != nil {
		logger.Error(err, "Failed to sort resources")
		r.updateStatus(ctx, tenant, 0, 0, 0, false, "SortError", err.Error())
		return ctrl.Result{}, err
	}

	// Apply resources
	applier := apply.NewApplier(r.Client, r.Scheme)
	checker := readiness.NewChecker(r.Client)
	templateEngine := template.NewEngine()

	var readyCount, failedCount int32
	totalResources := int32(len(sortedNodes))

	for _, node := range sortedNodes {
		resource := node.Resource

		// Render templates
		obj, err := r.renderResource(ctx, templateEngine, resource, vars, tenant)
		if err != nil {
			logger.Error(err, "Failed to render resource", "id", resource.ID)
			r.Recorder.Eventf(tenant, corev1.EventTypeWarning, "TemplateRenderError",
				"Failed to render resource %s: %v", resource.ID, err)
			failedCount++
			continue
		}

		// Handle CreationPolicy.Once
		if resource.CreationPolicy == tenantsv1.CreationPolicyOnce {
			// Check if resource already exists and has the "created-once" annotation
			exists, hasAnnotation, err := r.checkOnceCreated(ctx, obj)
			if err != nil {
				logger.Error(err, "Failed to check Once policy", "id", resource.ID)
				failedCount++
				continue
			}

			if exists && hasAnnotation {
				// Resource already created with Once policy, skip
				logger.Info("Skipping resource (CreationPolicy=Once, already created)", "id", resource.ID, "name", obj.GetName())
				readyCount++ // Count as ready since it exists
				continue
			}

			// Add annotation to track that this was created with Once policy
			annotations := obj.GetAnnotations()
			if annotations == nil {
				annotations = make(map[string]string)
			}
			annotations[AnnotationCreatedOnce] = "true"
			obj.SetAnnotations(annotations)
		}

		// Apply resource with specified patch strategy
		applyErr := applier.ApplyResource(ctx, obj, tenant, resource.ConflictPolicy, resource.PatchStrategy)

		// Record apply metrics
		kind := obj.GetKind()
		if kind == "" {
			kind = "Unknown"
		}
		applyResult := "success"
		if applyErr != nil {
			applyResult = "error"
		}
		metrics.ApplyAttemptsTotal.WithLabelValues(kind, applyResult, string(resource.ConflictPolicy)).Inc()

		if applyErr != nil {
			logger.Error(applyErr, "Failed to apply resource", "id", resource.ID)

			// Emit event based on error type
			if errors.IsConflict(applyErr) {
				r.Recorder.Eventf(tenant, corev1.EventTypeWarning, "ResourceConflict",
					"Resource %s conflict (policy=%s): %v", resource.ID, resource.ConflictPolicy, applyErr)
			} else {
				r.Recorder.Eventf(tenant, corev1.EventTypeWarning, "ApplyFailed",
					"Failed to apply resource %s: %v", resource.ID, applyErr)
			}

			failedCount++
			continue
		}

		// Wait for readiness if required
		if resource.WaitForReady != nil && *resource.WaitForReady {
			timeout := time.Duration(resource.TimeoutSeconds) * time.Second
			if timeout == 0 {
				timeout = 300 * time.Second
			}

			name := obj.GetName()
			namespace := obj.GetNamespace()

			if err := checker.WaitForReady(ctx, name, namespace, obj, timeout); err != nil {
				logger.Error(err, "Resource not ready within timeout", "id", resource.ID)
				r.Recorder.Eventf(tenant, corev1.EventTypeWarning, "ReadinessTimeout",
					"Resource %s not ready within %v: %v", resource.ID, timeout, err)
				failedCount++
				continue
			}
		}

		readyCount++
	}

	// Update status
	r.updateStatus(ctx, tenant, totalResources, readyCount, failedCount, failedCount == 0, "Reconciled", "Successfully reconciled all resources")

	// Record metrics
	result := "success"
	if failedCount > 0 {
		result = "partial_failure"
	}
	metrics.TenantReconcileDuration.WithLabelValues(result).Observe(time.Since(startTime).Seconds())
	metrics.TenantResourcesReady.WithLabelValues(tenant.Name, tenant.Namespace).Set(float64(readyCount))
	metrics.TenantResourcesDesired.WithLabelValues(tenant.Name, tenant.Namespace).Set(float64(totalResources))
	metrics.TenantResourcesFailed.WithLabelValues(tenant.Name, tenant.Namespace).Set(float64(failedCount))

	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// checkOnceCreated checks if a resource exists and has the "created-once" annotation
func (r *TenantReconciler) checkOnceCreated(ctx context.Context, obj *unstructured.Unstructured) (exists bool, hasAnnotation bool, err error) {
	// Try to get the resource
	current := obj.DeepCopy()
	key := client.ObjectKey{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	err = r.Get(ctx, key, current)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, false, nil
		}
		return false, false, err
	}

	// Resource exists, check for annotation
	annotations := current.GetAnnotations()
	if annotations != nil && annotations[AnnotationCreatedOnce] == "true" {
		return true, true, nil
	}

	return true, false, nil
}

// buildTemplateVariablesFromAnnotations builds template variables from Tenant annotations
func (r *TenantReconciler) buildTemplateVariablesFromAnnotations(tenant *tenantsv1.Tenant) (template.Variables, error) {
	// Get required values from annotations
	hostOrURL := tenant.Annotations["kubernetes-tenants.org/hostOrUrl"]
	if hostOrURL == "" {
		hostOrURL = tenant.Spec.UID
	}

	activate := tenant.Annotations["kubernetes-tenants.org/activate"]
	if activate == "" {
		activate = "true"
	}

	// Parse extra values from JSON
	extraJSON := tenant.Annotations["kubernetes-tenants.org/extra"]
	extraValues := make(map[string]string)
	if extraJSON != "" {
		if err := json.Unmarshal([]byte(extraJSON), &extraValues); err != nil {
			return nil, fmt.Errorf("failed to unmarshal extra values: %w", err)
		}
	}

	return template.BuildVariables(tenant.Spec.UID, hostOrURL, activate, extraValues), nil
}

// collectResourcesFromTenant collects all resources from Tenant.Spec
func (r *TenantReconciler) collectResourcesFromTenant(tenant *tenantsv1.Tenant) []tenantsv1.TResource {
	var resources []tenantsv1.TResource

	resources = append(resources, tenant.Spec.ServiceAccounts...)
	resources = append(resources, tenant.Spec.Deployments...)
	resources = append(resources, tenant.Spec.StatefulSets...)
	resources = append(resources, tenant.Spec.Services...)
	resources = append(resources, tenant.Spec.Ingresses...)
	resources = append(resources, tenant.Spec.ConfigMaps...)
	resources = append(resources, tenant.Spec.Secrets...)
	resources = append(resources, tenant.Spec.PersistentVolumeClaims...)
	resources = append(resources, tenant.Spec.Jobs...)
	resources = append(resources, tenant.Spec.CronJobs...)
	resources = append(resources, tenant.Spec.Manifests...)

	return resources
}

// renderResource renders a resource template
// Note: NameTemplate, LabelsTemplate, AnnotationsTemplate are already rendered by Registry controller
// We only need to render the spec (unstructured.Unstructured) contents which may contain template variables
func (r *TenantReconciler) renderResource(ctx context.Context, engine *template.Engine, resource tenantsv1.TResource, vars template.Variables, tenant *tenantsv1.Tenant) (*unstructured.Unstructured, error) {
	// Get spec (already an unstructured.Unstructured)
	obj := resource.Spec.DeepCopy()

	// Set metadata (use already-rendered values from resource)
	if resource.NameTemplate != "" {
		obj.SetName(resource.NameTemplate)
	}

	// Set namespace to Tenant CR's namespace
	// All resources are created in the same namespace as the Tenant CR
	obj.SetNamespace(tenant.Namespace)

	if len(resource.LabelsTemplate) > 0 {
		obj.SetLabels(resource.LabelsTemplate)
	}
	if len(resource.AnnotationsTemplate) > 0 {
		obj.SetAnnotations(resource.AnnotationsTemplate)
	}

	// Render spec recursively (for template variables inside the unstructured object)
	renderedSpec, err := r.renderUnstructured(ctx, obj.Object, engine, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render spec: %w", err)
	}
	obj.Object = renderedSpec

	return obj, nil
}

// renderUnstructured recursively renders template variables in unstructured data
func (r *TenantReconciler) renderUnstructured(ctx context.Context, data map[string]interface{}, engine *template.Engine, vars template.Variables) (map[string]interface{}, error) {
	logger := log.FromContext(ctx)
	result := make(map[string]interface{})

	for k, v := range data {
		switch val := v.(type) {
		case string:
			// Try to render as template
			rendered, err := engine.Render(val, vars)
			if err != nil {
				// Log warning but keep original value to allow reconciliation to continue
				logger.V(1).Info("Template rendering failed for field, keeping original value",
					"field", k,
					"template", val,
					"error", err.Error())
				result[k] = val
			} else {
				result[k] = rendered
			}
		case map[string]interface{}:
			// Recurse into nested maps
			rendered, err := r.renderUnstructured(ctx, val, engine, vars)
			if err != nil {
				logger.V(1).Info("Template rendering failed for nested object, keeping original",
					"field", k,
					"error", err.Error())
				result[k] = val
			} else {
				result[k] = rendered
			}
		case []interface{}:
			// Recurse into arrays
			renderedArray := make([]interface{}, len(val))
			for i, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					rendered, err := r.renderUnstructured(ctx, itemMap, engine, vars)
					if err != nil {
						logger.V(1).Info("Template rendering failed for array item, keeping original",
							"field", k,
							"index", i,
							"error", err.Error())
						renderedArray[i] = item
					} else {
						renderedArray[i] = rendered
					}
				} else if itemStr, ok := item.(string); ok {
					rendered, err := engine.Render(itemStr, vars)
					if err != nil {
						logger.V(1).Info("Template rendering failed for array string, keeping original",
							"field", k,
							"index", i,
							"template", itemStr,
							"error", err.Error())
						renderedArray[i] = item
					} else {
						renderedArray[i] = rendered
					}
				} else {
					renderedArray[i] = item
				}
			}
			result[k] = renderedArray
		default:
			result[k] = v
		}
	}

	return result, nil
}

// updateStatus updates Tenant status
func (r *TenantReconciler) updateStatus(ctx context.Context, tenant *tenantsv1.Tenant, desired, ready, failed int32, success bool, reason, message string) {
	tenant.Status.DesiredResources = desired
	tenant.Status.ReadyResources = ready
	tenant.Status.FailedResources = failed
	tenant.Status.ObservedGeneration = tenant.Generation

	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}
	if !success {
		condition.Status = metav1.ConditionFalse
	}

	// Update or append condition
	found := false
	for i := range tenant.Status.Conditions {
		if tenant.Status.Conditions[i].Type == condition.Type {
			tenant.Status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		tenant.Status.Conditions = append(tenant.Status.Conditions, condition)
	}

	if err := r.Status().Update(ctx, tenant); err != nil {
		log.FromContext(ctx).Error(err, "Failed to update Tenant status")
	}
}

// cleanupTenantResources handles resource cleanup according to DeletionPolicy
func (r *TenantReconciler) cleanupTenantResources(ctx context.Context, tenant *tenantsv1.Tenant) error {
	logger := log.FromContext(ctx)
	logger.Info("Starting tenant resource cleanup", "tenant", tenant.Name)

	applier := apply.NewApplier(r.Client, r.Scheme)
	templateEngine := template.NewEngine()

	// Build template variables from annotations
	vars, err := r.buildTemplateVariablesFromAnnotations(tenant)
	if err != nil {
		logger.Error(err, "Failed to build template variables for cleanup")
		// Continue with cleanup even if variables fail
		vars = template.Variables{}
	}

	// Collect all resources
	allResources := r.collectResourcesFromTenant(tenant)

	// Process each resource according to its DeletionPolicy
	for _, res := range allResources {
		// Render resource to get actual name/namespace
		rendered, err := r.renderResource(ctx, templateEngine, res, vars, tenant)
		if err != nil {
			logger.Error(err, "Failed to render resource for cleanup",
				"resource", res.ID,
				"kind", res.Spec.GetKind())
			// Continue with other resources
			continue
		}

		resourceName := rendered.GetName()
		resourceKind := rendered.GetKind()

		// Apply deletion policy
		switch res.DeletionPolicy {
		case tenantsv1.DeletionPolicyRetain:
			// Remove ownerReferences but keep resource
			logger.Info("Retaining resource (removing ownerReferences)",
				"resource", resourceName,
				"kind", resourceKind,
				"namespace", rendered.GetNamespace())

			if err := applier.DeleteResource(ctx, rendered, tenantsv1.DeletionPolicyRetain); err != nil {
				logger.Error(err, "Failed to retain resource",
					"resource", resourceName,
					"kind", resourceKind)
				// Continue with other resources
			} else {
				r.Recorder.Eventf(tenant, corev1.EventTypeNormal, "ResourceRetained",
					"Resource %s/%s retained (ownerReferences removed)", resourceKind, resourceName)
			}

		case tenantsv1.DeletionPolicyDelete, "":
			// Delete resource (default behavior)
			logger.V(1).Info("Deleting resource",
				"resource", resourceName,
				"kind", resourceKind,
				"namespace", rendered.GetNamespace())

			if err := applier.DeleteResource(ctx, rendered, tenantsv1.DeletionPolicyDelete); err != nil {
				// Log error but continue - ownerReferences will handle cleanup
				logger.V(1).Info("Resource deletion delegated to ownerReference garbage collection",
					"resource", resourceName,
					"kind", resourceKind,
					"error", err.Error())
			}
		}
	}

	logger.Info("Tenant resource cleanup completed", "tenant", tenant.Name, "resources", len(allResources))
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantsv1.Tenant{}).
		Named("tenant").
		// Watch owned resources for drift detection
		// When these resources are modified, the parent Tenant will be reconciled
		// Note: Namespace is not included here because it's cluster-scoped and
		// cannot have namespace-scoped owners. Instead, we use labels for tracking.
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&batchv1.Job{}).
		Owns(&batchv1.CronJob{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
