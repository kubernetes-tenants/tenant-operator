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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	lynqv1 "github.com/k8s-lynq/lynq/api/v1"
)

var _ = Describe("LynqNode Controller - IgnoreFields [Pending: Requires running controller]", func() {
	// NOTE: These integration tests require a running LynqNode controller
	// They are currently skipped because the test environment doesn't have controllers running
	// The core ignoreFields functionality is validated by unit tests in internal/fieldfilter/filter_test.go
	const (
		timeout          = time.Second * 30
		interval         = time.Millisecond * 250
		defaultNamespace = "default"
	)

	Context("When reconciling a LynqNode with ignoreFields for Deployment replicas", func() {
		PIt("should create resource with all fields initially, then preserve manually changed replicas while updating other fields", func() {
			ctx := context.Background()

			// ========================================
			// Scenario: HPA controls replicas, operator controls image
			// ========================================

			// Given: A LynqNode with a Deployment that ignores replicas field
			tenantName := fmt.Sprintf("tenant-ignore-replicas-%d", time.Now().UnixNano())
			namespace := defaultNamespace

			deploymentSpec := map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": tenantName + "-deployment",
				},
				"spec": map[string]interface{}{
					"replicas": int64(3), // Initial value
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": "test",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "test",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.20", // Initial image
								},
							},
						},
					},
				},
			}

			tenant := &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tenantName,
					Namespace: namespace,
				},
				Spec: lynqv1.LynqNodeSpec{
					Deployments: []lynqv1.TResource{
						{
							ID: "test-deployment",
							Spec: unstructured.Unstructured{
								Object: deploymentSpec,
							},
							CreationPolicy: lynqv1.CreationPolicyWhenNeeded, // Default - continues syncing
							IgnoreFields:   []string{"$.spec.replicas"},     // Ignore replicas (for HPA)
						},
					},
				},
			}

			// When: Creating the LynqNode
			Expect(k8sClient.Create(ctx, tenant)).To(Succeed())
			defer cleanupResource(ctx, tenant)

			// Then: Deployment should be created with initial replicas=3 and image=nginx:1.20
			deploymentKey := types.NamespacedName{
				Name:      tenantName + "-deployment",
				Namespace: namespace,
			}

			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, deploymentKey, deployment)
			}, timeout, interval).Should(Succeed())

			// Verify initial state
			Expect(*deployment.Spec.Replicas).To(Equal(int32(3)))
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal("nginx:1.20"))

			// ========================================
			// When: HPA (simulated) changes replicas to 5
			// ========================================
			By("Simulating HPA scaling replicas from 3 to 5")

			// Manually update the Deployment replicas (simulating HPA behavior)
			Eventually(func() error {
				if err := k8sClient.Get(ctx, deploymentKey, deployment); err != nil {
					return err
				}
				replicas := int32(5)
				deployment.Spec.Replicas = &replicas
				return k8sClient.Update(ctx, deployment)
			}, timeout, interval).Should(Succeed())

			// Verify replicas changed to 5
			Eventually(func() int32 {
				if err := k8sClient.Get(ctx, deploymentKey, deployment); err != nil {
					return 0
				}
				if deployment.Spec.Replicas == nil {
					return 0
				}
				return *deployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(5)))

			// ========================================
			// When: Template is updated (image change from nginx:1.20 to nginx:1.21)
			// ========================================
			By("Updating LynqNode template to change image to nginx:1.21")

			Eventually(func() error {
				// Get latest LynqNode
				latestLynqNode := &lynqv1.LynqNode{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: tenantName, Namespace: namespace}, latestLynqNode); err != nil {
					return err
				}

				// Update image in spec
				deploymentSpec := latestLynqNode.Spec.Deployments[0].Spec.Object
				templateSpec := deploymentSpec["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})
				containers := templateSpec["containers"].([]interface{})
				container := containers[0].(map[string]interface{})
				container["image"] = "nginx:1.21" // Update image

				return k8sClient.Update(ctx, latestLynqNode)
			}, timeout, interval).Should(Succeed())

			// Give controller time to reconcile
			time.Sleep(2 * time.Second)

			// ========================================
			// Then: Image should be updated to nginx:1.21, but replicas should remain 5
			// ========================================
			By("Verifying image is updated but replicas is preserved")

			Eventually(func() string {
				if err := k8sClient.Get(ctx, deploymentKey, deployment); err != nil {
					return ""
				}
				if len(deployment.Spec.Template.Spec.Containers) == 0 {
					return ""
				}
				return deployment.Spec.Template.Spec.Containers[0].Image
			}, timeout, interval).Should(Equal("nginx:1.21"))

			// Verify replicas is STILL 5 (not reverted to 3)
			Consistently(func() int32 {
				if err := k8sClient.Get(ctx, deploymentKey, deployment); err != nil {
					return 0
				}
				if deployment.Spec.Replicas == nil {
					return 0
				}
				return *deployment.Spec.Replicas
			}, time.Second*5, interval).Should(Equal(int32(5)))
		})
	})

	Context("When reconciling a LynqNode with multiple ignoreFields", func() {
		PIt("should preserve multiple manually changed fields while updating others", func() {
			ctx := context.Background()

			// ========================================
			// Scenario: Ignore both replicas and resources, but update image
			// ========================================

			// Given: A LynqNode with multiple ignored fields
			tenantName := fmt.Sprintf("tenant-multi-ignore-%d", time.Now().UnixNano())
			namespace := defaultNamespace

			deploymentSpec := map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": tenantName + "-deployment",
				},
				"spec": map[string]interface{}{
					"replicas": int64(2),
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": "multi-ignore",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "multi-ignore",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "app",
									"image": "nginx:1.20",
									"resources": map[string]interface{}{
										"limits": map[string]interface{}{
											"cpu":    "500m",
											"memory": "512Mi",
										},
									},
								},
							},
						},
					},
				},
			}

			tenant := &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tenantName,
					Namespace: namespace,
				},
				Spec: lynqv1.LynqNodeSpec{
					Deployments: []lynqv1.TResource{
						{
							ID: "multi-ignore-deployment",
							Spec: unstructured.Unstructured{
								Object: deploymentSpec,
							},
							CreationPolicy: lynqv1.CreationPolicyWhenNeeded,
							IgnoreFields: []string{
								"$.spec.replicas",
								"$.spec.template.spec.containers[0].resources",
							},
						},
					},
				},
			}

			// When: Creating the LynqNode
			Expect(k8sClient.Create(ctx, tenant)).To(Succeed())
			defer cleanupResource(ctx, tenant)

			// Then: Deployment should be created with initial values
			deploymentKey := types.NamespacedName{
				Name:      tenantName + "-deployment",
				Namespace: namespace,
			}

			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, deploymentKey, deployment)
			}, timeout, interval).Should(Succeed())

			// Verify initial state
			Expect(*deployment.Spec.Replicas).To(Equal(int32(2)))
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String()).To(Equal("500m"))

			// ========================================
			// When: Manually change both replicas and resources
			// ========================================
			By("Manually changing replicas to 7 and CPU to 1000m")

			Eventually(func() error {
				if err := k8sClient.Get(ctx, deploymentKey, deployment); err != nil {
					return err
				}
				replicas := int32(7)
				deployment.Spec.Replicas = &replicas
				deployment.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] = resource.MustParse("1000m")
				return k8sClient.Update(ctx, deployment)
			}, timeout, interval).Should(Succeed())

			// Verify changes applied
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, deploymentKey, deployment); err != nil {
					return false
				}
				return *deployment.Spec.Replicas == 7 &&
					deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String() == "1"
			}, timeout, interval).Should(BeTrue())

			// ========================================
			// When: Update template with image change
			// ========================================
			By("Updating LynqNode template to change image")

			Eventually(func() error {
				latestLynqNode := &lynqv1.LynqNode{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: tenantName, Namespace: namespace}, latestLynqNode); err != nil {
					return err
				}

				deploymentSpec := latestLynqNode.Spec.Deployments[0].Spec.Object
				templateSpec := deploymentSpec["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})
				containers := templateSpec["containers"].([]interface{})
				container := containers[0].(map[string]interface{})
				container["image"] = "nginx:1.22"

				return k8sClient.Update(ctx, latestLynqNode)
			}, timeout, interval).Should(Succeed())

			time.Sleep(2 * time.Second)

			// ========================================
			// Then: Image updated, but replicas and resources preserved
			// ========================================
			By("Verifying image updated but ignored fields preserved")

			Eventually(func() string {
				if err := k8sClient.Get(ctx, deploymentKey, deployment); err != nil {
					return ""
				}
				return deployment.Spec.Template.Spec.Containers[0].Image
			}, timeout, interval).Should(Equal("nginx:1.22"))

			// Both ignored fields should be preserved
			Consistently(func() bool {
				if err := k8sClient.Get(ctx, deploymentKey, deployment); err != nil {
					return false
				}
				return *deployment.Spec.Replicas == 7 &&
					deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String() == "1"
			}, time.Second*5, interval).Should(BeTrue())
		})
	})

	Context("When ignoreFields is used with CreationPolicy Once", func() {
		PIt("should behave like Once policy (ignoreFields has no effect)", func() {
			ctx := context.Background()

			// ========================================
			// Scenario: CreationPolicy=Once with ignoreFields should still be Once
			// (ignoreFields has no effect, warning should be logged)
			// ========================================

			// Given: A LynqNode with CreationPolicy=Once and ignoreFields
			tenantName := fmt.Sprintf("tenant-once-ignore-%d", time.Now().UnixNano())
			namespace := defaultNamespace

			deploymentSpec := map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": tenantName + "-deployment",
				},
				"spec": map[string]interface{}{
					"replicas": int64(4),
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": "once-test",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "once-test",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "app",
									"image": "nginx:1.19",
								},
							},
						},
					},
				},
			}

			tenant := &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tenantName,
					Namespace: namespace,
				},
				Spec: lynqv1.LynqNodeSpec{
					Deployments: []lynqv1.TResource{
						{
							ID: "once-deployment",
							Spec: unstructured.Unstructured{
								Object: deploymentSpec,
							},
							CreationPolicy: lynqv1.CreationPolicyOnce,   // Once policy
							IgnoreFields:   []string{"$.spec.replicas"}, // Should have no effect
						},
					},
				},
			}

			// When: Creating the LynqNode
			Expect(k8sClient.Create(ctx, tenant)).To(Succeed())
			defer cleanupResource(ctx, tenant)

			// Then: Deployment created with initial values
			deploymentKey := types.NamespacedName{
				Name:      tenantName + "-deployment",
				Namespace: namespace,
			}

			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, deploymentKey, deployment)
			}, timeout, interval).Should(Succeed())

			Expect(*deployment.Spec.Replicas).To(Equal(int32(4)))
			initialImage := deployment.Spec.Template.Spec.Containers[0].Image
			Expect(initialImage).To(Equal("nginx:1.19"))

			// ========================================
			// When: Update template (change image)
			// ========================================
			By("Updating template with new image")

			Eventually(func() error {
				latestLynqNode := &lynqv1.LynqNode{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: tenantName, Namespace: namespace}, latestLynqNode); err != nil {
					return err
				}

				deploymentSpec := latestLynqNode.Spec.Deployments[0].Spec.Object
				templateSpec := deploymentSpec["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})
				containers := templateSpec["containers"].([]interface{})
				container := containers[0].(map[string]interface{})
				container["image"] = "nginx:1.23"

				return k8sClient.Update(ctx, latestLynqNode)
			}, timeout, interval).Should(Succeed())

			time.Sleep(2 * time.Second)

			// ========================================
			// Then: Image should NOT be updated (Once policy behavior)
			// ========================================
			By("Verifying image is NOT updated (Once policy takes precedence)")

			Consistently(func() string {
				if err := k8sClient.Get(ctx, deploymentKey, deployment); err != nil {
					return ""
				}
				return deployment.Spec.Template.Spec.Containers[0].Image
			}, time.Second*5, interval).Should(Equal("nginx:1.19")) // Should remain initial value
		})
	})

	Context("When ignoreFields contains non-existent paths", func() {
		PIt("should not fail and continue reconciliation", func() {
			ctx := context.Background()

			// ========================================
			// Scenario: Graceful handling of non-existent ignore paths
			// ========================================

			// Given: A LynqNode with ignoreFields pointing to non-existent paths
			tenantName := fmt.Sprintf("tenant-nonexist-ignore-%d", time.Now().UnixNano())
			namespace := defaultNamespace

			serviceSpec := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": tenantName + "-svc",
				},
				"spec": map[string]interface{}{
					"type": "ClusterIP",
					"ports": []interface{}{
						map[string]interface{}{
							"port":       int64(80),
							"targetPort": int64(8080),
						},
					},
					"selector": map[string]interface{}{
						"app": "test",
					},
				},
			}

			tenant := &lynqv1.LynqNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tenantName,
					Namespace: namespace,
				},
				Spec: lynqv1.LynqNodeSpec{
					Services: []lynqv1.TResource{
						{
							ID: "test-service",
							Spec: unstructured.Unstructured{
								Object: serviceSpec,
							},
							IgnoreFields: []string{
								"$.spec.nonExistentField",          // Doesn't exist
								"$.spec.deeply.nested.nonExistent", // Path doesn't exist
								"$.spec.ports[5].protocol",         // Array index out of bounds
							},
						},
					},
				},
			}

			// When: Creating the LynqNode
			Expect(k8sClient.Create(ctx, tenant)).To(Succeed())
			defer cleanupResource(ctx, tenant)

			// Then: Service should be created successfully (ignore non-existent paths)
			serviceKey := types.NamespacedName{
				Name:      tenantName + "-svc",
				Namespace: namespace,
			}

			service := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, serviceKey, service)
			}, timeout, interval).Should(Succeed())

			// Verify service spec is intact
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
			Expect(service.Spec.Ports).To(HaveLen(1))
			Expect(service.Spec.Ports[0].Port).To(Equal(int32(80)))
		})
	})
})

// Helper function for cleanup
func cleanupResource(ctx context.Context, obj client.Object) {
	_ = k8sClient.Delete(ctx, obj)
}
