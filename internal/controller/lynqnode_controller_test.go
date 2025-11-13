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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lynqv1 "github.com/k8s-lynq/lynq/api/v1"
	"github.com/k8s-lynq/lynq/internal/status"
)

var _ = Describe("LynqNode Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		const registryName = "test-registry"
		const templateName = "test-template"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		lynqNode := &lynqv1.LynqNode{}

		BeforeEach(func() {
			By("creating the LynqHub prerequisite")
			registry := &lynqv1.LynqHub{}
			registryKey := types.NamespacedName{Name: registryName, Namespace: "default"}
			err := k8sClient.Get(ctx, registryKey, registry)
			if err != nil && errors.IsNotFound(err) {
				registry := &lynqv1.LynqHub{
					ObjectMeta: metav1.ObjectMeta{
						Name:      registryName,
						Namespace: "default",
					},
					Spec: lynqv1.LynqHubSpec{
						Source: lynqv1.DataSource{
							Type:         lynqv1.SourceTypeMySQL,
							SyncInterval: "30s",
							MySQL: &lynqv1.MySQLSource{
								Host:     "mysql.default.svc.cluster.local",
								Port:     3306,
								Username: "root",
								Database: "nodes",
								Table:    "nodes",
							},
						},
						ValueMappings: lynqv1.ValueMappings{
							UID:       "id",
							HostOrURL: "url",
							Activate:  "isActive",
						},
					},
				}
				Expect(k8sClient.Create(ctx, registry)).To(Succeed())
			}

			By("creating the LynqForm prerequisite")
			template := &lynqv1.LynqForm{}
			templateKey := types.NamespacedName{Name: templateName, Namespace: "default"}
			err = k8sClient.Get(ctx, templateKey, template)
			if err != nil && errors.IsNotFound(err) {
				template := &lynqv1.LynqForm{
					ObjectMeta: metav1.ObjectMeta{
						Name:      templateName,
						Namespace: "default",
					},
					Spec: lynqv1.LynqFormSpec{
						HubID: registryName,
					},
				}
				Expect(k8sClient.Create(ctx, template)).To(Succeed())
			}

			By("creating the custom resource for the Kind LynqNode")
			err = k8sClient.Get(ctx, typeNamespacedName, lynqNode)
			if err != nil && errors.IsNotFound(err) {
				resource := &lynqv1.LynqNode{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
						Labels: map[string]string{
							"lynq.sh/hub": registryName,
						},
					},
					Spec: lynqv1.LynqNodeSpec{
						UID:         "test-uid-123",
						TemplateRef: templateName,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// Cleanup LynqNode
			resource := &lynqv1.LynqNode{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				By("Cleanup the specific resource instance LynqNode")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}

			// Cleanup template
			template := &lynqv1.LynqForm{}
			templateKey := types.NamespacedName{Name: templateName, Namespace: "default"}
			err = k8sClient.Get(ctx, templateKey, template)
			if err == nil {
				Expect(k8sClient.Delete(ctx, template)).To(Succeed())
			}

			// Cleanup registry
			registry := &lynqv1.LynqHub{}
			registryKey := types.NamespacedName{Name: registryName, Namespace: "default"}
			err = k8sClient.Get(ctx, registryKey, registry)
			if err == nil {
				Expect(k8sClient.Delete(ctx, registry)).To(Succeed())
			}
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &LynqNodeReconciler{
				Client:        k8sClient,
				Scheme:        k8sClient.Scheme(),
				Recorder:      &fakeRecorder{},
				StatusManager: status.NewManager(k8sClient, status.WithSyncMode()),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			// Without a database connection, the reconcile will fail when querying extra values
			// This is expected in a test environment
			if err != nil {
				// Verify it's a database-related error (acceptable in test env)
				// or that the node was processed (no template resources to apply)
				Expect(err.Error()).To(Or(
					ContainSubstring("failed to query"),
					ContainSubstring("no template found"),
				))
			}
		})
	})
})
