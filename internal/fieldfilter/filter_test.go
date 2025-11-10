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

package fieldfilter_test

import (
	"testing"

	"github.com/kubernetes-tenants/tenant-operator/internal/fieldfilter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestFieldFilter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FieldFilter Suite")
}

var _ = Describe("FieldFilter with ojg/jp", func() {
	Describe("Validation", func() {
		Context("When creating a filter with valid JSONPath expressions", func() {
			It("should succeed for simple root field path", func() {
				// Given: A valid simple JSONPath expression
				paths := []string{"$.spec.replicas"}

				// When: Creating a new filter
				filter, err := fieldfilter.NewFilter(paths)

				// Then: No error should occur
				Expect(err).NotTo(HaveOccurred())
				Expect(filter).NotTo(BeNil())
			})

			It("should succeed for nested field path", func() {
				// Given: A valid nested JSONPath expression
				paths := []string{"$.spec.template.spec.containers[0].image"}

				// When: Creating a new filter
				filter, err := fieldfilter.NewFilter(paths)

				// Then: No error should occur
				Expect(err).NotTo(HaveOccurred())
				Expect(filter).NotTo(BeNil())
			})

			It("should succeed for map key with special characters", func() {
				// Given: A JSONPath with map key containing special chars
				paths := []string{"$.metadata.annotations['app.kubernetes.io/name']"}

				// When: Creating a new filter
				filter, err := fieldfilter.NewFilter(paths)

				// Then: No error should occur (ojg supports this!)
				Expect(err).NotTo(HaveOccurred())
				Expect(filter).NotTo(BeNil())
			})

			It("should succeed for wildcard expressions", func() {
				// Given: A JSONPath with wildcard
				paths := []string{"$.spec.template.spec.containers[*].image"}

				// When: Creating a new filter
				filter, err := fieldfilter.NewFilter(paths)

				// Then: No error should occur
				Expect(err).NotTo(HaveOccurred())
				Expect(filter).NotTo(BeNil())
			})

			It("should succeed for multiple valid paths", func() {
				// Given: Multiple valid JSONPath expressions
				paths := []string{
					"$.spec.replicas",
					"$.spec.template.spec.containers[0].resources",
					"$.metadata.annotations['custom-annotation']",
					"$.spec.template.spec.containers[*].env",
				}

				// When: Creating a new filter
				filter, err := fieldfilter.NewFilter(paths)

				// Then: No error should occur
				Expect(err).NotTo(HaveOccurred())
				Expect(filter).NotTo(BeNil())
			})

			It("should succeed with empty path list", func() {
				// Given: An empty path list
				paths := []string{}

				// When: Creating a new filter
				filter, err := fieldfilter.NewFilter(paths)

				// Then: No error should occur (no-op filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(filter).NotTo(BeNil())
			})
		})

		Context("When creating a filter with invalid JSONPath expressions", func() {
			It("should fail for completely malformed JSONPath", func() {
				// Given: A completely invalid JSONPath (unclosed bracket)
				paths := []string{"$[invalid"}

				// When: Creating a new filter
				filter, err := fieldfilter.NewFilter(paths)

				// Then: An error should occur
				Expect(err).To(HaveOccurred())
				Expect(filter).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid JSONPath"))
			})

			It("should fail for incomplete bracket syntax", func() {
				// Given: An incomplete bracket syntax
				paths := []string{"$['unclosed"}

				// When: Creating a new filter
				filter, err := fieldfilter.NewFilter(paths)

				// Then: An error should occur
				Expect(err).To(HaveOccurred())
				Expect(filter).To(BeNil())
			})
		})
	})

	Describe("Field Removal", func() {
		Context("When removing simple root-level fields", func() {
			It("should remove a single integer field", func() {
				// Given: A Deployment object with replicas field
				obj := &unstructured.Unstructured{
					Object: map[string]any{
						"apiVersion": "apps/v1",
						"kind":       "Deployment",
						"metadata": map[string]any{
							"name": "test-deployment",
						},
						"spec": map[string]any{
							"replicas": int64(3),
							"selector": map[string]any{
								"matchLabels": map[string]any{
									"app": "test",
								},
							},
						},
					},
				}

				// And: A filter that ignores replicas
				filter, err := fieldfilter.NewFilter([]string{"$.spec.replicas"})
				Expect(err).NotTo(HaveOccurred())

				// When: Removing ignored fields
				err = filter.RemoveIgnoredFields(obj)

				// Then: No error should occur
				Expect(err).NotTo(HaveOccurred())

				// And: replicas field should be removed
				spec, found, err := unstructured.NestedMap(obj.Object, "spec")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				_, replicasExists := spec["replicas"]
				Expect(replicasExists).To(BeFalse())

				// And: other fields should remain
				selector, found, err := unstructured.NestedMap(obj.Object, "spec", "selector")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(selector).To(HaveKey("matchLabels"))
			})
		})

		Context("When removing nested fields", func() {
			It("should remove a deeply nested field", func() {
				// Given: A Deployment with container image
				obj := &unstructured.Unstructured{
					Object: map[string]any{
						"apiVersion": "apps/v1",
						"kind":       "Deployment",
						"spec": map[string]any{
							"template": map[string]any{
								"spec": map[string]any{
									"containers": []any{
										map[string]any{
											"name":  "app",
											"image": "nginx:1.20",
											"ports": []any{
												map[string]any{
													"containerPort": int64(80),
												},
											},
										},
									},
								},
							},
						},
					},
				}

				// And: A filter that ignores container image
				filter, err := fieldfilter.NewFilter([]string{"$.spec.template.spec.containers[0].image"})
				Expect(err).NotTo(HaveOccurred())

				// When: Removing ignored fields
				err = filter.RemoveIgnoredFields(obj)

				// Then: image field should be removed
				Expect(err).NotTo(HaveOccurred())
				containers, found, err := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(containers).To(HaveLen(1))

				container := containers[0].(map[string]any)
				_, imageExists := container["image"]
				Expect(imageExists).To(BeFalse())

				// And: other container fields should remain
				Expect(container).To(HaveKey("name"))
				Expect(container).To(HaveKey("ports"))
			})
		})

		Context("When removing fields with wildcards", func() {
			It("should remove fields from all matching array elements", func() {
				// Given: A Deployment with multiple containers
				obj := &unstructured.Unstructured{
					Object: map[string]any{
						"spec": map[string]any{
							"template": map[string]any{
								"spec": map[string]any{
									"containers": []any{
										map[string]any{
											"name":  "app",
											"image": "nginx:1.20",
										},
										map[string]any{
											"name":  "sidecar",
											"image": "busybox:latest",
										},
									},
								},
							},
						},
					},
				}

				// And: A filter with wildcard that removes all images
				filter, err := fieldfilter.NewFilter([]string{"$.spec.template.spec.containers[*].image"})
				Expect(err).NotTo(HaveOccurred())

				// When: Removing ignored fields
				err = filter.RemoveIgnoredFields(obj)

				// Then: All images should be removed
				Expect(err).NotTo(HaveOccurred())
				containers, found, err := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(containers).To(HaveLen(2))

				// Both containers should have no image field
				for _, c := range containers {
					container := c.(map[string]any)
					_, imageExists := container["image"]
					Expect(imageExists).To(BeFalse())
					Expect(container).To(HaveKey("name")) // But name should remain
				}
			})
		})

		Context("When removing map keys with special characters", func() {
			It("should remove annotation with dots and slashes", func() {
				// Given: An object with complex annotation keys
				obj := &unstructured.Unstructured{
					Object: map[string]any{
						"metadata": map[string]any{
							"name": "test",
							"annotations": map[string]any{
								"app.kubernetes.io/name":    "myapp",
								"app.kubernetes.io/version": "v1.0.0",
								"simple":                    "value",
							},
						},
					},
				}

				// And: A filter that removes specific annotation
				filter, err := fieldfilter.NewFilter([]string{"$.metadata.annotations['app.kubernetes.io/name']"})
				Expect(err).NotTo(HaveOccurred())

				// When: Removing ignored fields
				err = filter.RemoveIgnoredFields(obj)

				// Then: Specific annotation should be removed
				Expect(err).NotTo(HaveOccurred())
				annotations, found, err := unstructured.NestedStringMap(obj.Object, "metadata", "annotations")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())

				_, nameExists := annotations["app.kubernetes.io/name"]
				Expect(nameExists).To(BeFalse())

				// And: other annotations should remain
				Expect(annotations).To(HaveKey("app.kubernetes.io/version"))
				Expect(annotations).To(HaveKey("simple"))
			})
		})

		Context("When removing multiple fields", func() {
			It("should remove multiple independent fields", func() {
				// Given: A Deployment with multiple fields to ignore
				obj := &unstructured.Unstructured{
					Object: map[string]any{
						"spec": map[string]any{
							"replicas": int64(3),
							"strategy": map[string]any{
								"type": "RollingUpdate",
							},
							"template": map[string]any{
								"spec": map[string]any{
									"containers": []any{
										map[string]any{
											"name":  "app",
											"image": "nginx:1.20",
											"resources": map[string]any{
												"limits": map[string]any{
													"cpu": "500m",
												},
											},
										},
									},
								},
							},
						},
					},
				}

				// And: A filter that ignores multiple fields
				filter, err := fieldfilter.NewFilter([]string{
					"$.spec.replicas",
					"$.spec.template.spec.containers[0].resources",
				})
				Expect(err).NotTo(HaveOccurred())

				// When: Removing ignored fields
				err = filter.RemoveIgnoredFields(obj)

				// Then: Both fields should be removed
				Expect(err).NotTo(HaveOccurred())

				spec, found, err := unstructured.NestedMap(obj.Object, "spec")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				_, replicasExists := spec["replicas"]
				Expect(replicasExists).To(BeFalse())

				containers, found, err := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				container := containers[0].(map[string]any)
				_, resourcesExist := container["resources"]
				Expect(resourcesExist).To(BeFalse())

				// And: other fields remain
				Expect(spec).To(HaveKey("strategy"))
				Expect(container).To(HaveKey("image"))
			})
		})

		Context("When handling edge cases", func() {
			It("should succeed when field doesn't exist", func() {
				// Given: A Deployment without the field to remove
				obj := &unstructured.Unstructured{
					Object: map[string]any{
						"spec": map[string]any{
							"selector": map[string]any{
								"matchLabels": map[string]any{
									"app": "test",
								},
							},
						},
					},
				}

				// And: A filter that tries to remove non-existent field
				filter, err := fieldfilter.NewFilter([]string{"$.spec.replicas"})
				Expect(err).NotTo(HaveOccurred())

				// When: Removing ignored fields
				err = filter.RemoveIgnoredFields(obj)

				// Then: No error should occur (gracefully handled)
				Expect(err).NotTo(HaveOccurred())

				// And: object remains unchanged
				spec, found, err := unstructured.NestedMap(obj.Object, "spec")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(spec).To(HaveKey("selector"))
			})

			It("should handle array index out of bounds gracefully", func() {
				// Given: A Deployment with one container
				obj := &unstructured.Unstructured{
					Object: map[string]any{
						"spec": map[string]any{
							"template": map[string]any{
								"spec": map[string]any{
									"containers": []any{
										map[string]any{
											"name": "app",
										},
									},
								},
							},
						},
					},
				}

				// And: A filter that tries to access non-existent second container
				filter, err := fieldfilter.NewFilter([]string{"$.spec.template.spec.containers[5].image"})
				Expect(err).NotTo(HaveOccurred())

				// When: Removing ignored fields
				err = filter.RemoveIgnoredFields(obj)

				// Then: No error should occur (gracefully handled)
				Expect(err).NotTo(HaveOccurred())

				// And: first container remains unchanged
				containers, found, err := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(containers).To(HaveLen(1))
			})
		})

		Context("When filter has no ignore paths", func() {
			It("should not modify the object", func() {
				// Given: A Deployment object
				original := &unstructured.Unstructured{
					Object: map[string]any{
						"spec": map[string]any{
							"replicas": int64(3),
							"selector": map[string]any{
								"matchLabels": map[string]any{
									"app": "test",
								},
							},
						},
					},
				}

				// And: A filter with no ignore paths (no-op)
				filter, err := fieldfilter.NewFilter([]string{})
				Expect(err).NotTo(HaveOccurred())

				// When: Removing ignored fields
				err = filter.RemoveIgnoredFields(original)

				// Then: Object should remain unchanged
				Expect(err).NotTo(HaveOccurred())
				spec, found, err := unstructured.NestedMap(original.Object, "spec")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(spec).To(HaveKey("replicas"))
				Expect(spec).To(HaveKey("selector"))
			})
		})
	})

	Describe("ValidateJSONPath helper", func() {
		It("should validate correct JSONPath", func() {
			err := fieldfilter.ValidateJSONPath("$.spec.replicas")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject invalid JSONPath", func() {
			err := fieldfilter.ValidateJSONPath("not-a-jsonpath")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetMatchingFields helper", func() {
		It("should return matching field values", func() {
			// Given: An object with values
			obj := &unstructured.Unstructured{
				Object: map[string]any{
					"spec": map[string]any{
						"replicas": int64(3),
						"selector": map[string]any{
							"app": "test",
						},
					},
				},
			}

			// And: A filter
			filter, err := fieldfilter.NewFilter([]string{"$.spec.replicas"})
			Expect(err).NotTo(HaveOccurred())

			// When: Getting matching fields
			matches, err := filter.GetMatchingFields(obj)

			// Then: Should return the values
			Expect(err).NotTo(HaveOccurred())
			Expect(matches).To(HaveKey("$.spec.replicas"))
			Expect(matches["$.spec.replicas"]).To(ContainElement(int64(3)))
		})
	})
})
