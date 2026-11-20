package launch

import (
	"fmt"
	"strings"

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
