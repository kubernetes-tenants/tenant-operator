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
	errorsStd "errors"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"

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
	// Annotation value for created resources
	AnnotationValueTrue = "true"

	// Finalizer for tenant cleanup
	TenantFinalizer = "tenant.operator.kubernetes-tenants.org/finalizer"

	// Condition types
	ConditionTypeReady = "Ready"

	// Resource formatting
	NoResourcesMessage = "no resources"
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
// NOTE: Cross-namespace resource support requires cluster-wide permissions for resource types
// The above RBAC rules allow the operator to create resources in any namespace when targetNamespace is specified

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
		r.updateDegradedCondition(ctx, tenant, true, "VariablesBuildError", err.Error())
		metrics.TenantDegradedStatus.WithLabelValues(tenant.Name, tenant.Namespace, "VariablesBuildError").Set(1)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Collect all resources from Tenant.Spec (already rendered by Registry controller)
	allResources := r.collectResourcesFromTenant(tenant)

	// Build dependency graph
	depGraph, err := graph.BuildGraph(allResources)
	if err != nil {
		logger.Error(err, "Failed to build dependency graph")
		r.updateStatus(ctx, tenant, 0, 0, 0, false, "DependencyError", err.Error())
		r.updateDegradedCondition(ctx, tenant, true, "DependencyCycle", "Dependency cycle detected in resource graph")
		metrics.TenantDegradedStatus.WithLabelValues(tenant.Name, tenant.Namespace, "DependencyCycle").Set(1)
		return ctrl.Result{}, err
	}

	// Get sorted resources
	sortedNodes, err := depGraph.TopologicalSort()
	if err != nil {
		logger.Error(err, "Failed to sort resources")
		r.updateStatus(ctx, tenant, 0, 0, 0, false, "SortError", err.Error())
		r.updateDegradedCondition(ctx, tenant, true, "DependencyCycle", err.Error())
		metrics.TenantDegradedStatus.WithLabelValues(tenant.Name, tenant.Namespace, "DependencyCycle").Set(1)
		return ctrl.Result{}, err
	}

	// Detect and cleanup orphaned resources (resources removed from template)
	// Build current desired resource keys
	currentKeys, err := r.buildAppliedResourceKeys(ctx, tenant)
	if err != nil {
		logger.Error(err, "Failed to build applied resource keys")
		// Continue with reconciliation even if orphan detection fails
		currentKeys = make(map[string]bool)
	}

	// Get previously applied resource keys from status
	previousKeys := tenant.Status.AppliedResources

	// Find orphaned resources
	orphanedKeys := r.findOrphanedResources(previousKeys, currentKeys)

	// Delete orphaned resources
	if len(orphanedKeys) > 0 {
		logger.Info("Found orphaned resources", "count", len(orphanedKeys))
		for _, orphanKey := range orphanedKeys {
			if err := r.deleteOrphanedResource(ctx, tenant, orphanKey); err != nil {
				logger.Error(err, "Failed to delete orphaned resource", "key", orphanKey)
				// Continue with other orphans even if one fails
			}
		}
	}

	// Apply resources and track changes
	readyCount, failedCount, changedCount, conflictedCount := r.applyResources(ctx, tenant, sortedNodes, vars)
	totalResources := int32(len(sortedNodes))

	// Update Conflicted condition based on conflict detection
	hasConflict := conflictedCount > 0
	r.updateConflictedCondition(ctx, tenant, hasConflict)

	// Determine if tenant is degraded (template errors, dependency errors, or conflicts)
	isDegraded := failedCount > 0 || hasConflict
	var degradedReason string
	if failedCount > 0 && hasConflict {
		degradedReason = "ResourceFailuresAndConflicts"
	} else if failedCount > 0 {
		degradedReason = "ResourceFailures"
	} else if hasConflict {
		degradedReason = "ResourceConflicts"
	}

	// Update Degraded condition
	if isDegraded {
		r.updateDegradedCondition(ctx, tenant, true, degradedReason, fmt.Sprintf("Tenant has %d failed and %d conflicted resources", failedCount, conflictedCount))
	} else {
		r.updateDegradedCondition(ctx, tenant, false, "Healthy", "All resources are healthy")
	}

	// Always update status after reconciliation with actual counts
	// This ensures status reflects reality without unnecessary resets
	r.updateStatus(ctx, tenant, totalResources, readyCount, failedCount, failedCount == 0, "Reconciled", "Successfully reconciled all resources")

	// Update AppliedResources in status to enable orphan detection
	// Convert map keys to slice
	appliedResourceKeys := make([]string, 0, len(currentKeys))
	for key := range currentKeys {
		appliedResourceKeys = append(appliedResourceKeys, key)
	}

	// Update status with applied resource keys using retry
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get fresh copy
		fresh := &tenantsv1.Tenant{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(tenant), fresh); err != nil {
			return err
		}

		// Update AppliedResources field
		fresh.Status.AppliedResources = appliedResourceKeys

		// Update status
		return r.Status().Update(ctx, fresh)
	})

	if err != nil {
		logger.Error(err, "Failed to update AppliedResources in status")
		// Non-fatal error, continue with reconciliation
	}

	// Emit completion event if resources were changed
	if changedCount > 0 {
		r.emitTemplateAppliedCompleteEvent(ctx, tenant, totalResources, readyCount, failedCount, changedCount)
		logger.Info("Reconciliation completed with changes", "changed", changedCount, "ready", readyCount, "failed", failedCount, "conflicted", conflictedCount)
	} else {
		logger.V(1).Info("Reconciliation completed without changes", "ready", readyCount, "failed", failedCount, "conflicted", conflictedCount)
	}

	// Record metrics
	result := "success"
	if failedCount > 0 {
		result = "partial_failure"
	}
	metrics.TenantReconcileDuration.WithLabelValues(result).Observe(time.Since(startTime).Seconds())
	metrics.TenantResourcesReady.WithLabelValues(tenant.Name, tenant.Namespace).Set(float64(readyCount))
	metrics.TenantResourcesDesired.WithLabelValues(tenant.Name, tenant.Namespace).Set(float64(totalResources))
	metrics.TenantResourcesFailed.WithLabelValues(tenant.Name, tenant.Namespace).Set(float64(failedCount))
	metrics.TenantResourcesConflicted.WithLabelValues(tenant.Name, tenant.Namespace).Set(float64(conflictedCount))

	// Record condition status metrics
	r.recordConditionMetrics(ctx, tenant)

	// Record degraded status metric
	if isDegraded {
		metrics.TenantDegradedStatus.WithLabelValues(tenant.Name, tenant.Namespace, degradedReason).Set(1)
	} else {
		// Reset all degraded metrics for this tenant
		metrics.TenantDegradedStatus.WithLabelValues(tenant.Name, tenant.Namespace, "ResourceFailures").Set(0)
		metrics.TenantDegradedStatus.WithLabelValues(tenant.Name, tenant.Namespace, "ResourceConflicts").Set(0)
		metrics.TenantDegradedStatus.WithLabelValues(tenant.Name, tenant.Namespace, "ResourceFailuresAndConflicts").Set(0)
		metrics.TenantDegradedStatus.WithLabelValues(tenant.Name, tenant.Namespace, "TemplateRenderError").Set(0)
		metrics.TenantDegradedStatus.WithLabelValues(tenant.Name, tenant.Namespace, "DependencyCycle").Set(0)
	}

	// Requeue after 30 seconds for faster resource status reflection
	// This ensures that status changes in child resources are detected more quickly
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// applyResources applies all resources and returns counts for ready, failed, changed, and conflicted resources
func (r *TenantReconciler) applyResources(ctx context.Context, tenant *tenantsv1.Tenant, sortedNodes []*graph.Node, vars template.Variables) (readyCount, failedCount, changedCount, conflictedCount int32) {
	logger := log.FromContext(ctx)
	applier := apply.NewApplier(r.Client, r.Scheme)
	checker := readiness.NewChecker(r.Client)
	templateEngine := template.NewEngine()

	totalResources := int32(len(sortedNodes))
	progressingSet := false
	templateAppliedEventEmitted := false

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
			annotations[AnnotationCreatedOnce] = AnnotationValueTrue
			obj.SetAnnotations(annotations)
		}

		// Apply resource with specified patch strategy and track changes
		changed, applyErr := applier.ApplyResource(ctx, obj, tenant, resource.ConflictPolicy, resource.PatchStrategy)

		// Track changes and emit events on first change
		if changed {
			changedCount++

			// On first change, update Progressing condition and emit event
			if !progressingSet {
				r.updateProgressingCondition(ctx, tenant, true, "Reconciling", "Reconciling changed resources")
				progressingSet = true

				// Emit detailed template applied event on first resource change
				if !templateAppliedEventEmitted {
					r.emitTemplateAppliedEvent(ctx, tenant, totalResources)
					templateAppliedEventEmitted = true
				}
			}
		}

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

			// Check if this is a ConflictError
			var conflictErr *apply.ConflictError
			if errorsStd.As(applyErr, &conflictErr) {
				// Resource conflict detected
				conflictedCount++
				// Increment conflict counter metric
				metrics.TenantConflictsTotal.WithLabelValues(tenant.Name, tenant.Namespace, kind, string(resource.ConflictPolicy)).Inc()
				r.Recorder.Eventf(tenant, corev1.EventTypeWarning, "ResourceConflict",
					"Resource conflict detected for %s/%s (Kind: %s, Policy: %s). "+
						"Another controller or user may be managing this resource. "+
						"Consider using ConflictPolicy=Force to take ownership or resolve the conflict manually. Error: %v",
					conflictErr.Namespace, conflictErr.ResourceName, conflictErr.Kind, resource.ConflictPolicy, conflictErr.Err)
			} else {
				// Other apply error
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

	return readyCount, failedCount, changedCount, conflictedCount
}

