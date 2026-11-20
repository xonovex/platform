package wiring

import (
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/controller"
	isoshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/test/testutil"
)

// newOfflineManager builds a manager whose client and cache are lazy, so
// registration runs without an API server.
func newOfflineManager(t *testing.T, scheme *runtime.Scheme) ctrl.Manager {
	t.Helper()
	mgr, err := ctrl.NewManager(&rest.Config{Host: "http://127.0.0.1:1"}, ctrl.Options{
		Scheme:  scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	return mgr
}

// Controller names are process-global, so this is the only test of the
// package that registers the controllers on a valid scheme.
func TestSetup_RegistersEveryControllerAndWebhook(t *testing.T) {
	mgr := newOfflineManager(t, testutil.NewScheme())

	if err := SetupControllers(mgr, DefaultConfig()); err != nil {
		t.Fatalf("SetupControllers() error = %v", err)
	}
	if err := SetupWebhooks(mgr); err != nil {
		t.Fatalf("SetupWebhooks() error = %v", err)
	}
}

func TestSetupControllers_NamesTheControllerThatFails(t *testing.T) {
	err := SetupControllers(newOfflineManager(t, runtime.NewScheme()), DefaultConfig())

	if err == nil || !strings.Contains(err.Error(), "AgentRun controller") {
		t.Errorf("SetupControllers() error = %v, want it to name the AgentRun controller", err)
	}
}

func TestDefaultConfig_UsesThePinnedInitImageAndTheDefaultJobPolicy(t *testing.T) {
	config := DefaultConfig()

	if config.WorkspaceInitImage != controller.DefaultWorkspaceInitImage || config.JobPolicy != isoshared.DefaultJobPolicy() {
		t.Errorf("DefaultConfig() = %+v", config)
	}
}

func TestSystemClock_ReadsTheCurrentTime(t *testing.T) {
	before := time.Now()

	now := systemClock{}.Now()

	if now.Before(before) || now.After(time.Now()) {
		t.Errorf("Now() = %v, want a time between %v and the return", now, before)
	}
}

func TestSetupWebhooks_NamesTheWebhookThatFails(t *testing.T) {
	err := SetupWebhooks(newOfflineManager(t, runtime.NewScheme()))

	if err == nil || !strings.Contains(err.Error(), "AgentRun webhook") {
		t.Errorf("SetupWebhooks() error = %v, want it to name the AgentRun webhook", err)
	}
}
