//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	crwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
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

func cleanupIntegrationNamespace(t *testing.T, namespace string) {
	t.Helper()

	cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cleanupCancel()

	resources := []client.Object{
		&agentv1alpha1.AgentRun{},
		&agentv1alpha1.AgentWorkspace{},
		&agentv1alpha1.AgentProvider{},
		&agentv1alpha1.AgentHarness{},
		&agentv1alpha1.AgentToolchain{},
		&agentv1alpha1.AgentPolicy{},
	}
	for _, resource := range resources {
		if err := k8sClient.DeleteAllOf(cleanupContext, resource, client.InNamespace(namespace)); client.IgnoreNotFound(err) != nil {
			t.Errorf("failed to delete %T resources from namespace %s: %v", resource, namespace, err)
			return
		}
	}

	if err := wait.PollUntilContextTimeout(cleanupContext, 20*time.Millisecond, 10*time.Second, true, func(pollContext context.Context) (bool, error) {
		var runs agentv1alpha1.AgentRunList
		if err := k8sClient.List(pollContext, &runs, client.InNamespace(namespace)); err != nil {
			return false, err
		}
		var workspaces agentv1alpha1.AgentWorkspaceList
		if err := k8sClient.List(pollContext, &workspaces, client.InNamespace(namespace)); err != nil {
			return false, err
		}
		var providers agentv1alpha1.AgentProviderList
		if err := k8sClient.List(pollContext, &providers, client.InNamespace(namespace)); err != nil {
			return false, err
		}
		return len(runs.Items) == 0 && len(workspaces.Items) == 0 && len(providers.Items) == 0, nil
	}); err != nil {
		t.Errorf("controllers did not observe resource cleanup in namespace %s: %v", namespace, err)
		return
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
	if err := k8sClient.Delete(cleanupContext, ns); client.IgnoreNotFound(err) != nil {
		t.Errorf("failed to delete namespace %s: %v", namespace, err)
	}
}

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
