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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	tenantsv1 "github.com/kubernetes-tenants/tenant-operator/api/v1"
)

const (
	// FieldManager is the name used for Server-Side Apply
	FieldManager = "tenant-operator"

	// Labels for cross-namespace resource tracking
	LabelTenantName      = "kubernetes-tenants.org/tenant"
	LabelTenantNamespace = "kubernetes-tenants.org/tenant-namespace"

	// Labels for orphaned resources (DeletionPolicy=Retain)
	LabelOrphaned       = "kubernetes-tenants.org/orphaned"
	LabelOrphanedAt     = "kubernetes-tenants.org/orphaned-at"
	LabelOrphanedReason = "kubernetes-tenants.org/orphaned-reason"
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
) (bool, error) {
	// Set owner reference or tracking labels based on namespace
	if owner != nil {
		isCrossNamespace := obj.GetNamespace() != owner.Namespace
		isNamespaceResource := obj.GetKind() == "Namespace"

		if isCrossNamespace || isNamespaceResource {
			// For cross-namespace resources or Namespace resources, use label-based tracking
			// ownerReferences cannot be used across namespaces in Kubernetes
			labels := obj.GetLabels()
			if labels == nil {
				labels = make(map[string]string)
			}
			labels[LabelTenantName] = owner.Name
			labels[LabelTenantNamespace] = owner.Namespace
			obj.SetLabels(labels)
		} else {
			// For same-namespace resources, use traditional ownerReference
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

	// Update labels: remove tracking labels and add orphan labels
	labels := current.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	// Remove tracking labels
	delete(labels, LabelTenantName)
	delete(labels, LabelTenantNamespace)

	// Add orphan labels
	labels[LabelOrphaned] = "true"
	labels[LabelOrphanedAt] = metav1.Now().Format("2006-01-02T15:04:05Z")
	if orphanReason != "" {
		labels[LabelOrphanedReason] = orphanReason
	}

	current.SetLabels(labels)

	// Update the resource
	if err := a.client.Update(ctx, current); err != nil {
		return fmt.Errorf("failed to remove owner references and labels: %w", err)
	}

	return nil
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
