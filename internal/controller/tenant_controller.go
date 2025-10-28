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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
	"github.com/kubernetes-tenants/tenant-operator/internal/apply"
	"github.com/kubernetes-tenants/tenant-operator/internal/graph"
	"github.com/kubernetes-tenants/tenant-operator/internal/readiness"
	"github.com/kubernetes-tenants/tenant-operator/internal/template"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenants/finalizers,verbs=update
// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenanttemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenantregistries,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts;services;configmaps;secrets;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments;statefulsets;daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs;cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile applies all resources for a tenant
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch Tenant
	tenant := &tenantsv1.Tenant{}
	if err := r.Get(ctx, req.NamespacedName, tenant); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Tenant")
		return ctrl.Result{}, err
	}

	// Get TenantTemplate
	tmpl, err := r.getTenantTemplate(ctx, tenant)
	if err != nil {
		logger.Error(err, "Failed to get TenantTemplate")
		r.updateStatus(ctx, tenant, 0, 0, 0, false, "TemplateNotFound", err.Error())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Get TenantRegistry to get extra values
	registry, extraValues, err := r.getTenantRegistry(ctx, tenant)
	if err != nil {
		logger.Error(err, "Failed to get TenantRegistry")
		r.updateStatus(ctx, tenant, 0, 0, 0, false, "RegistryNotFound", err.Error())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Build template variables
	vars := r.buildTemplateVariables(tenant, registry, extraValues)

	// Collect all resources
	allResources := r.collectResources(tmpl)

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
		obj, err := r.renderResource(ctx, templateEngine, resource, vars)
		if err != nil {
			logger.Error(err, "Failed to render resource", "id", resource.ID)
			failedCount++
			continue
		}

		// Apply resource
		if err := applier.ApplyResource(ctx, obj, tenant, resource.ConflictPolicy); err != nil {
			logger.Error(err, "Failed to apply resource", "id", resource.ID)
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
				failedCount++
				continue
			}
		}

		readyCount++
	}

	// Update status
	r.updateStatus(ctx, tenant, totalResources, readyCount, failedCount, failedCount == 0, "Reconciled", "Successfully reconciled all resources")

	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// getTenantTemplate retrieves the associated TenantTemplate
func (r *TenantReconciler) getTenantTemplate(ctx context.Context, tenant *tenantsv1.Tenant) (*tenantsv1.TenantTemplate, error) {
	// Find template by registry label
	registryName := tenant.Labels["tenants.ecube.dev/registry"]
	if registryName == "" {
		return nil, fmt.Errorf("tenant missing registry label")
	}

	templateList := &tenantsv1.TenantTemplateList{}
	if err := r.List(ctx, templateList, client.InNamespace(tenant.Namespace)); err != nil {
		return nil, err
	}

	for _, tmpl := range templateList.Items {
		if tmpl.Spec.RegistryID == registryName {
			return &tmpl, nil
		}
	}

	return nil, fmt.Errorf("no template found for registry: %s", registryName)
}

// getTenantRegistry retrieves the TenantRegistry and extra values
func (r *TenantReconciler) getTenantRegistry(ctx context.Context, tenant *tenantsv1.Tenant) (*tenantsv1.TenantRegistry, map[string]string, error) {
	registryName := tenant.Labels["tenants.ecube.dev/registry"]
	if registryName == "" {
		return nil, nil, fmt.Errorf("tenant missing registry label")
	}

	registry := &tenantsv1.TenantRegistry{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      registryName,
		Namespace: tenant.Namespace,
	}, registry); err != nil {
		return nil, nil, err
	}

	// Query database to get extra values for this tenant
	extraValues, err := r.queryTenantExtraValues(ctx, registry, tenant.Spec.UID)
	if err != nil {
		return registry, nil, err
	}

	return registry, extraValues, nil
}

// queryTenantExtraValues queries database for extra values
func (r *TenantReconciler) queryTenantExtraValues(ctx context.Context, registry *tenantsv1.TenantRegistry, uid string) (map[string]string, error) {
	// This would query the database similar to TenantRegistry controller
	// For now, return empty map
	return make(map[string]string), nil
}

