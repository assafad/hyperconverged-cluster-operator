package operands

import (
	"context"
	"reflect"

	"k8s.io/utils/pointer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	consolev1 "github.com/openshift/api/console/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hcov1beta1 "github.com/kubevirt/hyperconverged-cluster-operator/api/v1beta1"
	"github.com/kubevirt/hyperconverged-cluster-operator/controllers/common"
	"github.com/kubevirt/hyperconverged-cluster-operator/controllers/commontestutils"
	hcoutil "github.com/kubevirt/hyperconverged-cluster-operator/pkg/util"
)

var _ = Describe("Kubevirt Console Plugin", func() {
	Context("Console Plugin CR", func() {
		var hco *hcov1beta1.HyperConverged
		var req *common.HcoRequest

		var expectedConsoleConfig = &operatorv1.Console{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Console",
				APIVersion: "operator.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster",
			},
		}

		BeforeEach(func() {
			hco = commontestutils.NewHco()
			req = commontestutils.NewReq(hco)
		})

		It("should create plugin CR if not present", func() {
			expectedResource := NewKVConsolePlugin(hco)
			cl := commontestutils.InitClient([]client.Object{})
			handler, err := newKvUIPluginCRHandler(logger, cl, commontestutils.GetScheme(), hco)
			Expect(err).ToNot(HaveOccurred())

			res := handler[0].ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &consolev1.ConsolePlugin{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())
			Expect(foundResource.Name).To(Equal(expectedResource.Name))
			Expect(foundResource.Labels).Should(HaveKeyWithValue(hcoutil.AppLabel, commontestutils.Name))
			Expect(foundResource.Namespace).To(Equal(expectedResource.Namespace))
		})

		It("should find plugin CR if present", func() {
			expectedResource := NewKVConsolePlugin(hco)

			cl := commontestutils.InitClient([]client.Object{hco, expectedResource, expectedConsoleConfig})
			handler, err := newKvUIPluginCRHandler(logger, cl, commontestutils.GetScheme(), hco)
			Expect(err).ToNot(HaveOccurred())

			res := handler[0].ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Err).ToNot(HaveOccurred())

			// Check HCO's status
			Expect(hco.Status.RelatedObjects).To(Not(BeNil()))
			objectRef, err := reference.GetReference(commontestutils.GetScheme(), expectedResource)
			Expect(err).ToNot(HaveOccurred())
			// ObjectReference should have been added
			Expect(hco.Status.RelatedObjects).To(ContainElement(*objectRef))
		})

		It("should reconcile plugin spec to default if changed", func() {
			expectedResource := NewKVConsolePlugin(hco)
			outdatedResource := NewKVConsolePlugin(hco)

			outdatedResource.Spec.Backend.Service.Port = int32(6666)
			outdatedResource.Spec.DisplayName = "fake plugin name"

			cl := commontestutils.InitClient([]client.Object{hco, outdatedResource, expectedConsoleConfig})
			handler, err := newKvUIPluginCRHandler(logger, cl, commontestutils.GetScheme(), hco)
			Expect(err).ToNot(HaveOccurred())

			res := handler[0].ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Updated).To(BeTrue())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &consolev1.ConsolePlugin{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())

			Expect(reflect.DeepEqual(expectedResource.Spec, foundResource.Spec)).To(BeTrue())

			// ObjectReference should have been updated
			Expect(hco.Status.RelatedObjects).To(Not(BeNil()))
			objectRefOutdated, err := reference.GetReference(commontestutils.GetScheme(), outdatedResource)
			Expect(err).ToNot(HaveOccurred())
			objectRefFound, err := reference.GetReference(commontestutils.GetScheme(), foundResource)
			Expect(err).ToNot(HaveOccurred())
			Expect(hco.Status.RelatedObjects).To(Not(ContainElement(*objectRefOutdated)))
			Expect(hco.Status.RelatedObjects).To(ContainElement(*objectRefFound))
		})

		It("should reconcile plugin labels to default if changed", func() {
			expectedResource := NewKVConsolePlugin(hco)
			outdatedResource := NewKVConsolePlugin(hco)

			outdatedResource.Labels[hcoutil.AppLabel] = "changed"

			cl := commontestutils.InitClient([]client.Object{hco, outdatedResource, expectedConsoleConfig})
			handler, err := newKvUIPluginCRHandler(logger, cl, commontestutils.GetScheme(), hco)
			Expect(err).ToNot(HaveOccurred())

			res := handler[0].ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Updated).To(BeTrue())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &consolev1.ConsolePlugin{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())

			Expect(reflect.DeepEqual(foundResource.Labels, expectedResource.Labels)).To(BeTrue())
		})

		It("should reconcile plugin labels to default if added", func() {
			expectedResource := NewKVConsolePlugin(hco)
			outdatedResource := NewKVConsolePlugin(hco)

			outdatedResource.Labels["fake_label"] = "something"

			cl := commontestutils.InitClient([]client.Object{hco, outdatedResource, expectedConsoleConfig})
			handler, err := newKvUIPluginCRHandler(logger, cl, commontestutils.GetScheme(), hco)
			Expect(err).ToNot(HaveOccurred())

			res := handler[0].ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Updated).To(BeTrue())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &consolev1.ConsolePlugin{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())

			Expect(reflect.DeepEqual(foundResource.Labels, expectedResource.Labels)).To(BeTrue())
		})

		It("should reconcile plugin labels to default if deleted", func() {
			expectedResource := NewKVConsolePlugin(hco)
			outdatedResource := NewKVConsolePlugin(hco)

			delete(outdatedResource.Labels, hcoutil.AppLabel)

			cl := commontestutils.InitClient([]client.Object{hco, outdatedResource, expectedConsoleConfig})
			handler, err := newKvUIPluginCRHandler(logger, cl, commontestutils.GetScheme(), hco)
			Expect(err).ToNot(HaveOccurred())

			res := handler[0].ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Updated).To(BeTrue())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &consolev1.ConsolePlugin{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())

			Expect(reflect.DeepEqual(foundResource.Labels, expectedResource.Labels)).To(BeTrue())
		})
	})

	Context("Kubevirt Plugin Deployment", func() {
		var hco *hcov1beta1.HyperConverged
		var req *common.HcoRequest

		BeforeEach(func() {
			hco = commontestutils.NewHco()
			req = commontestutils.NewReq(hco)
		})

		It("should create if not present", func() {
			expectedResource, err := NewKvUIPluginDeplymnt(hco)
			Expect(err).ToNot(HaveOccurred())

			cl := commontestutils.InitClient([]client.Object{})
			handler, err := newKvUIPluginDeploymentHandler(logger, cl, commontestutils.GetScheme(), hco)
			Expect(err).ToNot(HaveOccurred())

			res := handler[0].ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &appsv1.Deployment{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())
			Expect(foundResource.Name).To(Equal(expectedResource.Name))
			Expect(foundResource.Labels).Should(HaveKeyWithValue(hcoutil.AppLabel, commontestutils.Name))
			Expect(foundResource.Spec.Template.Labels).Should(HaveKeyWithValue(hcoutil.AppLabel, commontestutils.Name))
			Expect(foundResource.Spec.Template.Labels).Should(HaveKeyWithValue(hcoutil.AppLabelComponent, string(hcoutil.AppComponentDeployment)))
			Expect(foundResource.Spec.Template.Labels).Should(HaveKeyWithValue(hcoutil.AppLabelManagedBy, hcoutil.OperatorName))
			Expect(foundResource.Spec.Template.Labels).Should(HaveKeyWithValue(hcoutil.AppLabelVersion, hcoutil.GetHcoKvIoVersion()))
			Expect(foundResource.Spec.Template.Labels).Should(HaveKeyWithValue(hcoutil.AppLabelPartOf, hcoutil.HyperConvergedCluster))
			Expect(foundResource.Namespace).To(Equal(expectedResource.Namespace))
			Expect(reflect.DeepEqual(expectedResource.Spec, foundResource.Spec)).To(BeTrue())
		})

		It("should find plugin deployment if present", func() {
			expectedResource, err := NewKvUIPluginDeplymnt(hco)
			Expect(err).ToNot(HaveOccurred())

			cl := commontestutils.InitClient([]client.Object{hco, expectedResource})
			handler, err := newKvUIPluginDeploymentHandler(logger, cl, commontestutils.GetScheme(), hco)
			Expect(err).ToNot(HaveOccurred())

			res := handler[0].ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &appsv1.Deployment{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())

			// Check HCO's status
			Expect(hco.Status.RelatedObjects).To(Not(BeNil()))
			objectRef, err := reference.GetReference(commontestutils.GetScheme(), foundResource)
			Expect(err).ToNot(HaveOccurred())
			// ObjectReference should have been added
			Expect(hco.Status.RelatedObjects).To(ContainElement(*objectRef))
		})

		It("should reconcile deployment to default if changed - (updatable fields)", func() {
			expectedResource, _ := NewKvUIPluginDeplymnt(hco)
			outdatedResource, _ := NewKvUIPluginDeplymnt(hco)

			outdatedResource.ObjectMeta.UID = "oldObjectUID"
			outdatedResource.ObjectMeta.ResourceVersion = "1234"

			outdatedResource.Spec.Replicas = pointer.Int32(123)
			outdatedResource.Spec.Template.Spec.Containers[0].Image = "quay.io/fake/image:latest"

			cl := commontestutils.InitClient([]client.Object{hco, outdatedResource})
			handler, err := newKvUIPluginDeploymentHandler(logger, cl, commontestutils.GetScheme(), hco)
			Expect(err).ToNot(HaveOccurred())
			res := handler[0].ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Updated).To(BeTrue())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &appsv1.Deployment{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())

			Expect(foundResource.Spec.Replicas).ToNot(Equal(outdatedResource.Spec.Replicas))
			Expect(foundResource.Spec.Replicas).To(Equal(expectedResource.Spec.Replicas))
			Expect(foundResource.Spec.Template.Spec.Containers[0].Image).To(Equal(expectedResource.Spec.Template.Spec.Containers[0].Image))
			Expect(reflect.DeepEqual(expectedResource.Spec, foundResource.Spec)).To(BeTrue())

			// ObjectReference should have been updated
			Expect(hco.Status.RelatedObjects).To(Not(BeNil()))
			objectRefOutdated, err := reference.GetReference(commontestutils.GetScheme(), outdatedResource)
			Expect(err).ToNot(HaveOccurred())
			objectRefFound, err := reference.GetReference(commontestutils.GetScheme(), foundResource)
			Expect(err).ToNot(HaveOccurred())
			Expect(hco.Status.RelatedObjects).To(Not(ContainElement(*objectRefOutdated)))
			Expect(hco.Status.RelatedObjects).To(ContainElement(*objectRefFound))

			// let's check the object UID to ensure that the object get updated and not deleted and recreated
			Expect(foundResource.GetUID()).To(Equal(types.UID("oldObjectUID")))
		})

		It("should reconcile deployment to default if changed - (immutable fields)", func() {
			expectedResource, _ := NewKvUIPluginDeplymnt(hco)
			outdatedResource, _ := NewKvUIPluginDeplymnt(hco)

			outdatedResource.ObjectMeta.UID = "oldObjectUID"
			outdatedResource.ObjectMeta.ResourceVersion = "1234"

			outdatedResource.ObjectMeta.Labels[hcoutil.AppLabel] = "wrong label"
			outdatedResource.Spec.Selector.MatchLabels[hcoutil.AppLabel] = "wrong label"
			outdatedResource.Spec.Template.Labels[hcoutil.AppLabel] = "wrong label"

			cl := commontestutils.InitClient([]client.Object{hco, outdatedResource})
			handler, err := newKvUIPluginDeploymentHandler(logger, cl, commontestutils.GetScheme(), hco)
			Expect(err).ToNot(HaveOccurred())
			res := handler[0].ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Updated).To(BeTrue())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &appsv1.Deployment{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())

			Expect(foundResource.ObjectMeta.Labels).ToNot(Equal(outdatedResource.ObjectMeta.Labels))
			Expect(foundResource.ObjectMeta.Labels).To(Equal(expectedResource.ObjectMeta.Labels))
			Expect(reflect.DeepEqual(expectedResource.Spec, foundResource.Spec)).To(BeTrue())

			// ObjectReference should have been updated
			Expect(hco.Status.RelatedObjects).To(Not(BeNil()))
			objectRefOutdated, err := reference.GetReference(commontestutils.GetScheme(), outdatedResource)
			Expect(err).ToNot(HaveOccurred())
			objectRefFound, err := reference.GetReference(commontestutils.GetScheme(), foundResource)
			Expect(err).ToNot(HaveOccurred())
			Expect(hco.Status.RelatedObjects).To(Not(ContainElement(*objectRefOutdated)))
			Expect(hco.Status.RelatedObjects).To(ContainElement(*objectRefFound))

			// let's check the object UID to ensure that the object get really deleted and recreated
			Expect(foundResource.GetUID()).ToNot(Equal(types.UID("oldObjectUID")))
		})
	})

	Context("Kubevirt Plugin Service", func() {
		var hco *hcov1beta1.HyperConverged
		var req *common.HcoRequest

		BeforeEach(func() {
			hco = commontestutils.NewHco()
			req = commontestutils.NewReq(hco)
		})

		It("should create plugin service if not present", func() {
			expectedResource := NewKvUIPluginSvc(hco)
			cl := commontestutils.InitClient([]client.Object{})
			handler := (*genericOperand)(newServiceHandler(cl, commontestutils.GetScheme(), NewKvUIPluginSvc))

			res := handler.ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &v1.Service{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())
			Expect(foundResource.Name).To(Equal(expectedResource.Name))
			Expect(foundResource.Labels).Should(HaveKeyWithValue(hcoutil.AppLabel, commontestutils.Name))
			Expect(foundResource.Namespace).To(Equal(expectedResource.Namespace))
		})

		It("should find plugin service if present", func() {
			expectedResource := NewKvUIPluginSvc(hco)

			cl := commontestutils.InitClient([]client.Object{hco, expectedResource})
			handler := (*genericOperand)(newServiceHandler(cl, commontestutils.GetScheme(), NewKvUIPluginSvc))

			res := handler.ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &v1.Service{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())

			// Check HCO's status
			Expect(hco.Status.RelatedObjects).To(Not(BeNil()))
			objectRef, err := reference.GetReference(commontestutils.GetScheme(), foundResource)
			Expect(err).ToNot(HaveOccurred())
			// ObjectReference should have been added
			Expect(hco.Status.RelatedObjects).To(ContainElement(*objectRef))
		})

		It("should reconcile service to default if changed", func() {
			expectedResource := NewKvUIPluginSvc(hco)
			outdatedResource := NewKvUIPluginSvc(hco)

			outdatedResource.ObjectMeta.Labels[hcoutil.AppLabel] = "wrong label"
			outdatedResource.Spec.Ports[0].Port = 6666

			cl := commontestutils.InitClient([]client.Object{hco, outdatedResource})
			handler := (*genericOperand)(newServiceHandler(cl, commontestutils.GetScheme(), NewKvUIPluginSvc))
			res := handler.ensure(req)
			Expect(res.UpgradeDone).To(BeFalse())
			Expect(res.Updated).To(BeTrue())
			Expect(res.Err).ToNot(HaveOccurred())

			foundResource := &v1.Service{}
			Expect(
				cl.Get(context.TODO(),
					types.NamespacedName{Name: expectedResource.Name, Namespace: expectedResource.Namespace},
					foundResource),
			).ToNot(HaveOccurred())

			Expect(foundResource.ObjectMeta.Labels).ToNot(Equal(outdatedResource.ObjectMeta.Labels))
			Expect(foundResource.ObjectMeta.Labels).To(Equal(expectedResource.ObjectMeta.Labels))
			Expect(foundResource.Spec.Ports).ToNot(Equal(outdatedResource.Spec.Ports))
			Expect(foundResource.Spec.Ports).To(Equal(expectedResource.Spec.Ports))

			// ObjectReference should have been updated
			Expect(hco.Status.RelatedObjects).To(Not(BeNil()))
			objectRefOutdated, err := reference.GetReference(commontestutils.GetScheme(), outdatedResource)
			Expect(err).ToNot(HaveOccurred())
			objectRefFound, err := reference.GetReference(commontestutils.GetScheme(), foundResource)
			Expect(err).ToNot(HaveOccurred())
			Expect(hco.Status.RelatedObjects).To(Not(ContainElement(*objectRefOutdated)))
			Expect(hco.Status.RelatedObjects).To(ContainElement(*objectRefFound))
		})
	})

})
