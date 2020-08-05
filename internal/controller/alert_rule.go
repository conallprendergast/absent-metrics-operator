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
	"fmt"
	"strings"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/promql"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

// metricNameExtractor is used to walk through a promql expression and extract
// time series names.
type metricNameExtractor struct {
	// expr is the promql expression that the metricNameExtractor is working on.
	expr string
	// This map contains metric names that were extracted from a promql.Node.
	// We only use the keys of the map and never depend on the presence of an
	// element in the map nor its value therefore an empty struct is better
	// than bool.
	found map[string]struct{}
}

// Visit implements the promql.Visitor interface.
func (mex *metricNameExtractor) Visit(node promql.Node) promql.Visitor {
	var name string
	switch n := node.(type) {
	case *promql.MatrixSelector:
		name = n.Name
	case *promql.VectorSelector:
		name = n.Name
	default:
		return mex
	}

	switch {
	case strings.Contains(mex.expr, fmt.Sprintf("absent(%s", name)):
		// Skip this metric if the there is already an
		// absent function for it in the original expression.
	case name == "up":
		// Skip "up" metric, it is automatically injected by
		// Prometheus to describe Prometheus scraping jobs.
	default:
		mex.found[name] = struct{}{}
	}
	return mex
}

// parseRuleGroups takes a slice of RuleGroup that has alert rules and returns
// a new slice of RuleGroup that has the corresponding absent metric alert rules.
//
// The original tier and service labels of alert rules will be carried over to
// the corresponding absent alert rule unless templating was used (i.e. $labels)
// for these labels in which case the provided default tier and service will be used.
//
// The rule group names for the absent metric alerts have the format:
//   promRuleName/originalGroupName.
func parseRuleGroups(promRuleName, defaultTier, defaultService string, in []monitoringv1.RuleGroup) ([]monitoringv1.RuleGroup, error) {
	out := make([]monitoringv1.RuleGroup, 0, len(in))
	for _, g := range in {
		var absentRules []monitoringv1.Rule
		for _, r := range g.Rules {
			rules, err := parseAlertRule(defaultTier, defaultService, r)
			if err != nil {
				return nil, err
			}

			if len(rules) > 0 {
				absentRules = append(absentRules, rules...)
			}
		}

		if len(absentRules) > 0 {
			out = append(out, monitoringv1.RuleGroup{
				Name:  fmt.Sprintf("%s/%s", promRuleName, g.Name),
				Rules: absentRules,
			})
		}
	}
	return out, nil
}

// parseAlertRule converts an alert rule to absent metric alert rules.
// Since an original alert expression can reference multiple time series therefore
// a slice of []monitoringv1.Rule is returned as the result would be multiple
// absent metric alert rules (one for each time series).
func parseAlertRule(tier, service string, in monitoringv1.Rule) ([]monitoringv1.Rule, error) {
	exprStr := in.Expr.String()
	exprNode, err := promql.ParseExpr(exprStr)
	if err != nil {
		// The returned error has the expression in last because
		// it could contain newline chracters.
		return nil, fmt.Errorf("could not parse rule expression: %s: %s",
			err.Error(), in.Expr.String())
	}

	mex := &metricNameExtractor{expr: exprStr, found: map[string]struct{}{}}
	promql.Walk(mex, exprNode)
	if len(mex.found) == 0 {
		return nil, nil
	}

	// Carry over labels from the original alert
	if origLab := in.Labels; origLab != nil {
		if v, ok := origLab["tier"]; ok && !strings.Contains(v, "$labels") {
			tier = v
		}
		if v, ok := origLab["service"]; ok && !strings.Contains(v, "$labels") {
			service = v
		}
	}
	lab := map[string]string{
		"tier":     tier,
		"service":  service,
		"severity": "info",
	}

	out := make([]monitoringv1.Rule, 0, len(mex.found))
	for metric := range mex.found {
		// Generate an alert name from metric name:
		//   network:tis_a_metric:rate5m -> AbsentTierServiceNetworkTisAMetricRate5m
		words := []string{"absent", tier, service}
		sL1 := strings.Split(metric, "_")
		for _, v := range sL1 {
			sL2 := strings.Split(v, ":")
			words = append(words, sL2...)
		}
		// Avoid name stuttering
		var alertName string
		var prev string
		for _, v := range words {
			l := strings.ToLower(v)
			if prev != l {
				prev = l
				alertName += strings.Title(l)
			}
		}

		ann := map[string]string{
			"summary":     fmt.Sprintf("missing %s", metric),
			"description": fmt.Sprintf("The metric '%s' is missing", metric),
		}

		out = append(out, monitoringv1.Rule{
			Alert:       alertName,
			Expr:        intstr.FromString(fmt.Sprintf("absent(%s)", metric)),
			For:         "10m",
			Labels:      lab,
			Annotations: ann,
		})
	}

	return out, nil
}
