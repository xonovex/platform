//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/controller"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

var (
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestMain(m *testing.M) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.Background())

	scheme := testutil.NewScheme()

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "bases"),
		},
		Scheme: scheme,
	}

	cfg, err := testEnv.Start()
	if err != nil {
		panic("failed to start envtest: " + err.Error())
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	if err != nil {
		panic("failed to create manager: " + err.Error())
	}

	if err := (&controller.AgentRunReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		panic("failed to setup AgentRun controller: " + err.Error())
	}

	if err := (&controller.AgentProviderReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		panic("failed to setup AgentProvider controller: " + err.Error())
	}

	if err := (&controller.AgentConfigReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		panic("failed to setup AgentConfig controller: " + err.Error())
	}

	go func() {
		if err := mgr.Start(ctx); err != nil {
			panic("failed to start manager: " + err.Error())
		}
	}()

	// Wait for caches to sync before running tests
	if !mgr.GetCache().WaitForCacheSync(ctx) {
		panic("failed to sync manager caches")
	}

	k8sClient = mgr.GetClient()

	code := m.Run()

	cancel()
	if err := testEnv.Stop(); err != nil {
		panic("failed to stop envtest: " + err.Error())
	}

	os.Exit(code)
}