// emitTemplateAppliedEvent emits a detailed event when template changes are being applied
func (r *TenantReconciler) emitTemplateAppliedEvent(ctx context.Context, tenant *tenantsv1.Tenant, totalResources int32) {
	logger := log.FromContext(ctx)

	// Get template information from tenant
	templateName := tenant.Spec.TemplateRef
	templateGeneration := tenant.Annotations["kubernetes-tenants.org/template-generation"]

	// Count resources by type
	resourceCounts := r.countTenantResourcesByType(tenant)
	resourceDetails := r.formatTenantResourceDetails(resourceCounts)

	// Get registry name from labels
	registryName := tenant.Labels["kubernetes-tenants.org/registry"]
	if registryName == "" {
		registryName = "unknown"
	}

	// Emit detailed event
	r.Recorder.Eventf(tenant, corev1.EventTypeNormal, "TemplateResourcesApplying",
		"Applying resources from TenantTemplate '%s' (generation: %s). "+
			"Reconciling %d total resources: %s. "+
			"Registry: %s, Tenant UID: %s, Namespace: %s. "+
			"Resources will be applied in dependency order with readiness checks.",
		templateName, templateGeneration,
		totalResources, resourceDetails,
		registryName, tenant.Spec.UID, tenant.Namespace)

	logger.Info("Applying template resources to cluster",
		"tenant", tenant.Name,
		"template", templateName,
		"generation", templateGeneration,
		"totalResources", totalResources,
		"registry", registryName)
}

