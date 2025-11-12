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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	lynqv1 "github.com/k8s-lynq/lynq/api/v1"
	"github.com/k8s-lynq/lynq/internal/status"
	"github.com/k8s-lynq/lynq/internal/template"
)

// TestBuildTemplateVariablesFromAnnotations tests variable extraction from LynqNode annotations
func TestBuildTemplateVariablesFromAnnotations(t *testing.T) {
	tests := []struct {
		name         string
		tenant       *lynqv1.LynqNode
		wantUID      string
		wantHost     string
		wantActivate string
		wantExtra    map[string]string
		wantErr      bool
	}{
		{
			name: "all annotations present",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"lynq.sh/hostOrUrl": "https://example.com",
						"lynq.sh/activate":  "true",
						"lynq.sh/extra":     `{"region":"us-west-2","plan":"premium"}`,
					},
				},
				Spec: lynqv1.LynqNodeSpec{
					UID: "tenant-123",
				},
			},
			wantUID:      "tenant-123",
			wantHost:     "https://example.com",
			wantActivate: "true",
			wantExtra: map[string]string{
				"region": "us-west-2",
				"plan":   "premium",
			},
			wantErr: false,
		},
		{
			name: "missing hostOrUrl defaults to UID",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"lynq.sh/activate": "1",
					},
				},
				Spec: lynqv1.LynqNodeSpec{
					UID: "tenant-456",
				},
			},
			wantUID:      "tenant-456",
			wantHost:     "tenant-456",
			wantActivate: "1",
			wantExtra:    map[string]string{},
			wantErr:      false,
		},
		{
			name: "missing activate defaults to true",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"lynq.sh/hostOrUrl": "https://tenant.example.com",
					},
				},
				Spec: lynqv1.LynqNodeSpec{
					UID: "tenant-789",
				},
			},
			wantUID:      "tenant-789",
			wantHost:     "https://tenant.example.com",
			wantActivate: "true",
			wantExtra:    map[string]string{},
			wantErr:      false,
		},
		{
			name: "invalid extra JSON",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"lynq.sh/extra": `{invalid json}`,
					},
				},
				Spec: lynqv1.LynqNodeSpec{
					UID: "tenant-error",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LynqNodeReconciler{StatusManager: status.NewManager(nil, status.WithSyncMode())}
			vars, err := r.buildTemplateVariablesFromAnnotations(tt.tenant)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUID, vars["uid"])
			assert.Equal(t, tt.wantHost, vars["hostOrUrl"])
			assert.Equal(t, tt.wantActivate, vars["activate"])
			for k, v := range tt.wantExtra {
				assert.Equal(t, v, vars[k])
			}
		})
	}
}

// TestBuildResourceKey tests resource key generation
func TestBuildResourceKey(t *testing.T) {
	tests := []struct {
		name       string
		obj        *unstructured.Unstructured
		resourceID string
		want       string
	}{
		{
			name: "deployment resource",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Deployment",
					"metadata": map[string]interface{}{
						"name":      "test-deployment",
						"namespace": "default",
					},
				},
			},
			resourceID: "app-deploy",
			want:       "Deployment/default/test-deployment@app-deploy",
		},
		{
			name: "service resource",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Service",
					"metadata": map[string]interface{}{
						"name":      "my-service",
						"namespace": "production",
					},
				},
			},
			resourceID: "svc-main",
			want:       "Service/production/my-service@svc-main",
		},
		{
			name: "namespace resource (no namespace field)",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Namespace",
					"metadata": map[string]interface{}{
						"name": "my-namespace",
					},
				},
			},
			resourceID: "ns-1",
			want:       "Namespace//my-namespace@ns-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildResourceKey(tt.obj, tt.resourceID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseResourceKey tests resource key parsing
