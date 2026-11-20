package launch

import (
	"encoding/json"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// ConfigView is the configuration of a run as a person inspects it.
type ConfigView struct {
	// Merged holds every merged value with the layers and the files that set it.
	Merged merge.Resolved
	// Sections are the merged sections agent-cli reads, by name.
	Sections map[string]json.RawMessage
}

// ShowConfig reads the configuration of the options as a run does, so the
// project content passes the trust gate first, and returns the merged view of
// the sections agent-cli reads.
func ShowConfig(deps Dependencies, options Options) (ConfigView, error) {
	loaded, err := LoadConfig(deps, options)
	if err != nil {
		return ConfigView{}, err
	}
	merged := loaded.Resolved.Merged()
	sections := map[string]json.RawMessage{}
	for _, section := range runconfig.Sections() {
		sections[section.Name] = merged.Section(section.Name)
	}
	return ConfigView{Merged: merged, Sections: sections}, nil
}
