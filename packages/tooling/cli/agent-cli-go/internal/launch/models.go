package launch

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// parseModelOverrides turns the merged "slot=id" items into overrides in the
// order a consumer applies them. A later item of the same slot wins, and a
// command line value that differs from a managed one fails, because a managed
// value is not changed by any other layer.
func parseModelOverrides(items []runconfig.Item) ([]providers.Override, error) {
	var overrides []providers.Override
	byslot := map[string]runconfig.Item{}
	for _, item := range items {
		slot, id, ok := strings.Cut(item.Text, "=")
		if !ok || slot == "" || id == "" {
			return nil, fmt.Errorf("model override %q from %s must be slot=id", item.Text, item.Source)
		}
		if previous, seen := byslot[slot]; seen && item.Source.Layer == spec.Managed && previous.Source.Layer == spec.Flags && previous.Text != item.Text {
			return nil, fmt.Errorf("%s on the command line contradicts %s from %s; the managed value cannot be changed", previous.Text, item.Text, item.Source)
		}
		byslot[slot] = item
		overrides = append(overrides, providers.Override{Slot: slot, ID: id, Source: item.Source.String()})
	}
	return overrides, nil
}
