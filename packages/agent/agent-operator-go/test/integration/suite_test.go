//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	crwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/controller"
	agentwebhook "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/webhook"
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
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{
				filepath.Join("..", "..", "config", "webhook", "manifests.yaml"),
			},
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
		WebhookServer: crwebhook.NewServer(crwebhook.Options{
			Host:    testEnv.WebhookInstallOptions.LocalServingHost,
			Port:    testEnv.WebhookInstallOptions.LocalServingPort,
			CertDir: testEnv.WebhookInstallOptions.LocalServingCertDir,
		}),
	})
	if err != nil {
		panic("failed to create manager: " + err.Error())
	}

	if err := (&agentwebhook.AgentRunWebhook{Client: mgr.GetClient()}).SetupWebhookWithManager(mgr); err != nil {
		panic("failed to setup AgentRun webhook: " + err.Error())
	}
	if err := (&agentwebhook.AgentHarnessWebhook{}).SetupWebhookWithManager(mgr); err != nil {
		panic("failed to setup AgentHarness webhook: " + err.Error())
	}
	if err := (&agentwebhook.AgentProviderWebhook{}).SetupWebhookWithManager(mgr); err != nil {
		panic("failed to setup AgentProvider webhook: " + err.Error())
	}
	if err := (&agentwebhook.AgentToolchainWebhook{}).SetupWebhookWithManager(mgr); err != nil {
		panic("failed to setup AgentToolchain webhook: " + err.Error())
	}
	if err := (&agentwebhook.AgentWorkspaceWebhook{}).SetupWebhookWithManager(mgr); err != nil {
		panic("failed to setup AgentWorkspace webhook: " + err.Error())
	}

	if err := (&controller.AgentRunReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorder("agentrun-controller"),
	}).SetupWithManager(mgr); err != nil {
		panic("failed to setup AgentRun controller: " + err.Error())
	}

	if err := (&controller.AgentProviderReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorder("agentprovider-controller"),
	}).SetupWithManager(mgr); err != nil {
		panic("failed to setup AgentProvider controller: " + err.Error())
	}

	if err := (&controller.AgentWorkspaceReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorder("agentworkspace-controller"),
	}).SetupWithManager(mgr); err != nil {
		panic("failed to setup AgentWorkspace controller: " + err.Error())
	}

	managerStopped := make(chan error, 1)
	go func() {
		managerStopped <- mgr.Start(ctx)
	}()

	webhookDeadline := time.Now().Add(10 * time.Second)
	for mgr.GetWebhookServer().StartedChecker()(nil) != nil {
		if time.Now().After(webhookDeadline) {
			panic("webhook server did not start")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Wait for caches to sync before running tests
	if !mgr.GetCache().WaitForCacheSync(ctx) {
		panic("failed to sync manager caches")
	}

	k8sClient = mgr.GetClient()

	code := m.Run()

	cancel()
	if err := <-managerStopped; err != nil {
		panic("manager stopped with an error: " + err.Error())
	}
	if err := testEnv.Stop(); err != nil {
		panic("failed to stop envtest: " + err.Error())
	}

	os.Exit(code)
}