// emitTemplateAppliedCompleteEvent emits a detailed completion event after template resources are applied
func (r *TenantReconciler) emitTemplateAppliedCompleteEvent(ctx context.Context, tenant *tenantsv1.Tenant, totalResources, readyCount, failedCount, changedCount int32) {
	logger := log.FromContext(ctx)

	// Get template information
	templateName := tenant.Spec.TemplateRef
	templateGeneration := tenant.Annotations["kubernetes-tenants.org/template-generation"]

	// Get registry name from labels
	registryName := tenant.Labels["kubernetes-tenants.org/registry"]
	if registryName == "" {
		registryName = "unknown"
	}

	// Determine event type and message based on results
	if failedCount > 0 {
		// Partial failure
		r.Recorder.Eventf(tenant, corev1.EventTypeWarning, "TemplateAppliedPartial",
			"Applied TenantTemplate '%s' (generation: %s) with partial success. "+
				"Changed: %d, Ready: %d, Failed: %d out of %d total resources. "+
				"Registry: %s, Tenant UID: %s. "+
				"Failed resources require attention.",
			templateName, templateGeneration,
			changedCount, readyCount, failedCount, totalResources,
			registryName, tenant.Spec.UID)

		logger.Error(nil, "Template application completed with failures",
			"tenant", tenant.Name,
			"template", templateName,
			"generation", templateGeneration,
			"changed", changedCount,
			"ready", readyCount,
			"failed", failedCount,
			"total", totalResources)
	} else {
		// Success
		r.Recorder.Eventf(tenant, corev1.EventTypeNormal, "TemplateAppliedSuccess",
			"Successfully applied TenantTemplate '%s' (generation: %s). "+
				"All %d resources reconciled successfully (%d changed, %d ready). "+
				"Registry: %s, Tenant UID: %s, Namespace: %s. "+
				"All resources are now in desired state.",
			templateName, templateGeneration,
			totalResources, changedCount, readyCount,
			registryName, tenant.Spec.UID, tenant.Namespace)

		logger.Info("Template application completed successfully",
			"tenant", tenant.Name,
			"template", templateName,
			"generation", templateGeneration,
			"changed", changedCount,
			"ready", readyCount,
			"total", totalResources)
	}
}

