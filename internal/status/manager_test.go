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

package status

import (
	"context"
	"testing"
	"time"

	lynqv1 "github.com/k8s-lynq/lynq/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestStatusUpdate_Apply(t *testing.T) {
	key := client.ObjectKey{Name: "test-tenant", Namespace: "default"}
	update := NewStatusUpdate(key)

	// Test resource counts update
	event1 := StatusEvent{
		Type:      EventResourceCountsUpdated,
		TenantKey: key,
		Payload: ResourceCountsPayload{
			Ready:      5,
			Failed:     1,
			Desired:    6,
			Conflicted: 0,
		},
		Timestamp: time.Now(),
	}
	update.Apply(event1)

	assert.NotNil(t, update.ReadyResources)
	assert.Equal(t, int32(5), *update.ReadyResources)
	assert.NotNil(t, update.FailedResources)
	assert.Equal(t, int32(1), *update.FailedResources)
	assert.NotNil(t, update.DesiredResources)
	assert.Equal(t, int32(6), *update.DesiredResources)

	// Test condition update
	event2 := StatusEvent{
		Type:      EventConditionChanged,
		TenantKey: key,
		Payload: ConditionPayload{
			Condition: metav1.Condition{
				Type:   "Ready",
				Status: metav1.ConditionTrue,
				Reason: "AllResourcesReady",
			},
		},
		Timestamp: time.Now(),
	}
	update.Apply(event2)

	assert.Len(t, update.Conditions, 1)
	assert.Equal(t, "Ready", update.Conditions["Ready"].Type)
	assert.Equal(t, metav1.ConditionTrue, update.Conditions["Ready"].Status)

	// Test applied resources update
	event3 := StatusEvent{
		Type:      EventAppliedResourcesUpdated,
		TenantKey: key,
		Payload: AppliedResourcesPayload{
			Keys: []string{"Deployment/default/app1@dep1", "Service/default/svc1@svc1"},
		},
		Timestamp: time.Now(),
	}
	update.Apply(event3)

	assert.Len(t, update.AppliedResources, 2)
	assert.Contains(t, update.AppliedResources, "Deployment/default/app1@dep1")
}

func TestStatusUpdate_HasChanges(t *testing.T) {
	key := client.ObjectKey{Name: "test-tenant", Namespace: "default"}

	// Empty update has no changes
	update := NewStatusUpdate(key)
	assert.False(t, update.HasChanges())

	// Update with resource counts has changes
	ready := int32(5)
	update.ReadyResources = &ready
	assert.True(t, update.HasChanges())

	// New update with conditions has changes
	update2 := NewStatusUpdate(key)
	update2.Conditions["Ready"] = metav1.Condition{Type: "Ready"}
	assert.True(t, update2.HasChanges())
}

func TestManager_PublishSync(t *testing.T) {
	// Setup
	scheme := runtime.NewScheme()
	err := lynqv1.AddToScheme(scheme)
	require.NoError(t, err)

	tenant := &lynqv1.LynqNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-tenant",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: lynqv1.LynqNodeSpec{
			UID:         "tenant-123",
			TemplateRef: "test-template",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(tenant).
		WithStatusSubresource(tenant).
		Build()

	// Create manager in sync mode for testing
	manager := NewManager(fakeClient, WithSyncMode())

	// Publish resource counts
	manager.PublishResourceCounts(tenant, 5, 1, 6, 0)

	// Verify status was updated
	updated := &lynqv1.LynqNode{}
	err = fakeClient.Get(context.Background(), client.ObjectKeyFromObject(tenant), updated)
	require.NoError(t, err)

	assert.Equal(t, int32(5), updated.Status.ReadyResources)
	assert.Equal(t, int32(1), updated.Status.FailedResources)
	assert.Equal(t, int32(6), updated.Status.DesiredResources)
}

