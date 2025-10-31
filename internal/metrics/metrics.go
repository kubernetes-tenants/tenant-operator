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
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// TenantReconcileDuration measures the duration of tenant reconciliation
	TenantReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tenant_reconcile_duration_seconds",
			Help:    "Duration of tenant reconciliation in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"result"}, // success or error
	)

	// TenantResourcesReady tracks the number of ready resources per tenant
	TenantResourcesReady = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tenant_resources_ready",
			Help: "Number of ready resources for a tenant",
		},
		[]string{"tenant", "namespace"},
	)

	// TenantResourcesDesired tracks the total number of desired resources per tenant
	TenantResourcesDesired = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tenant_resources_desired",
			Help: "Total number of desired resources for a tenant",
		},
		[]string{"tenant", "namespace"},
	)

	// TenantResourcesFailed tracks the number of failed resources per tenant
	TenantResourcesFailed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tenant_resources_failed",
			Help: "Number of failed resources for a tenant",
		},
		[]string{"tenant", "namespace"},
	)

	// RegistryDesired tracks the desired tenant count per registry
	RegistryDesired = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "registry_desired",
			Help: "Number of desired tenants from the registry data source",
		},
		[]string{"registry", "namespace"},
	)

	// RegistryReady tracks the ready tenant count per registry
	RegistryReady = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "registry_ready",
			Help: "Number of ready tenants for a registry",
		},
		[]string{"registry", "namespace"},
	)

	// RegistryFailed tracks the failed tenant count per registry
	RegistryFailed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "registry_failed",
			Help: "Number of failed tenants for a registry",
		},
		[]string{"registry", "namespace"},
	)

	// ApplyAttemptsTotal counts resource apply attempts
	ApplyAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "apply_attempts_total",
			Help: "Total number of resource apply attempts",
		},
		[]string{"kind", "result", "conflict_policy"},
	)

	// TenantConditionStatus tracks the status of tenant conditions
	// status: 0=False, 1=True, 2=Unknown
	TenantConditionStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tenant_condition_status",
			Help: "Status of tenant conditions (0=False, 1=True, 2=Unknown)",
		},
		[]string{"tenant", "namespace", "type"},
	)

	// TenantConflictsTotal counts the total number of conflicts encountered
	TenantConflictsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tenant_conflicts_total",
			Help: "Total number of resource conflicts encountered during reconciliation",
		},
		[]string{"tenant", "namespace", "resource_kind", "conflict_policy"},
	)

	// TenantResourcesConflicted tracks the current number of resources in conflict state
	TenantResourcesConflicted = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tenant_resources_conflicted",
			Help: "Number of resources currently in conflict state for a tenant",
		},
		[]string{"tenant", "namespace"},
	)

	// TenantDegradedStatus indicates if a tenant is in degraded state (1=degraded, 0=not degraded)
	TenantDegradedStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tenant_degraded_status",
			Help: "Indicates if a tenant is in degraded state (1=degraded, 0=not degraded)",
		},
		[]string{"tenant", "namespace", "reason"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(
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
	)
}