// countTenantResourcesByType counts resources by type in a Tenant
func (r *TenantReconciler) countTenantResourcesByType(tenant *tenantsv1.Tenant) map[string]int {
	counts := make(map[string]int)
	spec := &tenant.Spec

	if len(spec.ServiceAccounts) > 0 {
		counts["ServiceAccounts"] = len(spec.ServiceAccounts)
	}
	if len(spec.Deployments) > 0 {
		counts["Deployments"] = len(spec.Deployments)
	}
	if len(spec.StatefulSets) > 0 {
		counts["StatefulSets"] = len(spec.StatefulSets)
	}
	if len(spec.Services) > 0 {
		counts["Services"] = len(spec.Services)
	}
	if len(spec.Ingresses) > 0 {
		counts["Ingresses"] = len(spec.Ingresses)
	}
	if len(spec.ConfigMaps) > 0 {
		counts["ConfigMaps"] = len(spec.ConfigMaps)
	}
	if len(spec.Secrets) > 0 {
		counts["Secrets"] = len(spec.Secrets)
	}
	if len(spec.PersistentVolumeClaims) > 0 {
		counts["PVCs"] = len(spec.PersistentVolumeClaims)
	}
	if len(spec.Jobs) > 0 {
		counts["Jobs"] = len(spec.Jobs)
	}
	if len(spec.CronJobs) > 0 {
		counts["CronJobs"] = len(spec.CronJobs)
	}
	if len(spec.Manifests) > 0 {
		counts["Manifests"] = len(spec.Manifests)
	}

	return counts
}