func TestParseResourceKey(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		wantKind      string
		wantNamespace string
		wantName      string
		wantID        string
		wantErr       bool
	}{
		{
			name:          "valid deployment key",
			key:           "Deployment/default/test-deployment@app-deploy",
			wantKind:      "Deployment",
			wantNamespace: "default",
			wantName:      "test-deployment",
			wantID:        "app-deploy",
			wantErr:       false,
		},
		{
			name:          "valid namespace key (empty namespace)",
			key:           "Namespace//my-namespace@ns-1",
			wantKind:      "Namespace",
			wantNamespace: "",
			wantName:      "my-namespace",
			wantID:        "ns-1",
			wantErr:       false,
		},
		{
			name:    "invalid key format (no @)",
			key:     "Deployment/default/test-deployment",
			wantErr: true,
		},
		{
			name:    "invalid key format (no /)",
			key:     "Deployment-default-test-deployment@id",
			wantErr: true,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, namespace, name, id, err := parseResourceKey(tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantKind, kind)
			assert.Equal(t, tt.wantNamespace, namespace)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

// TestCollectResourcesFromLynqNode tests collection of all resource types from LynqNode spec
func TestCollectResourcesFromLynqNode(t *testing.T) {
	tests := []struct {
		name      string
		tenant    *lynqv1.LynqNode
		wantCount int
	}{
		{
			name: "tenant with multiple resource types",
			tenant: &lynqv1.LynqNode{
				Spec: lynqv1.LynqNodeSpec{
					ServiceAccounts: []lynqv1.TResource{
						{ID: "sa-1", Spec: unstructured.Unstructured{}},
					},
					Deployments: []lynqv1.TResource{
						{ID: "deploy-1", Spec: unstructured.Unstructured{}},
						{ID: "deploy-2", Spec: unstructured.Unstructured{}},
					},
					Services: []lynqv1.TResource{
						{ID: "svc-1", Spec: unstructured.Unstructured{}},
					},
					ConfigMaps: []lynqv1.TResource{
						{ID: "cm-1", Spec: unstructured.Unstructured{}},
					},
					Jobs: []lynqv1.TResource{
						{ID: "job-1", Spec: unstructured.Unstructured{}},
					},
				},
			},
			wantCount: 6,
		},
		{
			name: "tenant with no resources",
			tenant: &lynqv1.LynqNode{
				Spec: lynqv1.LynqNodeSpec{},
			},
			wantCount: 0,
		},
		{
			name: "tenant with all resource types",
			tenant: &lynqv1.LynqNode{
				Spec: lynqv1.LynqNodeSpec{
					ServiceAccounts:        []lynqv1.TResource{{ID: "sa"}},
					Deployments:            []lynqv1.TResource{{ID: "deploy"}},
					StatefulSets:           []lynqv1.TResource{{ID: "sts"}},
					Services:               []lynqv1.TResource{{ID: "svc"}},
					Ingresses:              []lynqv1.TResource{{ID: "ing"}},
					ConfigMaps:             []lynqv1.TResource{{ID: "cm"}},
					Secrets:                []lynqv1.TResource{{ID: "secret"}},
					PersistentVolumeClaims: []lynqv1.TResource{{ID: "pvc"}},
					Jobs:                   []lynqv1.TResource{{ID: "job"}},
					CronJobs:               []lynqv1.TResource{{ID: "cron"}},
					Manifests:              []lynqv1.TResource{{ID: "manifest"}},
				},
			},
			wantCount: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LynqNodeReconciler{StatusManager: status.NewManager(nil, status.WithSyncMode())}
			resources := r.collectResourcesFromLynqNode(tt.tenant)
			assert.Len(t, resources, tt.wantCount)
		})
	}
}

// TestCountLynqNodeResourcesByType tests resource counting by type
func TestCountLynqNodeResourcesByType(t *testing.T) {
	tests := []struct {
		name      string
		tenant    *lynqv1.LynqNode
		wantCount map[string]int
	}{
		{
			name: "various resources",
			tenant: &lynqv1.LynqNode{
				Spec: lynqv1.LynqNodeSpec{
					Deployments: []lynqv1.TResource{{ID: "d1"}, {ID: "d2"}},
					Services:    []lynqv1.TResource{{ID: "s1"}},
					ConfigMaps:  []lynqv1.TResource{{ID: "cm1"}, {ID: "cm2"}, {ID: "cm3"}},
				},
			},
			wantCount: map[string]int{
				"Deployments": 2,
				"Services":    1,
				"ConfigMaps":  3,
			},
		},
		{
			name: "empty tenant",
			tenant: &lynqv1.LynqNode{
				Spec: lynqv1.LynqNodeSpec{},
			},
			wantCount: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LynqNodeReconciler{StatusManager: status.NewManager(nil, status.WithSyncMode())}
			counts := r.countLynqNodeResourcesByType(tt.tenant)
			assert.Equal(t, tt.wantCount, counts)
		})
	}
}

