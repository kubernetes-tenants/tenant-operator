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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TenantTemplateSpec defines the desired state of TenantTemplate.
// Note: Namespace management has been removed. All resources are created in the same namespace as the Tenant CR.
// Users must create the target namespace before deploying the Tenant CR.
type TenantTemplateSpec struct {
	// RegistryID references the TenantRegistry that this template is associated with
	// +kubebuilder:validation:Required
	RegistryID string `json:"registryId"`

	// ServiceAccounts defines ServiceAccount resources to create
	// +optional
	ServiceAccounts []TResource `json:"serviceAccounts,omitempty"`

	// Deployments defines Deployment resources to create
	// +optional
	Deployments []TResource `json:"deployments,omitempty"`

	// StatefulSets defines StatefulSet resources to create
	// +optional
	StatefulSets []TResource `json:"statefulSets,omitempty"`

	// Services defines Service resources to create
	// +optional
	Services []TResource `json:"services,omitempty"`

	// Ingresses defines Ingress resources to create
	// +optional
	Ingresses []TResource `json:"ingresses,omitempty"`

	// ConfigMaps defines ConfigMap resources to create
	// +optional
	ConfigMaps []TResource `json:"configMaps,omitempty"`

	// Secrets defines Secret resources to create
	// +optional
	Secrets []TResource `json:"secrets,omitempty"`

	// PersistentVolumeClaims defines PVC resources to create
	// +optional
	PersistentVolumeClaims []TResource `json:"persistentVolumeClaims,omitempty"`

	// Jobs defines Job resources to create
	// +optional
	Jobs []TResource `json:"jobs,omitempty"`

	// CronJobs defines CronJob resources to create
	// +optional
	CronJobs []TResource `json:"cronJobs,omitempty"`

	// Manifests defines arbitrary Kubernetes resources as raw manifests
	// Use this for any resource type not explicitly supported above
	// +optional
	Manifests []TResource `json:"manifests,omitempty"`
}

// TenantTemplateStatus defines the observed state of TenantTemplate.
type TenantTemplateStatus struct {
	// ObservedGeneration is the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest available observations of the template's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Registry",type="string",JSONPath=".spec.registryId",description="TenantRegistry reference"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status",description="Ready status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// TenantTemplate is the Schema for the tenanttemplates API.
type TenantTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantTemplateSpec   `json:"spec,omitempty"`
	Status TenantTemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TenantTemplateList contains a list of TenantTemplate.
type TenantTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TenantTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TenantTemplate{}, &TenantTemplateList{})
}
