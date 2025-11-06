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

package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
)

// TestGetExistingTenants tests the getExistingTenants function
func TestGetExistingTenants(t *testing.T) {
	tests := []struct {
		name          string
		registry      *tenantsv1.TenantRegistry
		existingItems []tenantsv1.Tenant
		wantCount     int
		wantErr       bool
	}{
		{
			name: "no tenants found",
			registry: &tenantsv1.TenantRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
			},
			existingItems: []tenantsv1.Tenant{},
			wantCount:     0,
			wantErr:       false,
		},
		{
			name: "multiple tenants found with registry label",
			registry: &tenantsv1.TenantRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
			},
			existingItems: []tenantsv1.Tenant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "test-registry",
						},
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant1",
						TemplateRef: "web",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant2-web",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "test-registry",
						},
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant2",
						TemplateRef: "web",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant3-api",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "other-registry",
						},
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant3",
						TemplateRef: "api",
					},
				},
			},
			wantCount: 2, // Only tenant1 and tenant2 match the registry
			wantErr:   false,
		},
		{
			name: "tenants in different namespace not returned",
			registry: &tenantsv1.TenantRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
			},
			existingItems: []tenantsv1.Tenant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web",
						Namespace: "other-namespace",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "test-registry",
						},
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant1",
						TemplateRef: "web",
					},
				},
			},
			wantCount: 0, // Different namespace
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			require.NoError(t, tenantsv1.AddToScheme(scheme))

			// Convert []Tenant to []client.Object
			objects := []runtime.Object{tt.registry}
			for i := range tt.existingItems {
				objects = append(objects, &tt.existingItems[i])
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			r := &TenantRegistryReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			result, err := r.getExistingTenants(ctx, tt.registry)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(result.Items), "Expected %d tenants, got %d", tt.wantCount, len(result.Items))
		})
	}
}

// TestCountTenantStatus tests the countTenantStatus function
func TestCountTenantStatus(t *testing.T) {
	tests := []struct {
		name       string
		registry   *tenantsv1.TenantRegistry
		tenants    []tenantsv1.Tenant
		wantReady  int32
		wantFailed int32
	}{
		{
			name: "no tenants",
			registry: &tenantsv1.TenantRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
			},
			tenants:    []tenantsv1.Tenant{},
			wantReady:  0,
			wantFailed: 0,
		},
		{
			name: "all tenants ready",
			registry: &tenantsv1.TenantRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
			},
			tenants: []tenantsv1.Tenant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "test-registry",
						},
					},
					Status: tenantsv1.TenantStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant2-web",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "test-registry",
						},
					},
					Status: tenantsv1.TenantStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			wantReady:  2,
			wantFailed: 0,
		},
		{
			name: "mixed ready and failed tenants",
			registry: &tenantsv1.TenantRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
			},
			tenants: []tenantsv1.Tenant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "test-registry",
						},
					},
					Status: tenantsv1.TenantStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant2-web",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "test-registry",
						},
					},
					Status: tenantsv1.TenantStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionFalse,
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant3-web",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "test-registry",
						},
					},
					Status: tenantsv1.TenantStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionUnknown,
							},
						},
					},
				},
			},
			wantReady:  1,
			wantFailed: 2, // False and Unknown both count as failed
		},
		{
			name: "tenants without Ready condition",
			registry: &tenantsv1.TenantRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
			},
			tenants: []tenantsv1.Tenant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "test-registry",
						},
					},
					Status: tenantsv1.TenantStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Degraded",
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			wantReady:  0,
			wantFailed: 0, // No Ready condition means not counted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			require.NoError(t, tenantsv1.AddToScheme(scheme))

			objects := []runtime.Object{tt.registry}
			for i := range tt.tenants {
				objects = append(objects, &tt.tenants[i])
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			r := &TenantRegistryReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			ready, failed := r.countTenantStatus(ctx, tt.registry)

			assert.Equal(t, tt.wantReady, ready, "Expected %d ready tenants, got %d", tt.wantReady, ready)
			assert.Equal(t, tt.wantFailed, failed, "Expected %d failed tenants, got %d", tt.wantFailed, failed)
		})
	}
}

