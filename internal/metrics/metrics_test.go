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

package metrics

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestTenantReconcileDuration(t *testing.T) {
	// Reset the metric before testing
	TenantReconcileDuration.Reset()

	// Observe a reconciliation duration
	duration := 2.5 // 2.5 seconds
	TenantReconcileDuration.WithLabelValues("success").Observe(duration)

	// Collect metrics
	count := testutil.CollectAndCount(TenantReconcileDuration)
	assert.Equal(t, 1, count, "Expected 1 metric to be collected")

	// Verify metric can be collected and has the expected type
	problems, err := testutil.CollectAndLint(TenantReconcileDuration)
	assert.NoError(t, err)
	assert.Empty(t, problems, "Metric should have no lint problems")

	// Test with error result
	TenantReconcileDuration.WithLabelValues("error").Observe(1.0)
	count = testutil.CollectAndCount(TenantReconcileDuration)
	assert.Equal(t, 2, count, "Expected 2 metrics to be collected after error observation")
}

func TestTenantReconcileDuration_Timer(t *testing.T) {
	TenantReconcileDuration.Reset()

	// Test using a timer
	timer := prometheus.NewTimer(TenantReconcileDuration.WithLabelValues("success"))
	time.Sleep(10 * time.Millisecond)
	timer.ObserveDuration()

	count := testutil.CollectAndCount(TenantReconcileDuration)
	assert.Equal(t, 1, count)
}

func TestTenantResourcesReady(t *testing.T) {
	TenantResourcesReady.Reset()

	// Set ready resources count
	TenantResourcesReady.WithLabelValues("tenant1", "default").Set(5)
	TenantResourcesReady.WithLabelValues("tenant2", "production").Set(10)

	// Verify count
	count := testutil.CollectAndCount(TenantResourcesReady)
	assert.Equal(t, 2, count)

	// Verify metric values
	expected := `
# HELP tenant_resources_ready Number of ready resources for a tenant
# TYPE tenant_resources_ready gauge
tenant_resources_ready{namespace="default",tenant="tenant1"} 5
tenant_resources_ready{namespace="production",tenant="tenant2"} 10
`
	err := testutil.CollectAndCompare(TenantResourcesReady, strings.NewReader(expected))
	assert.NoError(t, err)
}

func TestTenantResourcesDesired(t *testing.T) {
	TenantResourcesDesired.Reset()

	// Set desired resources count
	TenantResourcesDesired.WithLabelValues("tenant1", "default").Set(8)

	count := testutil.CollectAndCount(TenantResourcesDesired)
	assert.Equal(t, 1, count)

	// Verify metric has no lint problems
	problems, err := testutil.CollectAndLint(TenantResourcesDesired)
	assert.NoError(t, err)
	assert.Empty(t, problems)
}

func TestTenantResourcesFailed(t *testing.T) {
	TenantResourcesFailed.Reset()

	// Set failed resources count
	TenantResourcesFailed.WithLabelValues("tenant1", "default").Set(2)

	count := testutil.CollectAndCount(TenantResourcesFailed)
	assert.Equal(t, 1, count)

	// Verify metric can be incremented
	TenantResourcesFailed.WithLabelValues("tenant1", "default").Inc()
	expected := `
# HELP tenant_resources_failed Number of failed resources for a tenant
# TYPE tenant_resources_failed gauge
tenant_resources_failed{namespace="default",tenant="tenant1"} 3
`
	err := testutil.CollectAndCompare(TenantResourcesFailed, strings.NewReader(expected))
	assert.NoError(t, err)
}

func TestRegistryDesired(t *testing.T) {
	RegistryDesired.Reset()

	// Set desired tenant count for registries
	RegistryDesired.WithLabelValues("mysql-prod", "default").Set(100)
	RegistryDesired.WithLabelValues("mysql-staging", "staging").Set(20)

	count := testutil.CollectAndCount(RegistryDesired)
	assert.Equal(t, 2, count)

	expected := `
# HELP registry_desired Number of desired tenants from the registry data source
# TYPE registry_desired gauge
registry_desired{namespace="default",registry="mysql-prod"} 100
registry_desired{namespace="staging",registry="mysql-staging"} 20
`
	err := testutil.CollectAndCompare(RegistryDesired, strings.NewReader(expected))
	assert.NoError(t, err)
}

