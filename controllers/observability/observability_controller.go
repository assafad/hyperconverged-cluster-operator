package observability

import (
	"context"
	hcoutil "github.com/kubevirt/hyperconverged-cluster-operator/pkg/util"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kubevirt/hyperconverged-cluster-operator/pkg/alertmanager"
)

var (
	log         = logf.Log.WithName("controller_observability")
	periodicity = 1 * time.Hour
)

type Reconciler struct {
	config    *rest.Config
	client    client.Client
	namespace string
	owner     metav1.OwnerReference
	events    chan event.GenericEvent

	amApi *alertmanager.Api
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Info("Reconciling Observability")

	//if err := r.ensurePodDisruptionBudgetAtLimitIsSilenced(); err != nil {
	//	return ctrl.Result{}, err
	//}

	if err := r.reconcileRules(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func NewReconciler(config *rest.Config, client client.Client, ci hcoutil.ClusterInfo) *Reconciler {
	deployment := ci.GetDeployment()
	namespace := deployment.Namespace
	owner := getDeploymentReference(deployment)

	return &Reconciler{
		config:    config,
		client:    client,
		namespace: namespace,
		owner:     owner,
		events:    make(chan event.GenericEvent, 1),
	}
}

func SetupWithManager(mgr ctrl.Manager, ci hcoutil.ClusterInfo) error {
	log.Info("Setting up controller")

	r := NewReconciler(mgr.GetConfig(), mgr.GetClient(), ci)
	r.startEventLoop()

	return ctrl.NewControllerManagedBy(mgr).
		Named("observability").
		WatchesRawSource(source.Channel(
			r.events,
			&handler.EnqueueRequestForObject{},
		)).
		Watches(&monitoringv1.PrometheusRule{}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

func (r *Reconciler) startEventLoop() {
	ticker := time.NewTicker(periodicity)

	go func() {
		r.events <- event.GenericEvent{
			Object: &metav1.PartialObjectMetadata{},
		}

		for range ticker.C {
			r.events <- event.GenericEvent{
				Object: &metav1.PartialObjectMetadata{},
			}
		}
	}()
}

func getDeploymentReference(deployment *appsv1.Deployment) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         appsv1.SchemeGroupVersion.String(),
		Kind:               "Deployment",
		Name:               deployment.GetName(),
		UID:                deployment.GetUID(),
		BlockOwnerDeletion: ptr.To(false),
		Controller:         ptr.To(false),
	}
}
