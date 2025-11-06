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

package apply

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
)

const (
	// FieldManager is the name used for Server-Side Apply
	FieldManager = "tenant-operator"

	// Labels for cross-namespace resource tracking
	LabelTenantName      = "kubernetes-tenants.org/tenant"
	LabelTenantNamespace = "kubernetes-tenants.org/tenant-namespace"

	// Label for orphaned resources (DeletionPolicy=Retain) - used for selectors
	LabelOrphaned = "kubernetes-tenants.org/orphaned"

	// Annotations for orphaned resources - detailed information
	AnnotationOrphanedAt     = "kubernetes-tenants.org/orphaned-at"
	AnnotationOrphanedReason = "kubernetes-tenants.org/orphaned-reason"

	// Annotation for storing DeletionPolicy on resources
	AnnotationDeletionPolicy = "kubernetes-tenants.org/deletion-policy"

	// OrphanedLabelValue is the value for orphaned label
	OrphanedLabelValue = "true"
)

// ConflictError represents a resource conflict error
type ConflictError struct {
	ResourceName string
	Namespace    string
	Kind         string
	Err          error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("resource conflict for %s/%s (%s): %v", e.Namespace, e.ResourceName, e.Kind, e.Err)
}

func (e *ConflictError) Unwrap() error {
	return e.Err
}

// Applier handles Server-Side Apply operations
type Applier struct {
	client client.Client
	scheme *runtime.Scheme
}

// NewApplier creates a new Applier
func NewApplier(c client.Client, scheme *runtime.Scheme) *Applier {
	return &Applier{
		client: c,
		scheme: scheme,
	}
}

