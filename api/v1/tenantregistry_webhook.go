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
)

// log is for logging in this package.
var tenantregistrylog = logf.Log.WithName("tenantregistry-resource")

// SetupWebhookWithManager sets up the webhook with the Manager.
func (r *TenantRegistry) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(&TenantRegistryDefaulter{}).
		WithValidator(&TenantRegistryValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-operator-kubernetes-tenants-org-v1-tenantregistry,mutating=true,failurePolicy=fail,sideEffects=None,groups=operator.kubernetes-tenants.org,resources=tenantregistries,verbs=create;update,versions=v1,name=mtenantregistry.kb.io,admissionReviewVersions=v1

// TenantRegistryDefaulter handles defaulting for TenantRegistry
type TenantRegistryDefaulter struct{}

var _ webhook.CustomDefaulter = &TenantRegistryDefaulter{}

// Default implements webhook.Defaulter
func (d *TenantRegistryDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	registry, ok := obj.(*TenantRegistry)
	if !ok {
		return fmt.Errorf("expected TenantRegistry but got %T", obj)
	}

	tenantregistrylog.Info("default", "name", registry.Name)

	// Set default SyncInterval
	if registry.Spec.Source.SyncInterval == "" {
		registry.Spec.Source.SyncInterval = "30s"
	}

	// Set default MySQL port
	if registry.Spec.Source.MySQL != nil && registry.Spec.Source.MySQL.Port == 0 {
		registry.Spec.Source.MySQL.Port = 3306
	}

	return nil
}

// +kubebuilder:webhook:path=/validate-operator-kubernetes-tenants-org-v1-tenantregistry,mutating=false,failurePolicy=fail,sideEffects=None,groups=operator.kubernetes-tenants.org,resources=tenantregistries,verbs=create;update,versions=v1,name=vtenantregistry.kb.io,admissionReviewVersions=v1

// TenantRegistryValidator handles validation for TenantRegistry
type TenantRegistryValidator struct{}

var _ webhook.CustomValidator = &TenantRegistryValidator{}

// ValidateCreate implements webhook.Validator
func (v *TenantRegistryValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	registry, ok := obj.(*TenantRegistry)
	if !ok {
		return nil, fmt.Errorf("expected TenantRegistry but got %T", obj)
	}

	tenantregistrylog.Info("validate create", "name", registry.Name)

	return v.validateTenantRegistry(ctx, registry)
}

// ValidateUpdate implements webhook.Validator
func (v *TenantRegistryValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	registry, ok := newObj.(*TenantRegistry)
	if !ok {
		return nil, fmt.Errorf("expected TenantRegistry but got %T", newObj)
	}

	tenantregistrylog.Info("validate update", "name", registry.Name)

	return v.validateTenantRegistry(ctx, registry)
}

// ValidateDelete implements webhook.Validator
func (v *TenantRegistryValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// No validation needed for deletion
	return nil, nil
}

// validateTenantRegistry performs all validation checks
//
//nolint:unparam // ctx kept for future validation that may need context
func (v *TenantRegistryValidator) validateTenantRegistry(ctx context.Context, registry *TenantRegistry) (admission.Warnings, error) {
	var warnings admission.Warnings

	// Validate required ValueMappings
	if registry.Spec.ValueMappings.UID == "" {
		return warnings, fmt.Errorf("valueMappings.uid is required")
	}
	if registry.Spec.ValueMappings.HostOrURL == "" {
		return warnings, fmt.Errorf("valueMappings.hostOrUrl is required")
	}
	if registry.Spec.ValueMappings.Activate == "" {
		return warnings, fmt.Errorf("valueMappings.activate is required")
	}

	// Validate source configuration
	if registry.Spec.Source.Type == SourceTypeMySQL {
		if registry.Spec.Source.MySQL == nil {
			return warnings, fmt.Errorf("mysql configuration is required when source type is mysql")
		}
		if registry.Spec.Source.MySQL.Host == "" {
			return warnings, fmt.Errorf("mysql.host is required")
		}
		if registry.Spec.Source.MySQL.Username == "" {
			return warnings, fmt.Errorf("mysql.username is required")
		}
		if registry.Spec.Source.MySQL.Database == "" {
			return warnings, fmt.Errorf("mysql.database is required")
		}
		if registry.Spec.Source.MySQL.Table == "" {
			return warnings, fmt.Errorf("mysql.table is required")
		}
	}

	return warnings, nil
}
