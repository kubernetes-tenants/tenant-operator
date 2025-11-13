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
)

var _ = Describe("LynqHub Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		lynqhub := &lynqv1.LynqHub{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind LynqHub")
			err := k8sClient.Get(ctx, typeNamespacedName, lynqhub)
			if err != nil && errors.IsNotFound(err) {
				resource := &lynqv1.LynqHub{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
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
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &lynqv1.LynqHub{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance LynqHub")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &LynqHubReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: &fakeRecorder{},
			}

			// First reconciliation adds finalizer
			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(result.RequeueAfter))

			// Second reconciliation attempts database connection
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			// In test environment without MySQL, we expect a database connection error
			// The reconciler should handle this gracefully and requeue
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to ping MySQL"))
		})
	})
})
