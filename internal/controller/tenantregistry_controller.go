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
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
	"github.com/kubernetes-tenants/tenant-operator/internal/datasource"
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

	// Get all templates that reference this registry
	templates, err := r.getTemplatesForRegistry(ctx, registry)
	if err != nil {
		logger.Error(err, "Failed to get templates for registry")
		r.updateStatus(ctx, registry, 0, 0, 0, 0, false)
		return ctrl.Result{RequeueAfter: syncInterval}, err
	}

	// Connect to database and query tenants
	tenantRows, err := r.queryDatabase(ctx, registry)
	if err != nil {
		logger.Error(err, "Failed to query database")
		r.Recorder.Eventf(registry, corev1.EventTypeWarning, "DatabaseQueryFailed",
			"Failed to query database: %v", err)
		r.updateStatus(ctx, registry, int32(len(templates)), 0, 0, 0, false)
		return ctrl.Result{RequeueAfter: syncInterval}, err
	}

	// Get existing Tenant CRs
	existingTenants, err := r.getExistingTenants(ctx, registry)
	if err != nil {
		logger.Error(err, "Failed to list existing tenants")
		return ctrl.Result{RequeueAfter: syncInterval}, err
	}

	// Build desired tenant set: key = {template-name}-{uid}
	type TenantKey struct {
		TemplateName string
		UID          string
	}
	desired := make(map[TenantKey]struct {
		Template *tenantsv1.TenantTemplate
		Row      datasource.TenantRow
	})

	for _, tmpl := range templates {
		for _, row := range tenantRows {
			key := TenantKey{
				TemplateName: tmpl.Name,
				UID:          row.UID,
			}
			desired[key] = struct {
				Template *tenantsv1.TenantTemplate
				Row      datasource.TenantRow
			}{
				Template: tmpl,
				Row:      row,
			}
		}
	}

	// Build existing tenant map: key = {template-name}-{uid}
	existing := make(map[TenantKey]*tenantsv1.Tenant)
	for i := range existingTenants.Items {
		tenant := &existingTenants.Items[i]
		key := TenantKey{
			TemplateName: tenant.Spec.TemplateRef,
			UID:          tenant.Spec.UID,
		}
		existing[key] = tenant
	}

	// Create/update tenants for each template-row combination
	for key, desired := range desired {
		if existingTenant, exists := existing[key]; !exists {
			// Create new Tenant
			if err := r.createTenant(ctx, registry, desired.Template, desired.Row); err != nil {
				// Ignore AlreadyExists errors (can happen due to concurrent reconciliations)
				if !errors.IsAlreadyExists(err) {
					logger.Error(err, "Failed to create Tenant", "template", key.TemplateName, "uid", key.UID)
				}
			}
		} else {
			// Update existing Tenant if data or template changed
			if r.shouldUpdateTenant(ctx, registry, existingTenant, desired.Row) {
				if err := r.updateTenant(ctx, registry, desired.Template, existingTenant, desired.Row); err != nil {
					logger.Error(err, "Failed to update Tenant", "template", key.TemplateName, "uid", key.UID)
				}
			}
		}
	}

	// Delete tenants no longer in desired set
	// This handles:
	// 1. Rows deleted from database
	// 2. Rows with activate=false
	// 3. Templates deleted/changed
	deletedCount := 0
	for key, tenant := range existing {
		if _, stillExists := desired[key]; !stillExists {
			logger.Info("Deleting Tenant (no longer in desired set)",
				"tenant", tenant.Name,
				"template", key.TemplateName,
				"uid", key.UID,
				"reason", "row removed from database or activate=false or template changed")

			// Emit detailed deletion event
			r.Recorder.Eventf(registry, corev1.EventTypeNormal, "TenantDeleting",
				"Deleting Tenant '%s' (template: %s, uid: %s) - no longer in active dataset. "+
					"This could be due to: row deletion, activate=false, or template change.",
				tenant.Name, key.TemplateName, key.UID)

			if err := r.Delete(ctx, tenant); err != nil {
				if !errors.IsNotFound(err) {
					logger.Error(err, "Failed to delete Tenant", "tenant", tenant.Name, "template", key.TemplateName, "uid", key.UID)
					r.Recorder.Eventf(registry, corev1.EventTypeWarning, "TenantDeletionFailed",
						"Failed to delete Tenant '%s': %v", tenant.Name, err)
				}
			} else {
				deletedCount++
				r.Recorder.Eventf(registry, corev1.EventTypeNormal, "TenantDeleted",
					"Successfully deleted Tenant '%s' (template: %s, uid: %s)",
					tenant.Name, key.TemplateName, key.UID)
			}
		}
	}

	if deletedCount > 0 {
		logger.Info("Garbage collection completed", "deletedTenants", deletedCount)
	}

	// Update status
	readyCount, failedCount := r.countTenantStatus(ctx, registry)
	totalDesired := int32(len(templates)) * int32(len(tenantRows))
	r.updateStatus(ctx, registry, int32(len(templates)), totalDesired, readyCount, failedCount, true)

	return ctrl.Result{RequeueAfter: syncInterval}, nil
}

