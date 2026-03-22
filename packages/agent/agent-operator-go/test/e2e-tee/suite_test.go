//go:build e2e_tee

package e2e_tee

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
	k8sClient   client.Client
	ctx         context.Context
	cancel      context.CancelFunc
	clusterName = "agent-operator-e2e-tee"
	tmpDir      string
)

func TestMain(m *testing.M) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.Background())

	var err error
	tmpDir, err = os.MkdirTemp("", "e2e-tee-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}

	// Create kind cluster
	kindConfig := filepath.Join(tmpDir, "kind-config.yaml")
	if err := os.WriteFile(kindConfig, []byte(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
`), 0644); err != nil {
		os.RemoveAll(tmpDir)
		panic("failed to write kind config: " + err.Error())
	}

	if err := runCmd("kind", "create", "cluster", "--name", clusterName, "--config", kindConfig, "--wait", "120s"); err != nil {
		os.RemoveAll(tmpDir)
		panic("failed to create kind cluster: " + err.Error())
	}

	// Install CRDs
	crdPath := filepath.Join("..", "..", "config", "crd", "bases")
	if err := runCmd("kubectl", "--context", "kind-"+clusterName, "apply", "-f", crdPath); err != nil {
		cleanup()
		panic("failed to install CRDs: " + err.Error())
	}

	// Create RuntimeClass for kata-cc (simulated — no actual TEE in kind)
	runtimeClassYAML := `apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: kata-cc
handler: runc
---
apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: kata-tdx
handler: runc
`
	runtimeClassPath := filepath.Join(tmpDir, "runtimeclass-tee.yaml")
	if err := os.WriteFile(runtimeClassPath, []byte(runtimeClassYAML), 0644); err != nil {
		cleanup()
		panic("failed to write RuntimeClass YAML: " + err.Error())
	}
	if err := runCmd("kubectl", "--context", "kind-"+clusterName, "apply", "-f", runtimeClassPath); err != nil {
		cleanup()
		panic("failed to create RuntimeClasses: " + err.Error())
	}

	// Label the kind node with the AKS confidential computing label
	nodeName := clusterName + "-control-plane"
	if err := runCmd("kubectl", "--context", "kind-"+clusterName,
		"label", "node", nodeName,
		"kubernetes.azure.com/confidential-computing=true", "--overwrite"); err != nil {
		cleanup()
		panic("failed to label node: " + err.Error())
	}

	// Build kubeconfig
	var buf bytes.Buffer
	cmd := exec.Command("kind", "get", "kubeconfig", "--name", clusterName)
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		cleanup()
		panic("failed to get kubeconfig: " + err.Error())
	}

	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig.yaml")
	if err := os.WriteFile(kubeconfigPath, buf.Bytes(), 0600); err != nil {
		cleanup()
		panic("failed to write kubeconfig: " + err.Error())
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		cleanup()
		panic("failed to build rest config: " + err.Error())
	}

	scheme := testutil.NewScheme()

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	if err != nil {
		cleanup()
		panic("failed to create manager: " + err.Error())
	}

	if err := (&controller.AgentRunReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Recorder: mgr.GetEventRecorderFor("agentrun-controller")}).SetupWithManager(mgr); err != nil {
		cleanup()
		panic("failed to setup AgentRun controller: " + err.Error())
	}
	if err := (&controller.AgentProviderReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Recorder: mgr.GetEventRecorderFor("agentprovider-controller")}).SetupWithManager(mgr); err != nil {
		cleanup()
		panic("failed to setup AgentProvider controller: " + err.Error())
	}
	if err := (&controller.AgentWorkspaceReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Recorder: mgr.GetEventRecorderFor("agentworkspace-controller")}).SetupWithManager(mgr); err != nil {
		cleanup()
		panic("failed to setup AgentWorkspace controller: " + err.Error())
	}

	go func() {
		if err := mgr.Start(ctx); err != nil {
			panic("failed to start manager: " + err.Error())
		}
	}()

	if !mgr.GetCache().WaitForCacheSync(ctx) {
		cleanup()
		panic("failed to sync manager caches")
	}

	k8sClient = mgr.GetClient()

	code := m.Run()

	cancel()
	cleanup()
	os.Exit(code)
}

func cleanup() {
	_ = runCmd("kind", "delete", "cluster", "--name", clusterName)
	os.RemoveAll(tmpDir)
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
