package launch

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	sharednix "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision/nix"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
)

// defaultFlakeShell is the devShell attribute used when none is requested.
const defaultFlakeShell = "default"

// Request is a run as its caller asks for it, with every configuration layer
// already applied. An empty axis selects its default.
type Request struct {
	Agent string
	// Harnesses holds the facts of every harness the run may start, with the
	// override files of every layer applied.
	Harnesses agents.Catalog
	Provider  string
	// Catalog holds the provider files of every layer, and ModelOverrides the
	// model slot overrides in the order they apply.
	Catalog        providers.Catalog
	ModelOverrides []providers.Override

	Isolation string
	Provision string
	Network   string
	// ExplicitAxes reports that the caller chose the isolation or the provision;
	// without that choice the pair yields to the pinned default under
	// RequirePinnedProvision.
	ExplicitAxes bool

	IsolationBwrapPassthrough bool
	IsolationDockerRuntime    string
	Image                     string
	InitCommands              []string
	NixSource                 string
	NixRev                    string
	NixPackages               []string
	NixShell                  string

	// HomeDir is the home directory of the sandbox; empty selects the home
	// directory PlanSpec receives.
	HomeDir     string
	BindPaths   []string
	RoBindPaths []string
	CustomEnv   []string
	// AllowHosts are the hosts a managed network run may reach.
	AllowHosts []string
	// Credentials bind environment variables to credential store entries, as
	// NAME=scheme:locator.
	Credentials []string

	Policy policy.SandboxPolicy
	// AgentArgs are the arguments of the user, which end the argument vector of
	// the agent process.
	AgentArgs []string
	// Interactive attaches the agent to the terminal of the caller.
	Interactive bool
}

// ResolveAgent resolves the agent from the harness catalog of the request and
// its provider route, and fails on a route the agent does not read. A provider
// that no catalog document defines fails with an error that wraps
// providers.ErrNotDefined.
func ResolveAgent(request Request) (*types.AgentConfig, *types.ModelProvider, error) {
	agent, err := request.Harnesses.Get(types.AgentType(request.Agent))
	if err != nil {
		return nil, nil, fmt.Errorf("%w; available agents: %v", err, request.Harnesses.Types())
	}
	provider, err := resolveProvider(request, agent)
	if err != nil {
		return nil, nil, err
	}
	if err := agents.CheckRoutes(agent, provider); err != nil {
		return nil, nil, err
	}
	return agent, provider, nil
}

// resolveProvider resolves the model provider by name from the catalog of the
// run. An agent that uses its own login takes no provider, so a named one fails
// before the lookup, and model overrides need a provider to apply to.
func resolveProvider(request Request, agent *types.AgentConfig) (*types.ModelProvider, error) {
	if request.Provider == "" {
		if len(request.ModelOverrides) > 0 {
			return nil, fmt.Errorf("model override %s needs a provider; select one with --provider", request.ModelOverrides[0].Slot+"="+request.ModelOverrides[0].ID)
		}
		return nil, nil
	}
	if !agents.TakesProvider(agent) {
		return nil, fmt.Errorf("provider %s: agent %s uses its own login and takes no provider", request.Provider, agent.Type)
	}
	return request.Catalog.Resolve(request.Provider, agent.Type, request.ModelOverrides...)
}

// PlanSpec resolves the confinement of a run of agent and provider in
// sourceDir, and selects the isolator and the provisioner from registry. It
// resolves every relative path against cwd, and the sandbox home directory
// defaults to home. With enforce it refuses a sandbox whose guarantees fall short
// of the policy; without it the guarantees are left to the caller. It computes
// the agent Process once, from the facts of the agent and the provider, and the
// boundary the selected confinement declares, and it binds sourceDir only into
// a confinement that accepts host binds.
func PlanSpec[I isoshared.Confinement](request Request, agent *types.AgentConfig, provider *types.ModelProvider, sourceDir, cwd, home string, registry *sandbox.Registry[I], enforce bool) (Spec, sandbox.Choice[I], error) {
	selected, err := selectAxes(request, registry, enforce)
	if err != nil {
		return Spec{}, sandbox.Choice[I]{}, err
	}
	input, err := provisionInput(selected.provision, request, agent, sourceDir)
	if err != nil {
		return Spec{}, sandbox.Choice[I]{}, err
	}
	customEnv, err := envutil.ParseCustomEnv(request.CustomEnv)
	if err != nil {
		return Spec{}, sandbox.Choice[I]{}, fmt.Errorf("invalid custom environment: %w", err)
	}
	allow, err := planAllowlist(selected.network, request)
	if err != nil {
		return Spec{}, sandbox.Choice[I]{}, err
	}
	bindings, err := credshared.ParseBindings(request.Credentials)
	if err != nil {
		return Spec{}, sandbox.Choice[I]{}, err
	}

	homeDir := home
	if request.HomeDir != "" {
		homeDir = request.HomeDir
	}
	confinement := selected.choice.Isolator
	bindPaths := slices.Clone(request.BindPaths)
	if confinement.AcceptsHostBinds() {
		// The source directory is bound as well, so a checkout can reach the
		// shared .git directory of the repository it was created from.
		bindPaths = append(bindPaths, sourceDir)
	}

	return Spec{
		Agent:         agent,
		Provider:      provider,
		IsolationName: selected.isolation,
		ProvisionName: selected.provision,
		Provision:     input,
		Allow:         allow,
		Credentials:   bindings,
		Capabilities:  selected.choice.Capabilities,
		Policy:        request.Policy,
		Run: isoshared.RunConfig{
			HomeDir: absolute(cwd, homeDir),
			RepoDir: sourceDir,
			Process: isoshared.Process{
				Argv: agentArgv(agent, provider, confinement.Bounded(), request.AgentArgs),
				Env:  envutil.EnvMapToSlice(customEnv),
				Dir:  sourceDir,
			},
			Interactive:     request.Interactive,
			Network:         selected.network,
			Egress:          plannedRoute(selected.network, confinement.EgressTransport(), allow),
			HostPassthrough: request.IsolationBwrapPassthrough,
			Image:           request.Image,
			Runtime:         request.IsolationDockerRuntime,
			BindPaths:       absoluteAll(cwd, bindPaths),
			RoBindPaths:     absoluteAll(cwd, request.RoBindPaths),
			Agent:           agent,
			Provider:        provider,
		},
	}, selected.choice, nil
}

