//go:build e2e_gvisor

package e2e_gvisor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	clusterName = "agent-operator-e2e-gvisor"
	tmpDir      string
)

func TestMain(m *testing.M) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.Background())

	var err error
	tmpDir, err = os.MkdirTemp("", "e2e-gvisor-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}

	// Download runsc and containerd-shim-runsc-v1 binaries
	runscPath := filepath.Join(tmpDir, "runsc")
	shimPath := filepath.Join(tmpDir, "containerd-shim-runsc-v1")
	if err := downloadGVisorBinary("runsc", runscPath); err != nil {
		os.RemoveAll(tmpDir)
		panic("failed to download runsc: " + err.Error())
	}
	if err := downloadGVisorBinary("containerd-shim-runsc-v1", shimPath); err != nil {
		os.RemoveAll(tmpDir)
		panic("failed to download containerd-shim-runsc-v1: " + err.Error())
	}

	// Create kind cluster (no extraMounts needed — we docker cp the binaries in)
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

	// Copy binaries into the kind node and configure containerd
	nodeName := clusterName + "-control-plane"
	for _, bin := range []struct{ src, dst string }{
		{runscPath, "/usr/local/bin/runsc"},
		{shimPath, "/usr/local/bin/containerd-shim-runsc-v1"},
	} {
		if err := runCmd("docker", "cp", bin.src, nodeName+":"+bin.dst); err != nil {
			cleanup()
			panic("failed to copy " + bin.src + " into kind node: " + err.Error())
		}
	}

	// Configure containerd to use runsc runtime and restart
	setupScript := `cat >> /etc/containerd/config.toml <<'TOML'

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runsc]
  runtime_type = "io.containerd.runsc.v1"
TOML
systemctl restart containerd`
	if err := runCmd("docker", "exec", nodeName, "sh", "-c", setupScript); err != nil {
		cleanup()
		panic("failed to configure containerd for runsc: " + err.Error())
	}

	// Wait for node to be ready again after containerd restart
	if err := runCmd("kubectl", "--context", "kind-"+clusterName, "wait", "--for=condition=Ready", "node/"+nodeName, "--timeout=60s"); err != nil {
		cleanup()
		panic("failed waiting for node ready: " + err.Error())
	}

	// Install CRDs
	crdPath := filepath.Join("..", "..", "config", "crd", "bases")
	if err := runCmd("kubectl", "--context", "kind-"+clusterName, "apply", "-f", crdPath); err != nil {
		cleanup()
		panic("failed to install CRDs: " + err.Error())
	}

	// Create RuntimeClass
	runtimeClassYAML := `apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: gvisor
handler: runsc
`
	runtimeClassPath := filepath.Join(tmpDir, "runtimeclass-gvisor.yaml")
	if err := os.WriteFile(runtimeClassPath, []byte(runtimeClassYAML), 0644); err != nil {
		cleanup()
		panic("failed to write RuntimeClass YAML: " + err.Error())
	}
	if err := runCmd("kubectl", "--context", "kind-"+clusterName, "apply", "-f", runtimeClassPath); err != nil {
		cleanup()
		panic("failed to create RuntimeClass: " + err.Error())
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

	if err := (&controller.AgentRunReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}).SetupWithManager(mgr); err != nil {
		cleanup()
		panic("failed to setup AgentRun controller: " + err.Error())
	}
	if err := (&controller.AgentProviderReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}).SetupWithManager(mgr); err != nil {
		cleanup()
		panic("failed to setup AgentProvider controller: " + err.Error())
	}
	if err := (&controller.AgentWorkspaceReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}).SetupWithManager(mgr); err != nil {
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

func downloadGVisorBinary(name, dest string) error {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}
	url := fmt.Sprintf("https://storage.googleapis.com/gvisor/releases/release/latest/%s/%s", arch, name)
	fmt.Printf("Downloading %s from %s\n", name, url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	return os.Chmod(dest, 0755)
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
