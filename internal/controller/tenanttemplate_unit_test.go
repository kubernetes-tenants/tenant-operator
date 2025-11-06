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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
)

// TestCheckTenantStatuses tests the checkTenantStatuses function
func TestCheckTenantStatuses(t *testing.T) {
	tests := []struct {
		name             string
		template         *tenantsv1.TenantTemplate
		tenants          []tenantsv1.Tenant
		wantTotalTenants int32
		wantReadyTenants int32
		wantErr          bool
	}{
		{
			name: "no tenants using template",
			template: &tenantsv1.TenantTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: tenantsv1.TenantTemplateSpec{
					RegistryID: "test-registry",
				},
			},
			tenants:          []tenantsv1.Tenant{},
			wantTotalTenants: 0,
			wantReadyTenants: 0,
			wantErr:          false,
		},
		{
			name: "all tenants using template are ready",
			template: &tenantsv1.TenantTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: tenantsv1.TenantTemplateSpec{
					RegistryID: "test-registry",
				},
			},
			tenants: []tenantsv1.Tenant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web-app",
						Namespace: "default",
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant1",
						TemplateRef: "web-app",
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
						Name:      "tenant2-web-app",
						Namespace: "default",
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant2",
						TemplateRef: "web-app",
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
			wantTotalTenants: 2,
			wantReadyTenants: 2,
			wantErr:          false,
		},
		{
			name: "mixed ready and not ready tenants",
			template: &tenantsv1.TenantTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: tenantsv1.TenantTemplateSpec{
					RegistryID: "test-registry",
				},
			},
			tenants: []tenantsv1.Tenant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web-app",
						Namespace: "default",
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant1",
						TemplateRef: "web-app",
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
						Name:      "tenant2-web-app",
						Namespace: "default",
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant2",
						TemplateRef: "web-app",
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
						Name:      "tenant3-web-app",
						Namespace: "default",
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant3",
						TemplateRef: "web-app",
					},
					Status: tenantsv1.TenantStatus{
						Conditions: []metav1.Condition{},
					},
				},
			},
			wantTotalTenants: 3,
			wantReadyTenants: 1,
			wantErr:          false,
		},
		{
			name: "exclude tenants using different template",
			template: &tenantsv1.TenantTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: tenantsv1.TenantTemplateSpec{
					RegistryID: "test-registry",
				},
			},
			tenants: []tenantsv1.Tenant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web-app",
						Namespace: "default",
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant1",
						TemplateRef: "web-app",
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
						Name:      "tenant2-worker",
						Namespace: "default",
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant2",
						TemplateRef: "worker", // Different template
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
			wantTotalTenants: 1, // Only tenant1 uses web-app template
			wantReadyTenants: 1,
			wantErr:          false,
		},
		{
			name: "exclude tenants in different namespace",
			template: &tenantsv1.TenantTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: tenantsv1.TenantTemplateSpec{
					RegistryID: "test-registry",
				},
			},
			tenants: []tenantsv1.Tenant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web-app",
						Namespace: "other-namespace",
					},
					Spec: tenantsv1.TenantSpec{
						UID:         "tenant1",
						TemplateRef: "web-app",
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
			wantTotalTenants: 0, // Different namespace
			wantReadyTenants: 0,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			require.NoError(t, tenantsv1.AddToScheme(scheme))

			objects := []runtime.Object{tt.template}
			for i := range tt.tenants {
				objects = append(objects, &tt.tenants[i])
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			r := &TenantTemplateReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			totalTenants, readyTenants, err := r.checkTenantStatuses(ctx, tt.template)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantTotalTenants, totalTenants, "Expected %d total tenants, got %d", tt.wantTotalTenants, totalTenants)
			assert.Equal(t, tt.wantReadyTenants, readyTenants, "Expected %d ready tenants, got %d", tt.wantReadyTenants, readyTenants)
		})
	}
}

// TestFindTemplateForTenant tests the findTemplateForTenant mapping function
func TestFindTemplateForTenant(t *testing.T) {
	tests := []struct {
		name           string
		tenant         *tenantsv1.Tenant
		wantRequests   []reconcile.Request
		wantNumResults int
	}{
		{
			name: "tenant with template reference",
			tenant: &tenantsv1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant1-web-app",
					Namespace: "default",
				},
				Spec: tenantsv1.TenantSpec{
					UID:         "tenant1",
					TemplateRef: "web-app",
				},
			},
			wantRequests: []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      "web-app",
						Namespace: "default",
					},
				},
			},
			wantNumResults: 1,
		},
		{
			name: "tenant without template reference",
			tenant: &tenantsv1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant1-orphan",
					Namespace: "default",
				},
				Spec: tenantsv1.TenantSpec{
					UID:         "tenant1",
					TemplateRef: "",
				},
			},
			wantRequests:   nil,
			wantNumResults: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			require.NoError(t, tenantsv1.AddToScheme(scheme))

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			r := &TenantTemplateReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			results := r.findTemplateForTenant(ctx, tt.tenant)

			assert.Equal(t, tt.wantNumResults, len(results))
			if tt.wantNumResults > 0 {
				assert.Equal(t, tt.wantRequests, results)
			}
		})
	}
}

// TestFindRegistryForTemplate tests the findRegistryForTemplate mapping function
func TestFindRegistryForTemplate(t *testing.T) {
	tests := []struct {
		name           string
		template       *tenantsv1.TenantTemplate
		wantRequests   []reconcile.Request
		wantNumResults int
	}{
		{
			name: "template with registry reference",
			template: &tenantsv1.TenantTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: tenantsv1.TenantTemplateSpec{
					RegistryID: "test-registry",
				},
			},
			wantRequests: []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      "test-registry",
						Namespace: "default",
					},
				},
			},
			wantNumResults: 1,
		},
		{
			name: "template with empty registry reference still triggers reconcile",
			template: &tenantsv1.TenantTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "standalone-template",
					Namespace: "default",
				},
				Spec: tenantsv1.TenantTemplateSpec{
					RegistryID: "", // Empty registry ID
				},
			},
			wantRequests: []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      "", // Empty name will be passed to registry reconciler
						Namespace: "default",
					},
				},
			},
			// Note: Mapping function always returns a request even for empty registryID
			// The registry reconciler will handle the NotFound case appropriately
			wantNumResults: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			require.NoError(t, tenantsv1.AddToScheme(scheme))

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			r := &TenantRegistryReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			results := r.findRegistryForTemplate(ctx, tt.template)

			assert.Equal(t, tt.wantNumResults, len(results))
			if tt.wantNumResults > 0 {
				assert.Equal(t, tt.wantRequests, results)
			}
		})
	}
}
