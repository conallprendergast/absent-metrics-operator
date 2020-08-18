// Copyright 2020 SAP SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// AbsentPrometheusRuleName returns the name of an AbsentPrometheusRule.
func AbsentPrometheusRuleName(prometheusServer string) string {
	return fmt.Sprintf("%s-absent-metric-alert-rules", prometheusServer)
}

// absentPrometheusRule is a wrapper around *monitoringv1.PrometheusRule with
// some additional info that we use for working with AbsentPrometheusRules.
//
// An absentPrometheusRule is the corresponding resource that is generated for
// a PrometheusRule resource for defining the absent metric alerts.
type absentPrometheusRule struct {
	*monitoringv1.PrometheusRule

	// Default values to use for absent metric alerts.
	// See parseRuleGroups() on why we need this.
	Tier    string
	Service string
}

// TODO: add tier and service as labels.
func (c *Controller) getAbsentPrometheusRule(namespace, prometheusServer string) (*absentPrometheusRule, error) {
	n := AbsentPrometheusRuleName(prometheusServer)
	pr, err := c.promClientset.MonitoringV1().PrometheusRules(namespace).Get(context.Background(), n, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	aPR := absentPrometheusRule{
		PrometheusRule: pr,
	}

	// Find default tier and service values for this Prometheus server in this
	// namespace.
	if c.keepLabel[labelTier] || c.keepLabel[labelService] {
		// Fast path: get values from resource labels
		t, s := aPR.Labels[labelTier], aPR.Labels[labelService]
		if t == "" || s == "" {
			// If we can't get the values from resource then we fall back to
			// the slower method of getting them by checking alert rules.
			t, s = getTierAndService(aPR.Spec.Groups)
		}
		if t == "" || s == "" {
			c.logger.Info("msg", fmt.Sprintf("could not find default tier and service for Prometheus server '%s' in namespace '%s'",
				prometheusServer, namespace))
		}
		if c.keepLabel[labelTier] {
			aPR.Tier = t
			aPR.Labels[labelTier] = t
		}
		if c.keepLabel[labelService] {
			aPR.Service = s
			aPR.Labels[labelService] = s
		}
	}

	return &aPR, nil
}

func (c *Controller) newAbsentPrometheusRule(namespace, prometheusServer string) (*absentPrometheusRule, error) {
	n := AbsentPrometheusRuleName(prometheusServer)
	aPR := absentPrometheusRule{
		PrometheusRule: &monitoringv1.PrometheusRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      n,
				Namespace: namespace,
				Labels: map[string]string{
					// Add a label that identifies that this PrometheusRule is
					// created and managed by this operator.
					labelOperatorManagedBy: "true",
					"prometheus":           prometheusServer,
					"type":                 "alerting-rules",
				},
			},
		},
	}

	// Find default tier and service values for this Prometheus server in this
	// namespace.
	if c.keepLabel[labelTier] || c.keepLabel[labelService] {
		prList, err := c.promRuleLister.List(labels.Everything())
		if err != nil {
			return nil, errors.Wrap(err, "could not list PrometheusRules")
		}
		var rg []monitoringv1.RuleGroup
		for _, pr := range prList {
			s := pr.Labels["prometheus"]
			if pr.Namespace == namespace && s == prometheusServer {
				rg = append(rg, pr.Spec.Groups...)
			}
		}
		t, s := getTierAndService(rg)
		if t == "" || s == "" {
			// Ideally, we shouldn't arrive at this point because this would
			// mean that there was not a single alert rule for the prometheus
			// server in this namespace that did not use templating for its
			// tier and service labels.
			c.logger.Info("msg", fmt.Sprintf("could not find default tier and service for Prometheus server '%s' in namespace '%s'",
				prometheusServer, namespace))
		}
		if c.keepLabel[labelTier] {
			aPR.Tier = t
			aPR.Labels[labelTier] = t
		}
		if c.keepLabel[labelService] {
			aPR.Service = s
			aPR.Labels[labelService] = s
		}
	}

	return &aPR, nil
}

