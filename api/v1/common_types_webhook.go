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
