/*
Copyright 2024.

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

package observability

import (
	"context"
	"fmt"
	"github.com/kubevirt/hyperconverged-cluster-operator/controllers/commontestutils"
	"github.com/kubevirt/hyperconverged-cluster-operator/pkg/monitoring/rules"
	hcoutil "github.com/kubevirt/hyperconverged-cluster-operator/pkg/util"
	"github.com/machadovilaca/operator-observability/pkg/operatorrules"
	"github.com/onsi/gomega/gstruct"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	"os"
	"testing"

	//nolint:golint
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Observability Controller Suite")
}

var _ = Describe("Observability controller", func() {
	var (
		ci     = commontestutils.ClusterInfoMock{}
		ctx    = context.Background()
		config = &rest.Config{}
		ns     *corev1.Namespace
	)

	Context("Rules reconciler", func() {

		//req := reconcile.Request{}

		BeforeEach(func() {
			ns = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: commontestutils.Namespace,
				},
			}

		})

		It("should create prometheusrule if missing", func() {
			cl := commontestutils.InitClient([]client.Object{ns})

			r := NewReconciler(config, cl, ci)
			Expect(r.reconcileRules(ctx)).To(Succeed())

			pr := &monitoringv1.PrometheusRule{}
			Expect(cl.Get(context.Background(), client.ObjectKey{Namespace: r.namespace, Name: ruleName}, pr)).To(Succeed())
		})

		It("should add the labels if they are missing", func() {
			owner := getDeploymentReference(ci.GetDeployment())
			existRule, err := rules.BuildPrometheusRule2(commontestutils.Namespace, owner)
			Expect(err).ToNot(HaveOccurred())
			existRule.Labels = nil

			cl := commontestutils.InitClient([]client.Object{ns, existRule})
			r := NewReconciler(config, cl, ci)

			Expect(r.reconcileRules(ctx)).To(Succeed())
			pr := &monitoringv1.PrometheusRule{}
			Expect(cl.Get(context.Background(), client.ObjectKey{Namespace: r.namespace, Name: ruleName}, pr)).To(Succeed())

			Expect(pr.Labels).To(Equal(hcoutil.GetLabels(hcoutil.HyperConvergedName, hcoutil.AppComponentMonitoring)))
		})

		It("should update the labels if modified", func() {
			err := rules.SetupRules()
			Expect(err).ToNot(HaveOccurred())

			owner := getDeploymentReference(ci.GetDeployment())
			existRule, err := rules.BuildPrometheusRule2(commontestutils.Namespace, owner)
			Expect(err).ToNot(HaveOccurred())
			existRule.Labels = map[string]string{
				"wrongKey1": "wrongValue1",
				"wrongKey2": "wrongValue2",
				"wrongKey3": "wrongValue3",
			}

			cl := commontestutils.InitClient([]client.Object{ns, existRule})
			r := NewReconciler(config, cl, ci)

			Expect(r.reconcileRules(ctx)).To(Succeed())

			pr := &monitoringv1.PrometheusRule{}
			Expect(cl.Get(context.Background(), client.ObjectKey{Namespace: r.namespace, Name: ruleName}, pr)).To(Succeed())

			Expect(pr.Labels).To(gstruct.MatchKeys(gstruct.IgnoreExtras, commontestutils.KeysFromSSMap(hcoutil.GetLabels(hcoutil.HyperConvergedName, hcoutil.AppComponentMonitoring))))
		})

		It("should add the OwnerReference if missing", func() {
			owner := metav1.OwnerReference{}
			existRule, err := rules.BuildPrometheusRule2(commontestutils.Namespace, owner)
			Expect(err).ToNot(HaveOccurred())
			existRule.OwnerReferences = nil
			cl := commontestutils.InitClient([]client.Object{ns, existRule})
			r := NewReconciler(config, cl, ci)

			Expect(r.reconcileRules(ctx)).To(Succeed())
			pr := &monitoringv1.PrometheusRule{}
			Expect(cl.Get(context.Background(), client.ObjectKey{Namespace: r.namespace, Name: ruleName}, pr)).To(Succeed())

			deployment := ci.GetDeployment()

			Expect(pr.OwnerReferences).To(HaveLen(1))
			Expect(pr.OwnerReferences[0].Name).To(Equal(deployment.Name))
			Expect(pr.OwnerReferences[0].Kind).To(Equal("Deployment"))
			Expect(pr.OwnerReferences[0].APIVersion).To(Equal(appsv1.GroupName + "/v1"))
			Expect(pr.OwnerReferences[0].UID).To(Equal(deployment.UID))
		})

		It("should update the referenceOwner if modified", func() {
			owner := metav1.OwnerReference{
				APIVersion:         "wrongAPIVersion",
				Kind:               "wrongKind",
				Name:               "wrongName",
				Controller:         ptr.To(true),
				BlockOwnerDeletion: ptr.To(true),
				UID:                "0987654321",
			}
			existRule, err := rules.BuildPrometheusRule2(commontestutils.Namespace, owner)
			Expect(err).ToNot(HaveOccurred())
			cl := commontestutils.InitClient([]client.Object{ns, existRule})
			r := NewReconciler(config, cl, ci)

			Expect(r.reconcileRules(ctx)).To(Succeed())
			pr := &monitoringv1.PrometheusRule{}
			Expect(cl.Get(context.Background(), client.ObjectKey{Namespace: r.namespace, Name: ruleName}, pr)).To(Succeed())

			deployment := ci.GetDeployment()

			Expect(pr.OwnerReferences).To(HaveLen(1))
			Expect(pr.OwnerReferences[0].Name).To(Equal(deployment.Name))
			Expect(pr.OwnerReferences[0].Kind).To(Equal("Deployment"))
			Expect(pr.OwnerReferences[0].APIVersion).To(Equal(appsv1.GroupName + "/v1"))
			Expect(pr.OwnerReferences[0].UID).To(Equal(deployment.UID))
		})

		It("should add the spec if it's missing", func() {
			owner := getDeploymentReference(ci.GetDeployment())
			existRule, err := rules.BuildPrometheusRule2(commontestutils.Namespace, owner)
			Expect(err).ToNot(HaveOccurred())

			existRule.Spec = monitoringv1.PrometheusRuleSpec{}

			cl := commontestutils.InitClient([]client.Object{ns, existRule})
			r := NewReconciler(config, cl, ci)

			Expect(r.reconcileRules(ctx)).To(Succeed())
			pr := &monitoringv1.PrometheusRule{}
			Expect(cl.Get(context.Background(), client.ObjectKey{Namespace: r.namespace, Name: ruleName}, pr)).To(Succeed())
			newRule, err := rules.BuildPrometheusRule(commontestutils.Namespace, owner)
			Expect(err).ToNot(HaveOccurred())
			Expect(pr.Spec).To(Equal(newRule.Spec))
		})

		It("should update the spec if modified", func() {
			owner := getDeploymentReference(ci.GetDeployment())
			existRule, err := rules.BuildPrometheusRule(commontestutils.Namespace, owner)
			Expect(err).ToNot(HaveOccurred())

			existRule.Spec.Groups[1].Rules = []monitoringv1.Rule{
				existRule.Spec.Groups[1].Rules[0],
				existRule.Spec.Groups[1].Rules[2],
				existRule.Spec.Groups[1].Rules[3],
			}
			// modify the first rule
			existRule.Spec.Groups[1].Rules[0].Alert = "modified alert"

			cl := commontestutils.InitClient([]client.Object{ns, existRule})
			r := NewReconciler(config, cl, ci)

			Expect(r.reconcileRules(ctx)).To(Succeed())
			pr := &monitoringv1.PrometheusRule{}
			Expect(cl.Get(context.Background(), client.ObjectKey{Namespace: r.namespace, Name: ruleName}, pr)).To(Succeed())
			newRule, err := rules.BuildPrometheusRule(commontestutils.Namespace, owner)
			Expect(err).ToNot(HaveOccurred())
			Expect(pr.Spec).To(Equal(newRule.Spec))
		})

		It("should use the default runbook URL template when no ENV Variable is set", func() {
			owner := getDeploymentReference(ci.GetDeployment())
			promRule, err := rules.BuildPrometheusRule2(commontestutils.Namespace, owner)
			Expect(err).ToNot(HaveOccurred())

			for _, group := range promRule.Spec.Groups {
				for _, rule := range group.Rules {
					if rule.Alert != "" {
						if rule.Annotations["runbook_url"] != "" {
							Expect(rule.Annotations["runbook_url"]).To(Equal(fmt.Sprintf(defaultRunbookURLTemplate, rule.Alert)))
						}
					}
				}
			}
		})

		It("should use the desired runbook URL template when its ENV Variable is set", func() {
			err := operatorrules.CleanRegistry()
			Expect(err).ToNot(HaveOccurred())

			desiredRunbookURLTemplate := "desired/runbookURL/template/%s"
			os.Setenv(runbookURLTemplateEnv, desiredRunbookURLTemplate)

			err = rules.SetupRules()
			Expect(err).ToNot(HaveOccurred())

			owner := getDeploymentReference(ci.GetDeployment())
			promRule, err := rules.BuildPrometheusRule2(commontestutils.Namespace, owner)
			Expect(err).ToNot(HaveOccurred())

			for _, group := range promRule.Spec.Groups {
				for _, rule := range group.Rules {
					if rule.Alert != "" {
						if rule.Annotations["runbook_url"] != "" {
							Expect(rule.Annotations["runbook_url"]).To(Equal(fmt.Sprintf(desiredRunbookURLTemplate, rule.Alert)))
						}
					}
				}
			}
		})
	})
})
