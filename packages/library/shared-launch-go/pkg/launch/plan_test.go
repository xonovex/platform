package launch

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	sharednix "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision/nix"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
)

const revision = "0123456789abcdef0123456789abcdef01234567"

// confinement declares a sandbox and starts nothing. The zero value is bounded,
// confines the network over a loopback proxy and accepts host binds.
type confinement struct {
	unbounded  bool
	noBinds    bool
	needsImage bool
}

func (confinement) PinnedProvision(_ provision.ProvisionMethod, pinned bool, _ string) bool {
	return pinned
}
func (confinement) HidesHost(passthrough bool, _ string) bool { return !passthrough }
func (confinement) KernelIsolated(runtime string) bool        { return runtime == "runsc" }
func (confinement) ConfinesNetwork() bool                     { return true }
func (confinement) EgressTransport() network.EgressTransport  { return network.TransportLoopback }
func (c confinement) Bounded() bool                           { return !c.unbounded }
func (c confinement) AcceptsHostBinds() bool                  { return !c.noBinds }
func (c confinement) RequiresImage() bool                     { return c.needsImage }

type provisioner struct{ pinned bool }

func (provisioner) Contribute(provshared.Input) (provision.Contribution, error) {
	return provision.Contribution{}, nil
}
func (p provisioner) Pinned() bool { return p.pinned }
func (provisioner) CheckAgent(provision.Contribution, *types.AgentConfig) (bool, error) {
	return false, nil
}

// registry registers the host as an unbounded isolation, bwrap and docker as
// bounded ones, and openshell as one without host binds.
func registry() *sandbox.Registry[isoshared.Confinement] {
	return sandbox.NewRegistry[isoshared.Confinement]().
		RegisterIsolator(isolation.IsolationNone, func() isoshared.Confinement { return confinement{unbounded: true} }).
		RegisterIsolator(isolation.IsolationBwrap, func() isoshared.Confinement { return confinement{} }).
		RegisterIsolator(isolation.IsolationDocker, func() isoshared.Confinement { return confinement{needsImage: true} }).
		RegisterIsolator(isolation.IsolationOpenShell, func() isoshared.Confinement { return confinement{noBinds: true, needsImage: true} }).
		RegisterProvisioner(provision.ProvisionNone, func() provshared.Provisioner { return provisioner{} }).
		RegisterProvisioner(provision.ProvisionCommand, func() provshared.Provisioner { return provisioner{} }).
		RegisterProvisioner(provision.ProvisionNix, func() provshared.Provisioner { return provisioner{pinned: true} })
}

func registeredAgent(t *testing.T, agentType types.AgentType) *types.AgentConfig {
	t.Helper()
	agent, err := agents.GetAgent(agentType)
	if err != nil {
		t.Fatalf("GetAgent(%s) error = %v", agentType, err)
	}
	return agent
}

func catalog(t *testing.T, raws ...string) providers.Catalog {
	t.Helper()
	var documents []providers.Document
	for _, raw := range raws {
		preset, err := providers.ParsePreset([]byte(raw))
		if err != nil {
			t.Fatalf("ParsePreset() error = %v", err)
		}
		documents = append(documents, providers.Document{Preset: preset, Layer: "user", Origin: "/providers/" + preset.Name + ".json", Raw: []byte(raw)})
	}
	return providers.NewCatalog(documents...)
}

// plan resolves the request against the shipped harness facts and plans it.
func plan(t *testing.T, request Request) (Spec, error) {
	t.Helper()
	request.Harnesses = agents.Embedded()
	agent, provider, err := ResolveAgent(request)
	if err != nil {
		t.Fatalf("ResolveAgent() error = %v", err)
	}
	spec, _, err := PlanSpec(request, agent, provider, "/work/repo", "/work", "/home/user", registry(), true)
	return spec, err
}

func mustPlan(t *testing.T, request Request) Spec {
	t.Helper()
	spec, err := plan(t, request)
	if err != nil {
		t.Fatalf("PlanSpec() error = %v", err)
	}
	return spec
}

