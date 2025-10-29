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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Policy types

// +kubebuilder:validation:Enum=Delete;Retain
type DeletionPolicy string

const (
	DeletionPolicyDelete DeletionPolicy = "Delete"
	DeletionPolicyRetain DeletionPolicy = "Retain"
)

// +kubebuilder:validation:Enum=Force;Stuck
type ConflictPolicy string

const (
	ConflictPolicyForce ConflictPolicy = "Force"
	ConflictPolicyStuck ConflictPolicy = "Stuck"
)

// +kubebuilder:validation:Enum=Once;WhenNeeded
type CreationPolicy string

const (
	CreationPolicyOnce       CreationPolicy = "Once"
	CreationPolicyWhenNeeded CreationPolicy = "WhenNeeded"
)

// +kubebuilder:validation:Enum=apply;merge;replace
type PatchStrategy string

const (
	PatchStrategyApply   PatchStrategy = "apply"
	PatchStrategyMerge   PatchStrategy = "merge"
	PatchStrategyReplace PatchStrategy = "replace"
)

// TResource defines a Kubernetes resource template with policies and dependencies
type TResource struct {
	// ID is a unique identifier within the template (used for dependencies and references)
	// +kubebuilder:validation:Required
	ID string `json:"id"`

	// Spec is the Kubernetes resource specification
	// Can be any Kubernetes native resource or custom resource
	// +kubebuilder:validation:Required
	// +kubebuilder:pruning:PreserveUnknownFields
	Spec unstructured.Unstructured `json:"spec"`

	// DependIds lists IDs of resources that must be ready before this resource is created
	// +optional
	DependIds []string `json:"dependIds,omitempty"`

	// CreationPolicy determines when the resource should be created
	// Default: WhenNeeded
	// +optional
	// +kubebuilder:default=WhenNeeded
	CreationPolicy CreationPolicy `json:"creationPolicy,omitempty"`

	// DeletionPolicy determines what happens to the resource when the Tenant is deleted
	// Default: Delete
	// +optional
	// +kubebuilder:default=Delete
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`

	// ConflictPolicy determines how to handle conflicts with existing resources
	// Default: Stuck (fail reconciliation if resource exists with different owner)
	// +optional
	// +kubebuilder:default=Stuck
	ConflictPolicy ConflictPolicy `json:"conflictPolicy,omitempty"`

	// NamespaceTemplate is a Go template for the namespace name
	// Template variables: .uid, .host, .hostOrUrl, and extraValueMappings
	// +optional
	NamespaceTemplate string `json:"namespaceTemplate,omitempty"`

	// NameTemplate is a Go template for the resource name
	// Template variables: .uid, .host, .hostOrUrl, and extraValueMappings
	// +optional
	NameTemplate string `json:"nameTemplate,omitempty"`

	// LabelsTemplate defines labels to apply to the resource (supports templates)
	// +optional
	LabelsTemplate map[string]string `json:"labelsTemplate,omitempty"`

	// AnnotationsTemplate defines annotations to apply to the resource (supports templates)
	// +optional
	AnnotationsTemplate map[string]string `json:"annotationsTemplate,omitempty"`

	// WaitForReady determines whether to wait for the resource to be ready before continuing
	// Default: true
	// +optional
	// +kubebuilder:default=true
	WaitForReady *bool `json:"waitForReady,omitempty"`

	// TimeoutSeconds is the maximum time to wait for the resource to be ready
	// Default: 300
	// +optional
	// +kubebuilder:default=300
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=3600
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`

	// PatchStrategy determines how to apply the resource
	// Default: apply (Server-Side Apply)
	// +optional
	// +kubebuilder:default=apply
	PatchStrategy PatchStrategy `json:"patchStrategy,omitempty"`
}

// SecretRef references a Kubernetes Secret
type SecretRef struct {
	// Name is the name of the Secret
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Key is the key within the Secret
	// +kubebuilder:validation:Required
	Key string `json:"key"`
}
