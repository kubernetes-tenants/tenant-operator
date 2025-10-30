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
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestIsResourceReady(t *testing.T) {
	tests := []struct {
		name  string
		obj   *unstructured.Unstructured
		want  bool
		setup func(*unstructured.Unstructured)
	}{
		{
			name: "ConfigMap - always ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
				},
			},
			want: true,
		},
		{
			name: "Secret - always ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Secret",
				},
			},
			want: true,
		},
		{
			name: "Service - always ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
				},
			},
			want: true,
		},
		{
			name: "Deployment - ready when replicas match",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"generation": int64(1),
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"observedGeneration": int64(1),
						"availableReplicas":  int64(3),
					},
				},
			},
			want: true,
		},
		{
			name: "Deployment - not ready when replicas don't match",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"generation": int64(1),
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"observedGeneration": int64(1),
						"availableReplicas":  int64(1),
					},
				},
			},
			want: false,
		},
		{
			name: "Deployment - not ready when generation mismatch",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"generation": int64(2),
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"observedGeneration": int64(1),
						"availableReplicas":  int64(3),
					},
				},
			},
			want: false,
		},
		{
			name: "StatefulSet - ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "StatefulSet",
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"readyReplicas": int64(3),
					},
				},
			},
			want: true,
		},
		{
			name: "StatefulSet - not ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "StatefulSet",
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"readyReplicas": int64(1),
					},
				},
			},
			want: false,
		},
		{
			name: "Job - completed",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "batch/v1",
					"kind":       "Job",
					"status": map[string]interface{}{
						"succeeded": int64(1),
					},
				},
			},
			want: true,
		},
		{
			name: "Job - not completed",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "batch/v1",
					"kind":       "Job",
					"status": map[string]interface{}{
						"succeeded": int64(0),
					},
				},
			},
			want: false,
		},
		{
			name: "resource with Ready condition - true",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "CustomResource",
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "resource with Ready condition - false",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "CustomResource",
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
							},
						},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(tt.obj)
			}
			if got := IsResourceReady(tt.obj); got != tt.want {
				t.Errorf("IsResourceReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetResourceMetadata(t *testing.T) {
	tests := []struct {
		name          string
		obj           *unstructured.Unstructured
		wantName      string
		wantNamespace string
		wantKind      string
		wantErr       bool
	}{
		{
			name: "complete metadata",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name":      "test-cm",
						"namespace": "default",
					},
				},
			},
			wantName:      "test-cm",
			wantNamespace: "default",
			wantKind:      "ConfigMap",
			wantErr:       false,
		},
		{
			name: "no namespace (cluster-scoped)",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]interface{}{
						"name": "test-ns",
					},
				},
			},
			wantName:      "test-ns",
			wantNamespace: "",
			wantKind:      "Namespace",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotNamespace, gotKind, err := GetResourceMetadata(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResourceMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("GetResourceMetadata() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotNamespace != tt.wantNamespace {
				t.Errorf("GetResourceMetadata() gotNamespace = %v, want %v", gotNamespace, tt.wantNamespace)
			}
			if gotKind != tt.wantKind {
				t.Errorf("GetResourceMetadata() gotKind = %v, want %v", gotKind, tt.wantKind)
			}
		})
	}
}

func TestIsDeploymentReady(t *testing.T) {
	tests := []struct {
		name string
		obj  *unstructured.Unstructured
		want bool
	}{
		{
			name: "ready - all replicas available",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"generation": int64(1),
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"observedGeneration": int64(1),
						"availableReplicas":  int64(3),
					},
				},
			},
			want: true,
		},
		{
			name: "not ready - generation mismatch",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"generation": int64(2),
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"observedGeneration": int64(1),
						"availableReplicas":  int64(3),
					},
				},
			},
			want: false,
		},
		{
			name: "not ready - fewer replicas available",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"generation": int64(1),
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"observedGeneration": int64(1),
						"availableReplicas":  int64(2),
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDeploymentReady(tt.obj); got != tt.want {
				t.Errorf("isDeploymentReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStatefulSetReady(t *testing.T) {
	tests := []struct {
		name string
		obj  *unstructured.Unstructured
		want bool
	}{
		{
			name: "ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"readyReplicas": int64(3),
					},
				},
			},
			want: true,
		},
		{
			name: "not ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"readyReplicas": int64(1),
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStatefulSetReady(tt.obj); got != tt.want {
				t.Errorf("isStatefulSetReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsJobReady(t *testing.T) {
	tests := []struct {
		name string
		obj  *unstructured.Unstructured
		want bool
	}{
		{
			name: "succeeded",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"succeeded": int64(1),
					},
				},
			},
			want: true,
		},
		{
			name: "not succeeded",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"succeeded": int64(0),
					},
				},
			},
			want: false,
		},
		{
			name: "no status",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isJobReady(tt.obj); got != tt.want {
				t.Errorf("isJobReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
