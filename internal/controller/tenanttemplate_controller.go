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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
	"github.com/kubernetes-tenants/tenant-operator/internal/graph"
)

// TenantTemplateReconciler reconciles a TenantTemplate object
type TenantTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenanttemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenanttemplates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenanttemplates/finalizers,verbs=update
// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenantregistries,verbs=get;list;watch

// Reconcile validates a TenantTemplate
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

	// Update status
	r.updateStatus(ctx, tmpl, validationErrors)

	if len(validationErrors) > 0 {
		logger.Info("TenantTemplate validation failed", "errors", validationErrors)
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// validate validates a TenantTemplate
func (r *TenantTemplateReconciler) validate(ctx context.Context, tmpl *tenantsv1.TenantTemplate) []string {
	var errors []string

	// 1. Check if TenantRegistry exists
	if err := r.validateRegistryExists(ctx, tmpl); err != nil {
		errors = append(errors, fmt.Sprintf("Registry validation failed: %v", err))
	}

	// 2. Check for duplicate resource IDs
	if dupes := r.findDuplicateIDs(tmpl); len(dupes) > 0 {
		errors = append(errors, fmt.Sprintf("Duplicate resource IDs: %v", dupes))
	}

	// 3. Validate dependency graph
	if err := r.validateDependencies(tmpl); err != nil {
		errors = append(errors, fmt.Sprintf("Dependency validation failed: %v", err))
	}

	return errors
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

// updateStatus updates TenantTemplate status
func (r *TenantTemplateReconciler) updateStatus(ctx context.Context, tmpl *tenantsv1.TenantTemplate, validationErrors []string) {
	tmpl.Status.ObservedGeneration = tmpl.Generation

	condition := metav1.Condition{
		Type:               "Valid",
		Status:             metav1.ConditionTrue,
		Reason:             "ValidationPassed",
		Message:            "Template validation passed",
		LastTransitionTime: metav1.Now(),
	}

	if len(validationErrors) > 0 {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "ValidationFailed"
		condition.Message = fmt.Sprintf("Validation errors: %v", validationErrors)
	}

	// Update or append condition
	found := false
	for i := range tmpl.Status.Conditions {
		if tmpl.Status.Conditions[i].Type == condition.Type {
			tmpl.Status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		tmpl.Status.Conditions = append(tmpl.Status.Conditions, condition)
	}

	_ = r.Status().Update(ctx, tmpl)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantsv1.TenantTemplate{}).
		Named("tenanttemplate").
		Complete(r)
}
