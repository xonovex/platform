package launch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
)

// ApprovalKind is the kind of the approval a project's provider material needs.
const ApprovalKind = "project-providers"

// ApprovalRequiredError reports that the provider of the run depends on provider
// material of the project, and the user did not approve its digest. The caller
// shows the material, asks the user, records the digest and runs again.
type ApprovalRequiredError struct {
	Material providers.Material
}

// Error names the digest to approve.
func (e *ApprovalRequiredError) Error() string {
	return fmt.Sprintf("the project supplies provider files or model overrides that are not approved; review them, then run agent-cli providers approve --approve-digest %s", e.Material.Digest())
}

// parseModelOverrides turns the merged "slot=id" items into overrides in the
// order a consumer applies them. A later item of the same slot wins, and a
// command line value that differs from a managed one fails, because a managed
// value is not changed by any other layer.
func parseModelOverrides(items []config.Item) ([]providers.Override, error) {
	var overrides []providers.Override
	byslot := map[string]config.Item{}
	for _, item := range items {
		slot, id, ok := strings.Cut(item.Text, "=")
		if !ok || slot == "" || id == "" {
			return nil, fmt.Errorf("model override %q from %s must be slot=id", item.Text, item.Source)
		}
		if previous, seen := byslot[slot]; seen && item.Source.Layer == config.LayerManaged && previous.Source.Layer == config.LayerFlags && previous.Text != item.Text {
			return nil, fmt.Errorf("%s on the command line contradicts %s from %s; the managed value cannot be changed", previous.Text, item.Text, item.Source)
		}
		byslot[slot] = item
		overrides = append(overrides, providers.Override{Slot: slot, ID: id, Source: item.Source.String()})
	}
	return overrides, nil
}

// needsApproval reports whether the provider the run selects depends on project
// material: it is defined, or refined, by a project file, or the project
// overrides its models.
func needsApproval(options Options, material providers.Material) bool {
	if options.Provider == "" || material.Empty() {
		return false
	}
	if len(material.Models) > 0 {
		return true
	}
	for _, layer := range options.Catalog.Layers(options.Provider, types.AgentType(options.Agent)) {
		if layer == string(config.LayerProject) {
			return true
		}
	}
	return false
}

// notDefined explains an undefined provider and where to add one.
func notDefined(options Options, agent *types.AgentConfig, cause error) error {
	hint := "add a file to a providers directory"
	for _, dir := range options.ProviderDirs {
		if dir.Layer == config.LayerUser {
			hint = fmt.Sprintf("add %s%s to %s", options.Provider, providers.Extension, dir.Path)
		}
	}
	defined := options.Catalog.Names(agent.Type)
	if len(defined) == 0 {
		return fmt.Errorf("provider %s is not defined; %s (no provider is defined for agent %s): %w", options.Provider, hint, agent.Type, cause)
	}
	return fmt.Errorf("provider %s is not defined; %s (defined for agent %s: %s): %w", options.Provider, hint, agent.Type, strings.Join(defined, ", "), cause)
}

// resolveProvider resolves the model provider by name from the catalog of the
// run. An agent that uses its own login takes no provider, so a named one fails
// before the lookup, and model overrides need a provider to apply to.
func resolveProvider(options Options, agent *types.AgentConfig) (*types.ModelProvider, error) {
	if options.Provider == "" {
		if len(options.ModelOverrides) > 0 {
			return nil, fmt.Errorf("model override %s needs a provider; select one with --provider", options.ModelOverrides[0].Slot+"="+options.ModelOverrides[0].ID)
		}
		return nil, nil
	}
	if !agents.TakesProvider(agent) {
		return nil, fmt.Errorf("provider %s: agent %s uses its own login and takes no provider", options.Provider, agent.Type)
	}
	provider, err := options.Catalog.Resolve(options.Provider, agent.Type, options.ModelOverrides...)
	if errors.Is(err, providers.ErrNotDefined) {
		return nil, notDefined(options, agent, err)
	}
	if err != nil {
		return nil, err
	}
	return provider, nil
}