// ApplyResource applies a resource using the specified patch strategy
// Returns true if the resource was changed, false if no change was needed
func (a *Applier) ApplyResource(
	ctx context.Context,
	obj *unstructured.Unstructured,
	owner *tenantsv1.Tenant,
	conflictPolicy tenantsv1.ConflictPolicy,
	patchStrategy tenantsv1.PatchStrategy,
	deletionPolicy tenantsv1.DeletionPolicy,
) (bool, error) {
	// Set owner reference or tracking labels based on namespace and deletion policy
	if owner != nil {
		isCrossNamespace := obj.GetNamespace() != owner.Namespace
		isNamespaceResource := obj.GetKind() == "Namespace"
		isRetainPolicy := deletionPolicy == tenantsv1.DeletionPolicyRetain

		// Use label-based tracking for:
		// 1. Cross-namespace resources (ownerReferences don't work across namespaces)
		// 2. Namespace resources (cannot have ownerReferences)
		// 3. Retain policy resources (to prevent automatic deletion by garbage collector)
		if isCrossNamespace || isNamespaceResource || isRetainPolicy {
			// Use label-based tracking instead of ownerReference
			labels := obj.GetLabels()
			if labels == nil {
				labels = make(map[string]string)
			}
			labels[LabelTenantName] = owner.Name
			labels[LabelTenantNamespace] = owner.Namespace
			obj.SetLabels(labels)
		} else {
			// For same-namespace resources with Delete policy, use traditional ownerReference
			// This enables automatic garbage collection when Tenant is deleted
			if err := controllerutil.SetControllerReference(owner, obj, a.scheme); err != nil {
				return false, fmt.Errorf("failed to set owner reference: %w", err)
			}
		}
	}

	// Get the existing resource to check for changes
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	existing := obj.DeepCopy()
	existsBeforeApply := true
	beforeResourceVersion := ""

	if err := a.client.Get(ctx, key, existing); err != nil {
		if errors.IsNotFound(err) {
			existsBeforeApply = false
		} else {
			return false, fmt.Errorf("failed to get existing resource: %w", err)
		}
	} else {
		beforeResourceVersion = existing.GetResourceVersion()

		// Remove orphan markers if present (resource is being re-added to management)
		// This must be done on the actual cluster resource, not just the in-memory object
		removed, err := a.removeOrphanMarkersFromCluster(ctx, existing)
		if err != nil {
			// Log but don't fail - orphan markers are metadata, not critical
			// The resource will still be applied correctly
			logger := log.FromContext(ctx)
			logger.V(1).Info("Failed to remove orphan markers, continuing anyway", "error", err)
		}
		_ = removed // Will be used for event logging in controller
	}

	// Apply resource based on patch strategy
	switch patchStrategy {
	case tenantsv1.PatchStrategyApply, "":
		// Server-Side Apply (default)
		force := conflictPolicy == tenantsv1.ConflictPolicyForce

		if err := a.client.Patch(ctx, obj, client.Apply, &client.PatchOptions{
			FieldManager: FieldManager,
			Force:        &force,
		}); err != nil {
			if errors.IsConflict(err) && conflictPolicy == tenantsv1.ConflictPolicyStuck {
				return false, &ConflictError{
					ResourceName: obj.GetName(),
					Namespace:    obj.GetNamespace(),
					Kind:         obj.GetKind(),
					Err:          err,
				}
			}
			return false, fmt.Errorf("failed to apply resource: %w", err)
		}

	case tenantsv1.PatchStrategyMerge:
		// Strategic Merge Patch
		if err := a.client.Patch(ctx, obj, client.Merge); err != nil {
			return false, fmt.Errorf("failed to merge resource: %w", err)
		}

	case tenantsv1.PatchStrategyReplace:
		// Full replacement via Update
		if !existsBeforeApply {
			// Create if not exists
			if err := a.client.Create(ctx, obj); err != nil {
				return false, fmt.Errorf("failed to create resource: %w", err)
			}
			return true, nil
		}

		// Preserve resourceVersion and update
		obj.SetResourceVersion(existing.GetResourceVersion())
		if err := a.client.Update(ctx, obj); err != nil {
			return false, fmt.Errorf("failed to replace resource: %w", err)
		}

	default:
		return false, fmt.Errorf("unsupported patch strategy: %s", patchStrategy)
	}

	// Check if resource was actually changed by comparing resourceVersion
	if !existsBeforeApply {
		// Resource was newly created
		return true, nil
	}

	// Get the resource after apply to check resourceVersion
	after := obj.DeepCopy()
	if err := a.client.Get(ctx, key, after); err != nil {
		// If we can't get the resource, assume it was changed
		return true, nil
	}

	afterResourceVersion := after.GetResourceVersion()
	changed := beforeResourceVersion != afterResourceVersion

	return changed, nil
}

// DeleteResource deletes a resource respecting deletion policy
func (a *Applier) DeleteResource(
	ctx context.Context,
	obj *unstructured.Unstructured,
	policy tenantsv1.DeletionPolicy,
	orphanReason string,
) error {
	if policy == tenantsv1.DeletionPolicyRetain {
		// Remove owner references and tracking labels but keep the resource
		// Add orphan labels to mark it as retained orphan
		return a.removeOwnerReferencesAndLabels(ctx, obj, orphanReason)
	}

	// Delete the resource
	if err := a.client.Delete(ctx, obj); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}

// GetResource retrieves a resource from the cluster
func (a *Applier) GetResource(
	ctx context.Context,
	name, namespace string,
	obj *unstructured.Unstructured,
) error {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	if err := a.client.Get(ctx, key, obj); err != nil {
		return err
	}

	return nil
}