func TestResolveAgentResolvesTheAgentAndItsProvider(t *testing.T) {
	request := Request{Agent: "codex", Harnesses: agents.Embedded(), Provider: "coder", Catalog: catalog(t, `{"name":"coder","agent":"codex","models":{"main":{"id":"model-code"}},"args":["--model","{model:main}"]}`)}

	agent, provider, err := ResolveAgent(request)

	if err != nil || agent.Type != types.AgentCodex || provider == nil || provider.Name != "coder" {
		t.Fatalf("ResolveAgent() = %v, %v, %v; want codex with the coder route", agent, provider, err)
	}
	if agent, provider, err := ResolveAgent(Request{Agent: "claude", Harnesses: agents.Embedded()}); err != nil || agent.Type != types.AgentClaude || provider != nil {
		t.Errorf("ResolveAgent(no provider) = %v, %v, %v; want claude alone", agent, provider, err)
	}
}

func TestResolveAgentReadsTheHarnessCatalogOfTheRequest(t *testing.T) {
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	harnesses, err := agents.Embedded().Apply(validator, agents.Override{Name: types.AgentClaude, Origin: "claude.json", Data: []byte(`{"binary":"claude-next"}`)})
	if err != nil {
		t.Fatal(err)
	}

	agent, _, err := ResolveAgent(Request{Agent: "claude", Harnesses: harnesses})

	if err != nil || agent.Binary != "claude-next" {
		t.Errorf("ResolveAgent() = %+v, %v, want the overridden binary", agent, err)
	}
	if _, _, err := ResolveAgent(Request{Agent: "claude"}); err == nil || !strings.Contains(err.Error(), "available agents: []") {
		t.Errorf("ResolveAgent() without harnesses error = %v, want none available", err)
	}
}

func TestResolveAgentRefusesWhatTheAgentCannotTake(t *testing.T) {
	argued := catalog(t, `{"name":"argued","agent":"claude","models":{"main":{"id":"m"}},"args":["--model","{model:main}"]}`)
	cases := map[string]struct {
		request Request
		want    string
	}{
		"unknown agent":           {Request{Agent: "ghost"}, "available agents"},
		"override without route":  {Request{Agent: "claude", ModelOverrides: []providers.Override{{Slot: "main", ID: "m"}}}, "model override main=m needs a provider"},
		"provider for a login":    {Request{Agent: "copilot", Provider: "route"}, "uses its own login"},
		"provider not defined":    {Request{Agent: "claude", Provider: "ghost", Catalog: argued}, "provider is not defined"},
		"route the agent ignores": {Request{Agent: "claude", Provider: "argued", Catalog: argued}, "does not take"},
	}
	for name, test := range cases {
		test.request.Harnesses = agents.Embedded()
		if _, _, err := ResolveAgent(test.request); err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("%s: ResolveAgent() error = %v, want %q", name, err, test.want)
		}
	}
	if _, _, err := ResolveAgent(Request{Agent: "claude", Harnesses: agents.Embedded(), Provider: "ghost"}); !errors.Is(err, providers.ErrNotDefined) {
		t.Errorf("ResolveAgent(undefined) error = %v, want %v", err, providers.ErrNotDefined)
	}
}

func TestAgentArgvFollowsTheDeclaredFactsOfEachHarness(t *testing.T) {
	tests := []struct {
		agent     types.AgentType
		bounded   []string
		unbounded []string
	}{
		{types.AgentClaude, []string{"claude", "--permission-mode", "bypassPermissions", "go"}, []string{"claude", "go"}},
		{types.AgentCodex, []string{"codex", "--sandbox", "danger-full-access", "go"}, []string{"codex", "go"}},
		{types.AgentKiro, []string{"kiro-cli-chat", "chat", "--agent-engine", "v3", "go"}, []string{"kiro-cli-chat", "chat", "--agent-engine", "v3", "go"}},
		{types.AgentCopilot, []string{"copilot", "go"}, []string{"copilot", "go"}},
		{types.AgentPi, []string{"pi", "go"}, []string{"pi", "go"}},
	}
	for _, test := range tests {
		t.Run(string(test.agent), func(t *testing.T) {
			agent := registeredAgent(t, test.agent)
			if got := agentArgv(agent, nil, sandboxArgs(agent, confinement{}), []string{"go"}); !slices.Equal(got, test.bounded) {
				t.Errorf("bounded argv = %v, want %v", got, test.bounded)
			}
			if got := agentArgv(agent, nil, sandboxArgs(agent, confinement{unbounded: true}), []string{"go"}); !slices.Equal(got, test.unbounded) {
				t.Errorf("unbounded argv = %v, want %v", got, test.unbounded)
			}
		})
	}
}