// TestUpdateStatus tests the updateStatus function
func TestUpdateStatus(t *testing.T) {
	tests := []struct {
		name                 string
		referencingTemplates int32
		desired              int32
		ready                int32
		failed               int32
		synced               bool
		wantConditionStatus  metav1.ConditionStatus
		wantConditionReason  string
	}{
		{
			name:                 "successful sync",
			referencingTemplates: 2,
			desired:              6,
			ready:                4,
			failed:               2,
			synced:               true,
			wantConditionStatus:  metav1.ConditionTrue,
			wantConditionReason:  "DatabaseConnected",
		},
		{
			name:                 "failed sync",
			referencingTemplates: 1,
			desired:              0,
			ready:                0,
			failed:               0,
			synced:               false,
			wantConditionStatus:  metav1.ConditionFalse,
			wantConditionReason:  "DatabaseConnectionFailed",
		},
		{
			name:                 "no templates referencing registry",
			referencingTemplates: 0,
			desired:              0,
			ready:                0,
			failed:               0,
			synced:               true,
			wantConditionStatus:  metav1.ConditionTrue,
			wantConditionReason:  "DatabaseConnected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			require.NoError(t, tenantsv1.AddToScheme(scheme))

			registry := &tenantsv1.TenantRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(registry).
				WithStatusSubresource(registry).
				Build()

			r := &TenantRegistryReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			// Call updateStatus
			r.updateStatus(ctx, registry, tt.referencingTemplates, tt.desired, tt.ready, tt.failed, tt.synced)

			// Verify status was updated
			updated := &tenantsv1.TenantRegistry{}
			err := fakeClient.Get(ctx, types.NamespacedName{Name: registry.Name, Namespace: registry.Namespace}, updated)
			require.NoError(t, err)

			assert.Equal(t, tt.referencingTemplates, updated.Status.ReferencingTemplates)
			assert.Equal(t, tt.desired, updated.Status.Desired)
			assert.Equal(t, tt.ready, updated.Status.Ready)
			assert.Equal(t, tt.failed, updated.Status.Failed)

			// Verify condition
			var readyCondition *metav1.Condition
			for i := range updated.Status.Conditions {
				if updated.Status.Conditions[i].Type == "Ready" {
					readyCondition = &updated.Status.Conditions[i]
					break
				}
			}

			require.NotNil(t, readyCondition, "Ready condition should be set")
			assert.Equal(t, tt.wantConditionStatus, readyCondition.Status)
			assert.Equal(t, tt.wantConditionReason, readyCondition.Reason)
		})
	}
}

