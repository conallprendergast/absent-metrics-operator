module github.com/sapcc/absent-metrics-operator

go 1.17

// Note: the "k8s.io/*" and "github.com/prometheus/*" dependencies should have the same
// versions as used by github.com/prometheus-operator/prometheus-operator.
//
// Instructions:
// 1. Delete the `require` and `replace` directives in this file.
// 2. Pick a Git tag on https://github.com/prometheus-operator/prometheus-operator/tags
// 3. Copy the `require` and `replace` directives from that tag's go.mod file. If there
//    are any local references in the `replace` directive, remove those.
// 4. Run `make tidy-deps`.
// 5. Open an issue, if the above doesn't work.

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-kit/log v0.2.0
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.55.1
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.55.1
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/prometheus v1.8.2-0.20220308163432-03831554a519
	github.com/sapcc/go-bits v0.0.0-20220327202728-0bec9049b55c
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/text v0.3.7
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.11.2
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/go-logr/zapr v1.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/goleak v1.1.12 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/net v0.0.0-20220114011407-0dd24b26b47d // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/apiextensions-apiserver v0.23.5 // indirect
	k8s.io/klog/v2 v2.40.1 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	// A replace directive is needed for github.com/prometheus/alertmanager
	// because Cortex (which is an indirect dependency through Thanos) has a
	// requirement on a more recent commit version but untagged.
	github.com/prometheus/alertmanager => github.com/prometheus/alertmanager v0.23.0
	// A replace directive is needed for k8s.io/client-go because Cortex (which
	// is an indirect dependency through Thanos) has a requirement on v12.0.0.
	k8s.io/client-go => k8s.io/client-go v0.23.0
	k8s.io/klog/v2 => github.com/simonpasquier/klog-gokit/v3 v3.1.0
)