// TestPlanSpecAddsTheSandboxArgumentsOnlyInsideABoundary proves the one
// decision of sandboxArgs through the planner, for every harness the shipped
// facts declare and for each class of isolation: none, a bounded isolation, and
// an isolation that confines without being a boundary.
func TestPlanSpecAddsTheSandboxArgumentsOnlyInsideABoundary(t *testing.T) {
	classes := []struct {
		name      string
		isolation string
		bounded   bool
	}{
		{"none", "none", false},
		{"bounded", "bwrap", true},
		{"not bounded", "seatbelt", false},
	}
	// The registry of the other planner tests plus seatbelt as an isolation
	// that confines the network and is not a boundary.
	reg := registry().RegisterIsolator(isolation.IsolationSeatbelt, func() isoshared.Confinement { return confinement{unbounded: true} })
	harnesses := agents.Embedded()
	declaring := 0
	for _, agentType := range harnesses.Types() {
		facts, err := harnesses.Get(agentType)
		if err != nil {
			t.Fatalf("Get(%s) error = %v", agentType, err)
		}
		if len(facts.SandboxArgs) > 0 {
			declaring++
		}
		for _, class := range classes {
			t.Run(string(agentType)+"/"+class.name, func(t *testing.T) {
				request := Request{Agent: string(agentType), Harnesses: harnesses, Isolation: class.isolation, AgentArgs: []string{"go"}}
				agent, provider, err := ResolveAgent(request)
				if err != nil {
					t.Fatalf("ResolveAgent() error = %v", err)
				}
				spec, _, err := PlanSpec(request, agent, provider, "/work/repo", "/work", "/home/user", reg, true)
				if err != nil {
					t.Fatalf("PlanSpec() error = %v", err)
				}
				want := append([]string{facts.Binary}, facts.LaunchArgs...)
				if class.bounded {
					want = append(want, facts.SandboxArgs...)
				}
				want = append(want, "go")
				if !slices.Equal(spec.Run.Process.Argv, want) {
					t.Errorf("argv = %v, want %v", spec.Run.Process.Argv, want)
				}
			})
		}
	}
	if declaring == 0 {
		t.Error("no harness declares sandbox arguments, so the table proves nothing")
	}
}

func TestAgentArgvPlacesTheProviderArgumentsBeforeTheUserArguments(t *testing.T) {
	provider := &types.ModelProvider{CliArgs: []string{"--model", "m"}}

	codex := agentArgv(registeredAgent(t, types.AgentCodex), provider, nil, []string{"go"})
	if !slices.Equal(codex, []string{"codex", "--model", "m", "go"}) {
		t.Errorf("codex argv = %v, want the provider arguments before go", codex)
	}
	claude := agentArgv(registeredAgent(t, types.AgentClaude), provider, nil, []string{"go"})
	if !slices.Equal(claude, []string{"claude", "go"}) {
		t.Errorf("claude argv = %v, want no provider arguments for an agent that reads none", claude)
	}
}

