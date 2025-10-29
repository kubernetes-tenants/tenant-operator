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
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
	"github.com/kubernetes-tenants/tenant-operator/internal/database"
	"github.com/kubernetes-tenants/tenant-operator/internal/metrics"
	"github.com/kubernetes-tenants/tenant-operator/internal/template"
)

const (
	// Finalizer for TenantRegistry
	FinalizerTenantRegistry = "kubernetes-tenants.org/registry-finalizer"
)

// TenantRegistryReconciler reconciles a TenantRegistry object
type TenantRegistryReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenantregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenantregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenantregistries/finalizers,verbs=update
// +kubebuilder:rbac:groups=operator.kubernetes-tenants.org,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

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

	// Handle finalizer logic
	if !registry.DeletionTimestamp.IsZero() {
		// Registry is being deleted
		if containsString(registry.Finalizers, FinalizerTenantRegistry) {
			// Run cleanup logic for DeletionPolicy.Retain resources
			if err := r.cleanupRetainResources(ctx, registry); err != nil {
				logger.Error(err, "Failed to cleanup retain resources")
				r.Recorder.Eventf(registry, corev1.EventTypeWarning, "CleanupFailed",
					"Failed to cleanup retain resources: %v", err)
				return ctrl.Result{RequeueAfter: 10 * time.Second}, err
			}

			// Remove finalizer
			registry.Finalizers = removeString(registry.Finalizers, FinalizerTenantRegistry)
			if err := r.Update(ctx, registry); err != nil {
				logger.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
			logger.Info("Finalizer removed, registry cleanup complete")
		}
		return ctrl.Result{}, nil
	}

	// Ensure finalizer is present
	if !containsString(registry.Finalizers, FinalizerTenantRegistry) {
		registry.Finalizers = append(registry.Finalizers, FinalizerTenantRegistry)
		if err := r.Update(ctx, registry); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		logger.Info("Finalizer added to registry")
		return ctrl.Result{Requeue: true}, nil
	}

	// Connect to database and query tenants
	tenantRows, err := r.queryDatabase(ctx, registry)
	if err != nil {
		logger.Error(err, "Failed to query database")
		r.Recorder.Eventf(registry, corev1.EventTypeWarning, "DatabaseQueryFailed",
			"Failed to query database: %v", err)
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
		if existingTenant, exists := existing[uid]; !exists {
			// Create new Tenant
			if err := r.createTenant(ctx, registry, row); err != nil {
				logger.Error(err, "Failed to create Tenant", "uid", uid)
			}
		} else {
			// Update existing Tenant if data changed
			if r.shouldUpdateTenant(existingTenant, row) {
				if err := r.updateTenant(ctx, registry, existingTenant, row); err != nil {
					logger.Error(err, "Failed to update Tenant", "uid", uid)
				}
			}
		}
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

// getTemplateForRegistry retrieves the TenantTemplate associated with this registry
func (r *TenantRegistryReconciler) getTemplateForRegistry(ctx context.Context, registry *tenantsv1.TenantRegistry) (*tenantsv1.TenantTemplate, error) {
	// List all templates in the same namespace
	templateList := &tenantsv1.TenantTemplateList{}
	if err := r.List(ctx, templateList, client.InNamespace(registry.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	// Find template with matching registryId
	for i := range templateList.Items {
		tmpl := &templateList.Items[i]
		if tmpl.Spec.RegistryID == registry.Name {
			return tmpl, nil
		}
	}

	return nil, fmt.Errorf("no template found for registry: %s", registry.Name)
}

// renderAllTemplateResources renders all resources from a template with the given variables
func (r *TenantRegistryReconciler) renderAllTemplateResources(
	tmpl *tenantsv1.TenantTemplate,
	vars template.Variables,
) (*tenantsv1.TenantSpec, error) {
	engine := template.NewEngine()

	spec := &tenantsv1.TenantSpec{
		ServiceAccounts:        make([]tenantsv1.TResource, 0),
		Deployments:            make([]tenantsv1.TResource, 0),
		StatefulSets:           make([]tenantsv1.TResource, 0),
		Services:               make([]tenantsv1.TResource, 0),
		Ingresses:              make([]tenantsv1.TResource, 0),
		ConfigMaps:             make([]tenantsv1.TResource, 0),
		Secrets:                make([]tenantsv1.TResource, 0),
		PersistentVolumeClaims: make([]tenantsv1.TResource, 0),
		Jobs:                   make([]tenantsv1.TResource, 0),
		CronJobs:               make([]tenantsv1.TResource, 0),
		Manifests:              make([]tenantsv1.TResource, 0),
	}

	// Render each resource type
	var err error

	spec.ServiceAccounts, err = r.renderResourceList(engine, tmpl.Spec.ServiceAccounts, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render serviceAccounts: %w", err)
	}

	spec.Deployments, err = r.renderResourceList(engine, tmpl.Spec.Deployments, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render deployments: %w", err)
	}

	spec.StatefulSets, err = r.renderResourceList(engine, tmpl.Spec.StatefulSets, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render statefulSets: %w", err)
	}

	spec.Services, err = r.renderResourceList(engine, tmpl.Spec.Services, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render services: %w", err)
	}

	spec.Ingresses, err = r.renderResourceList(engine, tmpl.Spec.Ingresses, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render ingresses: %w", err)
	}

	spec.ConfigMaps, err = r.renderResourceList(engine, tmpl.Spec.ConfigMaps, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render configMaps: %w", err)
	}

	spec.Secrets, err = r.renderResourceList(engine, tmpl.Spec.Secrets, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render secrets: %w", err)
	}

	spec.PersistentVolumeClaims, err = r.renderResourceList(engine, tmpl.Spec.PersistentVolumeClaims, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render persistentVolumeClaims: %w", err)
	}

	spec.Jobs, err = r.renderResourceList(engine, tmpl.Spec.Jobs, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render jobs: %w", err)
	}

	spec.CronJobs, err = r.renderResourceList(engine, tmpl.Spec.CronJobs, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render cronJobs: %w", err)
	}

	spec.Manifests, err = r.renderResourceList(engine, tmpl.Spec.Manifests, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render manifests: %w", err)
	}

	return spec, nil
}

// renderResourceList renders a list of template resources
func (r *TenantRegistryReconciler) renderResourceList(
	engine *template.Engine,
	resources []tenantsv1.TResource,
	vars template.Variables,
) ([]tenantsv1.TResource, error) {
	if len(resources) == 0 {
		return []tenantsv1.TResource{}, nil
	}

	rendered := make([]tenantsv1.TResource, len(resources))
	for i, resource := range resources {
		var err error
		rendered[i], err = r.renderResource(engine, resource, vars)
		if err != nil {
			return nil, fmt.Errorf("failed to render resource %s: %w", resource.ID, err)
		}
	}

	return rendered, nil
}

// renderResource renders a single template resource
func (r *TenantRegistryReconciler) renderResource(
	engine *template.Engine,
	resource tenantsv1.TResource,
	vars template.Variables,
) (tenantsv1.TResource, error) {
	rendered := resource

	// Render name template
	if resource.NameTemplate != "" {
		name, err := engine.Render(resource.NameTemplate, vars)
		if err != nil {
			return resource, fmt.Errorf("failed to render name template: %w", err)
		}
		rendered.NameTemplate = name
	}

	// Render labels template
	if len(resource.LabelsTemplate) > 0 {
		labels, err := engine.RenderMap(resource.LabelsTemplate, vars)
		if err != nil {
			return resource, fmt.Errorf("failed to render labels template: %w", err)
		}
		rendered.LabelsTemplate = labels
	}

	// Render annotations template
	if len(resource.AnnotationsTemplate) > 0 {
		annotations, err := engine.RenderMap(resource.AnnotationsTemplate, vars)
		if err != nil {
			return resource, fmt.Errorf("failed to render annotations template: %w", err)
		}
		rendered.AnnotationsTemplate = annotations
	}

	// Note: resource.Spec (unstructured.Unstructured) will be rendered by Tenant controller
	// when it actually applies the resources, as it needs deep recursive rendering

	return rendered, nil
}

// createTenant creates a new Tenant CR
func (r *TenantRegistryReconciler) createTenant(ctx context.Context, registry *tenantsv1.TenantRegistry, row database.TenantRow) error {
	logger := log.FromContext(ctx)

	// 1. Get TenantTemplate
	tmpl, err := r.getTemplateForRegistry(ctx, registry)
	if err != nil {
		logger.Error(err, "Failed to get template for registry", "registry", registry.Name)
		r.Recorder.Eventf(registry, corev1.EventTypeWarning, "TemplateNotFound",
			"No template found for registry %s: %v", registry.Name, err)
		return fmt.Errorf("failed to get template: %w", err)
	}

	// 2. Build template variables
	vars := template.BuildVariables(row.UID, row.HostOrURL, row.Activate, row.Extra)

	// 3. Render all template resources
	renderedSpec, err := r.renderAllTemplateResources(tmpl, vars)
	if err != nil {
		logger.Error(err, "Failed to render template resources", "tenant", row.UID)
		r.Recorder.Eventf(registry, corev1.EventTypeWarning, "TemplateRenderFailed",
			"Failed to render template for tenant %s: %v", row.UID, err)
		return fmt.Errorf("failed to render template: %w", err)
	}

	// 4. Marshal extra values to JSON for annotation
	extraJSON, err := json.Marshal(row.Extra)
	if err != nil {
		logger.Error(err, "Failed to marshal extra values", "tenant", row.UID)
		extraJSON = []byte("{}")
	}

	// 5. Create Tenant CR with rendered resources
	tenant := &tenantsv1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("tenant-%s", row.UID),
			Namespace: registry.Namespace,
			Labels: map[string]string{
				"kubernetes-tenants.org/registry": registry.Name,
				"kubernetes-tenants.org/uid":      row.UID,
			},
			Annotations: map[string]string{
				"kubernetes-tenants.org/hostOrUrl": row.HostOrURL,
				"kubernetes-tenants.org/activate":  row.Activate,
				"kubernetes-tenants.org/extra":     string(extraJSON),
			},
		},
		Spec: *renderedSpec,
	}

	// Set UID and TemplateRef
	tenant.Spec.UID = row.UID
	tenant.Spec.TemplateRef = tmpl.Name

	// Set owner reference
	if err := ctrl.SetControllerReference(registry, tenant, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	logger.Info("Creating Tenant", "tenant", tenant.Name, "uid", row.UID, "template", tmpl.Name)
	return r.Create(ctx, tenant)
}

// shouldUpdateTenant checks if a tenant needs to be updated based on data changes
func (r *TenantRegistryReconciler) shouldUpdateTenant(tenant *tenantsv1.Tenant, row database.TenantRow) bool {
	// Check if stored data differs from current row data
	storedHostOrURL := tenant.Annotations["kubernetes-tenants.org/hostOrUrl"]
	storedActivate := tenant.Annotations["kubernetes-tenants.org/activate"]
	storedExtraJSON := tenant.Annotations["kubernetes-tenants.org/extra"]

	if storedHostOrURL != row.HostOrURL || storedActivate != row.Activate {
		return true
	}

	// Compare extra values
	currentExtraJSON, err := json.Marshal(row.Extra)
	if err != nil {
		// If can't marshal, assume changed
		return true
	}

	if storedExtraJSON != string(currentExtraJSON) {
		return true
	}

	return false
}

// updateTenant updates an existing Tenant CR with new data from database
func (r *TenantRegistryReconciler) updateTenant(ctx context.Context, registry *tenantsv1.TenantRegistry, tenant *tenantsv1.Tenant, row database.TenantRow) error {
	logger := log.FromContext(ctx)

	// 1. Get TenantTemplate
	tmpl, err := r.getTemplateForRegistry(ctx, registry)
	if err != nil {
		logger.Error(err, "Failed to get template for registry", "registry", registry.Name)
		r.Recorder.Eventf(registry, corev1.EventTypeWarning, "TemplateNotFound",
			"No template found for registry %s: %v", registry.Name, err)
		return fmt.Errorf("failed to get template: %w", err)
	}

	// 2. Build template variables with new data
	vars := template.BuildVariables(row.UID, row.HostOrURL, row.Activate, row.Extra)

	// 3. Render all template resources
	renderedSpec, err := r.renderAllTemplateResources(tmpl, vars)
	if err != nil {
		logger.Error(err, "Failed to render template resources", "tenant", row.UID)
		r.Recorder.Eventf(registry, corev1.EventTypeWarning, "TemplateRenderFailed",
			"Failed to render template for tenant %s: %v", row.UID, err)
		return fmt.Errorf("failed to render template: %w", err)
	}

	// 4. Marshal extra values to JSON for annotation
	extraJSON, err := json.Marshal(row.Extra)
	if err != nil {
		logger.Error(err, "Failed to marshal extra values", "tenant", row.UID)
		extraJSON = []byte("{}")
	}

	// 5. Update Tenant annotations and spec
	tenant.Annotations["kubernetes-tenants.org/hostOrUrl"] = row.HostOrURL
	tenant.Annotations["kubernetes-tenants.org/activate"] = row.Activate
	tenant.Annotations["kubernetes-tenants.org/extra"] = string(extraJSON)

	// Update spec with newly rendered resources
	tenant.Spec = *renderedSpec
	tenant.Spec.UID = row.UID
	tenant.Spec.TemplateRef = tmpl.Name

	logger.Info("Updating Tenant", "tenant", tenant.Name, "uid", row.UID, "template", tmpl.Name)
	r.Recorder.Eventf(tenant, corev1.EventTypeNormal, "TenantUpdated",
		"Tenant updated with new data from registry")

	return r.Update(ctx, tenant)
}

// getExistingTenants lists Tenant CRs managed by this registry
func (r *TenantRegistryReconciler) getExistingTenants(ctx context.Context, registry *tenantsv1.TenantRegistry) (*tenantsv1.TenantList, error) {
	tenantList := &tenantsv1.TenantList{}
	if err := r.List(ctx, tenantList, client.InNamespace(registry.Namespace), client.MatchingLabels{
		"kubernetes-tenants.org/registry": registry.Name,
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
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "DatabaseConnected",
		Message:            "Successfully connected to database and queried tenant data",
		LastTransitionTime: metav1.Now(),
	}
	if !synced {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "DatabaseConnectionFailed"
		condition.Message = "Failed to connect to database or query tenant data"
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

	// Record metrics
	metrics.RegistryDesired.WithLabelValues(registry.Name, registry.Namespace).Set(float64(desired))
	metrics.RegistryReady.WithLabelValues(registry.Name, registry.Namespace).Set(float64(ready))
	metrics.RegistryFailed.WithLabelValues(registry.Name, registry.Namespace).Set(float64(failed))

	if err := r.Status().Update(ctx, registry); err != nil {
		log.FromContext(ctx).Error(err, "Failed to update TenantRegistry status")
	}
}

// cleanupRetainResources handles DeletionPolicy.Retain resources when Registry is deleted
func (r *TenantRegistryReconciler) cleanupRetainResources(ctx context.Context, registry *tenantsv1.TenantRegistry) error {
	logger := log.FromContext(ctx)

	// Get all tenants managed by this registry
	tenants, err := r.getExistingTenants(ctx, registry)
	if err != nil {
		return fmt.Errorf("failed to list tenants: %w", err)
	}

	logger.Info("Cleaning up retain resources", "registry", registry.Name, "tenantCount", len(tenants.Items))

	// For each tenant, process Retain resources
	for _, tenant := range tenants.Items {
		if err := r.processRetainResourcesForTenant(ctx, &tenant); err != nil {
			logger.Error(err, "Failed to process retain resources for tenant", "tenant", tenant.Name)
			// Continue with other tenants even if one fails
			r.Recorder.Eventf(registry, corev1.EventTypeWarning, "RetainResourceProcessFailed",
				"Failed to process retain resources for tenant %s: %v", tenant.Name, err)
		}
	}

	logger.Info("Cleanup complete for retain resources", "registry", registry.Name)
	return nil
}

// processRetainResourcesForTenant removes ownerReferences from Retain resources
func (r *TenantRegistryReconciler) processRetainResourcesForTenant(ctx context.Context, tenant *tenantsv1.Tenant) error {
	logger := log.FromContext(ctx)

	// Collect all resources with DeletionPolicy.Retain
	allResources := []tenantsv1.TResource{}
	allResources = append(allResources, tenant.Spec.ServiceAccounts...)
	allResources = append(allResources, tenant.Spec.Deployments...)
	allResources = append(allResources, tenant.Spec.StatefulSets...)
	allResources = append(allResources, tenant.Spec.Services...)
	allResources = append(allResources, tenant.Spec.Ingresses...)
	allResources = append(allResources, tenant.Spec.ConfigMaps...)
	allResources = append(allResources, tenant.Spec.Secrets...)
	allResources = append(allResources, tenant.Spec.PersistentVolumeClaims...)
	allResources = append(allResources, tenant.Spec.Jobs...)
	allResources = append(allResources, tenant.Spec.CronJobs...)
	allResources = append(allResources, tenant.Spec.Manifests...)

	// Process each resource with Retain policy
	for _, resource := range allResources {
		if resource.DeletionPolicy != tenantsv1.DeletionPolicyRetain {
			continue
		}

		logger.Info("Processing retain resource", "tenant", tenant.Name, "resourceId", resource.ID, "name", resource.NameTemplate)

		// Get the resource from cluster
		obj := resource.Spec.DeepCopy()
		obj.SetName(resource.NameTemplate)
		// All resources are in the same namespace as the Tenant CR
		obj.SetNamespace(tenant.Namespace)

		key := client.ObjectKey{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		}

		if err := r.Get(ctx, key, obj); err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Retain resource not found, skipping", "name", key.Name, "namespace", key.Namespace)
				continue
			}
			logger.Error(err, "Failed to get retain resource", "name", key.Name, "namespace", key.Namespace)
			continue
		}

		// Remove ownerReferences that point to this tenant
		ownerRefs := obj.GetOwnerReferences()
		newOwnerRefs := []metav1.OwnerReference{}
		for _, ref := range ownerRefs {
			if ref.UID != tenant.UID {
				newOwnerRefs = append(newOwnerRefs, ref)
			}
		}

		if len(newOwnerRefs) != len(ownerRefs) {
			obj.SetOwnerReferences(newOwnerRefs)
			if err := r.Update(ctx, obj); err != nil {
				logger.Error(err, "Failed to remove ownerReference from retain resource", "name", key.Name, "namespace", key.Namespace)
				continue
			}
			logger.Info("Removed ownerReference from retain resource", "name", key.Name, "namespace", key.Namespace)
			r.Recorder.Eventf(tenant, corev1.EventTypeNormal, "RetainResourcePreserved",
				"Removed ownerReference from resource %s (policy=Retain)", resource.ID)
		}
	}

	return nil
}

// containsString checks if a string is in a slice
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// removeString removes a string from a slice
func removeString(slice []string, str string) []string {
	result := []string{}
	for _, item := range slice {
		if item != str {
			result = append(result, item)
		}
	}
	return result
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantsv1.TenantRegistry{}).
		Owns(&tenantsv1.Tenant{}).
		Named("tenantregistry").
		Complete(r)
}