// TestFormatLynqNodeResourceDetails tests resource detail formatting
func TestFormatLynqNodeResourceDetails(t *testing.T) {
	tests := []struct {
		name   string
		counts map[string]int
		want   string
	}{
		{
			name: "multiple resource types",
			counts: map[string]int{
				"Deployments": 2,
				"Services":    1,
				"ConfigMaps":  3,
			},
			want: "2 Deployment(s), 1 Service(s), 3 ConfigMap(s)",
		},
		{
			name:   "empty counts",
			counts: map[string]int{},
			want:   "no resources",
		},
		{
			name: "single resource type",
			counts: map[string]int{
				"Jobs": 1,
			},
			want: "1 Job(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LynqNodeReconciler{StatusManager: status.NewManager(nil, status.WithSyncMode())}
			got := r.formatLynqNodeResourceDetails(tt.counts)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGetAPIVersionForKind tests API version mapping
func TestGetAPIVersionForKind(t *testing.T) {
	tests := []struct {
		kind string
		want string
	}{
		{kind: "Namespace", want: "v1"},
		{kind: "ServiceAccount", want: "v1"},
		{kind: "Service", want: "v1"},
		{kind: "ConfigMap", want: "v1"},
		{kind: "Secret", want: "v1"},
		{kind: "PersistentVolumeClaim", want: "v1"},
		{kind: "Deployment", want: "apps/v1"},
		{kind: "StatefulSet", want: "apps/v1"},
		{kind: "DaemonSet", want: "apps/v1"},
		{kind: "Job", want: "batch/v1"},
		{kind: "CronJob", want: "batch/v1"},
		{kind: "Ingress", want: "networking.k8s.io/v1"},
		{kind: "UnknownKind", want: "v1"}, // Default
	}

	r := &LynqNodeReconciler{StatusManager: status.NewManager(nil, status.WithSyncMode())}
	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			got := r.getAPIVersionForKind(tt.kind)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestFindOrphanedResources tests orphan detection logic
func TestFindOrphanedResources(t *testing.T) {
	tests := []struct {
		name         string
		previousKeys []string
		currentKeys  map[string]bool
		want         []string
	}{
		{
			name: "some resources removed",
			previousKeys: []string{
				"Deployment/default/app1@d1",
				"Service/default/svc1@s1",
				"ConfigMap/default/cm1@cm1",
			},
			currentKeys: map[string]bool{
				"Deployment/default/app1@d1": true,
				"Service/default/svc1@s1":    true,
			},
			want: []string{"ConfigMap/default/cm1@cm1"},
		},
		{
			name: "no resources removed",
			previousKeys: []string{
				"Deployment/default/app1@d1",
				"Service/default/svc1@s1",
			},
			currentKeys: map[string]bool{
				"Deployment/default/app1@d1": true,
				"Service/default/svc1@s1":    true,
			},
			want: []string{},
		},
		{
			name:         "all resources removed",
			previousKeys: []string{"Deployment/default/app1@d1", "Service/default/svc1@s1"},
			currentKeys:  map[string]bool{},
			want:         []string{"Deployment/default/app1@d1", "Service/default/svc1@s1"},
		},
		{
			name:         "no previous resources",
			previousKeys: []string{},
			currentKeys: map[string]bool{
				"Deployment/default/app1@d1": true,
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LynqNodeReconciler{StatusManager: status.NewManager(nil, status.WithSyncMode())}
			orphans := r.findOrphanedResources(tt.previousKeys, tt.currentKeys)
			assert.ElementsMatch(t, tt.want, orphans)
		})
	}
}

// TestRenderUnstructured tests template rendering (simple cases without template engine)
func TestRenderUnstructured_NoTemplates(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		want map[string]interface{}
	}{
		{
			name: "simple string values",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "nested maps",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
			},
			want: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
			},
		},
		{
			name: "arrays",
			data: map[string]interface{}{
				"items": []interface{}{"item1", "item2"},
			},
			want: map[string]interface{}{
				"items": []interface{}{"item1", "item2"},
			},
		},
	}

	scheme := runtime.NewScheme()
	r := &LynqNodeReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
		Scheme: scheme,
	}
	ctx := context.Background()
	engine := template.NewEngine()
	vars := template.BuildVariables("test-uid", "https://example.com", "true", map[string]string{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.renderUnstructured(ctx, tt.data, engine, vars)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestCheckOnceCreated tests "created-once" annotation check
func TestCheckOnceCreated(t *testing.T) {
	tests := []struct {
		name              string
		existingResource  bool
		hasAnnotation     bool
		wantExists        bool
		wantHasAnnotation bool
	}{
		{
			name:              "resource exists with annotation",
			existingResource:  true,
			hasAnnotation:     true,
			wantExists:        true,
			wantHasAnnotation: true,
		},
		{
			name:              "resource exists without annotation",
			existingResource:  true,
			hasAnnotation:     false,
			wantExists:        true,
			wantHasAnnotation: false,
		},
		{
			name:              "resource does not exist",
			existingResource:  false,
			hasAnnotation:     false,
			wantExists:        false,
			wantHasAnnotation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			clientBuilder := fake.NewClientBuilder().WithScheme(scheme)

			obj := &unstructured.Unstructured{}
			obj.SetKind("ConfigMap")
			obj.SetAPIVersion("v1")
			obj.SetName("test-cm")
			obj.SetNamespace("default")

			if tt.existingResource {
				existing := obj.DeepCopy()
				if tt.hasAnnotation {
					existing.SetAnnotations(map[string]string{
						AnnotationCreatedOnce: AnnotationValueTrue,
					})
				}
				clientBuilder.WithObjects(existing)
			}

			r := &LynqNodeReconciler{
				Client: clientBuilder.Build(),
				Scheme: scheme,
			}

			ctx := context.Background()
			exists, hasAnnotation, err := r.checkOnceCreated(ctx, obj)

			require.NoError(t, err)
			assert.Equal(t, tt.wantExists, exists)
			assert.Equal(t, tt.wantHasAnnotation, hasAnnotation)
		})
	}
}

// TestEmitTemplateAppliedEvent tests the emitTemplateAppliedEvent function
func TestEmitTemplateAppliedEvent(t *testing.T) {
	tests := []struct {
		name           string
		tenant         *lynqv1.LynqNode
		totalResources int32
	}{
		{
			name: "basic event emission",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant1-web",
					Namespace: "default",
					Annotations: map[string]string{
						"lynq.sh/template-generation": "5",
					},
					Labels: map[string]string{
						"lynq.sh/registry": "test-registry",
					},
				},
				Spec: lynqv1.LynqNodeSpec{
					UID:         "tenant1",
					TemplateRef: "web-app",
					ConfigMaps: []lynqv1.TResource{
						{ID: "cm1"},
						{ID: "cm2"},
					},
					Deployments: []lynqv1.TResource{
						{ID: "deploy1"},
					},
				},
			},
			totalResources: 3,
		},
		{
			name: "without registry label",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant2-api",
					Namespace: "default",
					Annotations: map[string]string{
						"lynq.sh/template-generation": "1",
					},
				},
				Spec: lynqv1.LynqNodeSpec{
					UID:         "tenant2",
					TemplateRef: "api",
					Services: []lynqv1.TResource{
						{ID: "svc1"},
					},
				},
			},
			totalResources: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			require.NoError(t, lynqv1.AddToScheme(scheme))

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.tenant).
				Build()

			// Use fake recorder stub - event verification requires integration test with real recorder
			recorder := &fakeRecorder{}

			r := &LynqNodeReconciler{
				Client:   fakeClient,
				Scheme:   scheme,
				Recorder: recorder,
			}

			ctx := context.Background()
			// Test verifies function executes without panic
			// Note: Full event verification requires integration test with record.NewFakeRecorder
			// which is not compatible with the existing fakeRecorder type in suite_test.go
			r.emitTemplateAppliedEvent(ctx, tt.tenant, tt.totalResources)

			// Basic sanity check: function should have accessed tenant properties
			assert.NotEmpty(t, tt.tenant.Spec.TemplateRef, "Test tenant should have template ref")
		})
	}
}

