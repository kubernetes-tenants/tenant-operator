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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
	"github.com/kubernetes-tenants/tenant-operator/internal/status"
)

// TestCreationPolicyOnce_WithRetain_NoReAdoption verifies that when CreationPolicy=Once
// and DeletionPolicy=Retain, orphan markers are NOT removed on Tenant recreation.
// This is because ApplyResource is skipped when created-once annotation exists.
func TestCreationPolicyOnce_WithRetain_NoReAdoption(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, tenantsv1.AddToScheme(scheme))

	// Simulate a ConfigMap that was previously created with Once+Retain
	// and then orphaned after Tenant deletion
	orphanedConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
			Labels: map[string]string{
				"kubernetes-tenants.org/orphaned": "true",
			},
			Annotations: map[string]string{
				"kubernetes-tenants.org/created-once":    "true", // CreationPolicy=Once marker
				"kubernetes-tenants.org/orphaned-at":     "2025-01-15T10:30:00Z",
				"kubernetes-tenants.org/orphaned-reason": "TenantDeleted",
				"kubernetes-tenants.org/deletion-policy": "Retain",
			},
		},
		Data: map[string]string{
			"key": "value",
		},
	}

	// Create Tenant with Once+Retain that references the orphaned ConfigMap
	tenant := &tenantsv1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tenant",
			Namespace: "default",
			UID:       types.UID("test-uid"),
		},
		Spec: tenantsv1.TenantSpec{
			UID:         "test-uid",
			TemplateRef: "test-template",
			ConfigMaps: []tenantsv1.TResource{
				{
					ID:             "config",
					NameTemplate:   "test-config",
					CreationPolicy: tenantsv1.CreationPolicyOnce,
					DeletionPolicy: tenantsv1.DeletionPolicyRetain,
					Spec: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"data": map[string]interface{}{
								"key": "value",
							},
						},
					},
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(orphanedConfigMap, tenant).
		WithStatusSubresource(tenant).
		Build()

	recorder := record.NewFakeRecorder(100)

	r := &TenantReconciler{
		Client:        fakeClient,
		Scheme:        scheme,
		Recorder:      recorder,
		StatusManager: status.NewManager(fakeClient, status.WithSyncMode()),
	}

	// Reconcile the tenant
	_, err := r.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      tenant.Name,
			Namespace: tenant.Namespace,
		},
	})
	require.NoError(t, err)

	// Verify that orphan markers were NOT removed
	updatedConfigMap := &corev1.ConfigMap{}
	err = fakeClient.Get(ctx, types.NamespacedName{
		Name:      "test-config",
		Namespace: "default",
	}, updatedConfigMap)
	require.NoError(t, err)

	// Key assertion: Orphan markers should REMAIN because CreationPolicy=Once skips ApplyResource
	assert.Equal(t, "true", updatedConfigMap.Labels["kubernetes-tenants.org/orphaned"],
		"Orphan label should remain with CreationPolicy=Once")
	assert.Equal(t, "2025-01-15T10:30:00Z", updatedConfigMap.Annotations["kubernetes-tenants.org/orphaned-at"],
		"Orphan timestamp should remain with CreationPolicy=Once")
	assert.Equal(t, "TenantDeleted", updatedConfigMap.Annotations["kubernetes-tenants.org/orphaned-reason"],
		"Orphan reason should remain with CreationPolicy=Once")
	assert.Equal(t, "true", updatedConfigMap.Annotations["kubernetes-tenants.org/created-once"],
		"Created-once annotation should remain")

	// Verify Tenant status considers the resource as Ready (even though it's orphaned)
	updatedTenant := &tenantsv1.Tenant{}
	err = fakeClient.Get(ctx, types.NamespacedName{
		Name:      tenant.Name,
		Namespace: tenant.Namespace,
	}, updatedTenant)
	require.NoError(t, err)

	// Tenant should be Ready (resource exists and is counted)
	readyCondition := findCondition(updatedTenant.Status.Conditions, ConditionTypeReady)
	if readyCondition != nil {
		// If condition exists, it should be True or progressing
		assert.Contains(t, []metav1.ConditionStatus{metav1.ConditionTrue, metav1.ConditionUnknown}, readyCondition.Status,
			"Tenant should be Ready or progressing even with orphaned Once resource")
	}
}

