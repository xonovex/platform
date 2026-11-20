// Package launch is the application core of agent-cli: a pure planner that
// turns the run options into a resolved Plan, and an executor that carries the
// Plan out through the ports it is given. It names no concrete isolator,
// provisioner, workspace variant or terminal, and it does no I/O of its own.
package launch

import (
	"fmt"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	sharednix "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision/nix"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/validation"
	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
	provshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

// defaultFlakeShell is the devShell attribute used when none is requested.
const defaultFlakeShell = "default"

// Options are the inputs of one run as the user gave them.
type Options struct {
	Agent    string
	Provider string

	Isolation string
	Provision string
	Network   string
	// IsolationSet and ProvisionSet report whether the user chose the axis; an
	// unchosen pair yields to the pinned default under RequirePinnedProvision.
	IsolationSet bool
	ProvisionSet bool

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

	Config      string
	BindPaths   []string
	RoBindPaths []string
	CustomEnv   []string

	Terminal        string
	TerminalSession string
	TerminalWindow  string
	TerminalDetach  bool

	DryRun  bool
	Verbose bool
	Policy  policy.SandboxPolicy

	AgentArgs []string
}

// Checkout is a worktree or workspace the run creates before it launches.
type Checkout struct {
	VCS    wsp.VCSType
	Config wsshared.Config
	// Dir is the absolute directory the checkout resolves to.
	Dir string
}

// Plan is a run with every choice resolved and every input validated.
type Plan struct {
	Agent         *types.AgentConfig
	Provider      *types.ModelProvider
	IsolationName isolation.IsolationMethod
	Isolator      isoshared.Isolator
	Provisioner   provshared.Provisioner
	Provision     provshared.Input
	// SourceDir is the absolute directory the run starts from; Checkout is nil
	// when the run uses SourceDir itself.
	SourceDir string
	Checkout  *Checkout
	// Run launches in SourceDir; the executor moves it into a created checkout.
	Run      isoshared.RunConfig
	Terminal *termshared.TerminalConfig
	DryRun   bool
	Verbose  bool
}

// NewPlan resolves options against the file configuration, the working
// directory and the home directory of the CLI, and selects the isolator and the
// provisioner from registry under the requested policy. It fails closed on any
// invalid choice or unmet guarantee and performs no I/O.
func NewPlan(options Options, file config.FileConfig, cwd, home string, registry *sandbox.Registry) (Plan, error) {
	agent, err := agents.GetAgent(types.AgentType(options.Agent))
	if err != nil {
		return Plan{}, fmt.Errorf("%w; available agents: %v", err, agents.GetAgentTypes())
	}
	provider, err := resolveProvider(agent, options.Provider, file.Provider)
	if err != nil {
		return Plan{}, err
	}
	if err := agents.CheckRoutes(agent, provider); err != nil {
		return Plan{}, err
	}

	sourceDir := absolute(cwd, options.WorkDir)
	checkout, err := planCheckout(options, sourceDir)
	if err != nil {
		return Plan{}, err
	}

	selection, err := selectAxes(options, file, registry)
	if err != nil {
		return Plan{}, err
	}

	input, err := provisionInput(selection.provision, options, agent, sourceDir)
	if err != nil {
		return Plan{}, err
	}
	customEnv, err := envutil.ParseCustomEnv(append(append([]string{}, file.CustomEnv...), options.CustomEnv...))
	if err != nil {
		return Plan{}, fmt.Errorf("invalid custom environment: %w", err)
	}

	homeDir := home
	if file.HomeDir != "" {
		homeDir = file.HomeDir
	}
	bindPaths := wsshared.BuildBindPaths(append(append([]string{}, file.BindPaths...), options.BindPaths...), sourceDir)

	return Plan{
		Agent:         agent,
		Provider:      provider,
		IsolationName: selection.isolation,
		Isolator:      selection.isolator,
		Provisioner:   selection.provisioner,
		Provision:     input,
		SourceDir:     sourceDir,
		Checkout:      checkout,
		Run: isoshared.RunConfig{
			HomeDir:         absolute(cwd, homeDir),
			WorkDir:         sourceDir,
			RepoDir:         sourceDir,
			Network:         selection.network,
			HostPassthrough: options.IsolationBwrapPassthrough,
			Image:           options.Image,
			Runtime:         options.IsolationDockerRuntime,
			BindPaths:       absoluteAll(cwd, bindPaths),
			RoBindPaths:     absoluteAll(cwd, append(append([]string{}, file.RoBindPaths...), options.RoBindPaths...)),
			CustomEnv:       envutil.EnvMapToSlice(customEnv),
			Agent:           agent,
			Provider:        provider,
			AgentArgs:       options.AgentArgs,
		},
		Terminal: planTerminal(options),
		DryRun:   options.DryRun,
		Verbose:  options.Verbose,
	}, nil
}

// DisplayDir is the directory a dry run reports: the checkout it would create,
// or the source directory.
func (p Plan) DisplayDir() string {
	if p.Checkout != nil {
		return p.Checkout.Dir
	}
	return p.SourceDir
}

// resolveProvider resolves the model provider from the flag, falling back to the
// file config's provider. An agent that uses its own login takes no provider, so
// a named one fails before the preset lookup.
func resolveProvider(agent *types.AgentConfig, optionProvider, fileProvider string) (*types.ModelProvider, error) {
	name := optionProvider
	if name == "" {
		name = fileProvider
	}
	if name == "" {
		return nil, nil
	}
	if !agents.TakesProvider(agent) {
		return nil, fmt.Errorf("provider %s: agent %s uses its own login and takes no provider", name, agent.Type)
	}
	provider, err := providers.GetProvider(name, agent.Type)
	if err != nil {
		return nil, fmt.Errorf("%w; available providers: %v", err, providers.GetProviderNames(agent.Type))
	}
	return provider, nil
}

// planCheckout validates the worktree options and resolves the checkout
// directory, or returns nil when no worktree branch is requested.
func planCheckout(options Options, sourceDir string) (*Checkout, error) {
	if options.WorktreeBranch == "" {
		return nil, nil
	}
	if err := validation.ValidateBranch(options.WorktreeBranch); err != nil {
		return nil, fmt.Errorf("invalid --worktree-branch: %w", err)
	}
	if options.WorktreeSourceBranch != "" {
		if err := validation.ValidateBranch(options.WorktreeSourceBranch); err != nil {
			return nil, fmt.Errorf("invalid --worktree-source-branch: %w", err)
		}
	}
	vcs := wsp.VCSType(options.VCS)
	if !vcs.IsValid() {
		return nil, fmt.Errorf("unknown --vcs %q; valid values: git, jj", options.VCS)
	}
	dir := options.WorktreeDir
	if dir == "" {
		dir = wsshared.GetDefaultDir(options.WorktreeBranch, filepath.Base(sourceDir))
	}
	return &Checkout{
		VCS: vcs,
		Config: wsshared.Config{
			SourceBranch: options.WorktreeSourceBranch,
			Branch:       options.WorktreeBranch,
			Dir:          dir,
		},
		Dir: absolute(sourceDir, dir),
	}, nil
}

type selection struct {
	isolation   isolation.IsolationMethod
	provision   provision.ProvisionMethod
	network     netshared.Mode
	isolator    isoshared.Isolator
	provisioner provshared.Provisioner
}

// selectAxes resolves the confinement axes. The registry is the source of truth
// for valid isolation and provision methods (Select rejects unregistered ones);
// network is a closed enum validated here. When a pinned toolchain is required
// but no axis was chosen explicitly, the pinned pair (bwrap, nix) is selected.
func selectAxes(options Options, file config.FileConfig, registry *sandbox.Registry) (selection, error) {
	isolationName := isolation.IsolationMethod(orDefault(options.Isolation, string(isolation.IsolationNone)))
	provisionName := provision.ProvisionMethod(orDefault(options.Provision, string(provision.ProvisionNone)))
	network, err := netshared.ParseMode(orDefault(options.Network, string(netshared.ModeHost)))
	if err != nil {
		return selection{}, err
	}
	if options.Policy.RequirePinnedProvision && !options.IsolationSet && !options.ProvisionSet {
		isolationName, provisionName = isolation.IsolationBwrap, provision.ProvisionNix
	}

	isolator, provisioner, err := sandbox.Select(registry, sandbox.Request{
		Isolation:      isolationName,
		Provision:      provisionName,
		Network:        network,
		Passthrough:    options.IsolationBwrapPassthrough,
		Runtime:        options.IsolationDockerRuntime,
		Image:          options.Image,
		HasCustomBinds: len(file.BindPaths)+len(options.BindPaths)+len(file.RoBindPaths)+len(options.RoBindPaths) > 0,
	}, options.Policy)
	if err != nil {
		return selection{}, err
	}
	return selection{
		isolation:   isolationName,
		provision:   provisionName,
		network:     network,
		isolator:    isolator,
		provisioner: provisioner,
	}, nil
}

// provisionInput assembles the neutral provisioner Input. The nix source is built
// only for the nix provisioner; the others ignore it. An invalid --nix-source is
// returned as an error here (naming the bad value) rather than degrading to a
// zero source that fails later with a misleading diagnostic.
func provisionInput(method provision.ProvisionMethod, options Options, agent *types.AgentConfig, sourceDir string) (provshared.Input, error) {
	in := provshared.Input{InitCommands: options.InitCommands}
	if method != provision.ProvisionNix {
		return in, nil
	}
	source, err := nixSource(options, agent, sourceDir)
	if err != nil {
		return provshared.Input{}, err
	}
	in.NixSource = source
	return in, nil
}

// nixSource maps the nix flags to a NixSource. The "packages" source always
// includes the package of the selected agent; the "flake" source uses the
// devShell of the flake in sourceDir.
func nixSource(options Options, agent *types.AgentConfig, sourceDir string) (sharednix.NixSource, error) {
	var source sharednix.NixSource
	switch options.NixSource {
	case "", "packages":
		packages := appendUnique(append([]string{}, options.NixPackages...), agent.NixPackage)
		source = sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: options.NixRev, Packages: packages}
	case "flake":
		source = sharednix.NixSource{Kind: sharednix.NixSourceProjectFlake, FlakeRef: sourceDir, Shell: orDefault(options.NixShell, defaultFlakeShell)}
	default:
		return sharednix.NixSource{}, fmt.Errorf("unknown nix source %q; valid: packages, flake", options.NixSource)
	}
	if err := sharednix.ValidateSource(source); err != nil {
		return sharednix.NixSource{}, err
	}
	return source, nil
}

// planTerminal returns the terminal wrapper configuration, or nil when the run
// uses no wrapper.
func planTerminal(options Options) *termshared.TerminalConfig {
	if options.Terminal == "" {
		return nil
	}
	return &termshared.TerminalConfig{
		Type:        termshared.TerminalType(options.Terminal),
		SessionName: options.TerminalSession,
		WindowName:  options.TerminalWindow,
		Detach:      options.TerminalDetach,
	}
}

func appendUnique(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, current := range values {
		if current == value {
			return values
		}
	}
	return append(values, value)
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// absolute resolves path against base and cleans it.
func absolute(base, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(base, path)
}

func absoluteAll(base string, paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	out := make([]string, len(paths))
	for i, path := range paths {
		out[i] = absolute(base, path)
	}
	return out
}