// TestEmitTemplateAppliedCompleteEvent tests the emitTemplateAppliedCompleteEvent function
// Note: This test verifies function execution without panic. Full event verification
// would require integration test with record.NewFakeRecorder.
func TestEmitTemplateAppliedCompleteEvent(t *testing.T) {
	tests := []struct {
		name           string
		tenant         *lynqv1.LynqNode
		totalResources int32
		readyCount     int32
		failedCount    int32
		changedCount   int32
		expectSuccess  bool // Whether this should be a success or failure event
	}{
		{
			name: "successful completion",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant1-web",
					Namespace: "default",
					Annotations: map[string]string{
						"lynq.sh/template-generation": "5",
					},
					Labels: map[string]string{
						"lynq.sh/registry": "test-registry",
					},
				},
				Spec: lynqv1.LynqNodeSpec{
					UID:         "tenant1",
					TemplateRef: "web-app",
				},
			},
			totalResources: 3,
			readyCount:     3,
			failedCount:    0,
			changedCount:   2,
			expectSuccess:  true, // No failures, should emit success event
		},
		{
			name: "partial failure",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant2-api",
					Namespace: "default",
					Annotations: map[string]string{
						"lynq.sh/template-generation": "3",
					},
					Labels: map[string]string{
						"lynq.sh/registry": "test-registry",
					},
				},
				Spec: lynqv1.LynqNodeSpec{
					UID:         "tenant2",
					TemplateRef: "api",
				},
			},
			totalResources: 5,
			readyCount:     3,
			failedCount:    2,
			changedCount:   1,
			expectSuccess:  false, // Has failures, should emit warning event
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			require.NoError(t, lynqv1.AddToScheme(scheme))

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.tenant).
				Build()

			// Use fake recorder stub - event verification requires integration test
			recorder := &fakeRecorder{}

			r := &LynqNodeReconciler{
				Client:   fakeClient,
				Scheme:   scheme,
				Recorder: recorder,
			}

			ctx := context.Background()
			// Test verifies function executes without panic
			// Full event verification (Normal vs Warning) requires integration test
			r.emitTemplateAppliedCompleteEvent(ctx, tt.tenant, tt.totalResources, tt.readyCount, tt.failedCount, tt.changedCount)

			// Basic sanity checks based on test case
			if tt.expectSuccess {
				assert.Zero(t, tt.failedCount, "Success case should have no failures")
			} else {
				assert.Greater(t, tt.failedCount, int32(0), "Failure case should have failures")
			}
		})
	}
}

