package launch

import (
	launchcore "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/launch"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// Options are the inputs of one run as the user gave them: the Request the
// planner core reads, plus what only this CLI does around the run.
type Options struct {
	launchcore.Request

	// Models are the "slot=id" overrides of the command line. UserProviderDir is
	// the provider directory of the user layer, named in the error of a provider
	// that is not defined.
	Models          []string
	UserProviderDir string

	// Explicit holds the keys the command line chose, as opposed to the defaults
	// of the flags.
	Explicit runconfig.Keys

	WorkDir              string
	WorktreeBranch       string
	WorktreeSourceBranch string
	WorktreeDir          string
	VCS                  string

	// Plugins are the installed plugins whose artifacts the run overlays, and
	// OverlayDirs are directories of artifacts the run overlays beside them. Store
	// is the plugin store, empty for the default one.
	Plugins     []string
	OverlayDirs []string
	Store       string

	// Project names the project agentic.json instead of the search from the
	// work directory, and ApproveDigest is the digest of project content the
	// user approved on the command line.
	Project       string
	ApproveDigest string

	Terminal        string
	TerminalSession string
	TerminalWindow  string
	TerminalDetach  bool

	DryRun  bool
	Verbose bool
}

// DefaultOptions are the options a run holds before any flag or layer changes
// them.
func DefaultOptions() Options {
	return Options{
		Request: launchcore.Request{
			Agent:     "claude",
			Isolation: "none",
			Provision: "none",
			Network:   "host",
			NixSource: "packages",
			NixShell:  "default",
		},
		VCS: "git",
	}
}

// ExplicitKeys returns the configuration keys whose flags the user set, as
// changed reports each flag by its name. Only these keys need the mark: a flag
// that keeps its default yields to the file layers, and every other flag has no
// default that could shadow a file.
func ExplicitKeys(changed func(flag string) bool) runconfig.Keys {
	chosen := runconfig.Keys{}
	for flag, key := range map[string]runconfig.Key{
		"agent":     runconfig.KeyAgent,
		"isolation": runconfig.KeyIsolation,
		"provision": runconfig.KeyProvision,
		"network":   runconfig.KeyNetwork,
	} {
		if changed(flag) {
			chosen[key] = struct{}{}
		}
	}
	return chosen
}
