package controller

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

// newOfflineManager builds a manager that registers controllers without
// contacting an API server: the client and cache are lazy, so wiring can be
// checked in the fast tier rather than only under envtest.
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

// setupFuncs is every reconciler's manager registration, which is what
// cmd/operator wires at startup.
func setupFuncs() map[string]func(ctrl.Manager) error {
	return map[string]func(ctrl.Manager) error{
		"AgentRun": func(mgr ctrl.Manager) error {
			return (&AgentRunReconciler{
				Client:   mgr.GetClient(),
				Scheme:   mgr.GetScheme(),
				Recorder: events.NewFakeRecorder(10),
			}).SetupWithManager(mgr)
		},
		"AgentProvider": func(mgr ctrl.Manager) error {
			return (&AgentProviderReconciler{
				Client:   mgr.GetClient(),
				Scheme:   mgr.GetScheme(),
				Recorder: events.NewFakeRecorder(10),
			}).SetupWithManager(mgr)
		},
		"AgentWorkspace": func(mgr ctrl.Manager) error {
			return (&AgentWorkspaceReconciler{
				Client:   mgr.GetClient(),
				Scheme:   mgr.GetScheme(),
				Recorder: events.NewFakeRecorder(10),
			}).SetupWithManager(mgr)
		},
	}
}

// Controller name validation is process-global rather than per-manager, so
// registering all three here also proves their names do not collide the way
// they would in cmd/operator.
func TestSetupWithManager_RegistersEveryReconciler(t *testing.T) {
	for name, setup := range setupFuncs() {
		t.Run(name, func(t *testing.T) {
			mgr := newOfflineManager(t, testutil.NewScheme())

			if err := setup(mgr); err != nil {
				t.Fatalf("SetupWithManager() error = %v", err)
			}
		})
	}
}

// Registration resolves each watched type through the scheme, so a reconciler
// watching a type the operator never registered fails here rather than at
// startup.
func TestSetupWithManager_FailsWhenTheWatchedTypesAreUnregistered(t *testing.T) {
	for name, setup := range setupFuncs() {
		t.Run(name, func(t *testing.T) {
			mgr := newOfflineManager(t, runtime.NewScheme())

			if err := setup(mgr); err == nil {
				t.Error("SetupWithManager() = nil, want an error on an empty scheme")
			}
		})
	}
}