// removeOwnerReferencesAndLabels removes all owner references and tracking labels from the resource
// and adds orphan labels to mark it as a retained orphan resource
func (a *Applier) removeOwnerReferencesAndLabels(ctx context.Context, obj *unstructured.Unstructured, orphanReason string) error {
	// Get current resource
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	current := obj.DeepCopy()
	if err := a.client.Get(ctx, key, current); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// Remove owner references
	current.SetOwnerReferences(nil)

	// Update labels: remove tracking labels and add orphan label
	labels := current.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	// Remove tracking labels
	delete(labels, LabelTenantName)
	delete(labels, LabelTenantNamespace)

	// Add orphan label (for selector queries)
	labels[LabelOrphaned] = OrphanedLabelValue

	current.SetLabels(labels)

	// Update annotations: add orphan metadata
	annotations := current.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// Add orphan annotations with timestamp and reason
	annotations[AnnotationOrphanedAt] = metav1.Now().Format(time.RFC3339)
	if orphanReason != "" {
		annotations[AnnotationOrphanedReason] = orphanReason
	}

	current.SetAnnotations(annotations)

	// Update the resource
	if err := a.client.Update(ctx, current); err != nil {
		return fmt.Errorf("failed to remove owner references and labels: %w", err)
	}

	// Log the orphaning
	logger := log.FromContext(ctx)
	logger.Info("Orphan markers added - resource retained",
		"kind", current.GetKind(),
		"name", current.GetName(),
		"namespace", current.GetNamespace(),
		"reason", orphanReason)

	// Create event on the resource
	message := fmt.Sprintf("Resource retained with orphan markers (reason: %s)", orphanReason)
	a.createEventForResource(ctx, current, corev1.EventTypeNormal, "OrphanMarkersAdded", message)

	return nil
}

// removeOrphanMarkersFromCluster removes orphan label and annotations from a cluster resource
// This is called when a previously orphaned resource is being re-added to management
// Returns true if markers were removed and resource was updated
func (a *Applier) removeOrphanMarkersFromCluster(ctx context.Context, obj *unstructured.Unstructured) (bool, error) {
	// Check if orphan markers are present
	labels := obj.GetLabels()
	annotations := obj.GetAnnotations()

	hasOrphanLabel := labels != nil && labels[LabelOrphaned] == OrphanedLabelValue
	hasOrphanAnnotations := annotations != nil && (annotations[AnnotationOrphanedAt] != "" || annotations[AnnotationOrphanedReason] != "")

	// If no orphan markers, nothing to do
	if !hasOrphanLabel && !hasOrphanAnnotations {
		return false, nil
	}

	// Get the current resource from cluster
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	current := obj.DeepCopy()
	if err := a.client.Get(ctx, key, current); err != nil {
		if errors.IsNotFound(err) {
			// Resource doesn't exist, nothing to clean
			return false, nil
		}
		return false, fmt.Errorf("failed to get resource for orphan marker cleanup: %w", err)
	}

	// Track if we made changes
	changed := false

	// Remove orphan label
	labels = current.GetLabels()
	if labels != nil && labels[LabelOrphaned] == OrphanedLabelValue {
		delete(labels, LabelOrphaned)
		current.SetLabels(labels)
		changed = true
	}

	// Remove orphan annotations
	annotations = current.GetAnnotations()
	if annotations != nil {
		if annotations[AnnotationOrphanedAt] != "" || annotations[AnnotationOrphanedReason] != "" {
			delete(annotations, AnnotationOrphanedAt)
			delete(annotations, AnnotationOrphanedReason)
			current.SetAnnotations(annotations)
			changed = true
		}
	}

	// Update the resource if we made changes
	if changed {
		if err := a.client.Update(ctx, current); err != nil {
			return false, fmt.Errorf("failed to remove orphan markers: %w", err)
		}

		// Log the re-adoption
		logger := log.FromContext(ctx)
		logger.Info("Orphan markers removed - resource re-adopted into management",
			"kind", current.GetKind(),
			"name", current.GetName(),
			"namespace", current.GetNamespace())

		// Create event on the resource
		a.createEventForResource(ctx, current, corev1.EventTypeNormal, "OrphanMarkersRemoved",
			"Resource re-adopted into management - orphan markers removed")
	}

	return changed, nil
}