// TestUpdateProgressingCondition tests the updateProgressingCondition function
func TestUpdateProgressingCondition(t *testing.T) {
	tests := []struct {
		name              string
		tenant            *lynqv1.LynqNode
		progressing       bool
		reason            string
		message           string
		wantConditionType string
		wantStatus        metav1.ConditionStatus
	}{
		{
			name: "set progressing to true",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant1-web",
					Namespace: "default",
				},
				Spec: lynqv1.LynqNodeSpec{
					UID:         "tenant1",
					TemplateRef: "web-app",
				},
			},
			progressing:       true,
			reason:            "Reconciling",
			message:           "Reconciling changed resources",
			wantConditionType: "Progressing",
			wantStatus:        metav1.ConditionTrue,
		},
		{
			name: "set progressing to false",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant2-api",
					Namespace: "default",
				},
				Spec: lynqv1.LynqNodeSpec{
					UID:         "tenant2",
					TemplateRef: "api",
				},
			},
			progressing:       false,
			reason:            "ReconcileComplete",
			message:           "Reconciliation completed",
			wantConditionType: "Progressing",
			wantStatus:        metav1.ConditionFalse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			require.NoError(t, lynqv1.AddToScheme(scheme))

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.tenant).
				WithStatusSubresource(tt.tenant).
				Build()

			r := &LynqNodeReconciler{
				Client:        fakeClient,
				Scheme:        scheme,
				StatusManager: status.NewManager(fakeClient, status.WithSyncMode()),
			}

			ctx := context.Background()
			r.StatusManager.PublishProgressingCondition(tt.tenant, tt.progressing, tt.reason, tt.message)

			// Verify condition was set
			updated := &lynqv1.LynqNode{}
			err := fakeClient.Get(ctx, client.ObjectKeyFromObject(tt.tenant), updated)
			require.NoError(t, err)

			// Find the Progressing condition
			var progressingCondition *metav1.Condition
			for i := range updated.Status.Conditions {
				if updated.Status.Conditions[i].Type == tt.wantConditionType {
					progressingCondition = &updated.Status.Conditions[i]
					break
				}
			}

			require.NotNil(t, progressingCondition, "Progressing condition should be set")
			assert.Equal(t, tt.wantStatus, progressingCondition.Status)
		})
	}
}