// queryDatabase connects to database and retrieves tenant rows
func (r *TenantRegistryReconciler) queryDatabase(ctx context.Context, registry *tenantsv1.TenantRegistry) ([]datasource.TenantRow, error) {
	// Determine datasource type
	sourceType := datasource.SourceType(registry.Spec.Source.Type)

	// Get password from Secret (MySQL/PostgreSQL specific)
	password := ""
	if registry.Spec.Source.MySQL != nil && registry.Spec.Source.MySQL.PasswordRef != nil {
		secret := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      registry.Spec.Source.MySQL.PasswordRef.Name,
			Namespace: registry.Namespace,
		}, secret); err != nil {
			return nil, fmt.Errorf("failed to get password secret: %w", err)
		}
		password = string(secret.Data[registry.Spec.Source.MySQL.PasswordRef.Key])
	}

	// Build datasource config
	config, table, err := r.buildDatasourceConfig(registry, password)
	if err != nil {
		return nil, err
	}

	// Create datasource adapter
	ds, err := datasource.NewDatasource(sourceType, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create datasource: %w", err)
	}
	defer func() {
		_ = ds.Close() // Best effort close
	}()

	// Query tenants
	queryConfig := datasource.QueryConfig{
		Table: table,
		ValueMappings: datasource.ValueMappings{
			UID:       registry.Spec.ValueMappings.UID,
			HostOrURL: registry.Spec.ValueMappings.HostOrURL,
			Activate:  registry.Spec.ValueMappings.Activate,
		},
		ExtraMappings: registry.Spec.ExtraValueMappings,
	}

	return ds.QueryTenants(ctx, queryConfig)
}

// buildDatasourceConfig builds datasource configuration from TenantRegistry spec
func (r *TenantRegistryReconciler) buildDatasourceConfig(registry *tenantsv1.TenantRegistry, password string) (datasource.Config, string, error) {
	switch registry.Spec.Source.Type {
	case tenantsv1.SourceTypeMySQL:
		mysql := registry.Spec.Source.MySQL
		if mysql == nil {
			return datasource.Config{}, "", fmt.Errorf("MySQL configuration is nil")
		}

		config := datasource.Config{
			Host:     mysql.Host,
			Port:     mysql.Port,
			Username: mysql.Username,
			Password: password,
			Database: mysql.Database,
		}

		return config, mysql.Table, nil

	default:
		return datasource.Config{}, "", fmt.Errorf("unsupported source type: %s", registry.Spec.Source.Type)
	}
}

