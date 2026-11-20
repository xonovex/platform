package launch

import (
	"fmt"
	"maps"
	"slices"

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
	return runconfig.Read(deps.ConfigEnv(), configRequest(options))
}

// LoadConfig reads the configuration of the options as a run does: the
// project content passes the trust gate first.
func LoadConfig(deps Dependencies, options Options) (runconfig.Loaded, error) {
	return runconfig.Load(deps.ConfigEnv(), configRequest(options))
}

// ListProviders reads the providers of every layer and marks each project
// provider with the approval state of the file that defines it.
func ListProviders(deps Dependencies, options Options) (ProviderListing, error) {
	loaded, err := ReadConfig(deps, options)
	if err != nil {
		return ProviderListing{}, err
	}
	// A project provider waits only while the pending content that defines it
	// is not approved.
	waiting := map[string]bool{}
	for _, p := range loaded.Pending {
		yes, err := runconfig.Approved(deps.ConfigEnv(), p)
		if err != nil {
			return ProviderListing{}, err
		}
		for _, origin := range p.Origins {
			waiting[origin] = !yes
		}
	}
	listing := ProviderListing{Dirs: loaded.ProviderDirs, Pending: loaded.Pending}
	for _, entry := range providers.NewCatalog(loaded.Providers...).List() {
		row := ProviderRow{Entry: entry, Document: definition(loaded.Providers, entry)}
		if entry.Layer == string(spec.Project) {
			row.Approval = "approved"
			if waiting[entry.Origin] {
				row.Approval = "not approved"
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

// ModelSlots lists the model slots of the provider as slot=id, sorted by slot.
func (r ProviderRow) ModelSlots() []string {
	models := r.Document.Preset.Models
	slots := make([]string, 0, len(models))
	for _, name := range slices.Sorted(maps.Keys(models)) {
		slots = append(slots, name+"="+models[name].ID)
	}
	return slots
}

// RouteHosts lists the hosts the route of the provider reaches.
func (r ProviderRow) RouteHosts() []string { return r.Document.Preset.Hosts() }

// ApproveProjectFiles lets each project file that the configuration of options
// reads only after approval through the trust gate, and returns the files it
// approved. approveDigest approves the one file whose digest it names. A
// project that supplies no such file approves nothing.
func ApproveProjectFiles(deps Dependencies, options Options) ([]runconfig.Pending, error) {
	listing, err := ListProviders(deps, options)
	if err != nil {
		return nil, err
	}
	if len(listing.Pending) == 0 {
		return nil, nil
	}
	if err := runconfig.Gate(deps.ConfigEnv(), listing.Pending, options.ApproveDigest); err != nil {
		return nil, err
	}
	return listing.Pending, nil
}

// RevokeProjectFiles removes from the approval ledger the approval of each
// project file that the configuration of options reads only after approval, and
// returns the files whose approval it removed.
func RevokeProjectFiles(deps Dependencies, options Options) ([]runconfig.Pending, error) {
	loaded, err := ReadConfig(deps, options)
	if err != nil {
		return nil, err
	}
	return runconfig.Revoke(deps.ConfigEnv(), loaded.Pending)
}

// RevokeDigest removes the approval of one digest from the approval ledger, and
// fails when the ledger holds none.
func RevokeDigest(deps Dependencies, digest string) error {
	held, err := runconfig.Ledger(deps.Host).Revoke(digest)
	if err != nil {
		return err
	}
	if !held {
		return fmt.Errorf("the approval ledger holds no approval of %s", digest)
	}
	return nil
}
