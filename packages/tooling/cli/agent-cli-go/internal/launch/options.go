package launch

import (
	launchcore "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/launch"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
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
	Explicit config.Keys

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

	// Config is the file of the user layer; empty selects the default file.
	Config string

	Terminal        string
	TerminalSession string
	TerminalWindow  string
	TerminalDetach  bool

	DryRun  bool
	Verbose bool
}
