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

// LynqFormSpec defines the desired state of LynqForm.
// Resources are created in the same namespace as the LynqNode CR by default.
// Use TResource.targetNamespace to create resources in different namespaces.
// Namespaces can be created using the dedicated 'namespaces' field or 'manifests' field.
type LynqFormSpec struct {
	// HubID references the LynqHub that this form is associated with
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="hubId is immutable"
	HubID string `json:"hubId"`

	// ServiceAccounts defines ServiceAccount resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	ServiceAccounts []TResource `json:"serviceAccounts,omitempty"`

	// Deployments defines Deployment resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	Deployments []TResource `json:"deployments,omitempty"`

	// StatefulSets defines StatefulSet resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	StatefulSets []TResource `json:"statefulSets,omitempty"`

	// DaemonSets defines DaemonSet resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	DaemonSets []TResource `json:"daemonSets,omitempty"`

	// Services defines Service resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	Services []TResource `json:"services,omitempty"`

	// Ingresses defines Ingress resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	Ingresses []TResource `json:"ingresses,omitempty"`

	// ConfigMaps defines ConfigMap resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	ConfigMaps []TResource `json:"configMaps,omitempty"`

	// Secrets defines Secret resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	Secrets []TResource `json:"secrets,omitempty"`

	// PersistentVolumeClaims defines PVC resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	PersistentVolumeClaims []TResource `json:"persistentVolumeClaims,omitempty"`

	// Jobs defines Job resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	Jobs []TResource `json:"jobs,omitempty"`

	// CronJobs defines CronJob resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	CronJobs []TResource `json:"cronJobs,omitempty"`

	// PodDisruptionBudgets defines PDB resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	PodDisruptionBudgets []TResource `json:"podDisruptionBudgets,omitempty"`

	// NetworkPolicies defines NetworkPolicy resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	NetworkPolicies []TResource `json:"networkPolicies,omitempty"`

	// HorizontalPodAutoscalers defines HPA resources to create
	// +optional
	// +listType=map
	// +listMapKey=id
	HorizontalPodAutoscalers []TResource `json:"horizontalPodAutoscalers,omitempty"`

	// Namespaces defines Namespace resources to create
	// Note: Namespaces are cluster-scoped and always use label-based tracking
	// The targetNamespace field is ignored for Namespace resources
	// +optional
	// +listType=map
	// +listMapKey=id
	Namespaces []TResource `json:"namespaces,omitempty"`

	// Manifests defines arbitrary Kubernetes resources as raw manifests
	// Use this for any resource type not explicitly supported above
	// +optional
	// +listType=map
	// +listMapKey=id
	Manifests []TResource `json:"manifests,omitempty"`
}

// LynqFormStatus defines the observed state of LynqForm.
type LynqFormStatus struct {
	// ObservedGeneration is the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// TotalNodes is the total number of LynqNodes using this form
	// +optional
	TotalNodes int32 `json:"totalNodes,omitempty"`

	// ReadyNodes is the number of Ready LynqNodes using this form
	// +optional
	ReadyNodes int32 `json:"readyNodes,omitempty"`

	// Conditions represent the latest available observations of the form's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Hub",type="string",JSONPath=".spec.hubId",description="LynqHub reference"
// +kubebuilder:printcolumn:name="Total",type="integer",JSONPath=".status.totalNodes",description="Total nodes using form"
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.readyNodes",description="Ready nodes"
// +kubebuilder:printcolumn:name="Applied",type="string",JSONPath=".status.conditions[?(@.type=='Applied')].status",description="Applied status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// LynqForm is the Schema for the lynqforms API.
type LynqForm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LynqFormSpec   `json:"spec,omitempty"`
	Status LynqFormStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LynqFormList contains a list of LynqForm.
type LynqFormList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LynqForm `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LynqForm{}, &LynqFormList{})
}
