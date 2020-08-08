package fixtures

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var kepLab = map[string]string{
	"tier":     "os",
	"service":  "keppel",
	"severity": "info",
}

// ResMgmtK8sAbsentPromRule represents the PrometheusRule that should be
// generated for the "kubernetes" Prometheus server in the "resmgmt" namespace.
var ResMgmtK8sAbsentPromRule = monitoringv1.PrometheusRule{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: "resmgmt",
		Name:      K8sAbsentPromRuleName,
		Labels: map[string]string{
			"prometheus":                         "kubernetes",
			"type":                               "alerting-rules",
			"absent-metrics-operator/managed-by": "true",
		},
	},
	Spec: monitoringv1.PrometheusRuleSpec{
		Groups: []monitoringv1.RuleGroup{
			{
				Name: "kubernetes-keppel.alerts/keppel.alerts",
				Rules: []monitoringv1.Rule{
					{
						Alert:  "AbsentOsKeppelKubePodFailedSchedulingMemoryTotal",
						Expr:   intstr.FromString("absent(kube_pod_failed_scheduling_memory_total)"),
						For:    "10m",
						Labels: kepLab,
						Annotations: map[string]string{
							"summary":     "missing kube_pod_failed_scheduling_memory_total",
							"description": "The metric 'kube_pod_failed_scheduling_memory_total' is missing",
						},
					},
					{
						Alert:  "AbsentOsKeppelContainerMemoryUsagePercent",
						Expr:   intstr.FromString("absent(keppel_container_memory_usage_percent)"),
						For:    "10m",
						Labels: kepLab,
						Annotations: map[string]string{
							"summary":     "missing keppel_container_memory_usage_percent",
							"description": "The metric 'keppel_container_memory_usage_percent' is missing",
						},
					},
				},
			},
		},
	},
}
