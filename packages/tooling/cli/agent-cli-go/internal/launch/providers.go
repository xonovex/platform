package launch

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// ProviderRow is one defined provider with the approval state of its file.
type ProviderRow struct {
	providers.Entry
	// Document is the provider document that defines the provider.
	Document providers.Document
	// Approval is empty for a file the user or an administrator supplies. For a
	// project file it is "approved" or "not approved".
	Approval string
}

// ProviderListing is every defined provider, the shard directories, and the
// project content behind the trust gate.
type ProviderListing struct {
	Rows    []ProviderRow
	Dirs    []runconfig.ProviderDir
	Pending []runconfig.Pending
}

// ReadConfig reads the configuration of the options without the trust gate,
// for a command that reports on the project content or approves it.
func ReadConfig(deps Dependencies, options Options) (runconfig.Loaded, error) {
	return readConfig(deps, options, runconfig.Read)
}

// LoadConfig reads the configuration of the options as a run does: the
// project content passes the trust gate first.
func LoadConfig(deps Dependencies, options Options) (runconfig.Loaded, error) {
	return readConfig(deps, options, runconfig.Load)
}

func readConfig(deps Dependencies, options Options, read func(runconfig.Env, runconfig.Request) (runconfig.Loaded, error)) (runconfig.Loaded, error) {
	return read(deps.ConfigEnv(), configRequest(options))
}

// ListProviders reads the providers of every layer and marks each project
// provider with the approval state of the file that defines it.
func ListProviders(deps Dependencies, options Options) (ProviderListing, error) {
	loaded, err := ReadConfig(deps, options)
	if err != nil {
		return ProviderListing{}, err
	}
	approved := map[string]bool{}
	for _, p := range loaded.Pending {
		yes, err := runconfig.Approved(deps.ConfigEnv(), p)
		if err != nil {
			return ProviderListing{}, err
		}
		for _, origin := range p.Origins {
			approved[origin] = yes
		}
	}
	listing := ProviderListing{Dirs: loaded.ProviderDirs, Pending: loaded.Pending}
	for _, entry := range providers.NewCatalog(loaded.Providers...).List() {
		row := ProviderRow{Entry: entry, Document: definition(loaded.Providers, entry)}
		if entry.Layer == string(spec.Project) {
			row.Approval = "not approved"
			if approved[entry.Origin] {
				row.Approval = "approved"
			}
		}
		listing.Rows = append(listing.Rows, row)
	}
	return listing, nil
}

// definition returns the document of the entry: the last one of its agent and
// name, as the catalog keeps it.
func definition(documents []providers.Document, entry providers.Entry) providers.Document {
	var found providers.Document
	for _, document := range documents {
		if document.Preset.Name == entry.Name && document.Preset.Agent == entry.Agent {
			found = document
		}
	}
	return found
}

// Named returns the rows of the providers with the name, one for each agent
// that defines it.
func (l ProviderListing) Named(name string) []ProviderRow {
	var rows []ProviderRow
	for _, row := range l.Rows {
		if row.Name == name {
			rows = append(rows, row)
		}
	}
	return rows
}
