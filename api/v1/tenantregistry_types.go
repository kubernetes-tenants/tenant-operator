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

// SourceType defines the type of external data source
// +kubebuilder:validation:Enum=mysql
type SourceType string

const (
	SourceTypeMySQL SourceType = "mysql"
)

// MySQLSource defines MySQL connection parameters
type MySQLSource struct {
	// Host is the MySQL server hostname or IP
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// Port is the MySQL server port
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=3306
	Port int32 `json:"port"`

	// Username is the MySQL username
	// +kubebuilder:validation:Required
	Username string `json:"username"`

	// PasswordRef references a Secret containing the MySQL password
	// +optional
	PasswordRef *SecretRef `json:"passwordRef,omitempty"`

	// Database is the MySQL database name
	// +kubebuilder:validation:Required
	Database string `json:"database"`

	// Table is the MySQL table name containing tenant data
	// +kubebuilder:validation:Required
	Table string `json:"table"`
}

// DataSource defines the external data source configuration
type DataSource struct {
	// Type is the type of data source
	// +kubebuilder:validation:Required
	Type SourceType `json:"type"`

	// SyncInterval is how often to sync from the data source
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[0-9]+(s|m|h)$`
	// +kubebuilder:default="30s"
	SyncInterval string `json:"syncInterval"`

	// MySQL contains MySQL-specific configuration
	// +optional
	MySQL *MySQLSource `json:"mysql,omitempty"`
}

// ValueMappings defines required column mappings
type ValueMappings struct {
	// UID is the column name for the tenant unique identifier
	// +kubebuilder:validation:Required
	UID string `json:"uid"`

	// HostOrURL is the column name for the tenant host or URL
	// +kubebuilder:validation:Required
	HostOrURL string `json:"hostOrUrl"`

	// Activate is the column name for the activation status
	// +kubebuilder:validation:Required
	Activate string `json:"activate"`
}

// TenantRegistrySpec defines the desired state of TenantRegistry.
type TenantRegistrySpec struct {
	// Source defines the external data source configuration
	// +kubebuilder:validation:Required
	Source DataSource `json:"source"`

	// ValueMappings defines required column to variable mappings
	// +kubebuilder:validation:Required
	ValueMappings ValueMappings `json:"valueMappings"`

	// ExtraValueMappings defines additional custom column to variable mappings
	// Keys become template variables, values are column names
	// +optional
	ExtraValueMappings map[string]string `json:"extraValueMappings,omitempty"`
}

// TenantRegistryStatus defines the observed state of TenantRegistry.
type TenantRegistryStatus struct {
	// ObservedGeneration is the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Desired is the number of active tenants from the data source
	// +optional
	Desired int32 `json:"desired,omitempty"`

	// Ready is the number of ready Tenant resources
	// +optional
	Ready int32 `json:"ready,omitempty"`

	// Failed is the number of failed Tenant resources
	// +optional
	Failed int32 `json:"failed,omitempty"`

	// Conditions represent the latest available observations of the registry's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TenantRegistry is the Schema for the tenantregistries API.
type TenantRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantRegistrySpec   `json:"spec,omitempty"`
	Status TenantRegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TenantRegistryList contains a list of TenantRegistry.
type TenantRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TenantRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TenantRegistry{}, &TenantRegistryList{})
}
