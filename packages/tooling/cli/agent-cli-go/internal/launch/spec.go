package launch

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	sharednix "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision/nix"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	credshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	provshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runspec"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox"
)

// defaultFlakeShell is the devShell attribute used when none is requested.
const defaultFlakeShell = "default"

// planAgent resolves the agent and its provider route, and fails on a route the
// agent does not read.
func planAgent(options Options) (*types.AgentConfig, *types.ModelProvider, error) {
	agent, err := agents.GetAgent(types.AgentType(options.Agent))
	if err != nil {
		return nil, nil, fmt.Errorf("%w; available agents: %v", err, agents.GetAgentTypes())
	}
	provider, err := resolveProvider(options, agent)
	if err != nil {
		return nil, nil, err
	}
	if err := agents.CheckRoutes(agent, provider); err != nil {
		return nil, nil, err
	}
	return agent, provider, nil
}

// planSpec resolves the confinement of a run of agent and provider in
// sourceDir, and selects the isolator and the provisioner from registry. It
// resolves every relative path against cwd, and the sandbox home directory
// defaults to home. With enforce it refuses a sandbox whose guarantees fall short
// of the policy; without it the guarantees are left to the caller. It computes
// the agent Process once, from the facts of the agent and the provider, and the
// boundary the selected confinement declares.
func planSpec[I isoshared.Confinement](options Options, agent *types.AgentConfig, provider *types.ModelProvider, sourceDir, cwd, home string, registry *sandbox.Registry[I], enforce bool) (runspec.Spec, sandbox.Choice[I], error) {
	selected, err := selectAxes(options, registry, enforce)
	if err != nil {
		return runspec.Spec{}, sandbox.Choice[I]{}, err
	}
	input, err := provisionInput(selected.provision, options, agent, sourceDir)
	if err != nil {
		return runspec.Spec{}, sandbox.Choice[I]{}, err
	}
	customEnv, err := envutil.ParseCustomEnv(options.CustomEnv)
	if err != nil {
		return runspec.Spec{}, sandbox.Choice[I]{}, fmt.Errorf("invalid custom environment: %w", err)
	}
	allow, err := planAllowlist(selected.network, options)
	if err != nil {
		return runspec.Spec{}, sandbox.Choice[I]{}, err
	}
	bindings, err := credshared.ParseBindings(options.Credentials)
	if err != nil {
		return runspec.Spec{}, sandbox.Choice[I]{}, err
	}

	homeDir := home
	if options.HomeDir != "" {
		homeDir = options.HomeDir
	}
	// The source directory is bound as well, so a checkout can reach the shared
	// .git directory of the repository it was created from.
	bindPaths := append(slices.Clone(options.BindPaths), sourceDir)
	confinement := selected.choice.Isolator

	return runspec.Spec{
		Agent:         agent,
		Provider:      provider,
		IsolationName: selected.isolation,
		ProvisionName: selected.provision,
		Provision:     input,
		Allow:         allow,
		Credentials:   bindings,
		Capabilities:  selected.choice.Capabilities,
		Policy:        options.Policy,
		Run: isoshared.RunConfig{
			HomeDir: absolute(cwd, homeDir),
			RepoDir: sourceDir,
			Process: isoshared.Process{
				Argv: agentArgv(agent, provider, confinement.Bounded(), options.AgentArgs),
				Env:  envutil.EnvMapToSlice(customEnv),
				Dir:  sourceDir,
			},
			Network:         selected.network,
			Egress:          plannedRoute(selected.network, confinement.EgressTransport(), allow),
			HostPassthrough: options.IsolationBwrapPassthrough,
			Image:           options.Image,
			Runtime:         options.IsolationDockerRuntime,
			BindPaths:       absoluteAll(cwd, bindPaths),
			RoBindPaths:     absoluteAll(cwd, options.RoBindPaths),
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
func planAllowlist(mode network.NetworkMethod, options Options) (egress.Allowlist, error) {
	texts := options.AllowHosts
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
func selectAxes[I isoshared.Confinement](options Options, registry *sandbox.Registry[I], enforce bool) (selection[I], error) {
	isolationName := isolation.IsolationMethod(orDefault(options.Isolation, string(isolation.IsolationNone)))
	provisionName := provision.ProvisionMethod(orDefault(options.Provision, string(provision.ProvisionNone)))
	mode, err := network.ParseMode(orDefault(options.Network, string(network.NetworkHost)))
	if err != nil {
		return selection[I]{}, err
	}
	if options.Policy.RequirePinnedProvision && !options.ExplicitAxes {
		isolationName, provisionName = isolation.IsolationBwrap, provision.ProvisionNix
	}

	request := sandbox.Request{
		Isolation:      isolationName,
		Provision:      provisionName,
		Network:        mode,
		Passthrough:    options.IsolationBwrapPassthrough,
		Runtime:        options.IsolationDockerRuntime,
		Image:          options.Image,
		HasCustomBinds: len(options.BindPaths)+len(options.RoBindPaths) > 0,
	}
	var choice sandbox.Choice[I]
	if enforce {
		choice, err = sandbox.Select(registry, request, options.Policy)
	} else {
		choice, err = sandbox.Evaluate(registry, request)
	}
	if err != nil {
		return selection[I]{}, err
	}
	return selection[I]{isolation: isolationName, provision: provisionName, network: mode, choice: choice}, nil
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

// notDefined explains an undefined provider and where to add one.
func notDefined(options Options, agent *types.AgentConfig, cause error) error {
	hint := "add a file to a providers directory"
	if options.UserProviderDir != "" {
		hint = fmt.Sprintf("add %s%s to %s", options.Provider, providers.Extension, options.UserProviderDir)
	}
	defined := options.Catalog.Names(agent.Type)
	if len(defined) == 0 {
		return fmt.Errorf("provider %s is not defined; %s (no provider is defined for agent %s): %w", options.Provider, hint, agent.Type, cause)
	}
	return fmt.Errorf("provider %s is not defined; %s (defined for agent %s: %s): %w", options.Provider, hint, agent.Type, strings.Join(defined, ", "), cause)
}

// resolveProvider resolves the model provider by name from the catalog of the
// run. An agent that uses its own login takes no provider, so a named one fails
// before the lookup, and model overrides need a provider to apply to.
func resolveProvider(options Options, agent *types.AgentConfig) (*types.ModelProvider, error) {
	if options.Provider == "" {
		if len(options.ModelOverrides) > 0 {
			return nil, fmt.Errorf("model override %s needs a provider; select one with --provider", options.ModelOverrides[0].Slot+"="+options.ModelOverrides[0].ID)
		}
		return nil, nil
	}
	if !agents.TakesProvider(agent) {
		return nil, fmt.Errorf("provider %s: agent %s uses its own login and takes no provider", options.Provider, agent.Type)
	}
	provider, err := options.Catalog.Resolve(options.Provider, agent.Type, options.ModelOverrides...)
	if errors.Is(err, providers.ErrNotDefined) {
		return nil, notDefined(options, agent, err)
	}
	if err != nil {
		return nil, err
	}
	return provider, nil
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
