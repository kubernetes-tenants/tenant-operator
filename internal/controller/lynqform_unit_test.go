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

	lynqv1 "github.com/k8s-lynq/lynq/api/v1"
)

// TestCheckLynqNodeStatuses tests the checkLynqNodeStatuses function
func TestCheckLynqNodeStatuses(t *testing.T) {
	tests := []struct {
		name           string
		template       *lynqv1.LynqForm
		tenants        []lynqv1.LynqNode
		wantTotalNodes int32
		wantReadyNodes int32
		wantErr        bool
	}{
		{
			name: "no tenants using template",
			template: &lynqv1.LynqForm{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: lynqv1.LynqFormSpec{
					RegistryID: "test-registry",
				},
			},
			tenants:        []lynqv1.LynqNode{},
			wantTotalNodes: 0,
			wantReadyNodes: 0,
			wantErr:        false,
		},
		{
			name: "all tenants using template are ready",
			template: &lynqv1.LynqForm{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: lynqv1.LynqFormSpec{
					RegistryID: "test-registry",
				},
			},
			tenants: []lynqv1.LynqNode{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web-app",
						Namespace: "default",
					},
					Spec: lynqv1.LynqNodeSpec{
						UID:         "tenant1",
						TemplateRef: "web-app",
					},
					Status: lynqv1.LynqNodeStatus{
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
					Spec: lynqv1.LynqNodeSpec{
						UID:         "tenant2",
						TemplateRef: "web-app",
					},
					Status: lynqv1.LynqNodeStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			wantTotalNodes: 2,
			wantReadyNodes: 2,
			wantErr:        false,
		},
		{
			name: "mixed ready and not ready nodes",
			template: &lynqv1.LynqForm{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: lynqv1.LynqFormSpec{
					RegistryID: "test-registry",
				},
			},
			tenants: []lynqv1.LynqNode{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web-app",
						Namespace: "default",
					},
					Spec: lynqv1.LynqNodeSpec{
						UID:         "tenant1",
						TemplateRef: "web-app",
					},
					Status: lynqv1.LynqNodeStatus{
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
					Spec: lynqv1.LynqNodeSpec{
						UID:         "tenant2",
						TemplateRef: "web-app",
					},
					Status: lynqv1.LynqNodeStatus{
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
					Spec: lynqv1.LynqNodeSpec{
						UID:         "tenant3",
						TemplateRef: "web-app",
					},
					Status: lynqv1.LynqNodeStatus{
						Conditions: []metav1.Condition{},
					},
				},
			},
			wantTotalNodes: 3,
			wantReadyNodes: 1,
			wantErr:        false,
		},
		{
			name: "exclude tenants using different template",
			template: &lynqv1.LynqForm{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: lynqv1.LynqFormSpec{
					RegistryID: "test-registry",
				},
			},
			tenants: []lynqv1.LynqNode{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web-app",
						Namespace: "default",
					},
					Spec: lynqv1.LynqNodeSpec{
						UID:         "tenant1",
						TemplateRef: "web-app",
					},
					Status: lynqv1.LynqNodeStatus{
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
					Spec: lynqv1.LynqNodeSpec{
						UID:         "tenant2",
						TemplateRef: "worker", // Different template
					},
					Status: lynqv1.LynqNodeStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			wantTotalNodes: 1, // Only tenant1 uses web-app template
			wantReadyNodes: 1,
			wantErr:        false,
		},
		{
			name: "exclude tenants in different namespace",
			template: &lynqv1.LynqForm{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: lynqv1.LynqFormSpec{
					RegistryID: "test-registry",
				},
			},
			tenants: []lynqv1.LynqNode{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant1-web-app",
						Namespace: "other-namespace",
					},
					Spec: lynqv1.LynqNodeSpec{
						UID:         "tenant1",
						TemplateRef: "web-app",
					},
					Status: lynqv1.LynqNodeStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			wantTotalNodes: 0, // Different namespace
			wantReadyNodes: 0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			require.NoError(t, lynqv1.AddToScheme(scheme))

			objects := []runtime.Object{tt.template}
			for i := range tt.tenants {
				objects = append(objects, &tt.tenants[i])
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			r := &LynqFormReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			totalLynqNodes, readyLynqNodes, err := r.checkLynqNodeStatuses(ctx, tt.template)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantTotalNodes, totalLynqNodes, "Expected %d total nodes, got %d", tt.wantTotalNodes, totalLynqNodes)
			assert.Equal(t, tt.wantReadyNodes, readyLynqNodes, "Expected %d ready nodes, got %d", tt.wantReadyNodes, readyLynqNodes)
		})
	}
}

// TestFindTemplateForLynqNode tests the findTemplateForLynqNode mapping function
func TestFindTemplateForLynqNode(t *testing.T) {
	tests := []struct {
		name           string
		tenant         *lynqv1.LynqNode
		wantRequests   []reconcile.Request
		wantNumResults int
	}{
		{
			name: "tenant with template reference",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant1-web-app",
					Namespace: "default",
				},
				Spec: lynqv1.LynqNodeSpec{
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
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant1-orphan",
					Namespace: "default",
				},
				Spec: lynqv1.LynqNodeSpec{
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
			require.NoError(t, lynqv1.AddToScheme(scheme))

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			r := &LynqFormReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			results := r.findTemplateForLynqNode(ctx, tt.tenant)

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
		template       *lynqv1.LynqForm
		wantRequests   []reconcile.Request
		wantNumResults int
	}{
		{
			name: "template with registry reference",
			template: &lynqv1.LynqForm{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: "default",
				},
				Spec: lynqv1.LynqFormSpec{
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
			template: &lynqv1.LynqForm{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "standalone-template",
					Namespace: "default",
				},
				Spec: lynqv1.LynqFormSpec{
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
			require.NoError(t, lynqv1.AddToScheme(scheme))

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			r := &LynqHubReconciler{
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