func TestPlanSpecFillsThePortableSpec(t *testing.T) {
	request := Request{
		Agent:       "claude",
		Isolation:   "bwrap",
		Network:     "managed",
		AllowHosts:  []string{"api.example.com"},
		Credentials: []string{"API_KEY=env:SOURCE"},
		CustomEnv:   []string{"CUSTOM=1"},
		BindPaths:   []string{"cache"},
		RoBindPaths: []string{"/opt/tools/"},
		Image:       "registry.example/agent:1",
		Policy:      policy.SandboxPolicy{RequireEgressRestricted: true},
		AgentArgs:   []string{"go"},
		Interactive: true,
	}

	spec := mustPlan(t, request)

	if spec.IsolationName != isolation.IsolationBwrap || spec.ProvisionName != provision.ProvisionNone || spec.Run.Network != network.NetworkManaged {
		t.Errorf("axes = %s %s %s", spec.IsolationName, spec.ProvisionName, spec.Run.Network)
	}
	if !spec.Allow.Allows("api.example.com", 443) || spec.Run.Egress.Transport != network.TransportLoopback || !spec.Run.Egress.Allow.Allows("api.example.com", 443) {
		t.Errorf("allowlist, route = %v %+v", spec.Allow, spec.Run.Egress)
	}
	if len(spec.Credentials) != 1 || spec.Credentials[0].Name != "API_KEY" {
		t.Errorf("Credentials = %+v, want the API_KEY binding", spec.Credentials)
	}
	if !spec.Capabilities.EgressRestricted || !spec.Policy.RequireEgressRestricted {
		t.Errorf("capabilities, policy = %+v %+v", spec.Capabilities, spec.Policy)
	}
	process := spec.Run.Process
	if !slices.Equal(process.Argv, []string{"claude", "--permission-mode", "bypassPermissions", "go"}) || process.Dir != "/work/repo" || !slices.Equal(process.Env, []string{"CUSTOM=1"}) {
		t.Errorf("Process = %+v", process)
	}
	if !slices.Equal(spec.Run.BindPaths, []string{"/work/cache", "/work/repo"}) || !slices.Equal(spec.Run.RoBindPaths, []string{"/opt/tools"}) {
		t.Errorf("binds = %v %v, want the absolute binds and the source directory", spec.Run.BindPaths, spec.Run.RoBindPaths)
	}
	if spec.Run.HomeDir != "/home/user" || spec.Run.RepoDir != "/work/repo" || spec.Run.Image != request.Image || spec.Run.Agent != spec.Agent {
		t.Errorf("Run = %+v", spec.Run)
	}
	if spec.Run.ProviderEnv != nil || spec.Run.Secrets != nil {
		t.Errorf("the plan resolved run-time values: %v %v", spec.Run.ProviderEnv, spec.Run.Secrets)
	}
	if !spec.Run.Interactive {
		t.Error("an interactive request plans an interactive run")
	}
}

func TestPlanSpecStartsTheAgentWithoutPromptsOnlyInsideABoundary(t *testing.T) {
	host := mustPlan(t, Request{Agent: "claude", AgentArgs: []string{"go"}})

	if !slices.Equal(host.Run.Process.Argv, []string{"claude", "go"}) || host.Run.Egress.Transport != "" {
		t.Errorf("host run = %v %+v, want the plain argv and no route", host.Run.Process.Argv, host.Run.Egress)
	}
	if host.Run.Interactive {
		t.Error("a request that asks for no terminal plans a run without one")
	}
}

func TestPlanSpecBindsTheSourceDirectoryOnlyWhereHostBindsAreAccepted(t *testing.T) {
	spec := mustPlan(t, Request{Agent: "claude", Isolation: "openshell", Image: "registry.example/agent:1"})

	if len(spec.Run.BindPaths) != 0 || spec.Run.Process.Dir != "/work/repo" {
		t.Errorf("openshell run = binds %v, dir %q; want no bind and the source directory as work directory", spec.Run.BindPaths, spec.Run.Process.Dir)
	}
	if _, err := plan(t, Request{Agent: "claude", Isolation: "openshell", Image: "registry.example/agent:1", BindPaths: []string{"/cache"}}); err == nil || !strings.Contains(err.Error(), "cannot bind a host path") {
		t.Errorf("PlanSpec(openshell with binds) error = %v, want the refused bind", err)
	}
}

func TestPlanSpecRefusesAnImageIsolationWithoutAnImage(t *testing.T) {
	for _, name := range []string{"docker", "openshell"} {
		if _, err := plan(t, Request{Agent: "claude", Isolation: name}); !errors.Is(err, isoshared.ErrNoImage) {
			t.Errorf("isolation %s without an image: error = %v, want ErrNoImage", name, err)
		}
	}
	if _, err := plan(t, Request{Agent: "claude", Isolation: "docker", Image: "registry.example/agent:1"}); err != nil {
		t.Errorf("docker with an image: error = %v", err)
	}
}

func TestPlanSpecHonorsTheHomeDirectoryOfTheRequest(t *testing.T) {
	spec := mustPlan(t, Request{Agent: "claude", HomeDir: "sandbox-home"})

	if spec.Run.HomeDir != "/work/sandbox-home" {
		t.Errorf("HomeDir = %q, want the requested home resolved against the working directory", spec.Run.HomeDir)
	}
}