// formatTenantResourceDetails formats resource counts into a readable string
func (r *TenantReconciler) formatTenantResourceDetails(counts map[string]int) string {
	var details []string

	if count, ok := counts["ServiceAccounts"]; ok {
		details = append(details, fmt.Sprintf("%d ServiceAccount(s)", count))
	}
	if count, ok := counts["Deployments"]; ok {
		details = append(details, fmt.Sprintf("%d Deployment(s)", count))
	}
	if count, ok := counts["StatefulSets"]; ok {
		details = append(details, fmt.Sprintf("%d StatefulSet(s)", count))
	}
	if count, ok := counts["Services"]; ok {
		details = append(details, fmt.Sprintf("%d Service(s)", count))
	}
	if count, ok := counts["Ingresses"]; ok {
		details = append(details, fmt.Sprintf("%d Ingress(es)", count))
	}
	if count, ok := counts["ConfigMaps"]; ok {
		details = append(details, fmt.Sprintf("%d ConfigMap(s)", count))
	}
	if count, ok := counts["Secrets"]; ok {
		details = append(details, fmt.Sprintf("%d Secret(s)", count))
	}
	if count, ok := counts["PVCs"]; ok {
		details = append(details, fmt.Sprintf("%d PVC(s)", count))
	}
	if count, ok := counts["Jobs"]; ok {
		details = append(details, fmt.Sprintf("%d Job(s)", count))
	}
	if count, ok := counts["CronJobs"]; ok {
		details = append(details, fmt.Sprintf("%d CronJob(s)", count))
	}
	if count, ok := counts["Manifests"]; ok {
		details = append(details, fmt.Sprintf("%d Manifest(s)", count))
	}

	if len(details) == 0 {
		return NoResourcesMessage
	}

	// Join all details with commas
	result := ""
	for i, detail := range details {
		if i > 0 {
			result += ", "
		}
		result += detail
	}
	return result
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
	if annotations != nil && annotations[AnnotationCreatedOnce] == AnnotationValueTrue {
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
		activate = AnnotationValueTrue
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
// Note: NameTemplate, LabelsTemplate, AnnotationsTemplate, TargetNamespace are already rendered by Registry controller
// We only need to render the spec (unstructured.Unstructured) contents which may contain template variables
func (r *TenantReconciler) renderResource(ctx context.Context, engine *template.Engine, resource tenantsv1.TResource, vars template.Variables, tenant *tenantsv1.Tenant) (*unstructured.Unstructured, error) {
	// Get spec (already an unstructured.Unstructured)
	obj := resource.Spec.DeepCopy()

	// Set metadata (use already-rendered values from resource)
	if resource.NameTemplate != "" {
		obj.SetName(resource.NameTemplate)
	}

	// Set namespace: use TargetNamespace if specified, otherwise use Tenant CR's namespace
	targetNamespace := tenant.Namespace
	if resource.TargetNamespace != "" {
		targetNamespace = resource.TargetNamespace
	}
	obj.SetNamespace(targetNamespace)

	// Set labels
	labels := resource.LabelsTemplate
	if labels == nil {
		labels = make(map[string]string)
	}

	// For cross-namespace resources or Namespaces, add tracking labels
	// since they cannot have ownerReferences
	isCrossNamespace := targetNamespace != tenant.Namespace
	isNamespaceResource := obj.GetKind() == "Namespace"
	if isCrossNamespace || isNamespaceResource {
		labels["kubernetes-tenants.org/tenant"] = tenant.Name
		labels["kubernetes-tenants.org/tenant-namespace"] = tenant.Namespace
	}

	if len(labels) > 0 {
		obj.SetLabels(labels)
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
//
//nolint:unparam // error return kept for future template rendering errors
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

// updateStatus updates Tenant status with retry on conflict
func (r *TenantReconciler) updateStatus(ctx context.Context, tenant *tenantsv1.Tenant, desired, ready, failed int32, success bool, reason, message string) {
	logger := log.FromContext(ctx)

	// Retry status update on conflict
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the latest version of the tenant
		key := client.ObjectKeyFromObject(tenant)
		latest := &tenantsv1.Tenant{}
		if err := r.Get(ctx, key, latest); err != nil {
			return err
		}

		// Update status fields
		latest.Status.DesiredResources = desired
		latest.Status.ReadyResources = ready
		latest.Status.FailedResources = failed
		latest.Status.ObservedGeneration = latest.Generation

		// Prepare Ready condition
		readyCondition := metav1.Condition{
			Type:               ConditionTypeReady,
			Status:             metav1.ConditionTrue,
			Reason:             reason,
			Message:            message,
			LastTransitionTime: metav1.Now(),
		}
		if !success {
			readyCondition.Status = metav1.ConditionFalse
		}

		// Update or append Ready condition
		foundReady := false
		for i := range latest.Status.Conditions {
			if latest.Status.Conditions[i].Type == ConditionTypeReady {
				latest.Status.Conditions[i] = readyCondition
				foundReady = true
				break
			}
		}
		if !foundReady {
			latest.Status.Conditions = append(latest.Status.Conditions, readyCondition)
		}

		// Update status subresource
		return r.Status().Update(ctx, latest)
	})

	if err != nil {
		logger.Error(err, "Failed to update Tenant status after retries")
	}
}

// updateProgressingCondition updates only the Progressing condition without touching other status fields
func (r *TenantReconciler) updateProgressingCondition(ctx context.Context, tenant *tenantsv1.Tenant, progressing bool, reason, message string) {
	logger := log.FromContext(ctx)

	// Retry status update on conflict
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the latest version of the tenant
		key := client.ObjectKeyFromObject(tenant)
		latest := &tenantsv1.Tenant{}
		if err := r.Get(ctx, key, latest); err != nil {
			return err
		}

		// Prepare Progressing condition
		progressingCondition := metav1.Condition{
			Type:               "Progressing",
			Status:             metav1.ConditionFalse,
			Reason:             "ReconcileComplete",
			Message:            "Reconciliation completed",
			LastTransitionTime: metav1.Now(),
		}

		if progressing {
			progressingCondition.Status = metav1.ConditionTrue
			progressingCondition.Reason = reason
			progressingCondition.Message = message
		}

		// Update or append Progressing condition
		found := false
		for i := range latest.Status.Conditions {
			if latest.Status.Conditions[i].Type == "Progressing" {
				// Only update if status actually changed to avoid unnecessary writes
				if latest.Status.Conditions[i].Status != progressingCondition.Status {
					latest.Status.Conditions[i] = progressingCondition
					found = true
					break
				}
				// No change needed
				return nil
			}
		}
		if !found {
			latest.Status.Conditions = append(latest.Status.Conditions, progressingCondition)
		}

		// Update status subresource
		return r.Status().Update(ctx, latest)
	})

	if err != nil {
		logger.Error(err, "Failed to update Progressing condition after retries")
	}
}

