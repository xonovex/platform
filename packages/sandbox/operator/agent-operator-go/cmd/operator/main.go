package main

import (
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/validator"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/wiring"
)

func main() {
	var probeAddr string
	var enableLeaderElection bool
	var providersDir string
	config := wiring.DefaultConfig()

	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	flag.StringVar(&config.WorkspaceInitImage, "workspace-init-image", config.WorkspaceInitImage, "Digest-pinned image used to initialize shared workspaces.")
	flag.DurationVar(&config.JobPolicy.WorkspaceInitTimeout, "workspace-init-timeout", config.JobPolicy.WorkspaceInitTimeout, "Active deadline of a workspace init Job.")
	flag.StringVar(&providersDir, "providers-dir", "", "Directory of provider files, <name>.json, that a presetRef may name; empty defines none.")
	ttl := flag.Int("job-ttl-seconds", int(config.JobPolicy.DefaultTTLSecondsAfterFinished), "Seconds a finished Job is kept when its run sets no TTL.")

	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	setupLog := ctrl.Log.WithName("setup")
	scheme, err := newScheme()
	if err != nil {
		setupLog.Error(err, "unable to register Kubernetes API schemes")
		os.Exit(1)
	}
	if err := validator.ValidatePinnedImageReference(config.WorkspaceInitImage); err != nil {
		setupLog.Error(err, "invalid workspace init image", "image", config.WorkspaceInitImage)
		os.Exit(1)
	}
	if *ttl < 0 || *ttl > math.MaxInt32 || config.JobPolicy.WorkspaceInitTimeout <= 0 {
		setupLog.Error(fmt.Errorf("job-ttl-seconds must be in [0, %d] and workspace-init-timeout positive", math.MaxInt32), "invalid Job policy")
		os.Exit(1)
	}
	config.JobPolicy.DefaultTTLSecondsAfterFinished = int32(*ttl)

	if providersDir != "" {
		if config.Presets, err = wiring.LoadPresets(osFiles{}, providersDir); err != nil {
			setupLog.Error(err, "unable to load provider files", "dir", providersDir)
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "agent-operator.xonovex.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := wiring.SetupControllers(mgr, config); err != nil {
		setupLog.Error(err, "unable to create controllers")
		os.Exit(1)
	}
	if err := wiring.SetupWebhooks(mgr, config.Presets); err != nil {
		setupLog.Error(err, "unable to set up webhooks")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func newScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("register Kubernetes core API: %w", err)
	}
	if err := agentv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("register agent API: %w", err)
	}
	return scheme, nil
}

// osFiles reads provider files from the local file system.
type osFiles struct{}

func (osFiles) ReadDir(dir string) ([]fs.DirEntry, error) { return os.ReadDir(dir) }
func (osFiles) ReadFile(file string) ([]byte, error)      { return os.ReadFile(file) }
