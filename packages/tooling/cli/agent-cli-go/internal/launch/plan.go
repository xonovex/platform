// Package launch is the application core of agent-cli: a pure planner that
// turns the run options into a resolved Plan, and an executor that carries the
// Plan out through the ports it is given. It names no concrete isolator,
// provisioner, workspace variant or terminal, and it does no I/O of its own.
package launch

import (
	"fmt"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	sharednix "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision/nix"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/validation"
	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
	credshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
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
	// Catalog holds the provider files of every layer, and ModelOverrides the model
	// slot overrides in the order they apply. Models are the "slot=id" overrides
	// of the command line. ProviderDirs are the directories the catalog was read
	// from, named in the error of a provider that is not defined.
	Catalog        providers.Catalog
	Models         []string
	ModelOverrides []providers.Override
	ProviderDirs   []config.ProviderDir

	Isolation string
	Provision string
	Network   string
	// Explicit holds the keys a layer chose, as opposed to the defaults of the
	// flags. An unchosen isolation and provision pair yields to the pinned default
	// under RequirePinnedProvision.
	Explicit config.Keys

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
	ProvisionName provision.ProvisionMethod
	Isolator      isoshared.Isolator
	Provisioner   provshared.Provisioner
	Provision     provshared.Input
	// SourceDir is the absolute directory the run starts from; Checkout is nil
	// when the run uses SourceDir itself.
	SourceDir string
	Checkout  *Checkout
	// Overlay is the artifact overlay of the run, nil when the run has none.
	Overlay *OverlayRequest
	// Allow is the egress allowlist of a managed network run, empty for any other.
	Allow egress.Allowlist
	// Credentials are the credential bindings the executor resolves at run time.
	Credentials []credshared.Binding
	// Capabilities are the guarantees the selected sandbox establishes, and Policy
	// the guarantees the run demands of it.
	Capabilities policy.Capabilities
	Policy       policy.SandboxPolicy
	// Run launches in SourceDir; the executor moves it into a created checkout.
	Run      isoshared.RunConfig
	Terminal *termshared.TerminalConfig
	DryRun   bool
	Verbose  bool
}

// NewPlan resolves options, which already carry every configuration layer,
// against the working directory and the home directory of the CLI, and selects
// the isolator and the provisioner from registry under the requested policy. It
// fails closed on any invalid choice or unmet guarantee and performs no I/O.
func NewPlan(options Options, cwd, home string, registry *sandbox.Registry) (Plan, error) {
	return buildPlan(options, cwd, home, registry, true)
}

// evaluatePolicy evaluates each guarantee of the policy against the capabilities
// of the selected sandbox.
func evaluatePolicy(plan Plan) []policy.Outcome {
	return policy.Evaluate(plan.Capabilities, plan.Policy)
}

// advisoryPlan is NewPlan without the demand: it resolves the same plan whether
// or not the sandbox meets the policy, and leaves the guarantees to the caller,
// which reports each of them. Every other invalid choice still fails.
func advisoryPlan(options Options, cwd, home string, registry *sandbox.Registry) (Plan, error) {
	return buildPlan(options, cwd, home, registry, false)
}

