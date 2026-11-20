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

	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/wiring"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/test/testutil"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	k8sClient          client.Client
	ctx                context.Context
	cancel             context.CancelFunc
	clusterName        = "agent-operator-e2e"
	useExistingCluster bool
	// busyboxImage names the image of the AgentRun fixtures by digest.
	busyboxImage = "busybox@" + busyboxRegistryDigest
)

// busyboxRegistryDigest is the digest of busybox:1.37 in Docker Hub, which an
// existing cluster pulls. A Kind cluster loads the image and uses the digest of
// that copy instead.
const busyboxRegistryDigest = "sha256:a7dbc319b1438115a207e11930dc4bb3e643dcb3ec0bdd694bb36cd8fad289ec"

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

		// Load the image into the Kind node to avoid Docker Hub rate limits and
		// slow pulls during test timeouts.
		if err := runCmd("docker", "pull", "busybox:1.37"); err != nil {
			panic("failed to pull busybox: " + err.Error())
		}
		pinned, err := testutil.LoadKindImage(clusterName, "busybox:1.37")
		if err != nil {
			_ = runCmd("kind", "delete", "cluster", "--name", clusterName)
			panic("failed to load busybox into the Kind cluster: " + err.Error())
		}
		busyboxImage = pinned
	}

	// Install CRDs
	crdPath := filepath.Join("..", "..", "config", "crd", "bases")
	if err := runCmd("kubectl", "apply", "-f", crdPath); err != nil {
		if !useExisting {
			_ = runCmd("kind", "delete", "cluster", "--name", clusterName)
		}
		panic("failed to install CRDs: " + err.Error())
	}

	// The AgentRun fixtures name the kata runtime class, which a Kind node lacks.
	// These tests check scheduling and garbage collection, not isolation, so the
	// class runs on the runc handler; the e2e-kata suite runs a real Kata VM.
	if !useExisting {
		runtimeClass := filepath.Join("..", "testdata", "runtimeclass-kata.yaml")
		if err := runCmd("kubectl", "apply", "-f", runtimeClass); err != nil {
			_ = runCmd("kind", "delete", "cluster", "--name", clusterName)
			panic("failed to create the kata RuntimeClass: " + err.Error())
		}
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

	if err := wiring.SetupControllers(mgr, operatorConfig()); err != nil {
		panic("failed to set up controllers: " + err.Error())
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

// operatorConfig is the default configuration of the operator with the
// workspace init image the deployment of the operator names.
func operatorConfig() wiring.Config {
	config := wiring.DefaultConfig()
	config.WorkspaceInitImage = testutil.WorkspaceInitImage
	return config
}
