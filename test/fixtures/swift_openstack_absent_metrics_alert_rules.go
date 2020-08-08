package fixtures

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var swiftLab = map[string]string{
	"tier":     "os",
	"service":  "swift",
	"severity": "info",
}

// SwiftOSAbsentPromRule represents the PrometheusRule that should be generated
// for the "openstack" Prometheus server in the "swift" namespace.
var SwiftOSAbsentPromRule = monitoringv1.PrometheusRule{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: "swift",
		Name:      OSAbsentPromRuleName,
		Labels: map[string]string{
			"prometheus":                         "openstack",
			"type":                               "alerting-rules",
			"absent-metrics-operator/managed-by": "true",
		},
	},
	Spec: monitoringv1.PrometheusRuleSpec{
		Groups: []monitoringv1.RuleGroup{
			{
				Name: "openstack-swift.alerts/swift.alerts",
				Rules: []monitoringv1.Rule{
					{
						Alert:  "AbsentOsSwiftGlobalSwiftClusterStorageUsedPercentAverage",
						Expr:   intstr.FromString("absent(global:swift_cluster_storage_used_percent_average)"),
						For:    "10m",
						Labels: swiftLab,
						Annotations: map[string]string{
							"summary":     "missing global:swift_cluster_storage_used_percent_average",
							"description": "The metric 'global:swift_cluster_storage_used_percent_average' is missing",
						},
					},
					{
						Alert:  "AbsentOsSwiftDispersionTaskExitCode",
						Expr:   intstr.FromString("absent(swift_dispersion_task_exit_code)"),
						For:    "10m",
						Labels: swiftLab,
						Annotations: map[string]string{
							"summary":     "missing swift_dispersion_task_exit_code",
							"description": "The metric 'swift_dispersion_task_exit_code' is missing",
						},
					},
					{
						Alert:  "AbsentOsSwiftReconTaskExitCode",
						Expr:   intstr.FromString("absent(swift_recon_task_exit_code)"),
						For:    "10m",
						Labels: swiftLab,
						Annotations: map[string]string{
							"summary":     "missing swift_recon_task_exit_code",
							"description": "The metric 'swift_recon_task_exit_code' is missing",
						},
					},
				},
			},
		},
	},
}