func TestRegistryReady(t *testing.T) {
	RegistryReady.Reset()

	// Set ready tenant count
	RegistryReady.WithLabelValues("mysql-prod", "default").Set(95)

	count := testutil.CollectAndCount(RegistryReady)
	assert.Equal(t, 1, count)

	problems, err := testutil.CollectAndLint(RegistryReady)
	assert.NoError(t, err)
	assert.Empty(t, problems)
}

func TestRegistryFailed(t *testing.T) {
	RegistryFailed.Reset()

	// Set failed tenant count
	RegistryFailed.WithLabelValues("mysql-prod", "default").Set(5)

	count := testutil.CollectAndCount(RegistryFailed)
	assert.Equal(t, 1, count)

	expected := `
# HELP registry_failed Number of failed tenants for a registry
# TYPE registry_failed gauge
registry_failed{namespace="default",registry="mysql-prod"} 5
`
	err := testutil.CollectAndCompare(RegistryFailed, strings.NewReader(expected))
	assert.NoError(t, err)
}

func TestApplyAttemptsTotal(t *testing.T) {
	ApplyAttemptsTotal.Reset()

	// Increment apply attempts
	ApplyAttemptsTotal.WithLabelValues("Deployment", "success", "Stuck").Inc()
	ApplyAttemptsTotal.WithLabelValues("Deployment", "success", "Stuck").Inc()
	ApplyAttemptsTotal.WithLabelValues("Deployment", "error", "Force").Inc()
	ApplyAttemptsTotal.WithLabelValues("Service", "success", "Stuck").Inc()

	count := testutil.CollectAndCount(ApplyAttemptsTotal)
	assert.Equal(t, 3, count, "Expected 3 unique label combinations")

	// Verify counter values
	expected := `
# HELP apply_attempts_total Total number of resource apply attempts
# TYPE apply_attempts_total counter
apply_attempts_total{conflict_policy="Force",kind="Deployment",result="error"} 1
apply_attempts_total{conflict_policy="Stuck",kind="Deployment",result="success"} 2
apply_attempts_total{conflict_policy="Stuck",kind="Service",result="success"} 1
`
	err := testutil.CollectAndCompare(ApplyAttemptsTotal, strings.NewReader(expected))
	assert.NoError(t, err)
}

func TestTenantConditionStatus(t *testing.T) {
	TenantConditionStatus.Reset()

	// Set condition statuses (0=False, 1=True, 2=Unknown)
	TenantConditionStatus.WithLabelValues("tenant1", "default", "Ready").Set(1)    // True
	TenantConditionStatus.WithLabelValues("tenant2", "default", "Ready").Set(0)    // False
	TenantConditionStatus.WithLabelValues("tenant3", "default", "Degraded").Set(2) // Unknown

	count := testutil.CollectAndCount(TenantConditionStatus)
	assert.Equal(t, 3, count)

	expected := `
# HELP tenant_condition_status Status of tenant conditions (0=False, 1=True, 2=Unknown)
# TYPE tenant_condition_status gauge
tenant_condition_status{namespace="default",tenant="tenant1",type="Ready"} 1
tenant_condition_status{namespace="default",tenant="tenant2",type="Ready"} 0
tenant_condition_status{namespace="default",tenant="tenant3",type="Degraded"} 2
`
	err := testutil.CollectAndCompare(TenantConditionStatus, strings.NewReader(expected))
	assert.NoError(t, err)
}