func buildPlan(options Options, cwd, home string, registry *sandbox.Registry, enforce bool) (Plan, error) {
	agent, err := agents.GetAgent(types.AgentType(options.Agent))
	if err != nil {
		return Plan{}, fmt.Errorf("%w; available agents: %v", err, agents.GetAgentTypes())
	}
	provider, err := resolveProvider(options, agent)
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
	overlayRequest, err := planOverlay(options, agent, cwd, sourceDir)
	if err != nil {
		return Plan{}, err
	}

	selection, err := selectAxes(options, registry, enforce)
	if err != nil {
		return Plan{}, err
	}

	input, err := provisionInput(selection.provision, options, agent, sourceDir)
	if err != nil {
		return Plan{}, err
	}
	customEnv, err := envutil.ParseCustomEnv(options.CustomEnv)
	if err != nil {
		return Plan{}, fmt.Errorf("invalid custom environment: %w", err)
	}

	allow, err := planAllowlist(selection, options)
	if err != nil {
		return Plan{}, err
	}
	bindings, err := credshared.ParseBindings(options.Credentials)
	if err != nil {
		return Plan{}, err
	}

	homeDir := home
	if options.HomeDir != "" {
		homeDir = options.HomeDir
	}
	bindPaths := wsshared.BuildBindPaths(options.BindPaths, sourceDir)

	return Plan{
		Agent:         agent,
		Provider:      provider,
		IsolationName: selection.isolation,
		ProvisionName: selection.provision,
		Isolator:      selection.isolator,
		Provisioner:   selection.provisioner,
		Provision:     input,
		SourceDir:     sourceDir,
		Checkout:      checkout,
		Overlay:       overlayRequest,
		Allow:         allow,
		Credentials:   bindings,
		Capabilities:  selection.capabilities,
		Policy:        options.Policy,
		Run: isoshared.RunConfig{
			HomeDir:         absolute(cwd, homeDir),
			WorkDir:         sourceDir,
			RepoDir:         sourceDir,
			Network:         selection.network,
			Egress:          plannedRoute(selection, allow),
			HostPassthrough: options.IsolationBwrapPassthrough,
			Image:           options.Image,
			Runtime:         options.IsolationDockerRuntime,
			BindPaths:       absoluteAll(cwd, bindPaths),
			RoBindPaths:     absoluteAll(cwd, options.RoBindPaths),
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

// planAllowlist parses the allowed hosts. A managed run needs at least one, and
// no other network mode takes any, because the hosts would then restrict nothing.
func planAllowlist(selected selection, options Options) (egress.Allowlist, error) {
	texts := options.AllowHosts
	if selected.network != network.NetworkManaged {
		if len(texts) > 0 {
			return egress.Allowlist{}, fmt.Errorf("allowed hosts need --network managed, not %s", selected.network)
		}
		return egress.Allowlist{}, nil
	}
	allow, err := egress.NewAllowlist(texts)
	if err != nil {
		return egress.Allowlist{}, fmt.Errorf("--network managed: %w; pass --allow-host", err)
	}
	return allow, nil
}

// plannedRoute is the route a managed run starts from: the transport the
// isolator declared and the allowlist. The executor fills in the socket or the
// proxy address when it starts the proxy.
func plannedRoute(selected selection, allow egress.Allowlist) isoshared.EgressRoute {
	if selected.network != network.NetworkManaged {
		return isoshared.EgressRoute{}
	}
	return isoshared.EgressRoute{Transport: selected.isolator.EgressTransport(), Allow: allow}
}

// DisplayDir is the directory a dry run reports: the checkout it would create,
// or the source directory.
func (p Plan) DisplayDir() string {
	if p.Checkout != nil {
		return p.Checkout.Dir
	}
	return p.SourceDir
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
	isolation    isolation.IsolationMethod
	provision    provision.ProvisionMethod
	network      network.NetworkMethod
	isolator     isoshared.Isolator
	provisioner  provshared.Provisioner
	capabilities policy.Capabilities
}

// selectAxes resolves the confinement axes. The registry is the source of truth
// for valid isolation and provision methods (Select rejects unregistered ones);
// network is a closed enum validated here. When a pinned toolchain is required
// but no axis was chosen explicitly, the pinned pair (bwrap, nix) is selected.
func selectAxes(options Options, registry *sandbox.Registry, enforce bool) (selection, error) {
	isolationName := isolation.IsolationMethod(orDefault(options.Isolation, string(isolation.IsolationNone)))
	provisionName := provision.ProvisionMethod(orDefault(options.Provision, string(provision.ProvisionNone)))
	network, err := network.ParseMode(orDefault(options.Network, string(network.NetworkHost)))
	if err != nil {
		return selection{}, err
	}
	if options.Policy.RequirePinnedProvision && !options.Explicit.Has(config.KeyIsolation) && !options.Explicit.Has(config.KeyProvision) {
		isolationName, provisionName = isolation.IsolationBwrap, provision.ProvisionNix
	}

	request := sandbox.Request{
		Isolation:      isolationName,
		Provision:      provisionName,
		Network:        network,
		Passthrough:    options.IsolationBwrapPassthrough,
		Runtime:        options.IsolationDockerRuntime,
		Image:          options.Image,
		HasCustomBinds: len(options.BindPaths)+len(options.RoBindPaths) > 0,
	}
	choose := sandbox.Evaluate
	if enforce {
		choose = func(registry *sandbox.Registry, request sandbox.Request) (sandbox.Choice, error) {
			return sandbox.Select(registry, request, options.Policy)
		}
	}
	choice, err := choose(registry, request)
	if err != nil {
		return selection{}, err
	}
	return selection{
		isolation:    isolationName,
		provision:    provisionName,
		network:      network,
		isolator:     choice.Isolator,
		provisioner:  choice.Provisioner,
		capabilities: choice.Capabilities,
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