// agentArgv is the argument vector that starts the agent: its executable, its
// launch arguments, its sandbox arguments when a boundary surrounds it, the
// arguments of the provider route, and then args, which always come last.
func agentArgv(agent *types.AgentConfig, provider *types.ModelProvider, bounded bool, args []string) []string {
	var providerArgs []string
	if provider != nil {
		providerArgs = providers.GetProviderCliArgs(provider)
	}
	built := agents.BuildArgs(agent, args, types.AgentExecOptions{Sandbox: bounded, ProviderCliArgs: providerArgs})
	return append([]string{agent.Binary}, built...)
}

// planAllowlist parses the allowed hosts. A managed run needs at least one, and
// no other network mode takes any, because the hosts would then restrict nothing.
func planAllowlist(mode network.NetworkMethod, request Request) (egress.Allowlist, error) {
	texts := request.AllowHosts
	if mode != network.NetworkManaged {
		if len(texts) > 0 {
			return egress.Allowlist{}, fmt.Errorf("allowed hosts need --network managed, not %s", mode)
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
func plannedRoute(mode network.NetworkMethod, transport network.EgressTransport, allow egress.Allowlist) isoshared.EgressRoute {
	if mode != network.NetworkManaged {
		return isoshared.EgressRoute{}
	}
	return isoshared.EgressRoute{Transport: transport, Allow: allow}
}

type selection[I isoshared.Confinement] struct {
	isolation isolation.IsolationMethod
	provision provision.ProvisionMethod
	network   network.NetworkMethod
	choice    sandbox.Choice[I]
}

// selectAxes resolves the confinement axes. The registry is the source of truth
// for valid isolation and provision methods (Select rejects unregistered ones);
// network is a closed enum validated here. When a pinned toolchain is required
// but no axis was chosen explicitly, the pinned pair (bwrap, nix) is selected.
func selectAxes[I isoshared.Confinement](request Request, registry *sandbox.Registry[I], enforce bool) (selection[I], error) {
	isolationName := isolation.IsolationMethod(orDefault(request.Isolation, string(isolation.IsolationNone)))
	provisionName := provision.ProvisionMethod(orDefault(request.Provision, string(provision.ProvisionNone)))
	mode, err := network.ParseMode(orDefault(request.Network, string(network.NetworkHost)))
	if err != nil {
		return selection[I]{}, err
	}
	if request.Policy.RequirePinnedProvision && !request.ExplicitAxes {
		isolationName, provisionName = isolation.IsolationBwrap, provision.ProvisionNix
	}

	sandboxRequest := sandbox.Request{
		Isolation:      isolationName,
		Provision:      provisionName,
		Network:        mode,
		Passthrough:    request.IsolationBwrapPassthrough,
		Runtime:        request.IsolationDockerRuntime,
		Image:          request.Image,
		HasCustomBinds: len(request.BindPaths)+len(request.RoBindPaths) > 0,
	}
	var choice sandbox.Choice[I]
	if enforce {
		choice, err = sandbox.Select(registry, sandboxRequest, request.Policy)
	} else {
		choice, err = sandbox.Evaluate(registry, sandboxRequest)
	}
	if err != nil {
		return selection[I]{}, err
	}
	return selection[I]{isolation: isolationName, provision: provisionName, network: mode, choice: choice}, nil
}

// provisionInput assembles the neutral provisioner Input. The nix source is built
// only for the nix provisioner; the others ignore it. An invalid nix source is
// returned as an error here (naming the bad value) rather than degrading to a
// zero source that fails later with a misleading diagnostic.
func provisionInput(method provision.ProvisionMethod, request Request, agent *types.AgentConfig, sourceDir string) (provshared.Input, error) {
	in := provshared.Input{InitCommands: request.InitCommands}
	if method != provision.ProvisionNix {
		return in, nil
	}
	source, err := nixSource(request, agent, sourceDir)
	if err != nil {
		return provshared.Input{}, err
	}
	in.NixSource = source
	return in, nil
}

// nixSource maps the nix choices to a NixSource. The "packages" source always
// includes the package of the selected agent; the "flake" source uses the
// devShell of the flake in sourceDir.
func nixSource(request Request, agent *types.AgentConfig, sourceDir string) (sharednix.NixSource, error) {
	var source sharednix.NixSource
	switch request.NixSource {
	case "", "packages":
		packages := appendUnique(append([]string{}, request.NixPackages...), agent.NixPackage)
		source = sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: request.NixRev, Packages: packages}
	case "flake":
		source = sharednix.NixSource{Kind: sharednix.NixSourceProjectFlake, FlakeRef: sourceDir, Shell: orDefault(request.NixShell, defaultFlakeShell)}
	default:
		return sharednix.NixSource{}, fmt.Errorf("unknown nix source %q; valid: packages, flake", request.NixSource)
	}
	if err := sharednix.ValidateSource(source); err != nil {
		return sharednix.NixSource{}, err
	}
	return source, nil
}

func appendUnique(values []string, value string) []string {
	if value == "" || slices.Contains(values, value) {
		return values
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