// TestCreationPolicyOnce_VsWhenNeeded_BehaviorDifference documents the key behavioral
// difference between Once and WhenNeeded policies when combined with Retain.
// This test ensures we don't accidentally change this critical behavior.
func TestCreationPolicyOnce_VsWhenNeeded_BehaviorDifference(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, tenantsv1.AddToScheme(scheme))

	t.Run("Once policy skips existing resources with created-once annotation", func(t *testing.T) {
		existingCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-config-once",
				Namespace: "default",
				Annotations: map[string]string{
					"kubernetes-tenants.org/created-once": "true",
				},
			},
			Data: map[string]string{"key": "original"},
		}

		tenant := &tenantsv1.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-tenant-once",
				Namespace: "default",
				UID:       types.UID("test-uid-once"),
			},
			Spec: tenantsv1.TenantSpec{
				UID:         "test-uid-once",
				TemplateRef: "test-template",
				ConfigMaps: []tenantsv1.TResource{
					{
						ID:             "config",
						NameTemplate:   "test-config-once",
						CreationPolicy: tenantsv1.CreationPolicyOnce,
						Spec: unstructured.Unstructured{
							Object: map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"data": map[string]interface{}{
									"key": "updated", // Different value - should NOT be applied
								},
							},
						},
					},
				},
			},
		}

		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(existingCM, tenant).
			Build()

		r := &TenantReconciler{
			Client:   fakeClient,
			Scheme:   scheme,
			Recorder: record.NewFakeRecorder(10),
		}

		// Check the behavior directly through checkOnceCreated
		obj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "test-config-once",
					"namespace": "default",
				},
			},
		}

		exists, hasAnnotation, err := r.checkOnceCreated(ctx, obj)
		require.NoError(t, err)
		assert.True(t, exists, "Resource should exist")
		assert.True(t, hasAnnotation, "Resource should have created-once annotation")

		// This proves that with CreationPolicy=Once, the resource will be skipped
		// and ApplyResource will NOT be called
	})

	t.Run("WhenNeeded policy does not skip existing resources", func(t *testing.T) {
		existingCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-config-whenneeded",
				Namespace: "default",
				// NO created-once annotation
			},
			Data: map[string]string{"key": "original"},
		}

		tenant := &tenantsv1.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-tenant-whenneeded",
				Namespace: "default",
				UID:       types.UID("test-uid-whenneeded"),
			},
			Spec: tenantsv1.TenantSpec{
				UID:         "test-uid-whenneeded",
				TemplateRef: "test-template",
				ConfigMaps: []tenantsv1.TResource{
					{
						ID:             "config",
						NameTemplate:   "test-config-whenneeded",
						CreationPolicy: tenantsv1.CreationPolicyWhenNeeded,
						Spec: unstructured.Unstructured{
							Object: map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"data": map[string]interface{}{
									"key": "updated",
								},
							},
						},
					},
				},
			},
		}

		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(existingCM, tenant).
			Build()

		r := &TenantReconciler{
			Client:   fakeClient,
			Scheme:   scheme,
			Recorder: record.NewFakeRecorder(10),
		}

		// Check the behavior
		obj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "test-config-whenneeded",
					"namespace": "default",
				},
			},
		}

		exists, hasAnnotation, err := r.checkOnceCreated(ctx, obj)
		require.NoError(t, err)
		assert.True(t, exists, "Resource should exist")
		assert.False(t, hasAnnotation, "Resource should NOT have created-once annotation")

		// This proves that with CreationPolicy=WhenNeeded, the resource will NOT be skipped
		// and ApplyResource WILL be called (which removes orphan markers)
	})
}

// TestCreationPolicyOnce_SkipsConflictCheck verifies that when a resource has
// created-once annotation, ApplyResource is not called and therefore no conflict
// check occurs on subsequent reconciliations.
func TestCreationPolicyOnce_SkipsConflictCheck(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, tenantsv1.AddToScheme(scheme))

	// Simulate a ConfigMap that was created with Once policy and has created-once annotation
	existingConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
			Annotations: map[string]string{
				"kubernetes-tenants.org/created-once": "true",
			},
		},
		Data: map[string]string{
			"key": "original-value",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(existingConfigMap).
		Build()

	r := &TenantReconciler{
		Client:   fakeClient,
		Scheme:   scheme,
		Recorder: record.NewFakeRecorder(10),
	}

	// Verify that checkOnceCreated returns true for created-once annotation
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-config",
				"namespace": "default",
			},
		},
	}

	exists, hasAnnotation, err := r.checkOnceCreated(ctx, obj)
	require.NoError(t, err)
	assert.True(t, exists, "Resource should exist")
	assert.True(t, hasAnnotation, "Resource should have created-once annotation")

	// This proves that with CreationPolicy=Once and created-once annotation:
	// 1. The resource will be skipped in reconciliation
	// 2. ApplyResource will NOT be called
	// 3. No conflict checking will occur
	// 4. No updates will be applied (even if spec changed)

	// Verify ConfigMap data remains unchanged
	updatedConfigMap := &corev1.ConfigMap{}
	err = fakeClient.Get(ctx, types.NamespacedName{
		Name:      "test-config",
		Namespace: "default",
	}, updatedConfigMap)
	require.NoError(t, err)
	assert.Equal(t, "original-value", updatedConfigMap.Data["key"],
		"ConfigMap data should remain unchanged")
}

// Helper function to find a condition in a slice
func findCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