// updateAbsentPrometheusRule updates an AbsentPrometheusRule with the provided
// slice of RuleGroup.
func (c *Controller) updateAbsentPrometheusRule(
	absentPromRule *absentPrometheusRule,
	absentAlertRuleGroups []monitoringv1.RuleGroup) error {

	// Check if the AbsentPrometheusRule already has these rule groups.
	// Update if it does, otherwise append.
	old := absentPromRule.Spec.Groups
	var new []monitoringv1.RuleGroup
	updated := make(map[string]bool)
OuterLoop:
	for _, oldG := range old {
		for _, g := range absentAlertRuleGroups {
			if oldG.Name == g.Name {
				// Add the new updated RuleGroup.
				new = append(new, g)
				updated[g.Name] = true
				continue OuterLoop
			}
		}
		// This RuleGroup should be carried over as is.
		new = append(new, oldG)
	}
	// Add the pending RuleGroups.
	for _, g := range absentAlertRuleGroups {
		if !updated[g.Name] {
			new = append(new, g)
		}
	}

	// No need to update if old and new rule groups are exactly the same.
	if reflect.DeepEqual(old, new) {
		return nil
	}

	absentPromRule.Spec.Groups = new
	_, err := c.promClientset.MonitoringV1().PrometheusRules(absentPromRule.Namespace).
		Update(context.Background(), absentPromRule.PrometheusRule, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "could not update AbsentPrometheusRule")
	}

	c.logger.Info("msg", "successfully updated absent metric alert rules",
		"key", fmt.Sprintf("%s/%s", absentPromRule.Namespace, absentPromRule.Name))
	return nil
}

// cleanUpOrphanedAbsentAlertsNamespace deletes orphaned absent alert rules
// concerning a specific PrometheusRule from a namespace.
func (c *Controller) cleanUpOrphanedAbsentAlertsNamespace(namespace, promRuleName string) error {
	prList, err := c.promClientset.MonitoringV1().PrometheusRules(namespace).
		List(context.Background(), metav1.ListOptions{LabelSelector: labelOperatorManagedBy})
	if err != nil {
		return errors.Wrap(err, "could not list AbsentPrometheusRules")
	}

	for _, pr := range prList.Items {
		aPR := &absentPrometheusRule{PrometheusRule: pr}
		err := c.cleanUpOrphanedAbsentAlerts(promRuleName, aPR)
		if err != nil {
			return err
		}
	}

	return nil
}

// cleanUpOrphanedAbsentAlerts deletes orphaned absent alert rules concerning a
// specific PrometheusRule from a specific AbsentPrometheusRule.
func (c *Controller) cleanUpOrphanedAbsentAlerts(promRuleName string, absentPromRule *absentPrometheusRule) error {
	old := absentPromRule.Spec.Groups
	new := make([]monitoringv1.RuleGroup, 0, len(old))
	for _, g := range old {
		// The rule group names for AbsentPrometheusRule have the format:
		// originPromRuleName/ruleGroupName.
		sL := strings.Split(g.Name, "/")
		if len(sL) > 0 && sL[0] == promRuleName {
			continue
		}
		new = append(new, g)
	}
	if reflect.DeepEqual(old, new) {
		return nil
	}

	var err error
	absentPromRule.Spec.Groups = new
	if len(absentPromRule.Spec.Groups) == 0 {
		err = c.promClientset.MonitoringV1().PrometheusRules(absentPromRule.Namespace).
			Delete(context.Background(), absentPromRule.Name, metav1.DeleteOptions{})
		if err == nil {
			c.logger.Info("msg", "successfully deleted orphaned AbsentPrometheusRule",
				"key", fmt.Sprintf("%s/%s", absentPromRule.Namespace, absentPromRule.Name))
		}
	} else {
		_, err = c.promClientset.MonitoringV1().PrometheusRules(absentPromRule.Namespace).
			Update(context.Background(), absentPromRule.PrometheusRule, metav1.UpdateOptions{})
		if err == nil {
			c.logger.Info("msg", "successfully cleaned up orphaned absent metric alert rules",
				"key", fmt.Sprintf("%s/%s", absentPromRule.Namespace, absentPromRule.Name))
		}
	}
	if err != nil {
		return errors.Wrap(err, "could not clean up orphaned absent metric alert rules")
	}

	return nil
}