func TestPlanSpecSelectsThePinnedPairUnlessAnAxisIsExplicit(t *testing.T) {
	request := Request{Agent: "claude", NixRev: revision, Policy: policy.SandboxPolicy{RequirePinnedProvision: true}}

	pinned := mustPlan(t, request)
	if pinned.IsolationName != isolation.IsolationBwrap || pinned.ProvisionName != provision.ProvisionNix {
		t.Fatalf("axes = %s %s, want the pinned pair", pinned.IsolationName, pinned.ProvisionName)
	}
	if !slices.Equal(pinned.Provision.NixSource.Packages, []string{"claude-code"}) || pinned.Provision.NixSource.Rev != revision {
		t.Errorf("NixSource = %+v, want the agent package at the revision", pinned.Provision.NixSource)
	}

	request.ExplicitAxes = true
	if _, err := plan(t, request); !errors.Is(err, policy.ErrPinnedProvisionUnmet) {
		t.Errorf("PlanSpec(explicit axes) error = %v, want %v", err, policy.ErrPinnedProvisionUnmet)
	}
}

func TestPlanSpecWithoutEnforcementLeavesTheGuaranteesToTheCaller(t *testing.T) {
	request := Request{Agent: "claude", ExplicitAxes: true, Policy: policy.SandboxPolicy{RequireKernelIsolation: true}}
	agent := registeredAgent(t, types.AgentClaude)

	spec, choice, err := PlanSpec(request, agent, nil, "/work/repo", "/work", "/home/user", registry(), false)

	if err != nil || spec.Capabilities.KernelIsolated || choice.Isolator == nil {
		t.Fatalf("PlanSpec(advisory) = %+v, %v; want the unmet guarantee reported, not enforced", spec.Capabilities, err)
	}
}

func TestPlanSpecBuildsTheNixSource(t *testing.T) {
	request := Request{Agent: "claude", Provision: "nix", NixPackages: []string{"git", "claude-code"}, NixRev: revision}
	packages := mustPlan(t, request)
	if packages.Provision.NixSource.Kind != sharednix.NixSourcePackages || !slices.Equal(packages.Provision.NixSource.Packages, []string{"git", "claude-code"}) {
		t.Errorf("packages source = %+v, want the agent package once", packages.Provision.NixSource)
	}

	request.NixSource = "flake"
	flake := mustPlan(t, request)
	if source := flake.Provision.NixSource; source.Kind != sharednix.NixSourceProjectFlake || source.FlakeRef != "/work/repo" || source.Shell != "default" {
		t.Errorf("flake source = %+v, want the default shell of the source directory", flake.Provision.NixSource)
	}
	request.NixShell = "ci"
	if got := mustPlan(t, request).Provision.NixSource.Shell; got != "ci" {
		t.Errorf("flake shell = %q, want the requested shell", got)
	}

	request = Request{Agent: "claude", Provision: "command", InitCommands: []string{"make"}, NixSource: "bogus"}
	if got := mustPlan(t, request).Provision; !slices.Equal(got.InitCommands, []string{"make"}) || got.NixSource.Kind != "" {
		t.Errorf("command input = %+v, want the commands and no nix source", got)
	}
}

func TestPlanSpecRefusesAnInvalidRequest(t *testing.T) {
	cases := map[string]struct {
		request Request
		want    string
	}{
		"unknown network":        {Request{Agent: "claude", Network: "radio"}, "radio"},
		"unregistered isolation": {Request{Agent: "claude", Isolation: "firejail"}, "no isolator registered"},
		"unknown nix source":     {Request{Agent: "claude", Provision: "nix", NixSource: "tarball"}, "unknown nix source"},
		"unpinned nix packages":  {Request{Agent: "claude", Provision: "nix"}, "revision"},
		"invalid custom env":     {Request{Agent: "claude", CustomEnv: []string{"no-equals"}}, "invalid custom environment"},
		"hosts without managed":  {Request{Agent: "claude", AllowHosts: []string{"api.example.com"}}, "allowed hosts need --network managed"},
		"managed without hosts":  {Request{Agent: "claude", Isolation: "bwrap", Network: "managed"}, "pass --allow-host"},
		"invalid credential":     {Request{Agent: "claude", Credentials: []string{"nope"}}, "NAME=scheme:locator"},
	}
	for name, test := range cases {
		if _, err := plan(t, test.request); err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("%s: PlanSpec() error = %v, want %q", name, err, test.want)
		}
	}
}
