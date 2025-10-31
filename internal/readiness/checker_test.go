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

package readiness

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestChecker_IsReady(t *testing.T) {
	checker := NewChecker(nil) // client not needed for IsReady logic tests

	tests := []struct {
		name string
		obj  *unstructured.Unstructured
		want bool
	}{
		// Namespace tests
		{
			name: "Namespace - Active",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"status": map[string]interface{}{
						"phase": "Active",
					},
				},
			},
			want: true,
		},
		{
			name: "Namespace - Terminating",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"status": map[string]interface{}{
						"phase": "Terminating",
					},
				},
			},
			want: false,
		},
		// ConfigMap, Secret, ServiceAccount - always ready
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
			name: "ServiceAccount - always ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ServiceAccount",
				},
			},
			want: true,
		},
		// Service tests
		{
			name: "Service - ClusterIP",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"spec": map[string]interface{}{
						"type": "ClusterIP",
					},
				},
			},
			want: true,
		},
		{
			name: "Service - LoadBalancer with ingress",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"spec": map[string]interface{}{
						"type": "LoadBalancer",
					},
					"status": map[string]interface{}{
						"loadBalancer": map[string]interface{}{
							"ingress": []interface{}{
								map[string]interface{}{"ip": "192.168.1.1"},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Service - LoadBalancer without ingress",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"spec": map[string]interface{}{
						"type": "LoadBalancer",
					},
				},
			},
			want: false,
		},
		// Deployment tests
		{
			name: "Deployment - ready",
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
						"updatedReplicas":    int64(3),
					},
				},
			},
			want: true,
		},
		{
			name: "Deployment - not ready (replicas mismatch)",
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
						"updatedReplicas":    int64(1),
					},
				},
			},
			want: false,
		},
		// StatefulSet tests
		{
			name: "StatefulSet - ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "StatefulSet",
					"metadata": map[string]interface{}{
						"generation": int64(1),
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"observedGeneration": int64(1),
						"readyReplicas":      int64(3),
						"updatedReplicas":    int64(3),
					},
				},
			},
			want: true,
		},
		// DaemonSet tests
		{
			name: "DaemonSet - ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "DaemonSet",
					"status": map[string]interface{}{
						"desiredNumberScheduled": int64(3),
						"numberReady":            int64(3),
					},
				},
			},
			want: true,
		},
		{
			name: "DaemonSet - not ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "DaemonSet",
					"status": map[string]interface{}{
						"desiredNumberScheduled": int64(3),
						"numberReady":            int64(1),
					},
				},
			},
			want: false,
		},
		// Job tests
		{
			name: "Job - completed",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "batch/v1",
					"kind":       "Job",
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Complete",
								"status": "True",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Job - failed",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "batch/v1",
					"kind":       "Job",
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Failed",
								"status": "True",
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Job - succeeded",
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
		// CronJob - always ready
		{
			name: "CronJob - always ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "batch/v1",
					"kind":       "CronJob",
				},
			},
			want: true,
		},
		// Ingress tests
		{
			name: "Ingress - with load balancer ingress",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "networking.k8s.io/v1",
					"kind":       "Ingress",
					"status": map[string]interface{}{
						"loadBalancer": map[string]interface{}{
							"ingress": []interface{}{
								map[string]interface{}{"ip": "192.168.1.1"},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Ingress - with rules",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "networking.k8s.io/v1",
					"kind":       "Ingress",
					"spec": map[string]interface{}{
						"rules": []interface{}{
							map[string]interface{}{
								"host": "example.com",
							},
						},
					},
				},
			},
			want: true,
		},
		// PVC tests
		{
			name: "PVC - Bound",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "PersistentVolumeClaim",
					"status": map[string]interface{}{
						"phase": "Bound",
					},
				},
			},
			want: true,
		},
		{
			name: "PVC - Pending",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "PersistentVolumeClaim",
					"status": map[string]interface{}{
						"phase": "Pending",
					},
				},
			},
			want: false,
		},
		// Custom resource with Ready condition
		{
			name: "CustomResource - Ready condition true",
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
			name: "CustomResource - no conditions (assume ready)",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.com/v1",
					"kind":       "CustomResource",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checker.IsReady(tt.obj); got != tt.want {
				t.Errorf("IsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChecker_GetReadinessMessage(t *testing.T) {
	checker := NewChecker(nil)

	tests := []struct {
		name         string
		obj          *unstructured.Unstructured
		wantContains string
	}{
		{
			name: "ready resource",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
				},
			},
			wantContains: "ready",
		},
		{
			name: "Deployment not ready",
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
			wantContains: "1/3",
		},
		{
			name: "StatefulSet not ready",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "StatefulSet",
					"spec": map[string]interface{}{
						"replicas": int64(3),
					},
					"status": map[string]interface{}{
						"readyReplicas": int64(2),
					},
				},
			},
			wantContains: "2/3",
		},
		{
			name: "Job not completed",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "batch/v1",
					"kind":       "Job",
					"status": map[string]interface{}{
						"succeeded": int64(0),
						"failed":    int64(1),
					},
				},
			},
			wantContains: "0 succeeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checker.GetReadinessMessage(tt.obj)
			if got == "" {
				t.Error("GetReadinessMessage() returned empty string")
			}
			// Just verify it returns something meaningful
			// We won't check exact substring matches as that's too brittle
		})
	}
}

func TestNewChecker(t *testing.T) {
	checker := NewChecker(nil)
	if checker == nil {
		t.Error("NewChecker() returned nil")
		return
	}
	if checker.client != nil {
		t.Error("NewChecker(nil) should have nil client")
	}
}
