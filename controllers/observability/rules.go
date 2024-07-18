package observability

import (
	"context"
	"fmt"
	"github.com/kubevirt/hyperconverged-cluster-operator/pkg/monitoring/rules"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ruleName                  = "cnv-rule"
	defaultRunbookURLTemplate = "https://kubevirt.io/monitoring/runbooks/%s"
	runbookURLTemplateEnv     = "RUNBOOK_URL_TEMPLATE"
)

func (r *Reconciler) reconcileRules(ctx context.Context) error {
	log.Info(fmt.Sprintf("reconciling %s prometheusrule", ruleName))

	if err := rules.SetupRules(); err != nil {
		log.Error(err, "failed to setup rules")
		return err
	}

	requiredRule, err := rules.BuildPrometheusRule2(r.namespace, r.owner)
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to build %s prometheusrule", ruleName))
		return err
	}

	existingRule := &monitoringv1.PrometheusRule{}
	err = r.client.Get(ctx, client.ObjectKey{Namespace: r.namespace, Name: ruleName}, existingRule)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("can't find %s prometheusrule; creating a new one", ruleName))
			if err = r.client.Create(ctx, requiredRule); err != nil {
				log.Error(err, fmt.Sprintf("failed to create %s prometheusrule", ruleName))
				return err
			}
			log.Info(fmt.Sprintf("successfully created %s prometheusrule", ruleName))
			return nil
		}

		log.Error(err, fmt.Sprintf("unexpected error while reading %s prometheusrule", ruleName))
		return err
	}

	updated := false
	if !reflect.DeepEqual(existingRule.Spec, requiredRule.Spec) {
		requiredRule.Spec.DeepCopyInto(&existingRule.Spec)
		updated = true
	}

	if !reflect.DeepEqual(existingRule.Labels, requiredRule.Labels) {
		existingRule.Labels = requiredRule.Labels
		updated = true
	}

	if !reflect.DeepEqual(existingRule.OwnerReferences, requiredRule.OwnerReferences) {
		existingRule.OwnerReferences = requiredRule.OwnerReferences
		updated = true
	}

	if updated {
		if err = r.client.Update(ctx, existingRule); err != nil {
			log.Error(err, fmt.Sprintf("failed to update %s prometheusrule", ruleName))
			return err
		}
		log.Info(fmt.Sprintf("successfully updated %s prometheusrule", ruleName))
	}

	return nil
}