// updateConflictedCondition updates the Conflicted condition based on conflict detection
func (r *TenantReconciler) updateConflictedCondition(ctx context.Context, tenant *tenantsv1.Tenant, hasConflict bool) {
	logger := log.FromContext(ctx)

	// Retry status update on conflict
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the latest version of the tenant
		key := client.ObjectKeyFromObject(tenant)
		latest := &tenantsv1.Tenant{}
		if err := r.Get(ctx, key, latest); err != nil {
			return err
		}

		// Prepare Conflicted condition
		conflictedCondition := metav1.Condition{
			Type:               "Conflicted",
			Status:             metav1.ConditionFalse,
			Reason:             "NoConflict",
			Message:            "No resource conflicts detected",
			LastTransitionTime: metav1.Now(),
		}

		if hasConflict {
			conflictedCondition.Status = metav1.ConditionTrue
			conflictedCondition.Reason = "ResourceConflict"
			conflictedCondition.Message = "One or more resources are in conflict. Check events for details."
		}

		// Update or append Conflicted condition
		foundConflicted := false
		for i := range latest.Status.Conditions {
			if latest.Status.Conditions[i].Type == "Conflicted" {
				// Only update if the status changed to avoid unnecessary updates
				if latest.Status.Conditions[i].Status != conflictedCondition.Status {
					latest.Status.Conditions[i] = conflictedCondition
					foundConflicted = true
					break
				}
				// Status hasn't changed, no update needed
				return nil
			}
		}
		if !foundConflicted {
			latest.Status.Conditions = append(latest.Status.Conditions, conflictedCondition)
		}

		// Update status subresource
		return r.Status().Update(ctx, latest)
	})

	if err != nil {
		logger.Error(err, "Failed to update Conflicted condition after retries")
	}
}

// updateDegradedCondition updates the Degraded condition based on tenant health
func (r *TenantReconciler) updateDegradedCondition(ctx context.Context, tenant *tenantsv1.Tenant, isDegraded bool, reason, message string) {
	logger := log.FromContext(ctx)

	// Retry status update on conflict
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the latest version of the tenant
		key := client.ObjectKeyFromObject(tenant)
		latest := &tenantsv1.Tenant{}
		if err := r.Get(ctx, key, latest); err != nil {
			return err
		}

		// Prepare Degraded condition
		degradedCondition := metav1.Condition{
			Type:               "Degraded",
			Status:             metav1.ConditionFalse,
			Reason:             reason,
			Message:            message,
			LastTransitionTime: metav1.Now(),
		}

		if isDegraded {
			degradedCondition.Status = metav1.ConditionTrue
		}

		// Update or append Degraded condition
		foundDegraded := false
		for i := range latest.Status.Conditions {
			if latest.Status.Conditions[i].Type == "Degraded" {
				// Only update if the status changed to avoid unnecessary updates
				if latest.Status.Conditions[i].Status != degradedCondition.Status {
					latest.Status.Conditions[i] = degradedCondition
					foundDegraded = true
					break
				}
				// Status hasn't changed, no update needed
				return nil
			}
		}
		if !foundDegraded {
			latest.Status.Conditions = append(latest.Status.Conditions, degradedCondition)
		}

		// Update status subresource
		return r.Status().Update(ctx, latest)
	})

	if err != nil {
		logger.Error(err, "Failed to update Degraded condition after retries")
	}
}

// recordConditionMetrics records condition status metrics for all tenant conditions
func (r *TenantReconciler) recordConditionMetrics(ctx context.Context, tenant *tenantsv1.Tenant) {
	// Fetch the latest tenant to get current conditions
	key := client.ObjectKeyFromObject(tenant)
	latest := &tenantsv1.Tenant{}
	if err := r.Get(ctx, key, latest); err != nil {
		return
	}

	// Track which condition types we've seen
	conditionsSeen := make(map[string]bool)

	// Record metrics for each condition
	for _, condition := range latest.Status.Conditions {
		conditionsSeen[condition.Type] = true

		var statusValue float64
		switch condition.Status {
		case metav1.ConditionTrue:
			statusValue = 1
		case metav1.ConditionFalse:
			statusValue = 0
		case metav1.ConditionUnknown:
			statusValue = 2
		default:
			statusValue = 2 // Unknown
		}

		metrics.TenantConditionStatus.WithLabelValues(
			tenant.Name,
			tenant.Namespace,
			condition.Type,
		).Set(statusValue)
	}

	// Ensure all expected condition types are recorded (default to Unknown if missing)
	expectedConditions := []string{"Ready", "Progressing", "Conflicted", "Degraded"}
	for _, condType := range expectedConditions {
		if !conditionsSeen[condType] {
			metrics.TenantConditionStatus.WithLabelValues(
				tenant.Name,
				tenant.Namespace,
				condType,
			).Set(2) // Unknown
		}
	}
}