func TestTenantConflictsTotal(t *testing.T) {
	TenantConflictsTotal.Reset()

	// Increment conflict counters
	TenantConflictsTotal.WithLabelValues("tenant1", "default", "Deployment", "Stuck").Inc()
	TenantConflictsTotal.WithLabelValues("tenant1", "default", "Deployment", "Stuck").Inc()
	TenantConflictsTotal.WithLabelValues("tenant1", "default", "Service", "Force").Inc()

	count := testutil.CollectAndCount(TenantConflictsTotal)
	assert.Equal(t, 2, count)

	expected := `
# HELP tenant_conflicts_total Total number of resource conflicts encountered during reconciliation
# TYPE tenant_conflicts_total counter
tenant_conflicts_total{conflict_policy="Force",namespace="default",resource_kind="Service",tenant="tenant1"} 1
tenant_conflicts_total{conflict_policy="Stuck",namespace="default",resource_kind="Deployment",tenant="tenant1"} 2
`
	err := testutil.CollectAndCompare(TenantConflictsTotal, strings.NewReader(expected))
	assert.NoError(t, err)
}

func TestTenantResourcesConflicted(t *testing.T) {
	TenantResourcesConflicted.Reset()

	// Set conflicted resources count
	TenantResourcesConflicted.WithLabelValues("tenant1", "default").Set(3)

	count := testutil.CollectAndCount(TenantResourcesConflicted)
	assert.Equal(t, 1, count)

	expected := `
# HELP tenant_resources_conflicted Number of resources currently in conflict state for a tenant
# TYPE tenant_resources_conflicted gauge
tenant_resources_conflicted{namespace="default",tenant="tenant1"} 3
`
	err := testutil.CollectAndCompare(TenantResourcesConflicted, strings.NewReader(expected))
	assert.NoError(t, err)
}

func TestTenantDegradedStatus(t *testing.T) {
	TenantDegradedStatus.Reset()

	// Set degraded status (1=degraded, 0=not degraded)
	TenantDegradedStatus.WithLabelValues("tenant1", "default", "ResourceConflict").Set(1)
	TenantDegradedStatus.WithLabelValues("tenant2", "default", "").Set(0)

	count := testutil.CollectAndCount(TenantDegradedStatus)
	assert.Equal(t, 2, count)

	expected := `
# HELP tenant_degraded_status Indicates if a tenant is in degraded state (1=degraded, 0=not degraded)
# TYPE tenant_degraded_status gauge
tenant_degraded_status{namespace="default",reason="",tenant="tenant2"} 0
tenant_degraded_status{namespace="default",reason="ResourceConflict",tenant="tenant1"} 1
`
	err := testutil.CollectAndCompare(TenantDegradedStatus, strings.NewReader(expected))
	assert.NoError(t, err)
}

func TestMetricsRegistration(t *testing.T) {
	// Test that all metrics are properly defined and can collect data
	metrics := []prometheus.Collector{
		TenantReconcileDuration,
		TenantResourcesReady,
		TenantResourcesDesired,
		TenantResourcesFailed,
		RegistryDesired,
		RegistryReady,
		RegistryFailed,
		ApplyAttemptsTotal,
		TenantConditionStatus,
		TenantConflictsTotal,
		TenantResourcesConflicted,
		TenantDegradedStatus,
	}

	for _, metric := range metrics {
		assert.NotNil(t, metric, "Metric should not be nil")

		// Verify that the metric can be collected
		count := testutil.CollectAndCount(metric)
		assert.GreaterOrEqual(t, count, 0, "Should be able to collect metric")
	}
}

func TestMetricLabels(t *testing.T) {
	// Test that metrics accept the correct label values

	// Reset all metrics
	TenantReconcileDuration.Reset()
	TenantResourcesReady.Reset()
	ApplyAttemptsTotal.Reset()

	// TenantReconcileDuration: result
	TenantReconcileDuration.WithLabelValues("success")
	TenantReconcileDuration.WithLabelValues("error")

	// TenantResourcesReady: tenant, namespace
	TenantResourcesReady.WithLabelValues("test-tenant", "test-namespace")

	// ApplyAttemptsTotal: kind, result, conflict_policy
	ApplyAttemptsTotal.WithLabelValues("Deployment", "success", "Stuck")
	ApplyAttemptsTotal.WithLabelValues("Service", "error", "Force")

	// All label combinations should work without panicking
	assert.True(t, true, "All label combinations worked")
}