func TestManager_PublishConditionSync(t *testing.T) {
	// Setup
	scheme := runtime.NewScheme()
	err := lynqv1.AddToScheme(scheme)
	require.NoError(t, err)

	tenant := &lynqv1.LynqNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-tenant",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: lynqv1.LynqNodeSpec{
			UID:         "tenant-123",
			TemplateRef: "test-template",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(tenant).
		WithStatusSubresource(tenant).
		Build()

	manager := NewManager(fakeClient, WithSyncMode())

	// Publish Ready condition
	manager.PublishReadyCondition(tenant, true, "AllResourcesReady", "All 6 resources are ready")

	// Verify condition was updated
	updated := &lynqv1.LynqNode{}
	err = fakeClient.Get(context.Background(), client.ObjectKeyFromObject(tenant), updated)
	require.NoError(t, err)

	require.Len(t, updated.Status.Conditions, 1)
	assert.Equal(t, "Ready", updated.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionTrue, updated.Status.Conditions[0].Status)
	assert.Equal(t, "AllResourcesReady", updated.Status.Conditions[0].Reason)
}

func TestManager_PublishMultipleConditionsSync(t *testing.T) {
	// Setup
	scheme := runtime.NewScheme()
	err := lynqv1.AddToScheme(scheme)
	require.NoError(t, err)

	tenant := &lynqv1.LynqNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-tenant",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: lynqv1.LynqNodeSpec{
			UID:         "tenant-123",
			TemplateRef: "test-template",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(tenant).
		WithStatusSubresource(tenant).
		Build()

	manager := NewManager(fakeClient, WithSyncMode())

	// Publish multiple conditions
	manager.PublishReadyCondition(tenant, true, "AllResourcesReady", "All resources ready")
	manager.PublishProgressingCondition(tenant, false, "ReconcileComplete", "Reconciliation completed")
	manager.PublishConflictedCondition(tenant, false)
	manager.PublishDegradedCondition(tenant, false, "Healthy", "All resources healthy")

	// Verify all conditions were updated
	updated := &lynqv1.LynqNode{}
	err = fakeClient.Get(context.Background(), client.ObjectKeyFromObject(tenant), updated)
	require.NoError(t, err)

	assert.Len(t, updated.Status.Conditions, 4)

	// Check each condition
	conditionMap := make(map[string]metav1.Condition)
	for _, cond := range updated.Status.Conditions {
		conditionMap[cond.Type] = cond
	}

	assert.Equal(t, metav1.ConditionTrue, conditionMap["Ready"].Status)
	assert.Equal(t, metav1.ConditionFalse, conditionMap["Progressing"].Status)
	assert.Equal(t, metav1.ConditionFalse, conditionMap["Conflicted"].Status)
	assert.Equal(t, metav1.ConditionFalse, conditionMap["Degraded"].Status)
}

func TestManager_PublishAppliedResourcesSync(t *testing.T) {
	// Setup
	scheme := runtime.NewScheme()
	err := lynqv1.AddToScheme(scheme)
	require.NoError(t, err)

	tenant := &lynqv1.LynqNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-tenant",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: lynqv1.LynqNodeSpec{
			UID:         "tenant-123",
			TemplateRef: "test-template",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(tenant).
		WithStatusSubresource(tenant).
		Build()

	manager := NewManager(fakeClient, WithSyncMode())

	// Publish applied resources
	keys := []string{
		"Deployment/default/app1@dep1",
		"Service/default/svc1@svc1",
	}
	manager.PublishAppliedResources(tenant, keys)

	// Verify applied resources were updated
	updated := &lynqv1.LynqNode{}
	err = fakeClient.Get(context.Background(), client.ObjectKeyFromObject(tenant), updated)
	require.NoError(t, err)

	assert.Len(t, updated.Status.AppliedResources, 2)
	assert.Contains(t, updated.Status.AppliedResources, "Deployment/default/app1@dep1")
	assert.Contains(t, updated.Status.AppliedResources, "Service/default/svc1@svc1")
}

