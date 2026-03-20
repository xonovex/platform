//go:build e2e_kata

package e2e_kata

import (
	"bytes"
	"context"
	"encoding/json"
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
	clusterName = "agent-operator-e2e-kata"
	tmpDir      string
)

func TestMain(m *testing.M) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.Background())

	// Kata requires /dev/kvm
	if _, err := os.Stat("/dev/kvm"); err != nil {
		fmt.Println("SKIP: /dev/kvm not available — Kata Containers requires hardware virtualization")
		os.Exit(0)
	}

	var err error
	tmpDir, err = os.MkdirTemp("", "e2e-kata-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}

	// Download and extract Kata static release
	kataDir := filepath.Join(tmpDir, "kata")
	if err := downloadKata(kataDir); err != nil {
		os.RemoveAll(tmpDir)
		panic("failed to download Kata: " + err.Error())
	}

	shimPath := filepath.Join(kataDir, "opt", "kata", "bin", "containerd-shim-kata-v2")
	if _, err := os.Stat(shimPath); err != nil {
		os.RemoveAll(tmpDir)
		panic("containerd-shim-kata-v2 not found at " + shimPath)
	}

	// Create kind cluster with Kata and /dev/kvm mounted
	kindConfig := filepath.Join(tmpDir, "kind-config.yaml")
	if err := os.WriteFile(kindConfig, []byte(fmt.Sprintf(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraMounts:
      - hostPath: %s/opt/kata
        containerPath: /opt/kata
      - hostPath: /dev/kvm
        containerPath: /dev/kvm
`, kataDir)), 0644); err != nil {
		os.RemoveAll(tmpDir)
		panic("failed to write kind config: " + err.Error())
	}

	if err := runCmd("kind", "create", "cluster", "--name", clusterName, "--config", kindConfig, "--wait", "120s"); err != nil {
		os.RemoveAll(tmpDir)
		panic("failed to create kind cluster: " + err.Error())
	}

	// Symlink shim inside the kind node and configure containerd
	setupScript := `
set -e
ln -sf /opt/kata/bin/containerd-shim-kata-v2 /usr/local/bin/containerd-shim-kata-v2
cat >> /etc/containerd/config.toml <<'TOML'

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.kata]
  runtime_type = "io.containerd.kata.v2"
  [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.kata.options]
    ConfigPath = "/opt/kata/share/defaults/kata-containers/configuration.toml"
TOML
systemctl restart containerd
`
	if err := runCmd("docker", "exec", clusterName+"-control-plane", "sh", "-c", setupScript); err != nil {
		cleanup()
		panic("failed to configure containerd for Kata: " + err.Error())
	}

	// Wait for node to be ready after containerd restart
	if err := runCmd("kubectl", "--context", "kind-"+clusterName, "wait", "--for=condition=Ready", "node/"+clusterName+"-control-plane", "--timeout=60s"); err != nil {
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
  name: kata
handler: kata
`
	runtimeClassPath := filepath.Join(tmpDir, "runtimeclass-kata.yaml")
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
	if err := (&controller.AgentConfigReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}).SetupWithManager(mgr); err != nil {
		cleanup()
		panic("failed to setup AgentConfig controller: " + err.Error())
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

func downloadKata(destDir string) error {
	arch := runtime.GOARCH // amd64, arm64, etc.

	// Get latest release tag
	resp, err := http.Get("https://api.github.com/repos/kata-containers/kata-containers/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release JSON: %w", err)
	}

	url := fmt.Sprintf("https://github.com/kata-containers/kata-containers/releases/download/%s/kata-static-%s-%s.tar.zst",
		release.TagName, release.TagName, arch)
	fmt.Printf("Downloading Kata %s from %s\n", release.TagName, url)

	tarball := filepath.Join(destDir, "kata.tar.zst")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	dlResp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", url, dlResp.StatusCode)
	}

	f, err := os.Create(tarball)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, dlResp.Body); err != nil {
		f.Close()
		return err
	}
	f.Close()

	// Extract (tar auto-detects zstd compression)
	cmd := exec.Command("tar", "xf", tarball, "-C", destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tar extract: %w", err)
	}

	return os.Remove(tarball)
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