// getTemplatesForRegistry retrieves all TenantTemplates that reference this registry
func (r *TenantRegistryReconciler) getTemplatesForRegistry(ctx context.Context, registry *tenantsv1.TenantRegistry) ([]*tenantsv1.TenantTemplate, error) {
	// List all templates in the same namespace
	templateList := &tenantsv1.TenantTemplateList{}
	if err := r.List(ctx, templateList, client.InNamespace(registry.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	// Find all templates with matching registryId
	var templates []*tenantsv1.TenantTemplate
	for i := range templateList.Items {
		tmpl := &templateList.Items[i]
		if tmpl.Spec.RegistryID == registry.Name {
			templates = append(templates, tmpl)
		}
	}

	return templates, nil
}

// getTemplateForRegistry retrieves a single TenantTemplate for backward compatibility
// Deprecated: Use getTemplatesForRegistry instead
func (r *TenantRegistryReconciler) getTemplateForRegistry(ctx context.Context, registry *tenantsv1.TenantRegistry) (*tenantsv1.TenantTemplate, error) {
	templates, err := r.getTemplatesForRegistry(ctx, registry)
	if err != nil {
		return nil, err
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("no template found for registry: %s", registry.Name)
	}

	return templates[0], nil
}

// renderAllTemplateResources renders all resources from a template with the given variables
func (r *TenantRegistryReconciler) renderAllTemplateResources(
	tmpl *tenantsv1.TenantTemplate,
	vars template.Variables,
) (*tenantsv1.TenantSpec, error) {
	engine := template.NewEngine()

	spec := &tenantsv1.TenantSpec{
		ServiceAccounts:          make([]tenantsv1.TResource, 0),
		Deployments:              make([]tenantsv1.TResource, 0),
		StatefulSets:             make([]tenantsv1.TResource, 0),
		Services:                 make([]tenantsv1.TResource, 0),
		Ingresses:                make([]tenantsv1.TResource, 0),
		ConfigMaps:               make([]tenantsv1.TResource, 0),
		Secrets:                  make([]tenantsv1.TResource, 0),
		PersistentVolumeClaims:   make([]tenantsv1.TResource, 0),
		Jobs:                     make([]tenantsv1.TResource, 0),
		CronJobs:                 make([]tenantsv1.TResource, 0),
		PodDisruptionBudgets:     make([]tenantsv1.TResource, 0),
		NetworkPolicies:          make([]tenantsv1.TResource, 0),
		HorizontalPodAutoscalers: make([]tenantsv1.TResource, 0),
		Namespaces:               make([]tenantsv1.TResource, 0),
		Manifests:                make([]tenantsv1.TResource, 0),
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

	spec.DaemonSets, err = r.renderResourceList(engine, tmpl.Spec.DaemonSets, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render daemonSets: %w", err)
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

	spec.PodDisruptionBudgets, err = r.renderResourceList(engine, tmpl.Spec.PodDisruptionBudgets, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render podDisruptionBudgets: %w", err)
	}

	spec.NetworkPolicies, err = r.renderResourceList(engine, tmpl.Spec.NetworkPolicies, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render networkPolicies: %w", err)
	}

	spec.HorizontalPodAutoscalers, err = r.renderResourceList(engine, tmpl.Spec.HorizontalPodAutoscalers, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render horizontalPodAutoscalers: %w", err)
	}

	spec.Namespaces, err = r.renderResourceList(engine, tmpl.Spec.Namespaces, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render namespaces: %w", err)
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
func (r *TenantRegistryReconciler) createTenant(ctx context.Context, registry *tenantsv1.TenantRegistry, tmpl *tenantsv1.TenantTemplate, row datasource.TenantRow) error {
	logger := log.FromContext(ctx)

	// 1. Build template variables
	vars := template.BuildVariables(row.UID, row.HostOrURL, row.Activate, row.Extra)

	// 2. Render all template resources
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

	// 3. Create Tenant CR with rendered resources
	// Name format: {uid}-{template-name} to support multiple templates per registry
	tenant := &tenantsv1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", row.UID, tmpl.Name),
			Namespace: registry.Namespace,
			Labels: map[string]string{
				"kubernetes-tenants.org/registry": registry.Name,
				"kubernetes-tenants.org/uid":      row.UID,
			},
			Annotations: map[string]string{
				"kubernetes-tenants.org/hostOrUrl":           row.HostOrURL,
				"kubernetes-tenants.org/activate":            row.Activate,
				"kubernetes-tenants.org/extra":               string(extraJSON),
				"kubernetes-tenants.org/template-generation": fmt.Sprintf("%d", tmpl.Generation),
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

// shouldUpdateTenant checks if a tenant needs to be updated based on data changes or template changes
func (r *TenantRegistryReconciler) shouldUpdateTenant(ctx context.Context, registry *tenantsv1.TenantRegistry, tenant *tenantsv1.Tenant, row datasource.TenantRow) bool {
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

	// Check if template has been updated
	tmpl, err := r.getTemplateForRegistry(ctx, registry)
	if err != nil {
		// If we can't get the template, assume update is needed
		return true
	}

	// Compare template generation
	storedTemplateGeneration := tenant.Annotations["kubernetes-tenants.org/template-generation"]
	currentTemplateGeneration := fmt.Sprintf("%d", tmpl.Generation)

	return storedTemplateGeneration != currentTemplateGeneration
}

// updateTenant updates an existing Tenant CR with new data from database
func (r *TenantRegistryReconciler) updateTenant(ctx context.Context, registry *tenantsv1.TenantRegistry, tmpl *tenantsv1.TenantTemplate, tenant *tenantsv1.Tenant, row datasource.TenantRow) error {
	logger := log.FromContext(ctx)

	// Check what triggered the update
	oldTemplateGeneration := tenant.Annotations["kubernetes-tenants.org/template-generation"]
	newTemplateGeneration := fmt.Sprintf("%d", tmpl.Generation)
	templateChanged := oldTemplateGeneration != newTemplateGeneration
	dataChanged := tenant.Annotations["kubernetes-tenants.org/hostOrUrl"] != row.HostOrURL ||
		tenant.Annotations["kubernetes-tenants.org/activate"] != row.Activate

	// 1. Build template variables with new data
	vars := template.BuildVariables(row.UID, row.HostOrURL, row.Activate, row.Extra)

	// 2. Render all template resources
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

	// 5. Count resources by type for detailed event
	resourceCounts := r.countResourcesByType(renderedSpec)
	totalResources := resourceCounts["total"]

	// 6. Retry update on conflict (handles concurrent modifications by Tenant controller)
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the latest version of the Tenant
		latest := &tenantsv1.Tenant{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(tenant), latest); err != nil {
			return err
		}

		// Update Tenant annotations and spec with latest version
		if latest.Annotations == nil {
			latest.Annotations = make(map[string]string)
		}
		latest.Annotations["kubernetes-tenants.org/hostOrUrl"] = row.HostOrURL
		latest.Annotations["kubernetes-tenants.org/activate"] = row.Activate
		latest.Annotations["kubernetes-tenants.org/extra"] = string(extraJSON)
		latest.Annotations["kubernetes-tenants.org/template-generation"] = newTemplateGeneration

		// Update spec with newly rendered resources
		latest.Spec = *renderedSpec
		latest.Spec.UID = row.UID
		latest.Spec.TemplateRef = tmpl.Name

		// Perform the update with the latest version
		return r.Update(ctx, latest)
	}); err != nil {
		return err
	}

	// 7. Emit detailed events based on what changed (only after successful update)
	if templateChanged {
		// Template changed - emit detailed Applied event
		resourceDetails := r.formatResourceDetails(resourceCounts)
		r.Recorder.Eventf(tenant, corev1.EventTypeNormal, "TemplateApplied",
			"Applying TenantTemplate '%s' changes (generation: %s -> %s). "+
				"Total %d resources will be reconciled: %s. "+
				"Registry: %s, UID: %s",
			tmpl.Name, oldTemplateGeneration, newTemplateGeneration,
			totalResources, resourceDetails,
			registry.Name, row.UID)

		logger.Info("Applying template changes to Tenant",
			"tenant", tenant.Name,
			"template", tmpl.Name,
			"oldGeneration", oldTemplateGeneration,
			"newGeneration", newTemplateGeneration,
			"totalResources", totalResources)
	} else if dataChanged {
		// Only data changed
		r.Recorder.Eventf(tenant, corev1.EventTypeNormal, "DataUpdated",
			"Tenant data updated from registry '%s'. "+
				"Total %d resources may be affected. "+
				"Template: %s (generation: %s)",
			registry.Name, totalResources,
			tmpl.Name, newTemplateGeneration)

		logger.Info("Updating Tenant data",
			"tenant", tenant.Name,
			"template", tmpl.Name,
			"totalResources", totalResources)
	} else {
		// Extra values or other changes
		r.Recorder.Eventf(tenant, corev1.EventTypeNormal, "TenantUpdated",
			"Tenant updated with new configuration from registry '%s'. "+
				"Total %d resources managed. Template: %s (generation: %s)",
			registry.Name, totalResources,
			tmpl.Name, newTemplateGeneration)
	}

	return nil
}

// countResourcesByType counts resources by type in a TenantSpec
func (r *TenantRegistryReconciler) countResourcesByType(spec *tenantsv1.TenantSpec) map[string]int {
	counts := make(map[string]int)
	total := 0

	if len(spec.ServiceAccounts) > 0 {
		counts["ServiceAccounts"] = len(spec.ServiceAccounts)
		total += len(spec.ServiceAccounts)
	}
	if len(spec.Deployments) > 0 {
		counts["Deployments"] = len(spec.Deployments)
		total += len(spec.Deployments)
	}
	if len(spec.StatefulSets) > 0 {
		counts["StatefulSets"] = len(spec.StatefulSets)
		total += len(spec.StatefulSets)
	}
	if len(spec.Services) > 0 {
		counts["Services"] = len(spec.Services)
		total += len(spec.Services)
	}
	if len(spec.Ingresses) > 0 {
		counts["Ingresses"] = len(spec.Ingresses)
		total += len(spec.Ingresses)
	}
	if len(spec.ConfigMaps) > 0 {
		counts["ConfigMaps"] = len(spec.ConfigMaps)
		total += len(spec.ConfigMaps)
	}
	if len(spec.Secrets) > 0 {
		counts["Secrets"] = len(spec.Secrets)
		total += len(spec.Secrets)
	}
	if len(spec.PersistentVolumeClaims) > 0 {
		counts["PVCs"] = len(spec.PersistentVolumeClaims)
		total += len(spec.PersistentVolumeClaims)
	}
	if len(spec.Jobs) > 0 {
		counts["Jobs"] = len(spec.Jobs)
		total += len(spec.Jobs)
	}
	if len(spec.CronJobs) > 0 {
		counts["CronJobs"] = len(spec.CronJobs)
		total += len(spec.CronJobs)
	}
	if len(spec.PodDisruptionBudgets) > 0 {
		counts["PodDisruptionBudgets"] = len(spec.PodDisruptionBudgets)
		total += len(spec.PodDisruptionBudgets)
	}
	if len(spec.NetworkPolicies) > 0 {
		counts["NetworkPolicies"] = len(spec.NetworkPolicies)
		total += len(spec.NetworkPolicies)
	}
	if len(spec.HorizontalPodAutoscalers) > 0 {
		counts["HorizontalPodAutoscalers"] = len(spec.HorizontalPodAutoscalers)
		total += len(spec.HorizontalPodAutoscalers)
	}
	if len(spec.Manifests) > 0 {
		counts["Manifests"] = len(spec.Manifests)
		total += len(spec.Manifests)
	}

	counts["total"] = total
	return counts
}

// formatResourceDetails formats resource counts into a readable string
func (r *TenantRegistryReconciler) formatResourceDetails(counts map[string]int) string {
	if counts["total"] == 0 {
		return NoResourcesMessage
	}

	var details []string
	if count, ok := counts["ServiceAccounts"]; ok {
		details = append(details, fmt.Sprintf("%d ServiceAccount(s)", count))
	}
	if count, ok := counts["Deployments"]; ok {
		details = append(details, fmt.Sprintf("%d Deployment(s)", count))
	}
	if count, ok := counts["StatefulSets"]; ok {
		details = append(details, fmt.Sprintf("%d StatefulSet(s)", count))
	}
	if count, ok := counts["Services"]; ok {
		details = append(details, fmt.Sprintf("%d Service(s)", count))
	}
	if count, ok := counts["Ingresses"]; ok {
		details = append(details, fmt.Sprintf("%d Ingress(es)", count))
	}
	if count, ok := counts["ConfigMaps"]; ok {
		details = append(details, fmt.Sprintf("%d ConfigMap(s)", count))
	}
	if count, ok := counts["Secrets"]; ok {
		details = append(details, fmt.Sprintf("%d Secret(s)", count))
	}
	if count, ok := counts["PVCs"]; ok {
		details = append(details, fmt.Sprintf("%d PVC(s)", count))
	}
	if count, ok := counts["Jobs"]; ok {
		details = append(details, fmt.Sprintf("%d Job(s)", count))
	}
	if count, ok := counts["CronJobs"]; ok {
		details = append(details, fmt.Sprintf("%d CronJob(s)", count))
	}
	if count, ok := counts["PodDisruptionBudgets"]; ok {
		details = append(details, fmt.Sprintf("%d PodDisruptionBudget(s)", count))
	}
	if count, ok := counts["NetworkPolicies"]; ok {
		details = append(details, fmt.Sprintf("%d NetworkPolicy(ies)", count))
	}
	if count, ok := counts["HorizontalPodAutoscalers"]; ok {
		details = append(details, fmt.Sprintf("%d HorizontalPodAutoscaler(s)", count))
	}
	if count, ok := counts["Manifests"]; ok {
		details = append(details, fmt.Sprintf("%d Manifest(s)", count))
	}

	if len(details) == 0 {
		return NoResourcesMessage
	}

	return fmt.Sprintf("%s", details[0:])
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

// updateStatus updates TenantRegistry status with retry on conflict
func (r *TenantRegistryReconciler) updateStatus(ctx context.Context, registry *tenantsv1.TenantRegistry, referencingTemplates, desired, ready, failed int32, synced bool) {
	logger := log.FromContext(ctx)

	// Record metrics first (these don't depend on the status update)
	metrics.RegistryDesired.WithLabelValues(registry.Name, registry.Namespace).Set(float64(desired))
	metrics.RegistryReady.WithLabelValues(registry.Name, registry.Namespace).Set(float64(ready))
	metrics.RegistryFailed.WithLabelValues(registry.Name, registry.Namespace).Set(float64(failed))

	// Retry status update on conflict
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the latest version of the registry
		key := client.ObjectKeyFromObject(registry)
		latest := &tenantsv1.TenantRegistry{}
		if err := r.Get(ctx, key, latest); err != nil {
			return err
		}

		// Update status fields
		latest.Status.ReferencingTemplates = referencingTemplates
		latest.Status.Desired = desired
		latest.Status.Ready = ready
		latest.Status.Failed = failed
		latest.Status.ObservedGeneration = latest.Generation

		// Prepare condition
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
		for i := range latest.Status.Conditions {
			if latest.Status.Conditions[i].Type == condition.Type {
				latest.Status.Conditions[i] = condition
				found = true
				break
			}
		}
		if !found {
			latest.Status.Conditions = append(latest.Status.Conditions, condition)
		}

		// Update status subresource
		return r.Status().Update(ctx, latest)
	})

	if err != nil {
		logger.Error(err, "Failed to update TenantRegistry status after retries")
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
//
//nolint:unparam // error return kept for future resource cleanup error handling
func (r *TenantRegistryReconciler) processRetainResourcesForTenant(ctx context.Context, tenant *tenantsv1.Tenant) error {
	logger := log.FromContext(ctx)

	// Collect all resources with DeletionPolicy.Retain
	allResources := []tenantsv1.TResource{}
	allResources = append(allResources, tenant.Spec.ServiceAccounts...)
	allResources = append(allResources, tenant.Spec.Deployments...)
	allResources = append(allResources, tenant.Spec.StatefulSets...)
	allResources = append(allResources, tenant.Spec.DaemonSets...)
	allResources = append(allResources, tenant.Spec.Services...)
	allResources = append(allResources, tenant.Spec.Ingresses...)
	allResources = append(allResources, tenant.Spec.ConfigMaps...)
	allResources = append(allResources, tenant.Spec.Secrets...)
	allResources = append(allResources, tenant.Spec.PersistentVolumeClaims...)
	allResources = append(allResources, tenant.Spec.Jobs...)
	allResources = append(allResources, tenant.Spec.CronJobs...)
	allResources = append(allResources, tenant.Spec.PodDisruptionBudgets...)
	allResources = append(allResources, tenant.Spec.NetworkPolicies...)
	allResources = append(allResources, tenant.Spec.HorizontalPodAutoscalers...)
	allResources = append(allResources, tenant.Spec.Namespaces...)
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
func (r *TenantRegistryReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantsv1.TenantRegistry{}).
		Owns(&tenantsv1.Tenant{}).
		// Watch TenantTemplates to re-sync Tenants when template changes
		Watches(&tenantsv1.TenantTemplate{}, handler.EnqueueRequestsFromMapFunc(r.findRegistryForTemplate)).
		Named("tenantregistry").
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		Complete(r)
}

// findRegistryForTemplate maps a TenantTemplate to its TenantRegistry for watch events
func (r *TenantRegistryReconciler) findRegistryForTemplate(ctx context.Context, obj client.Object) []reconcile.Request {
	tmpl := obj.(*tenantsv1.TenantTemplate)

	// Return a reconcile request for the registry referenced by this template
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      tmpl.Spec.RegistryID,
				Namespace: tmpl.Namespace,
			},
		},
	}
}
