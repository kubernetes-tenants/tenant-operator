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
var lynqhublog = logf.Log.WithName("lynqhub-resource")

// SetupWebhookWithManager sets up the webhook with the Manager.
func (r *LynqHub) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(&LynqHubDefaulter{}).
		WithValidator(&LynqHubValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-operator-lynq-sh-v1-lynqhub,mutating=true,failurePolicy=fail,sideEffects=None,groups=operator.lynq.sh,resources=lynqhubs,verbs=create;update,versions=v1,name=mlynqhub.kb.io,admissionReviewVersions=v1

// LynqHubDefaulter handles defaulting for LynqHub
type LynqHubDefaulter struct{}

var _ webhook.CustomDefaulter = &LynqHubDefaulter{}

// Default implements webhook.Defaulter
func (d *LynqHubDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	registry, ok := obj.(*LynqHub)
	if !ok {
		return fmt.Errorf("expected LynqHub but got %T", obj)
	}

	lynqhublog.Info("default", "name", registry.Name)

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

// +kubebuilder:webhook:path=/validate-operator-lynq-sh-v1-lynqhub,mutating=false,failurePolicy=fail,sideEffects=None,groups=operator.lynq.sh,resources=lynqhubs,verbs=create;update,versions=v1,name=vlynqhub.kb.io,admissionReviewVersions=v1

// LynqHubValidator handles validation for LynqHub
type LynqHubValidator struct{}

var _ webhook.CustomValidator = &LynqHubValidator{}

// ValidateCreate implements webhook.Validator
func (v *LynqHubValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	registry, ok := obj.(*LynqHub)
	if !ok {
		return nil, fmt.Errorf("expected LynqHub but got %T", obj)
	}

	lynqhublog.Info("validate create", "name", registry.Name)

	return v.validateLynqHub(ctx, registry)
}

// ValidateUpdate implements webhook.Validator
func (v *LynqHubValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	registry, ok := newObj.(*LynqHub)
	if !ok {
		return nil, fmt.Errorf("expected LynqHub but got %T", newObj)
	}

	lynqhublog.Info("validate update", "name", registry.Name)

	return v.validateLynqHub(ctx, registry)
}

// ValidateDelete implements webhook.Validator
func (v *LynqHubValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// No validation needed for deletion
	return nil, nil
}

// validateLynqHub performs all validation checks
//
//nolint:unparam // ctx kept for future validation that may need context
func (v *LynqHubValidator) validateLynqHub(ctx context.Context, registry *LynqHub) (admission.Warnings, error) {
	var warnings admission.Warnings

	// Validate required ValueMappings
	if registry.Spec.ValueMappings.UID == "" {
		return warnings, fmt.Errorf("valueMappings.uid is required")
	}
	if registry.Spec.ValueMappings.Activate == "" {
		return warnings, fmt.Errorf("valueMappings.activate is required")
	}

	// Deprecation warning for hostOrUrl
	if registry.Spec.ValueMappings.HostOrURL != "" {
		warnings = append(warnings,
			"valueMappings.hostOrUrl is deprecated since v1.1.11 and will be removed in v1.3.0. "+
				"Use extraValueMappings with the toHost() template function instead.")
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
