package launch

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
)

// Options are the inputs of one run as the user gave them.
type Options struct {
	Agent    string
	Provider string
	// Catalog holds the provider files of every layer, and ModelOverrides the model
	// slot overrides in the order they apply. Models are the "slot=id" overrides
	// of the command line. UserProviderDir is the provider directory of the user
	// layer, named in the error of a provider that is not defined.
	Catalog         providers.Catalog
	Models          []string
	ModelOverrides  []providers.Override
	UserProviderDir string

	Isolation string
	Provision string
	Network   string
	// Explicit holds the keys the command line chose, as opposed to the defaults
	// of the flags. ExplicitAxes reports that a layer chose the isolation or the
	// provision; without that choice the pair yields to the pinned default under
	// RequirePinnedProvision.
	Explicit     config.Keys
	ExplicitAxes bool

	IsolationBwrapPassthrough bool
	IsolationDockerRuntime    string
	Image                     string
	InitCommands              []string
	NixSource                 string
	NixRev                    string
	NixPackages               []string
	NixShell                  string

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
	// HomeDir is the home directory of the sandbox; empty selects the user's.
	HomeDir     string
	BindPaths   []string
	RoBindPaths []string
	CustomEnv   []string
	// AllowHosts are the hosts a managed network run may reach.
	AllowHosts []string
	// Credentials bind environment variables to credential store entries, as
	// NAME=scheme:locator.
	Credentials []string

	Terminal        string
	TerminalSession string
	TerminalWindow  string
	TerminalDetach  bool

	DryRun  bool
	Verbose bool
	Policy  policy.SandboxPolicy

	AgentArgs []string
}
