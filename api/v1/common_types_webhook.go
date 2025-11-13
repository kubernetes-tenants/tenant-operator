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
	"fmt"

	"github.com/k8s-lynq/lynq/internal/fieldfilter"
)

// SetDefaultsForTResource sets default values for a TResource
func SetDefaultsForTResource(r *TResource) {
	// Set default CreationPolicy
	if r.CreationPolicy == "" {
		r.CreationPolicy = CreationPolicyWhenNeeded
	}

	// Set default DeletionPolicy
	if r.DeletionPolicy == "" {
		r.DeletionPolicy = DeletionPolicyDelete
	}

	// Set default ConflictPolicy
	if r.ConflictPolicy == "" {
		r.ConflictPolicy = ConflictPolicyStuck
	}

	// Set default WaitForReady
	if r.WaitForReady == nil {
		trueVal := true
		r.WaitForReady = &trueVal
	}

	// Set default TimeoutSeconds
	if r.TimeoutSeconds == 0 {
		r.TimeoutSeconds = 300
	}

	// Set default PatchStrategy
	if r.PatchStrategy == "" {
		r.PatchStrategy = PatchStrategyApply
	}
}

// SetDefaultsForTResourceList sets default values for a list of TResources
func SetDefaultsForTResourceList(resources []TResource) {
	for i := range resources {
		SetDefaultsForTResource(&resources[i])
	}
}

// ValidateTResource validates a TResource, including IgnoreFields
func ValidateTResource(r *TResource) error {
	// Validate IgnoreFields JSONPath syntax using ojg/jp library
	if len(r.IgnoreFields) > 0 {
		if err := validateIgnoreFields(r.IgnoreFields); err != nil {
			return fmt.Errorf("invalid ignoreFields in resource '%s': %w", r.ID, err)
		}

		// Note: If IgnoreFields is used with CreationPolicy=Once, it has no effect
		// since the resource is only created once. This is intentional behavior,
		// not an error. The validation passes but the fields will not be ignored
		// in practice because the resource is never reconciled after creation.
	}

	return nil
}

// ValidateTResourceList validates a list of TResources
func ValidateTResourceList(resources []TResource) error {
	for i := range resources {
		if err := ValidateTResource(&resources[i]); err != nil {
			return err
		}
	}
	return nil
}

// validateIgnoreFields validates JSONPath expressions in ignoreFields
// Uses ojg/jp library for complete JSONPath standard validation
func validateIgnoreFields(paths []string) error {
	for _, path := range paths {
		// Use the fieldfilter package's validation which uses ojg/jp
		if err := fieldfilter.ValidateJSONPath(path); err != nil {
			return err
		}
	}
	return nil
}