// createEventForResource creates a Kubernetes Event for a resource
func (a *Applier) createEventForResource(ctx context.Context, obj *unstructured.Unstructured, eventType, reason, message string) {
	logger := log.FromContext(ctx)

	// Create Event object
	now := metav1.Now()
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s.%x", obj.GetName(), now.Unix()),
			Namespace: obj.GetNamespace(),
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion: obj.GetAPIVersion(),
			Kind:       obj.GetKind(),
			Name:       obj.GetName(),
			Namespace:  obj.GetNamespace(),
			UID:        obj.GetUID(),
		},
		Reason:  reason,
		Message: message,
		Source: corev1.EventSource{
			Component: "tenant-operator",
		},
		FirstTimestamp: now,
		LastTimestamp:  now,
		Count:          1,
		Type:           eventType,
	}

	// Try to create the event
	if err := a.client.Create(ctx, event); err != nil {
		// Log but don't fail - events are best-effort
		logger.V(1).Info("Failed to create event for resource",
			"kind", obj.GetKind(),
			"name", obj.GetName(),
			"reason", reason,
			"error", err.Error())
	}
}

// ResourceExists checks if a resource exists
func (a *Applier) ResourceExists(ctx context.Context, name, namespace string, obj *unstructured.Unstructured) (bool, error) {
	err := a.GetResource(ctx, name, namespace, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsResourceReady checks if a resource is ready (basic check using status.conditions)
func IsResourceReady(obj *unstructured.Unstructured) bool {
	// Try to get status.conditions
	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil || !found {
		// No conditions found, check if it's a simple resource type
		return isSimpleResourceReady(obj)
	}

	// Check for Ready condition
	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condMap, "type")
		condStatus, _, _ := unstructured.NestedString(condMap, "status")

		if condType == "Ready" && condStatus == string(metav1.ConditionTrue) {
			return true
		}
	}

	return false
}

// isSimpleResourceReady checks readiness for resources without conditions
func isSimpleResourceReady(obj *unstructured.Unstructured) bool {
	gvk := obj.GroupVersionKind()

	switch gvk.Kind {
	case "Namespace", "ConfigMap", "Secret", "Service", "ServiceAccount":
		// These resources are ready immediately after creation
		return true
	case "Deployment":
		return isDeploymentReady(obj)
	case "StatefulSet":
		return isStatefulSetReady(obj)
	case "Job":
		return isJobReady(obj)
	default:
		// Unknown resource type, assume ready if it exists
		return true
	}
}

// isDeploymentReady checks if a Deployment is ready
func isDeploymentReady(obj *unstructured.Unstructured) bool {
	generation, _, _ := unstructured.NestedInt64(obj.Object, "metadata", "generation")
	observedGeneration, _, _ := unstructured.NestedInt64(obj.Object, "status", "observedGeneration")

	if generation != observedGeneration {
		return false
	}

	replicas, _, _ := unstructured.NestedInt64(obj.Object, "spec", "replicas")
	availableReplicas, _, _ := unstructured.NestedInt64(obj.Object, "status", "availableReplicas")

	return availableReplicas >= replicas
}

// isStatefulSetReady checks if a StatefulSet is ready
func isStatefulSetReady(obj *unstructured.Unstructured) bool {
	replicas, _, _ := unstructured.NestedInt64(obj.Object, "spec", "replicas")
	readyReplicas, _, _ := unstructured.NestedInt64(obj.Object, "status", "readyReplicas")

	return readyReplicas >= replicas
}

// isJobReady checks if a Job is complete
func isJobReady(obj *unstructured.Unstructured) bool {
	succeeded, _, _ := unstructured.NestedInt64(obj.Object, "status", "succeeded")
	return succeeded > 0
}

// GetResourceMetadata extracts metadata from an unstructured object
func GetResourceMetadata(obj *unstructured.Unstructured) (name, namespace, kind string, err error) {
	name = obj.GetName()
	namespace = obj.GetNamespace()

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return "", "", "", err
	}

	gvk := obj.GroupVersionKind()
	kind = gvk.Kind

	_ = accessor // Use accessor if needed

	return name, namespace, kind, nil
}
