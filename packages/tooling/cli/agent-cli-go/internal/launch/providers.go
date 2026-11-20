package launch

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
)

// ProviderRow is one defined provider with the approval state of its file.
type ProviderRow struct {
	providers.Entry
	// Approval is empty for a file the user or an administrator supplies. For a
	// project file it is "approved" or "not approved".
	Approval string
}

// ProviderListing is every defined provider and the directories read for them.
type ProviderListing struct {
	Rows []ProviderRow
	Dirs []config.ProviderDir
	// Material is the provider content the project supplies.
	Material providers.Material
}

// ProjectMaterial returns the provider content the project of the options
// supplies: its provider files and its model overrides.
func ProjectMaterial(deps Dependencies, options Options) (providers.Material, error) {
	listing, err := ListProviders(deps, options)
	return listing.Material, err
}

// ListProviders reads the provider files of every layer and the project model
// overrides, and marks each project provider with the approval state of the
// project material.
func ListProviders(deps Dependencies, options Options) (ProviderListing, error) {
	cwd, err := deps.Host.Getwd()
	if err != nil {
		return ProviderListing{}, fmt.Errorf("resolve current working directory: %w", err)
	}
	projectDir := absolute(cwd, options.WorkDir)
	inputs, err := config.LoadLayers(deps.Host, options.Config, projectDir)
	if err != nil {
		return ProviderListing{}, fmt.Errorf("failed to load config: %w", err)
	}
	set, err := config.LoadProviders(deps.Host, options.Config, projectDir)
	if err != nil {
		return ProviderListing{}, fmt.Errorf("failed to load providers: %w", err)
	}
	material := providers.Material{Documents: set.Project, Models: config.ProjectModels(inputs)}
	approved, err := approvedDigest(deps.Approvals, material.Digest())
	if err != nil {
		return ProviderListing{}, fmt.Errorf("read the approvals: %w", err)
	}
	listing := ProviderListing{Dirs: set.Dirs, Material: material}
	for _, entry := range providers.NewCatalog(set.Documents...).List() {
		row := ProviderRow{Entry: entry}
		if entry.Layer == string(config.LayerProject) {
			row.Approval = "not approved"
			if approved {
				row.Approval = "approved"
			}
		}
		listing.Rows = append(listing.Rows, row)
	}
	return listing, nil
}
