//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/controller"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	k8sClient          client.Client
	ctx                context.Context
	cancel             context.CancelFunc
	clusterName        = "agent-operator-e2e"
	useExistingCluster bool
)

func TestMain(m *testing.M) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.Background())

	useExistingCluster = os.Getenv("USE_EXISTING_CLUSTER") == "true"

	useExisting := useExistingCluster

	if !useExisting {
		kindConfig := filepath.Join("..", "testdata", "kind-config.yaml")
		if err := runCmd("kind", "create", "cluster", "--name", clusterName, "--config", kindConfig, "--wait", "120s"); err != nil {
			panic("failed to create Kind cluster: " + err.Error())
		}
	}

	// Install CRDs
	crdPath := filepath.Join("..", "..", "config", "crd", "bases")
	if err := runCmd("kubectl", "apply", "-f", crdPath); err != nil {
		if !useExisting {
			_ = runCmd("kind", "delete", "cluster", "--name", clusterName)
		}
		panic("failed to install CRDs: " + err.Error())
	}

	// Build kubeconfig
	var kubeconfigPath string
	if useExisting {
		kubeconfigPath = os.Getenv("KUBECONFIG")
		if kubeconfigPath == "" {
			home, _ := os.UserHomeDir()
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	} else {
		// Get Kind kubeconfig
		var buf bytes.Buffer
		cmd := exec.Command("kind", "get", "kubeconfig", "--name", clusterName)
		cmd.Stdout = &buf
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			panic("failed to get Kind kubeconfig: " + err.Error())
		}

		tmpFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
		if err != nil {
			panic("failed to create temp kubeconfig: " + err.Error())
		}
		kubeconfigPath = tmpFile.Name()
		if _, err := tmpFile.Write(buf.Bytes()); err != nil {
			panic("failed to write kubeconfig: " + err.Error())
		}
		tmpFile.Close()
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic("failed to build rest config: " + err.Error())
	}

	scheme := testutil.NewScheme()

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

	if !mgr.GetCache().WaitForCacheSync(ctx) {
		panic("failed to sync manager caches")
	}

	k8sClient = mgr.GetClient()

	code := m.Run()

	cancel()

	if !useExisting {
		_ = runCmd("kind", "delete", "cluster", "--name", clusterName)
		_ = os.Remove(kubeconfigPath)
	}

	os.Exit(code)
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
