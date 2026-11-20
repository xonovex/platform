// Package wiring is the composition root of the operator: it binds the
// controller-runtime adapters and the plugin registry to the ports of the
// reconcilers and the admission webhooks.
package wiring

import (
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/controller"
	isoshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/kube"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/plugins"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/webhook"
)

// eventSource names the operator as the reporting controller of its events.
const eventSource = "agent-operator"

// Config holds the operator settings that the command line selects. An empty
// WorkspaceInitImage selects controller.DefaultWorkspaceInitImage.
type Config struct {
	WorkspaceInitImage string
	JobPolicy          isoshared.JobPolicy
	// Presets are the portable providers a presetRef names; the zero catalog
	// defines none, and a presetRef then fails validation.
	Presets providers.Catalog
}

// DefaultConfig returns the settings the operator uses when no flag overrides
// them.
func DefaultConfig() Config {
	return Config{WorkspaceInitImage: controller.DefaultWorkspaceInitImage, JobPolicy: isoshared.DefaultJobPolicy()}
}

// LoadPresets reads the provider files of dir into the catalog a presetRef
// resolves against.
func LoadPresets(files providers.Reader, dir string) (providers.Catalog, error) {
	documents, err := providers.LoadDir(files, dir, "operator")
	if err != nil {
		return providers.Catalog{}, err
	}
	return providers.NewCatalog(documents...), nil
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// SetupControllers registers the AgentRun, AgentProvider and AgentWorkspace
// reconcilers with the manager.
func SetupControllers(mgr ctrl.Manager, config Config) error {
	store := kube.NewStore(mgr.GetClient())
	events := kube.NewEventRecorder(mgr.GetEventRecorder(eventSource))
	registry := plugins.Registry{}
	reconcilers := []struct {
		name  string
		setup func(ctrl.Manager) error
	}{
		{name: "AgentRun", setup: controller.NewAgentRunReconciler(controller.AgentRunDependencies{
			Store:      store,
			Toolchains: registry,
			Harnesses:  registry,
			VCS:        registry,
			Events:     events,
			Clock:      systemClock{},
			JobPolicy:  config.JobPolicy,
			Presets:    config.Presets,
		}).SetupWithManager},
		{name: "AgentProvider", setup: controller.NewAgentProviderReconciler(controller.AgentProviderDependencies{
			Store:  store,
			Events: events,
			Clock:  systemClock{},
		}).SetupWithManager},
		{name: "AgentWorkspace", setup: controller.NewAgentWorkspaceReconciler(controller.AgentWorkspaceDependencies{
			Store:              store,
			VCS:                registry,
			Events:             events,
			Clock:              systemClock{},
			JobPolicy:          config.JobPolicy,
			WorkspaceInitImage: config.WorkspaceInitImage,
		}).SetupWithManager},
	}
	for _, reconciler := range reconcilers {
		if err := reconciler.setup(mgr); err != nil {
			return fmt.Errorf("set up %s controller: %w", reconciler.name, err)
		}
	}
	return nil
}

// SetupWebhooks registers the defaulting and validating webhooks of every
// agent resource with the manager. presets are the portable providers a presetRef
// may name.
func SetupWebhooks(mgr ctrl.Manager, presets providers.Catalog) error {
	store := kube.NewStore(mgr.GetClient())
	registry := plugins.Registry{}
	webhooks := []struct {
		name  string
		setup func(ctrl.Manager) error
	}{
		{name: "AgentRun", setup: webhook.NewAgentRunWebhook(store, store, registry, presets).SetupWebhookWithManager},
		{name: "AgentHarness", setup: webhook.NewAgentHarnessWebhook(registry).SetupWebhookWithManager},
		{name: "AgentProvider", setup: (&webhook.AgentProviderWebhook{Presets: presets}).SetupWebhookWithManager},
		{name: "AgentToolchain", setup: webhook.NewAgentToolchainWebhook(registry).SetupWebhookWithManager},
		{name: "AgentWorkspace", setup: webhook.NewAgentWorkspaceWebhook(store, registry).SetupWebhookWithManager},
	}
	for _, hook := range webhooks {
		if err := hook.setup(mgr); err != nil {
			return fmt.Errorf("set up %s webhook: %w", hook.name, err)
		}
	}
	return nil
}
