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

import "github.com/prometheus/client_golang/prometheus"

// Metrics represents the metrics associated with the controller.
type Metrics struct {
	SuccessfulPrometheusRuleReconcileTime *prometheus.GaugeVec
}

// NewMetrics returns a new Metrics and registers them with the provided
// Registerer.
func NewMetrics(r prometheus.Registerer) *Metrics {
	m := Metrics{
		SuccessfulPrometheusRuleReconcileTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "absent_metrics_operator_successful_reconcile_time",
				Help: "The time at which a specific PrometheusRule was successfully reconciled by the operator.",
			},
			[]string{"namespace", "name"},
		),
	}
	r.MustRegister(m.SuccessfulPrometheusRuleReconcileTime)
	return &m
}