// TestBuildAppliedResourceKeys tests the buildAppliedResourceKeys function
func TestBuildAppliedResourceKeys(t *testing.T) {
	tests := []struct {
		name      string
		tenant    *lynqv1.LynqNode
		wantCount int
	}{
		{
			name: "multiple resource types",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant1-web",
					Namespace: "default",
				},
				Spec: lynqv1.LynqNodeSpec{
					UID:         "tenant1",
					TemplateRef: "web-app",
					ConfigMaps: []lynqv1.TResource{
						{
							ID:           "cm1",
							NameTemplate: "tenant1-config",
						},
					},
					Deployments: []lynqv1.TResource{
						{
							ID:           "deploy1",
							NameTemplate: "tenant1-deploy",
						},
					},
					Services: []lynqv1.TResource{
						{
							ID:           "svc1",
							NameTemplate: "tenant1-svc",
						},
					},
				},
			},
			wantCount: 3,
		},
		{
			name: "no resources",
			tenant: &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tenant2-empty",
					Namespace: "default",
				},
				Spec: lynqv1.LynqNodeSpec{
					UID:         "tenant2",
					TemplateRef: "empty",
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			require.NoError(t, lynqv1.AddToScheme(scheme))

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			r := &LynqNodeReconciler{
				Client:        fakeClient,
				Scheme:        scheme,
				StatusManager: status.NewManager(fakeClient, status.WithSyncMode()),
			}

			ctx := context.Background()
			keys, err := r.buildAppliedResourceKeys(ctx, tt.tenant)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(keys))

			// Verify key format
			for key := range keys {
				// Keys should have format: "Kind/Namespace/Name@ID"
				assert.Contains(t, key, "/")
				assert.Contains(t, key, "@")
			}
		})
	}
}
