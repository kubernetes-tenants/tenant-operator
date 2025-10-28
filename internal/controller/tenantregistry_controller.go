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
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
	"github.com/kubernetes-tenants/tenant-operator/internal/database"
)

// TenantRegistryReconciler reconciles a TenantRegistry object
type TenantRegistryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenantregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenantregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenantregistries/finalizers,verbs=update
// +kubebuilder:rbac:groups=tenants.tenants.ecube.dev,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile syncs tenants from external data source to Kubernetes
func (r *TenantRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch TenantRegistry
	registry := &tenantsv1.TenantRegistry{}
	if err := r.Get(ctx, req.NamespacedName, registry); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get TenantRegistry")
		return ctrl.Result{}, err
	}

	// Parse syncInterval
	syncInterval, err := time.ParseDuration(registry.Spec.Source.SyncInterval)
	if err != nil {
		logger.Error(err, "Invalid syncInterval", "syncInterval", registry.Spec.Source.SyncInterval)
		syncInterval = 30 * time.Second // Default
	}

	// Connect to database and query tenants
	tenantRows, err := r.queryDatabase(ctx, registry)
	if err != nil {
		logger.Error(err, "Failed to query database")
		r.updateStatus(ctx, registry, 0, 0, 0, false)
		return ctrl.Result{RequeueAfter: syncInterval}, err
	}

	// Get existing Tenant CRs
	existingTenants, err := r.getExistingTenants(ctx, registry)
	if err != nil {
		logger.Error(err, "Failed to list existing tenants")
		return ctrl.Result{RequeueAfter: syncInterval}, err
	}

	// Reconcile tenants
	desired := make(map[string]database.TenantRow)
	for _, row := range tenantRows {
		desired[row.UID] = row
	}

	existing := make(map[string]*tenantsv1.Tenant)
	for i := range existingTenants.Items {
		tenant := &existingTenants.Items[i]
		existing[tenant.Spec.UID] = tenant
	}

	// Create/update tenants
	for uid, row := range desired {
		if _, exists := existing[uid]; !exists {
			// Create new Tenant
			if err := r.createTenant(ctx, registry, row); err != nil {
				logger.Error(err, "Failed to create Tenant", "uid", uid)
			}
		}
		// TODO: Update logic (check if template changed, etc.)
	}

	// Delete tenants no longer in database
	for uid, tenant := range existing {
		if _, stillExists := desired[uid]; !stillExists {
			if err := r.Delete(ctx, tenant); err != nil {
				logger.Error(err, "Failed to delete Tenant", "uid", uid)
			}
		}
	}

	// Update status
	readyCount, failedCount := r.countTenantStatus(ctx, registry)
	r.updateStatus(ctx, registry, int32(len(desired)), readyCount, failedCount, true)

	return ctrl.Result{RequeueAfter: syncInterval}, nil
}

// queryDatabase connects to database and retrieves tenant rows
func (r *TenantRegistryReconciler) queryDatabase(ctx context.Context, registry *tenantsv1.TenantRegistry) ([]database.TenantRow, error) {
	if registry.Spec.Source.Type != tenantsv1.SourceTypeMySQL {
		return nil, fmt.Errorf("unsupported source type: %s", registry.Spec.Source.Type)
	}

	mysql := registry.Spec.Source.MySQL
	if mysql == nil {
		return nil, fmt.Errorf("MySQL configuration is nil")
	}

	// Get password from Secret
	password := ""
	if mysql.PasswordRef != nil {
		secret := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      mysql.PasswordRef.Name,
			Namespace: registry.Namespace,
		}, secret); err != nil {
			return nil, fmt.Errorf("failed to get password secret: %w", err)
		}
		password = string(secret.Data[mysql.PasswordRef.Key])
	}

	// Connect to MySQL
	dbClient, err := database.NewMySQLClient(database.MySQLConfig{
		Host:     mysql.Host,
		Port:     mysql.Port,
		Username: mysql.Username,
		Password: password,
		Database: mysql.Database,
	})
	if err != nil {
		return nil, err
	}
	defer dbClient.Close()

	// Query tenants
	valueMappings := database.ValueMappings{
		UID:       registry.Spec.ValueMappings.UID,
		HostOrURL: registry.Spec.ValueMappings.HostOrURL,
		Activate:  registry.Spec.ValueMappings.Activate,
	}

	return dbClient.QueryTenants(ctx, mysql.Table, valueMappings, registry.Spec.ExtraValueMappings)
}

// createTenant creates a new Tenant CR
func (r *TenantRegistryReconciler) createTenant(ctx context.Context, registry *tenantsv1.TenantRegistry, row database.TenantRow) error {
	tenant := &tenantsv1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("tenant-%s", row.UID),
			Namespace: registry.Namespace,
			Labels: map[string]string{
				"tenants.ecube.dev/registry": registry.Name,
				"tenants.ecube.dev/uid":      row.UID,
			},
		},
		Spec: tenantsv1.TenantSpec{
			UID:         row.UID,
			TemplateRef: "", // Will be resolved by TenantTemplate controller
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(registry, tenant, r.Scheme); err != nil {
		return err
	}

	return r.Create(ctx, tenant)
}

// getExistingTenants lists Tenant CRs managed by this registry
func (r *TenantRegistryReconciler) getExistingTenants(ctx context.Context, registry *tenantsv1.TenantRegistry) (*tenantsv1.TenantList, error) {
	tenantList := &tenantsv1.TenantList{}
	if err := r.List(ctx, tenantList, client.InNamespace(registry.Namespace), client.MatchingLabels{
		"tenants.ecube.dev/registry": registry.Name,
	}); err != nil {
		return nil, err
	}
	return tenantList, nil
}

// countTenantStatus counts ready and failed tenants
func (r *TenantRegistryReconciler) countTenantStatus(ctx context.Context, registry *tenantsv1.TenantRegistry) (int32, int32) {
	tenants, err := r.getExistingTenants(ctx, registry)
	if err != nil {
		return 0, 0
	}

	var ready, failed int32
	for _, tenant := range tenants.Items {
		for _, cond := range tenant.Status.Conditions {
			if cond.Type == "Ready" {
				if cond.Status == metav1.ConditionTrue {
					ready++
				} else {
					failed++
				}
				break
			}
		}
	}

	return ready, failed
}

// updateStatus updates TenantRegistry status
func (r *TenantRegistryReconciler) updateStatus(ctx context.Context, registry *tenantsv1.TenantRegistry, desired, ready, failed int32, synced bool) {
	registry.Status.Desired = desired
	registry.Status.Ready = ready
	registry.Status.Failed = failed
	registry.Status.ObservedGeneration = registry.Generation

	condition := metav1.Condition{
		Type:               "Synced",
		Status:             metav1.ConditionTrue,
		Reason:             "SuccessfulSync",
		Message:            "Successfully synced tenants from data source",
		LastTransitionTime: metav1.Now(),
	}
	if !synced {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SyncFailed"
		condition.Message = "Failed to sync tenants from data source"
	}

	// Update or append condition
	found := false
	for i := range registry.Status.Conditions {
		if registry.Status.Conditions[i].Type == condition.Type {
			registry.Status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		registry.Status.Conditions = append(registry.Status.Conditions, condition)
	}

	_ = r.Status().Update(ctx, registry)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantsv1.TenantRegistry{}).
		Owns(&tenantsv1.Tenant{}).
		Named("tenantregistry").
		Complete(r)
}
