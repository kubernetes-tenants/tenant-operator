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
type TenantTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of TenantTemplate. Edit tenanttemplate_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// TenantTemplateStatus defines the observed state of TenantTemplate.
type TenantTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

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