func TestManager_PublishFullStatusSync(t *testing.T) {
	// Setup
	scheme := runtime.NewScheme()
	err := lynqv1.AddToScheme(scheme)
	require.NoError(t, err)

	tenant := &lynqv1.LynqNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-tenant",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: lynqv1.LynqNodeSpec{
			UID:         "tenant-123",
			TemplateRef: "test-template",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(tenant).
		WithStatusSubresource(tenant).
		Build()

	manager := NewManager(fakeClient, WithSyncMode())

	// Publish full status
	conditions := []metav1.Condition{
		{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			Reason:             "AllResourcesReady",
			Message:            "All resources ready",
			LastTransitionTime: metav1.Now(),
		},
		{
			Type:               "Progressing",
			Status:             metav1.ConditionFalse,
			Reason:             "ReconcileComplete",
			Message:            "Reconciliation completed",
			LastTransitionTime: metav1.Now(),
		},
	}
	appliedKeys := []string{"Deployment/default/app1@dep1"}

	manager.PublishFullStatus(tenant, 5, 1, 6, 0, conditions, appliedKeys, false, "")

	// Verify everything was updated
	updated := &lynqv1.LynqNode{}
	err = fakeClient.Get(context.Background(), client.ObjectKeyFromObject(tenant), updated)
	require.NoError(t, err)

	// Check resource counts
	assert.Equal(t, int32(5), updated.Status.ReadyResources)
	assert.Equal(t, int32(1), updated.Status.FailedResources)
	assert.Equal(t, int32(6), updated.Status.DesiredResources)

	// Check conditions
	assert.Len(t, updated.Status.Conditions, 2)

	// Check applied resources
	assert.Len(t, updated.Status.AppliedResources, 1)
	assert.Contains(t, updated.Status.AppliedResources, "Deployment/default/app1@dep1")
}

func TestManager_UpdateConditionDeduplication(t *testing.T) {
	// Setup
	scheme := runtime.NewScheme()
	err := lynqv1.AddToScheme(scheme)
	require.NoError(t, err)

	tenant := &lynqv1.LynqNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-tenant",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: lynqv1.LynqNodeSpec{
			UID:         "tenant-123",
			TemplateRef: "test-template",
		},
		Status: lynqv1.LynqNodeStatus{
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionFalse,
					Reason:             "Progressing",
					Message:            "Resources are being applied",
					LastTransitionTime: metav1.Now(),
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(tenant).
		WithStatusSubresource(tenant).
		Build()

	manager := NewManager(fakeClient, WithSyncMode())

	// Update the same condition with different status
	manager.PublishReadyCondition(tenant, true, "AllResourcesReady", "All resources ready")

	// Verify condition was updated (not duplicated)
	updated := &lynqv1.LynqNode{}
	err = fakeClient.Get(context.Background(), client.ObjectKeyFromObject(tenant), updated)
	require.NoError(t, err)

	assert.Len(t, updated.Status.Conditions, 1)
	assert.Equal(t, "Ready", updated.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionTrue, updated.Status.Conditions[0].Status)
	assert.Equal(t, "AllResourcesReady", updated.Status.Conditions[0].Reason)
}

func TestManager_HandleDeletedLynqNode(t *testing.T) {
	// Setup
	scheme := runtime.NewScheme()
	err := lynqv1.AddToScheme(scheme)
	require.NoError(t, err)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	manager := NewManager(fakeClient, WithSyncMode())

	// Try to publish to a non-existent tenant
	tenant := &lynqv1.LynqNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deleted-tenant",
			Namespace: "default",
		},
	}

	// This should not error - it should gracefully handle missing tenant
	manager.PublishResourceCounts(tenant, 5, 1, 6, 0)

	// Verify tenant still doesn't exist
	updated := &lynqv1.LynqNode{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "deleted-tenant",
		Namespace: "default",
	}, updated)
	assert.Error(t, err)
}

func TestNewStatusUpdate(t *testing.T) {
	key := client.ObjectKey{Name: "test", Namespace: "default"}
	update := NewStatusUpdate(key)

	assert.Equal(t, key, update.Key)
	assert.NotNil(t, update.Conditions)
	assert.Empty(t, update.Conditions)
	assert.Nil(t, update.ReadyResources)
	assert.Nil(t, update.FailedResources)
}
