package test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/sapcc/absent-metrics-operator/internal/controller"
)

var (
	cfg       *rest.Config
	testEnv   *envtest.Environment
	k8sClient client.Client

	wg     *errgroup.Group
	cancel context.CancelFunc
)

//nolint:unused
func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	// Use "bin" subdirectory for control binaries. By default, envtest looks
	// for these binaries in "/usr/local/kubebuilder/bin".
	p, err := filepath.Abs("bin")
	Expect(err).ToNot(HaveOccurred())
	err = os.Setenv("KUBEBUILDER_ASSETS", p)
	Expect(err).ToNot(HaveOccurred())

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("fixtures", "crd")},
	}
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())

	err = monitoringv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())

	// NOTE: We start the controller before adding objects since the items are
	// queued by the controller sequentially and we depend on this behavior in
	// our mock assertion.
	By("starting controller")
	l := log.NewLogfmtLogger(log.NewSyncWriter(GinkgoWriter))
	l = level.NewFilter(l, level.AllowAll())
	l = log.With(l,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
	c, err := controller.New(cfg, 1*time.Second, l)
	Expect(err).ToNot(HaveOccurred())

	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	wg, ctx = errgroup.WithContext(ctx)
	wg.Go(func() error { return c.Run(ctx.Done()) })

	By("adding mock PrometheusRule resources")
	mockDir := filepath.Join("fixtures", "prometheusrules")
	files, err := ioutil.ReadDir(mockDir)
	Expect(err).ToNot(HaveOccurred())
	for _, file := range files {
		b, err := ioutil.ReadFile(filepath.Join(mockDir, file.Name()))
		Expect(err).ToNot(HaveOccurred())

		var pr monitoringv1.PrometheusRule
		err = json.Unmarshal(b, &pr)
		Expect(err).ToNot(HaveOccurred())

		// Create namespace if it doesn't exist already.
		ns := corev1.Namespace{}
		err = k8sClient.Get(ctx, client.ObjectKey{Name: pr.Namespace}, &ns)
		if err != nil && apierrors.IsNotFound(err) {
			ns.Name = pr.Namespace
			err = k8sClient.Create(ctx, &ns)
		}
		Expect(err).ToNot(HaveOccurred())

		// Create PrometheusRule resource
		err = k8sClient.Create(ctx, &pr)
		Expect(err).ToNot(HaveOccurred())
	}
})

var _ = AfterSuite(func() {
	By("stopping controller")
	cancel()
	Expect(wg.Wait()).To(Succeed())

	By("tearing down the test environment")
	Expect(testEnv.Stop()).To(Succeed())
})
