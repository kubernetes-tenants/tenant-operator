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
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
	"github.com/kubernetes-tenants/tenant-operator/internal/graph"
)

// TenantTemplateReconciler reconciles a TenantTemplate object
type TenantTemplateReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenanttemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenanttemplates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenanttemplates/finalizers,verbs=update
// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenantregistries,verbs=get;list;watch
// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenants,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile validates a TenantTemplate and checks tenant statuses
func (r *TenantTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch TenantTemplate
	tmpl := &tenantsv1.TenantTemplate{}
	if err := r.Get(ctx, req.NamespacedName, tmpl); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get TenantTemplate")
		return ctrl.Result{}, err
	}

	// Validate
	validationErrors := r.validate(ctx, tmpl)

	if len(validationErrors) > 0 {
		logger.Info("TenantTemplate validation failed", "errors", validationErrors)
		// Update status with validation errors
		r.updateStatus(ctx, tmpl, validationErrors, 0, 0)
		// Emit warning event for validation failure
		r.Recorder.Eventf(tmpl, corev1.EventTypeWarning, "ValidationFailed",
			"Template validation failed: %v", validationErrors)
		return ctrl.Result{}, nil
	}

	// Emit normal event for validation success
	r.Recorder.Event(tmpl, corev1.EventTypeNormal, "ValidationPassed",
		"Template validation passed successfully")

	// Check tenant statuses
	totalTenants, readyTenants, err := r.checkTenantStatuses(ctx, tmpl)
	if err != nil {
		logger.Error(err, "Failed to check tenant statuses")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Update status with tenant counts
	r.updateStatus(ctx, tmpl, validationErrors, totalTenants, readyTenants)

	return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
}

// validate validates a TenantTemplate
func (r *TenantTemplateReconciler) validate(ctx context.Context, tmpl *tenantsv1.TenantTemplate) []string {
	var validationErrors []string

	// 1. Check if TenantRegistry exists
	if err := r.validateRegistryExists(ctx, tmpl); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("Registry validation failed: %v", err))
		// Emit specific event for registry not found
		r.Recorder.Eventf(tmpl, corev1.EventTypeWarning, "RegistryNotFound",
			"Referenced TenantRegistry '%s' not found in namespace '%s'",
			tmpl.Spec.RegistryID, tmpl.Namespace)
	}

	// 2. Check for duplicate resource IDs
	if dupes := r.findDuplicateIDs(tmpl); len(dupes) > 0 {
		validationErrors = append(validationErrors, fmt.Sprintf("Duplicate resource IDs: %v", dupes))
		r.Recorder.Eventf(tmpl, corev1.EventTypeWarning, "DuplicateResourceIDs",
			"Found duplicate resource IDs: %v", dupes)
	}

	// 3. Validate dependency graph
	if err := r.validateDependencies(tmpl); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("Dependency validation failed: %v", err))
		r.Recorder.Eventf(tmpl, corev1.EventTypeWarning, "DependencyValidationFailed",
			"Dependency graph validation failed: %v", err)
	}

	return validationErrors
}

// validateRegistryExists checks if the referenced TenantRegistry exists
func (r *TenantTemplateReconciler) validateRegistryExists(ctx context.Context, tmpl *tenantsv1.TenantTemplate) error {
	registry := &tenantsv1.TenantRegistry{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      tmpl.Spec.RegistryID,
		Namespace: tmpl.Namespace,
	}, registry); err != nil {
		return fmt.Errorf("registry '%s' not found: %w", tmpl.Spec.RegistryID, err)
	}
	return nil
}

// findDuplicateIDs finds duplicate resource IDs
func (r *TenantTemplateReconciler) findDuplicateIDs(tmpl *tenantsv1.TenantTemplate) []string {
	seen := make(map[string]bool)
	var duplicates []string

	allResources := r.collectAllResources(tmpl)

	for _, resource := range allResources {
		if resource.ID == "" {
			continue
		}
		if seen[resource.ID] {
			duplicates = append(duplicates, resource.ID)
		}
		seen[resource.ID] = true
	}

	return duplicates
}

// validateDependencies validates the dependency graph
func (r *TenantTemplateReconciler) validateDependencies(tmpl *tenantsv1.TenantTemplate) error {
	allResources := r.collectAllResources(tmpl)

	// Build dependency graph
	depGraph, err := graph.BuildGraph(allResources)
	if err != nil {
		return err
	}

	// Validate (checks for cycles and missing dependencies)
	if err := depGraph.Validate(); err != nil {
		return err
	}

	return nil
}

// collectAllResources collects all resources from the template
func (r *TenantTemplateReconciler) collectAllResources(tmpl *tenantsv1.TenantTemplate) []tenantsv1.TResource {
	var resources []tenantsv1.TResource

	resources = append(resources, tmpl.Spec.ServiceAccounts...)
	resources = append(resources, tmpl.Spec.Deployments...)
	resources = append(resources, tmpl.Spec.StatefulSets...)
	resources = append(resources, tmpl.Spec.DaemonSets...)
	resources = append(resources, tmpl.Spec.Services...)
	resources = append(resources, tmpl.Spec.Ingresses...)
	resources = append(resources, tmpl.Spec.ConfigMaps...)
	resources = append(resources, tmpl.Spec.Secrets...)
	resources = append(resources, tmpl.Spec.PersistentVolumeClaims...)
	resources = append(resources, tmpl.Spec.Jobs...)
	resources = append(resources, tmpl.Spec.CronJobs...)
	resources = append(resources, tmpl.Spec.PodDisruptionBudgets...)
	resources = append(resources, tmpl.Spec.NetworkPolicies...)
	resources = append(resources, tmpl.Spec.HorizontalPodAutoscalers...)
	resources = append(resources, tmpl.Spec.Manifests...)

	return resources
}