// TestCleanupRetainResources tests the cleanupRetainResources function
func TestCleanupRetainResources(t *testing.T) {
	tests := []struct {
		name     string
		registry *tenantsv1.TenantRegistry
		tenants  []tenantsv1.Tenant
		wantErr  bool
	}{
		{
			name: "no tenants to cleanup",
			registry: &tenantsv1.TenantRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
			},
			tenants: []tenantsv1.Tenant{},
			wantErr: false,
		},
		{
			name: "cleanup multiple tenants",
			registry: &tenantsv1.TenantRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
			},
			tenants: []tenantsv1.Tenant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "test-registry",
						},
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant1",
						TemplateRef: "web",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant2-web",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes-tenants.org/registry": "test-registry",
						},
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant2",
						TemplateRef: "web",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			require.NoError(t, tenantsv1.AddToScheme(scheme))
			require.NoError(t, corev1.AddToScheme(scheme))

			objects := []runtime.Object{tt.registry}
			for i := range tt.tenants {
				objects = append(objects, &tt.tenants[i])
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			recorder := record.NewFakeRecorder(100)

			r := &TenantRegistryReconciler{
				Client:   fakeClient,
				Scheme:   scheme,
				Recorder: recorder,
			}

			err := r.cleanupRetainResources(ctx, tt.registry)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProcessRetainResourcesForTenant tests the processRetainResourcesForTenant function
func TestProcessRetainResourcesForTenant(t *testing.T) {
	tests := []struct {
		name            string
		tenant          *tenantsv1.Tenant
		existingObjects []runtime.Object
		wantErr         bool
	}{
		{
			name: "no retain resources",
			tenant: &tenantsv1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant1-web",
					Namespace: "default",
					UID:       "tenant1-uid",
				},
				Spec: tenantsv1.TenantSpec{
					UID:         "tenant1",
					TemplateRef: "web",
					ConfigMaps: []tenantsv1.TResource{
						{
							ID:             "cm1",
							NameTemplate:   "tenant1-config",
							DeletionPolicy: tenantsv1.DeletionPolicyDelete,
							Spec: unstructured.Unstructured{
								Object: map[string]interface{}{
									"apiVersion": "v1",
									"kind":       "ConfigMap",
								},
							},
						},
					},
				},
			},
			existingObjects: []runtime.Object{},
			wantErr:         false,
		},
		{
			name: "retain resource exists - remove ownerReference",
			tenant: &tenantsv1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant1-web",
					Namespace: "default",
					UID:       "tenant1-uid",
				},
				Spec: tenantsv1.TenantSpec{
					UID:         "tenant1",
					TemplateRef: "web",
					ConfigMaps: []tenantsv1.TResource{
						{
							ID:             "cm1",
							NameTemplate:   "tenant1-config",
							DeletionPolicy: tenantsv1.DeletionPolicyRetain,
							Spec: unstructured.Unstructured{
								Object: map[string]interface{}{
									"apiVersion": "v1",
									"kind":       "ConfigMap",
								},
							},
						},
					},
				},
			},
			existingObjects: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-config",
						Namespace: "default",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "operator.kubernetes-tenants.org/v1",
								Kind:       "Tenant",
								Name:       "tenant1-web",
								UID:        "tenant1-uid",
							},
						},
					},
					Data: map[string]string{
						"key": "value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "retain resource not found - skip gracefully",
			tenant: &tenantsv1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant1-web",
					Namespace: "default",
					UID:       "tenant1-uid",
				},
				Spec: tenantsv1.TenantSpec{
					UID:         "tenant1",
					TemplateRef: "web",
					Secrets: []tenantsv1.TResource{
						{
							ID:             "secret1",
							NameTemplate:   "tenant1-secret",
							DeletionPolicy: tenantsv1.DeletionPolicyRetain,
							Spec: unstructured.Unstructured{
								Object: map[string]interface{}{
									"apiVersion": "v1",
									"kind":       "Secret",
								},
							},
						},
					},
				},
			},
			existingObjects: []runtime.Object{},
			wantErr:         false, // Should not error when resource not found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			require.NoError(t, tenantsv1.AddToScheme(scheme))
			require.NoError(t, corev1.AddToScheme(scheme))

			objects := append([]runtime.Object{tt.tenant}, tt.existingObjects...)

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			recorder := record.NewFakeRecorder(100)

			r := &TenantRegistryReconciler{
				Client:   fakeClient,
				Scheme:   scheme,
				Recorder: recorder,
			}

			err := r.processRetainResourcesForTenant(ctx, tt.tenant)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestContainsString tests the containsString helper function
func TestContainsString(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		str   string
		want  bool
	}{
		{
			name:  "string found in slice",
			slice: []string{"a", "b", "c"},
			str:   "b",
			want:  true,
		},
		{
			name:  "string not found in slice",
			slice: []string{"a", "b", "c"},
			str:   "d",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			str:   "a",
			want:  false,
		},
		{
			name:  "empty string found",
			slice: []string{"a", "", "c"},
			str:   "",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsString(tt.slice, tt.str)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestRemoveString tests the removeString helper function
func TestRemoveString(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		str   string
		want  []string
	}{
		{
			name:  "remove existing string",
			slice: []string{"a", "b", "c"},
			str:   "b",
			want:  []string{"a", "c"},
		},
		{
			name:  "remove non-existing string",
			slice: []string{"a", "b", "c"},
			str:   "d",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "remove from empty slice",
			slice: []string{},
			str:   "a",
			want:  []string{},
		},
		{
			name:  "remove multiple occurrences",
			slice: []string{"a", "b", "a", "c", "a"},
			str:   "a",
			want:  []string{"b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeString(tt.slice, tt.str)
			assert.Equal(t, tt.want, result)
		})
	}
}