// buildTemplateVariables builds template variables
func (r *TenantReconciler) buildTemplateVariables(tenant *tenantsv1.Tenant, registry *tenantsv1.TenantRegistry, extraValues map[string]string) template.Variables {
	// Get host from tenant labels or use UID
	hostOrURL := tenant.Annotations["tenants.ecube.dev/hostOrUrl"]
	if hostOrURL == "" {
		hostOrURL = tenant.Spec.UID
	}

	return template.BuildVariables(tenant.Spec.UID, hostOrURL, "true", extraValues)
}

// collectResources collects all resources from template
func (r *TenantReconciler) collectResources(tmpl *tenantsv1.TenantTemplate) []tenantsv1.TResource {
	var resources []tenantsv1.TResource

	resources = append(resources, tmpl.Spec.Namespaces...)
	resources = append(resources, tmpl.Spec.ServiceAccounts...)
	resources = append(resources, tmpl.Spec.Deployments...)
	resources = append(resources, tmpl.Spec.StatefulSets...)
	resources = append(resources, tmpl.Spec.Services...)
	resources = append(resources, tmpl.Spec.Ingresses...)
	resources = append(resources, tmpl.Spec.ConfigMaps...)
	resources = append(resources, tmpl.Spec.Secrets...)
	resources = append(resources, tmpl.Spec.PersistentVolumeClaims...)
	resources = append(resources, tmpl.Spec.Jobs...)
	resources = append(resources, tmpl.Spec.CronJobs...)
	resources = append(resources, tmpl.Spec.Manifests...)

	return resources
}

// renderResource renders a resource template
func (r *TenantReconciler) renderResource(ctx context.Context, engine *template.Engine, resource tenantsv1.TResource, vars template.Variables) (*unstructured.Unstructured, error) {
	// Render name
	name, err := engine.Render(resource.NameTemplate, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render name: %w", err)
	}

	// Render namespace
	namespace, err := engine.Render(resource.NamespaceTemplate, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render namespace: %w", err)
	}

	// Render labels
	labels, err := engine.RenderMap(resource.LabelsTemplate, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render labels: %w", err)
	}

	// Render annotations
	annotations, err := engine.RenderMap(resource.AnnotationsTemplate, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render annotations: %w", err)
	}

	// Parse spec
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal(resource.Spec.Raw, obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
	}

	// Set metadata
	obj.SetName(name)
	obj.SetNamespace(namespace)
	if len(labels) > 0 {
		obj.SetLabels(labels)
	}
	if len(annotations) > 0 {
		obj.SetAnnotations(annotations)
	}

	// Render spec recursively (for template variables in spec)
	renderedSpec, err := r.renderUnstructured(obj.Object, engine, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render spec: %w", err)
	}
	obj.Object = renderedSpec

	return obj, nil
}

// renderUnstructured recursively renders template variables in unstructured data
func (r *TenantReconciler) renderUnstructured(data map[string]interface{}, engine *template.Engine, vars template.Variables) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for k, v := range data {
		switch val := v.(type) {
		case string:
			// Try to render as template
			rendered, err := engine.Render(val, vars)
			if err != nil {
				// If rendering fails, keep original
				result[k] = val
			} else {
				result[k] = rendered
			}
		case map[string]interface{}:
			// Recurse into nested maps
			rendered, err := r.renderUnstructured(val, engine, vars)
			if err != nil {
				result[k] = val
			} else {
				result[k] = rendered
			}
		case []interface{}:
			// Recurse into arrays
			renderedArray := make([]interface{}, len(val))
			for i, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					rendered, err := r.renderUnstructured(itemMap, engine, vars)
					if err != nil {
						renderedArray[i] = item
					} else {
						renderedArray[i] = rendered
					}
				} else if itemStr, ok := item.(string); ok {
					rendered, err := engine.Render(itemStr, vars)
					if err != nil {
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

	_ = r.Status().Update(ctx, tenant)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantsv1.Tenant{}).
		Named("tenant").
		Complete(r)
}
