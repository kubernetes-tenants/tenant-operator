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

package v1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kubernetes-tenants/tenant-operator/internal/template"
)

// log is for logging in this package.
var tenanttemplatelog = logf.Log.WithName("tenanttemplate-resource")

// SetupWebhookWithManager sets up the webhook with the Manager.
func (r *TenantTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(&TenantTemplateDefaulter{}).
		WithValidator(&TenantTemplateValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-operator-kubernetes-tenants-org-v1-tenanttemplate,mutating=true,failurePolicy=fail,sideEffects=None,groups=operator.kubernetes-tenants.org,resources=tenanttemplates,verbs=create;update,versions=v1,name=mtenanttemplate.kb.io,admissionReviewVersions=v1

// TenantTemplateDefaulter handles defaulting for TenantTemplate
type TenantTemplateDefaulter struct{}

var _ webhook.CustomDefaulter = &TenantTemplateDefaulter{}

// Default implements webhook.Defaulter
func (d *TenantTemplateDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	tmpl, ok := obj.(*TenantTemplate)
	if !ok {
		return fmt.Errorf("expected TenantTemplate but got %T", obj)
	}

	tenanttemplatelog.Info("default", "name", tmpl.Name)

	// Set defaults for all resource types
	SetDefaultsForTResourceList(tmpl.Spec.ServiceAccounts)
	SetDefaultsForTResourceList(tmpl.Spec.Deployments)
	SetDefaultsForTResourceList(tmpl.Spec.StatefulSets)
	SetDefaultsForTResourceList(tmpl.Spec.Services)
	SetDefaultsForTResourceList(tmpl.Spec.Ingresses)
	SetDefaultsForTResourceList(tmpl.Spec.ConfigMaps)
	SetDefaultsForTResourceList(tmpl.Spec.Secrets)
	SetDefaultsForTResourceList(tmpl.Spec.PersistentVolumeClaims)
	SetDefaultsForTResourceList(tmpl.Spec.Jobs)
	SetDefaultsForTResourceList(tmpl.Spec.CronJobs)
	SetDefaultsForTResourceList(tmpl.Spec.Manifests)

	return nil
}

// +kubebuilder:webhook:path=/validate-operator-kubernetes-tenants-org-v1-tenanttemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=operator.kubernetes-tenants.org,resources=tenanttemplates,verbs=create;update,versions=v1,name=vtenanttemplate.kb.io,admissionReviewVersions=v1

// TenantTemplateValidator handles validation for TenantTemplate
type TenantTemplateValidator struct{}

var _ webhook.CustomValidator = &TenantTemplateValidator{}

// ValidateCreate implements webhook.Validator
func (v *TenantTemplateValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	tmpl, ok := obj.(*TenantTemplate)
	if !ok {
		return nil, fmt.Errorf("expected TenantTemplate but got %T", obj)
	}

	tenanttemplatelog.Info("validate create", "name", tmpl.Name)

	return v.validateTenantTemplate(ctx, tmpl)
}

// ValidateUpdate implements webhook.Validator
func (v *TenantTemplateValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	tmpl, ok := newObj.(*TenantTemplate)
	if !ok {
		return nil, fmt.Errorf("expected TenantTemplate but got %T", newObj)
	}

	tenanttemplatelog.Info("validate update", "name", tmpl.Name)

	return v.validateTenantTemplate(ctx, tmpl)
}

// ValidateDelete implements webhook.Validator
func (v *TenantTemplateValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// No validation needed for deletion
	return nil, nil
}

// validateTenantTemplate performs all validation checks
func (v *TenantTemplateValidator) validateTenantTemplate(ctx context.Context, tmpl *TenantTemplate) (admission.Warnings, error) {
	var warnings admission.Warnings

	// 1. Validate registryId is not empty
	if tmpl.Spec.RegistryID == "" {
		return warnings, fmt.Errorf("registryId is required")
	}

	// Note: Registry existence will be validated by TenantTemplate controller

	// 2. Check for duplicate resource IDs
	if dupes := v.findDuplicateIDs(tmpl); len(dupes) > 0 {
		return warnings, fmt.Errorf("duplicate resource IDs found: %v", dupes)
	}

	// 3. Validate resource IDs are not empty
	if err := v.validateResourceIDs(tmpl); err != nil {
		return warnings, err
	}

	// 4. Validate dependency graph (no cycles, no missing dependencies)
	if err := v.validateDependencies(tmpl); err != nil {
		return warnings, fmt.Errorf("dependency validation failed: %w", err)
	}

	// 5. Validate template syntax
	if err := v.validateTemplateSyntax(tmpl); err != nil {
		return warnings, fmt.Errorf("template validation failed: %w", err)
	}

	return warnings, nil
}

// findDuplicateIDs finds duplicate resource IDs
func (v *TenantTemplateValidator) findDuplicateIDs(tmpl *TenantTemplate) []string {
	seen := make(map[string]bool)
	var duplicates []string

	allResources := v.collectAllResources(tmpl)

	for _, resource := range allResources {
		if resource.ID == "" {
			continue
		}
		if seen[resource.ID] {
			if !contains(duplicates, resource.ID) {
				duplicates = append(duplicates, resource.ID)
			}
		}
		seen[resource.ID] = true
	}

	return duplicates
}

// validateResourceIDs ensures all resources have non-empty IDs
func (v *TenantTemplateValidator) validateResourceIDs(tmpl *TenantTemplate) error {
	allResources := v.collectAllResources(tmpl)

	for _, resource := range allResources {
		if resource.ID == "" {
			return fmt.Errorf("resource must have a non-empty ID")
		}
	}

	return nil
}

// validateDependencies validates the dependency graph
func (v *TenantTemplateValidator) validateDependencies(tmpl *TenantTemplate) error {
	allResources := v.collectAllResources(tmpl)

	// Build ID set for quick lookup
	idSet := make(map[string]bool)
	for _, resource := range allResources {
		idSet[resource.ID] = true
	}

	// Check that all dependencies exist
	for _, resource := range allResources {
		for _, depID := range resource.DependIds {
			if !idSet[depID] {
				return fmt.Errorf("resource '%s' depends on non-existent resource '%s'", resource.ID, depID)
			}
		}
	}

	// Check for cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	// Build adjacency list
	adjList := make(map[string][]string)
	for _, resource := range allResources {
		adjList[resource.ID] = resource.DependIds
	}

	// Check for cycles starting from each node
	for _, resource := range allResources {
		if err := v.detectCycle(resource.ID, adjList, visited, recStack); err != nil {
			return err
		}
	}

	return nil
}

// detectCycle performs DFS to detect cycles
func (v *TenantTemplateValidator) detectCycle(id string, adjList map[string][]string, visited, recStack map[string]bool) error {
	visited[id] = true
	recStack[id] = true

	for _, depID := range adjList[id] {
		if !visited[depID] {
			if err := v.detectCycle(depID, adjList, visited, recStack); err != nil {
				return err
			}
		} else if recStack[depID] {
			return fmt.Errorf("circular dependency detected: %s -> %s", id, depID)
		}
	}

	recStack[id] = false
	return nil
}

// collectAllResources collects all resources from the template
func (v *TenantTemplateValidator) collectAllResources(tmpl *TenantTemplate) []TResource {
	var resources []TResource

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

// validateTemplateSyntax validates that all template strings are valid Go templates
func (v *TenantTemplateValidator) validateTemplateSyntax(tmpl *TenantTemplate) error {
	engine := template.NewEngine()

	// Sample variables for validation
	sampleVars := template.Variables{
		"uid":       "test-tenant",
		"hostOrUrl": "https://example.com",
		"host":      "example.com",
		"activate":  "true",
	}

	allResources := v.collectAllResources(tmpl)

	for _, res := range allResources {
		// Validate NameTemplate
		if res.NameTemplate != "" {
			if _, err := engine.Render(res.NameTemplate, sampleVars); err != nil {
				return fmt.Errorf("invalid NameTemplate in resource '%s': %w", res.ID, err)
			}
		}

		// Validate LabelsTemplate
		for key, tmplStr := range res.LabelsTemplate {
			if _, err := engine.Render(tmplStr, sampleVars); err != nil {
				return fmt.Errorf("invalid LabelsTemplate[%s] in resource '%s': %w", key, res.ID, err)
			}
		}

		// Validate AnnotationsTemplate
		for key, tmplStr := range res.AnnotationsTemplate {
			if _, err := engine.Render(tmplStr, sampleVars); err != nil {
				return fmt.Errorf("invalid AnnotationsTemplate[%s] in resource '%s': %w", key, res.ID, err)
			}
		}
	}

	return nil
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