// cleanupTenantResources handles resource cleanup according to DeletionPolicy
//
//nolint:unparam // error return kept for future cleanup error handling
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
	// Create predicates for owned resources to reduce unnecessary reconciliations
	// Only trigger reconciliation on Generation changes (spec updates) or status updates
	ownedResourcePredicate := predicate.Or(
		predicate.GenerationChangedPredicate{},
		predicate.AnnotationChangedPredicate{},
	)

	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantsv1.Tenant{}).
		Named("tenant").
		// Watch owned resources for drift detection with predicates (same-namespace with ownerReference)
		// When these resources are modified, the parent Tenant will be reconciled
		// Predicates ensure we only react to meaningful changes (generation/annotations)
		Owns(&corev1.ServiceAccount{}, builder.WithPredicates(ownedResourcePredicate)).
		Owns(&corev1.Service{}, builder.WithPredicates(ownedResourcePredicate)).
		Owns(&corev1.ConfigMap{}, builder.WithPredicates(ownedResourcePredicate)).
		Owns(&corev1.Secret{}, builder.WithPredicates(ownedResourcePredicate)).
		Owns(&corev1.PersistentVolumeClaim{}, builder.WithPredicates(ownedResourcePredicate)).
		Owns(&appsv1.Deployment{}, builder.WithPredicates(ownedResourcePredicate)).
		Owns(&appsv1.StatefulSet{}, builder.WithPredicates(ownedResourcePredicate)).
		Owns(&appsv1.DaemonSet{}, builder.WithPredicates(ownedResourcePredicate)).
		Owns(&batchv1.Job{}, builder.WithPredicates(ownedResourcePredicate)).
		Owns(&batchv1.CronJob{}, builder.WithPredicates(ownedResourcePredicate)).
		Owns(&networkingv1.Ingress{}, builder.WithPredicates(ownedResourcePredicate)).
		// Watch resources with label-based tracking (cross-namespace or resources without ownerReference support)
		// These use labels for tracking: kubernetes-tenants.org/tenant and kubernetes-tenants.org/tenant-namespace
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Watches(
			&corev1.ServiceAccount{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Watches(
			&corev1.Service{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Watches(
			&corev1.PersistentVolumeClaim{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Watches(
			&appsv1.Deployment{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Watches(
			&appsv1.StatefulSet{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Watches(
			&appsv1.DaemonSet{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Watches(
			&batchv1.Job{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Watches(
			&batchv1.CronJob{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Watches(
			&networkingv1.Ingress{},
			handler.EnqueueRequestsFromMapFunc(r.findTenantForLabeledResource),
			builder.WithPredicates(ownedResourcePredicate),
		).
		Complete(r)
}

// findTenantForLabeledResource maps any resource to its Tenant using tracking labels
// This supports cross-namespace resources and resources without ownerReference support (like Namespaces)
func (r *TenantReconciler) findTenantForLabeledResource(ctx context.Context, obj client.Object) []ctrl.Request {
	// Check if this resource has our tracking labels
	labels := obj.GetLabels()
	if labels == nil {
		return nil
	}

	tenantName := labels["kubernetes-tenants.org/tenant"]
	tenantNamespace := labels["kubernetes-tenants.org/tenant-namespace"]

	if tenantName == "" || tenantNamespace == "" {
		return nil
	}

	return []ctrl.Request{
		{
			NamespacedName: client.ObjectKey{
				Name:      tenantName,
				Namespace: tenantNamespace,
			},
		},
	}
}

// buildResourceKey generates a unique key for a resource
// Format: "kind/namespace/name@id"
// Example: "Deployment/default/myapp@app-deployment"
func buildResourceKey(obj *unstructured.Unstructured, resourceID string) string {
	kind := obj.GetKind()
	namespace := obj.GetNamespace()
	name := obj.GetName()
	return fmt.Sprintf("%s/%s/%s@%s", kind, namespace, name, resourceID)
}

// parseResourceKey parses a resource key into its components
// Returns: kind, namespace, name, resourceID, error
func parseResourceKey(key string) (string, string, string, string, error) {
	// Split by '@' first to separate resource ID
	parts := strings.Split(key, "@")
	if len(parts) != 2 {
		return "", "", "", "", fmt.Errorf("invalid resource key format: %s (expected format: kind/namespace/name@id)", key)
	}
	resourceID := parts[1]

	// Split the first part by '/' to get kind/namespace/name
	resourceParts := strings.Split(parts[0], "/")
	if len(resourceParts) != 3 {
		return "", "", "", "", fmt.Errorf("invalid resource key format: %s (expected format: kind/namespace/name@id)", key)
	}

	kind := resourceParts[0]
	namespace := resourceParts[1]
	name := resourceParts[2]

	return kind, namespace, name, resourceID, nil
}

// buildAppliedResourceKeys builds a set of resource keys from current Tenant.Spec
func (r *TenantReconciler) buildAppliedResourceKeys(ctx context.Context, tenant *tenantsv1.Tenant) (map[string]bool, error) {
	keys := make(map[string]bool)
	templateEngine := template.NewEngine()

	// Build template variables
	vars, err := r.buildTemplateVariablesFromAnnotations(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to build template variables: %w", err)
	}

	// Collect all resources
	allResources := r.collectResourcesFromTenant(tenant)

	// Render each resource and build key
	for _, res := range allResources {
		rendered, err := r.renderResource(ctx, templateEngine, res, vars, tenant)
		if err != nil {
			// Skip resources that fail to render (they won't be applied)
			continue
		}

		key := buildResourceKey(rendered, res.ID)
		keys[key] = true
	}

	return keys, nil
}

// findOrphanedResources finds resources that were previously applied but are no longer in the spec
func (r *TenantReconciler) findOrphanedResources(previousKeys []string, currentKeys map[string]bool) []string {
	var orphans []string

	for _, prevKey := range previousKeys {
		if !currentKeys[prevKey] {
			orphans = append(orphans, prevKey)
		}
	}

	return orphans
}

// deleteOrphanedResource deletes a resource identified by its key
func (r *TenantReconciler) deleteOrphanedResource(ctx context.Context, tenant *tenantsv1.Tenant, key string) error {
	logger := log.FromContext(ctx)

	// Parse the key
	kind, namespace, name, resourceID, err := parseResourceKey(key)
	if err != nil {
		logger.Error(err, "Failed to parse resource key", "key", key)
		return err
	}

	// Find the resource definition in the tenant spec to get DeletionPolicy
	// If not found, default to Delete policy
	deletionPolicy := tenantsv1.DeletionPolicyDelete
	allResources := r.collectResourcesFromTenant(tenant)
	for _, res := range allResources {
		if res.ID == resourceID {
			if res.DeletionPolicy != "" {
				deletionPolicy = res.DeletionPolicy
			}
			break
		}
	}

	// If DeletionPolicy is Retain, don't delete the orphaned resource
	if deletionPolicy == tenantsv1.DeletionPolicyRetain {
		logger.V(1).Info("Orphaned resource has DeletionPolicy=Retain, skipping deletion",
			"key", key,
			"kind", kind,
			"namespace", namespace,
			"name", name)
		return nil
	}

	// Create an unstructured object to represent the resource
	obj := &unstructured.Unstructured{}
	obj.SetKind(kind)
	obj.SetNamespace(namespace)
	obj.SetName(name)

	// Set appropriate API version based on kind
	apiVersion := r.getAPIVersionForKind(kind)
	obj.SetAPIVersion(apiVersion)

	// Delete the resource
	applier := apply.NewApplier(r.Client, r.Scheme)
	if err := applier.DeleteResource(ctx, obj, tenantsv1.DeletionPolicyDelete); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Failed to delete orphaned resource",
				"key", key,
				"kind", kind,
				"namespace", namespace,
				"name", name)
			return err
		}
		// Resource already gone, treat as success
	}

	logger.Info("Deleted orphaned resource",
		"key", key,
		"kind", kind,
		"namespace", namespace,
		"name", name,
		"resourceID", resourceID)

	r.Recorder.Eventf(tenant, corev1.EventTypeNormal, "OrphanedResourceDeleted",
		"Deleted orphaned resource %s/%s (ID: %s) - removed from template", kind, name, resourceID)

	return nil
}

// getAPIVersionForKind returns the API version for a given kind string
func (r *TenantReconciler) getAPIVersionForKind(kind string) string {
	switch kind {
	case "Namespace", "ServiceAccount", "Service", "ConfigMap", "Secret", "PersistentVolumeClaim":
		return "v1"
	case "Deployment", "StatefulSet", "DaemonSet":
		return "apps/v1"
	case "Job", "CronJob":
		return "batch/v1"
	case "Ingress":
		return "networking.k8s.io/v1"
	default:
		// For unknown kinds, return v1 as default
		return "v1"
	}
}
