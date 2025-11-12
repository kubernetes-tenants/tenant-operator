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
	"time"

	lynqv1 "github.com/k8s-lynq/lynq/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PublishResourceCounts is a helper to publish resource count updates
func (m *Manager) PublishResourceCounts(tenant *lynqv1.LynqNode, ready, failed, desired, conflicted int32) {
	m.Publish(StatusEvent{
		Type:      EventResourceCountsUpdated,
		TenantKey: client.ObjectKeyFromObject(tenant),
		Payload: ResourceCountsPayload{
			Ready:      ready,
			Failed:     failed,
			Desired:    desired,
			Conflicted: conflicted,
		},
		Timestamp: time.Now(),
	})
}

// PublishCondition is a helper to publish condition updates
func (m *Manager) PublishCondition(tenant *lynqv1.LynqNode, conditionType string, status metav1.ConditionStatus, reason, message string) {
	m.Publish(StatusEvent{
		Type:      EventConditionChanged,
		TenantKey: client.ObjectKeyFromObject(tenant),
		Payload: ConditionPayload{
			Condition: metav1.Condition{
				Type:               conditionType,
				Status:             status,
				Reason:             reason,
				Message:            message,
				LastTransitionTime: metav1.Now(),
			},
		},
		Timestamp: time.Now(),
	})
}

// PublishReadyCondition is a helper to publish the Ready condition
func (m *Manager) PublishReadyCondition(tenant *lynqv1.LynqNode, isReady bool, reason, message string) {
	status := metav1.ConditionTrue
	if !isReady {
		status = metav1.ConditionFalse
	}
	m.PublishCondition(tenant, "Ready", status, reason, message)
}

// PublishProgressingCondition is a helper to publish the Progressing condition
func (m *Manager) PublishProgressingCondition(tenant *lynqv1.LynqNode, isProgressing bool, reason, message string) {
	status := metav1.ConditionTrue
	if !isProgressing {
		status = metav1.ConditionFalse
	}
	if reason == "" {
		reason = "ReconcileComplete"
	}
	if message == "" {
		message = "Reconciliation completed"
	}
	m.PublishCondition(tenant, "Progressing", status, reason, message)
}

// PublishConflictedCondition is a helper to publish the Conflicted condition
func (m *Manager) PublishConflictedCondition(tenant *lynqv1.LynqNode, hasConflict bool) {
	status := metav1.ConditionFalse
	reason := "NoConflict"
	message := "No resource conflicts detected"

	if hasConflict {
		status = metav1.ConditionTrue
		reason = "ResourceConflict"
		message = "One or more resources are in conflict. Check events for details."
	}

	m.PublishCondition(tenant, "Conflicted", status, reason, message)
}

// PublishDegradedCondition is a helper to publish the Degraded condition
func (m *Manager) PublishDegradedCondition(tenant *lynqv1.LynqNode, isDegraded bool, reason, message string) {
	status := metav1.ConditionFalse
	if isDegraded {
		status = metav1.ConditionTrue
	}
	m.PublishCondition(tenant, "Degraded", status, reason, message)
}

// PublishObservedGeneration is a helper to publish ObservedGeneration updates
func (m *Manager) PublishObservedGeneration(tenant *lynqv1.LynqNode, generation int64) {
	m.Publish(StatusEvent{
		Type:      EventObservedGenerationUpdated,
		TenantKey: client.ObjectKeyFromObject(tenant),
		Payload: ObservedGenerationPayload{
			ObservedGeneration: generation,
		},
		Timestamp: time.Now(),
	})
}

// PublishAppliedResources is a helper to publish applied resources list
func (m *Manager) PublishAppliedResources(tenant *lynqv1.LynqNode, keys []string) {
	m.Publish(StatusEvent{
		Type:      EventAppliedResourcesUpdated,
		TenantKey: client.ObjectKeyFromObject(tenant),
		Payload: AppliedResourcesPayload{
			Keys: keys,
		},
		Timestamp: time.Now(),
	})
}

// PublishMetrics is a helper to publish all metrics at once
func (m *Manager) PublishMetrics(tenant *lynqv1.LynqNode, ready, failed, desired, conflicted int32, conditions []metav1.Condition, isDegraded bool, degradedReason string) {
	m.Publish(StatusEvent{
		Type:      EventMetricsUpdate,
		TenantKey: client.ObjectKeyFromObject(tenant),
		Payload: MetricsPayload{
			Ready:          ready,
			Failed:         failed,
			Desired:        desired,
			Conflicted:     conflicted,
			Conditions:     conditions,
			IsDegraded:     isDegraded,
			DegradedReason: degradedReason,
		},
		Timestamp: time.Now(),
	})
}

// PublishFullStatus is a helper to publish all status updates at once
// This is useful at the end of reconciliation to update everything together
func (m *Manager) PublishFullStatus(tenant *lynqv1.LynqNode, ready, failed, desired, conflicted int32, conditions []metav1.Condition, appliedKeys []string, isDegraded bool, degradedReason string) {
	key := client.ObjectKeyFromObject(tenant)
	now := time.Now()

	// Publish resource counts
	m.Publish(StatusEvent{
		Type:      EventResourceCountsUpdated,
		TenantKey: key,
		Payload: ResourceCountsPayload{
			Ready:      ready,
			Failed:     failed,
			Desired:    desired,
			Conflicted: conflicted,
		},
		Timestamp: now,
	})

	// Publish all conditions
	for _, cond := range conditions {
		m.Publish(StatusEvent{
			Type:      EventConditionChanged,
			TenantKey: key,
			Payload: ConditionPayload{
				Condition: cond,
			},
			Timestamp: now,
		})
	}

	// Publish applied resources
	if appliedKeys != nil {
		m.Publish(StatusEvent{
			Type:      EventAppliedResourcesUpdated,
			TenantKey: key,
			Payload: AppliedResourcesPayload{
				Keys: appliedKeys,
			},
			Timestamp: now,
		})
	}

	// Publish metrics
	m.Publish(StatusEvent{
		Type:      EventMetricsUpdate,
		TenantKey: key,
		Payload: MetricsPayload{
			Ready:          ready,
			Failed:         failed,
			Desired:        desired,
			Conflicted:     conflicted,
			Conditions:     conditions,
			IsDegraded:     isDegraded,
			DegradedReason: degradedReason,
		},
		Timestamp: now,
	})
}