// checkTenantStatuses checks the status of all Tenants using this template
func (r *TenantTemplateReconciler) checkTenantStatuses(ctx context.Context, tmpl *tenantsv1.TenantTemplate) (totalTenants, readyTenants int32, err error) {
	// List all Tenants that reference this template
	tenantList := &tenantsv1.TenantList{}
	if err := r.List(ctx, tenantList, client.InNamespace(tmpl.Namespace)); err != nil {
		return 0, 0, fmt.Errorf("failed to list tenants: %w", err)
	}

	// Filter tenants that use this template
	for _, tenant := range tenantList.Items {
		if tenant.Spec.TemplateRef == tmpl.Name {
			totalTenants++

			// Check if tenant is Ready
			for _, condition := range tenant.Status.Conditions {
				if condition.Type == ConditionTypeReady && condition.Status == metav1.ConditionTrue {
					readyTenants++
					break
				}
			}
		}
	}

	return totalTenants, readyTenants, nil
}

// updateStatus updates TenantTemplate status with retry on conflict
func (r *TenantTemplateReconciler) updateStatus(ctx context.Context, tmpl *tenantsv1.TenantTemplate, validationErrors []string, totalTenants, readyTenants int32) {
	logger := log.FromContext(ctx)

	// Retry status update on conflict
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the latest version of the template
		key := client.ObjectKeyFromObject(tmpl)
		latest := &tenantsv1.TenantTemplate{}
		if err := r.Get(ctx, key, latest); err != nil {
			return err
		}

		// Update status fields
		latest.Status.ObservedGeneration = latest.Generation
		latest.Status.TotalTenants = totalTenants
		latest.Status.ReadyTenants = readyTenants

		// Prepare Valid condition
		validCondition := metav1.Condition{
			Type:               "Valid",
			Status:             metav1.ConditionTrue,
			Reason:             "ValidationPassed",
			Message:            "Template validation passed",
			LastTransitionTime: metav1.Now(),
		}

		if len(validationErrors) > 0 {
			validCondition.Status = metav1.ConditionFalse
			validCondition.Reason = "ValidationFailed"
			validCondition.Message = fmt.Sprintf("Validation errors: %v", validationErrors)
		}

		// Prepare Applied condition
		appliedCondition := metav1.Condition{
			Type:               "Applied",
			Status:             metav1.ConditionFalse,
			Reason:             "NotAllTenantsReady",
			Message:            fmt.Sprintf("%d/%d tenants ready", readyTenants, totalTenants),
			LastTransitionTime: metav1.Now(),
		}

		if totalTenants > 0 && readyTenants == totalTenants {
			appliedCondition.Status = metav1.ConditionTrue
			appliedCondition.Reason = "AllTenantsReady"
			appliedCondition.Message = fmt.Sprintf("All %d tenants ready", totalTenants)
		} else if totalTenants == 0 {
			appliedCondition.Reason = "NoTenants"
			appliedCondition.Message = "No tenants using this template"
		}

		// Update or append Valid condition
		foundValid := false
		for i := range latest.Status.Conditions {
			if latest.Status.Conditions[i].Type == "Valid" {
				latest.Status.Conditions[i] = validCondition
				foundValid = true
				break
			}
		}
		if !foundValid {
			latest.Status.Conditions = append(latest.Status.Conditions, validCondition)
		}

		// Update or append Applied condition
		foundApplied := false
		for i := range latest.Status.Conditions {
			if latest.Status.Conditions[i].Type == "Applied" {
				latest.Status.Conditions[i] = appliedCondition
				foundApplied = true
				break
			}
		}
		if !foundApplied {
			latest.Status.Conditions = append(latest.Status.Conditions, appliedCondition)
		}

		// Update status subresource
		return r.Status().Update(ctx, latest)
	})

	if err != nil {
		logger.Error(err, "Failed to update TenantTemplate status after retries")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantTemplateReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantsv1.TenantTemplate{}).
		Named("tenanttemplate").
		// Watch Tenants to update template Applied status when tenant status changes
		Watches(&tenantsv1.Tenant{}, handler.EnqueueRequestsFromMapFunc(r.findTemplateForTenant)).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		Complete(r)
}

// findTemplateForTenant maps a Tenant to its TenantTemplate for watch events
func (r *TenantTemplateReconciler) findTemplateForTenant(ctx context.Context, tenant client.Object) []reconcile.Request {
	t := tenant.(*tenantsv1.Tenant)
	if t.Spec.TemplateRef == "" {
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      t.Spec.TemplateRef,
				Namespace: t.Namespace,
			},
		},
	}
}
