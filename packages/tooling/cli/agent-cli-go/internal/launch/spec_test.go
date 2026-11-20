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
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	provshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox"
)

func registeredAgent(t *testing.T, agentType types.AgentType) *types.AgentConfig {
	t.Helper()
	agent, err := agents.GetAgent(agentType)
	if err != nil {
		t.Fatalf("GetAgent(%s) error = %v", agentType, err)
	}
	return agent
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
			if got := agentArgv(agent, nil, true, []string{"go"}); !slices.Equal(got, test.bounded) {
				t.Errorf("bounded argv = %v, want %v", got, test.bounded)
			}
			if got := agentArgv(agent, nil, false, []string{"go"}); !slices.Equal(got, test.unbounded) {
				t.Errorf("unbounded argv = %v, want %v", got, test.unbounded)
			}
		})
	}
}

func TestAgentArgvPlacesTheProviderArgumentsBeforeTheUserArguments(t *testing.T) {
	provider := &types.ModelProvider{CliArgs: []string{"--model", "m"}}

	codex := agentArgv(registeredAgent(t, types.AgentCodex), provider, false, []string{"go"})
	if !slices.Equal(codex, []string{"codex", "--model", "m", "go"}) {
		t.Errorf("codex argv = %v, want the provider arguments before go", codex)
	}
	claude := agentArgv(registeredAgent(t, types.AgentClaude), provider, false, []string{"go"})
	if !slices.Equal(claude, []string{"claude", "go"}) {
		t.Errorf("claude argv = %v, want no provider arguments for an agent that reads none", claude)
	}
}

func TestNewPlanStartsTheAgentWithoutPromptsOnlyInsideABoundary(t *testing.T) {
	options := baseOptions()
	options.AgentArgs = []string{"go"}

	host := mustPlan(t, options, config.FileConfig{})
	if !slices.Equal(host.Run.Process.Argv, []string{"claude", "go"}) {
		t.Errorf("unbounded argv = %v, want claude go", host.Run.Process.Argv)
	}

	options.Isolation = "bwrap"
	bounded := mustPlan(t, options, config.FileConfig{})
	if !slices.Equal(bounded.Run.Process.Argv, []string{"claude", "--permission-mode", "bypassPermissions", "go"}) {
		t.Errorf("bounded argv = %v, want the sandbox arguments", bounded.Run.Process.Argv)
	}
}

func TestNewPlanCarriesTheProviderArgumentsInTheProcess(t *testing.T) {
	options := baseOptions()
	options.Agent = "codex"
	options.Provider = "gamma"
	options.AgentArgs = []string{"go"}

	p := mustPlan(t, options, config.FileConfig{})

	if !slices.Equal(p.Run.Process.Argv, []string{"codex", "--model", "model-main", "go"}) {
		t.Errorf("argv = %v, want the provider route before the user arguments", p.Run.Process.Argv)
	}
}

func TestNewPlanFillsThePortableSpec(t *testing.T) {
	options := bwrapOptions()
	options.Network = "managed"
	options.AllowHosts = []string{"api.example.com"}
	options.Credentials = []string{"API_KEY=env:SOURCE"}
	options.Policy = policy.SandboxPolicy{RequireEgressRestricted: true}

	p := mustPlan(t, options, config.FileConfig{})

	if p.IsolationName != isolation.IsolationBwrap || p.ProvisionName != provision.ProvisionNone || p.Run.Network != network.NetworkManaged {
		t.Errorf("axes = %s %s %s", p.IsolationName, p.ProvisionName, p.Run.Network)
	}
	if !p.Allow.Allows("api.example.com", 443) || p.Run.Egress.Transport != network.TransportUnixSocket || !p.Run.Egress.Allow.Allows("api.example.com", 443) {
		t.Errorf("allowlist, route = %v %+v", p.Allow, p.Run.Egress)
	}
	if len(p.Credentials) != 1 || p.Credentials[0].Name != "API_KEY" {
		t.Errorf("Credentials = %+v, want the API_KEY binding", p.Credentials)
	}
	if !p.Capabilities.EgressRestricted || !p.Policy.RequireEgressRestricted {
		t.Errorf("capabilities, policy = %+v %+v", p.Capabilities, p.Policy)
	}
	if p.Run.ProviderEnv != nil || p.Run.Secrets != nil {
		t.Errorf("the plan resolved run-time values: %v %v", p.Run.ProviderEnv, p.Run.Secrets)
	}
}

// specConfinement declares a sandbox and starts nothing.
type specConfinement struct{}

func (specConfinement) PinnedProvision(_ provision.ProvisionMethod, pinned bool, _ string) bool {
	return pinned
}
func (specConfinement) HidesHost(bool, string) bool              { return true }
func (specConfinement) KernelIsolated(string) bool               { return true }
func (specConfinement) ConfinesNetwork() bool                    { return true }
func (specConfinement) EgressTransport() network.EgressTransport { return network.TransportLoopback }
func (specConfinement) Bounded() bool                            { return true }

func TestPlanSpecNeedsOnlyAConfinement(t *testing.T) {
	registry := sandbox.NewRegistry[isoshared.Confinement]().
		RegisterIsolator(isolation.IsolationDocker, func() isoshared.Confinement { return specConfinement{} }).
		RegisterProvisioner(provision.ProvisionNone, func() provshared.Provisioner { return &fakeProvisioner{} })
	options := baseOptions()
	options.Isolation = "docker"
	options.AgentArgs = []string{"go"}
	agent := registeredAgent(t, types.AgentClaude)

	spec, choice, err := planSpec(options, agent, nil, "/work/repo", "/work", "/home/user", registry, true)

	if err != nil {
		t.Fatalf("planSpec() error = %v", err)
	}
	if _, ok := choice.Isolator.(specConfinement); !ok || spec.IsolationName != isolation.IsolationDocker {
		t.Errorf("choice = %T %s, want the registered confinement", choice.Isolator, spec.IsolationName)
	}
	if !slices.Equal(spec.Run.Process.Argv, []string{"claude", "--permission-mode", "bypassPermissions", "go"}) || spec.Run.Process.Dir != "/work/repo" {
		t.Errorf("process = %+v", spec.Run.Process)
	}
	if !spec.Capabilities.KernelIsolated {
		t.Errorf("Capabilities = %+v, want the declared kernel boundary", spec.Capabilities)
	}
}

func TestNewPlanYieldsThePinnedPairToAnExplicitAxis(t *testing.T) {
	options := baseOptions()
	options.NixRev = testRevision
	options.Policy.RequirePinnedProvision = true

	pinned, err := NewPlan(options, "/work/repo", "/home/user", newTestSandbox().registry)
	if err != nil || pinned.IsolationName != isolation.IsolationBwrap || pinned.ProvisionName != provision.ProvisionNix {
		t.Fatalf("NewPlan() = %s %s, %v; want the pinned pair", pinned.IsolationName, pinned.ProvisionName, err)
	}

	options.ExplicitAxes = true
	if _, err := NewPlan(options, "/work/repo", "/home/user", newTestSandbox().registry); !errors.Is(err, policy.ErrPinnedProvisionUnmet) {
		t.Fatalf("NewPlan(explicit axes) error = %v, want %v", err, policy.ErrPinnedProvisionUnmet)
	}
}

func TestApplyResolvedMarksAnAxisTheCommandLineOrALayerChose(t *testing.T) {
	merge := func(options Options, file config.FileConfig) Options {
		return fold(t, options, file)
	}
	if merge(baseOptions(), config.FileConfig{}).ExplicitAxes {
		t.Error("ExplicitAxes = true without a chosen axis")
	}
	if !merge(baseOptions(), config.FileConfig{Provision: "nix"}).ExplicitAxes {
		t.Error("ExplicitAxes = false for a provision the user layer chose")
	}
	flagged := baseOptions()
	flagged.Explicit = config.KeysOf(config.KeyIsolation)
	if !merge(flagged, config.FileConfig{}).ExplicitAxes {
		t.Error("ExplicitAxes = false for an isolation the command line chose")
	}
	network := baseOptions()
	network.Explicit = config.KeysOf(config.KeyNetwork)
	if merge(network, config.FileConfig{}).ExplicitAxes {
		t.Error("ExplicitAxes = true for a network choice")
	}
}

func TestUserProviderDirNamesTheLastUserLayer(t *testing.T) {
	dirs := []config.ProviderDir{
		{Layer: config.LayerManaged, Path: "/etc/providers"},
		{Layer: config.LayerUser, Path: "/home/user/providers"},
		{Layer: config.LayerProject, Path: "/work/repo/providers"},
	}
	if got := userProviderDir(dirs); got != "/home/user/providers" {
		t.Errorf("userProviderDir() = %q, want the user directory", got)
	}
	if got := userProviderDir(dirs[:1]); got != "" {
		t.Errorf("userProviderDir(managed only) = %q, want empty", got)
	}
}

func TestInsertBeforePlacesTheOverlayArgumentsBeforeTheUserArguments(t *testing.T) {
	argv := []string{"claude", "--flag", "review"}

	if got := insertBefore(argv, 1, []string{"--add-dir", "/o"}); !slices.Equal(got, []string{"claude", "--flag", "--add-dir", "/o", "review"}) {
		t.Errorf("insertBefore(1) = %v", got)
	}
	if got := insertBefore(argv, 0, []string{"--add-dir", "/o"}); !slices.Equal(got, []string{"claude", "--flag", "review", "--add-dir", "/o"}) {
		t.Errorf("insertBefore(0) = %v", got)
	}
	if !slices.Equal(argv, []string{"claude", "--flag", "review"}) {
		t.Errorf("insertBefore changed its input: %v", argv)
	}
}

func TestExecuteResolvesTheProviderEnvironmentForTheIsolator(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(userProviders+"/routed.json", `{"name":"routed","agent":"claude","models":{"main":{"id":"m"}},"credential":{"sourceEnv":"ROUTE_TOKEN","targetEnv":"ROUTE_AUTH"}}`, 0o644)
	options := bwrapOptions()
	options.Provider = "routed"

	err := Execute(f.deps(), options)
	if err == nil || !strings.HasPrefix(err.Error(), "prepare sandbox command: ") || !strings.Contains(err.Error(), "ROUTE_TOKEN") {
		t.Fatalf("Execute(no token) error = %v, want the missing provider token", err)
	}
	if f.sandbox.isolator.ran {
		t.Error("the agent ran without its provider token")
	}

	f.host.Env["ROUTE_TOKEN"] = "t0ken"
	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := f.sandbox.isolator.commandRun.ProviderEnv["ROUTE_AUTH"]; got != "t0ken" {
		t.Errorf("ProviderEnv = %v, want ROUTE_AUTH from the host", f.sandbox.isolator.commandRun.ProviderEnv)
	}
}

func TestNewPlanRejectsARouteTheAgentDoesNotRead(t *testing.T) {
	raw := `{"name":"argued","agent":"claude","models":{"main":{"id":"m"}},"args":["--model","{model:main}"]}`
	preset, err := providers.ParsePreset([]byte(raw))
	if err != nil {
		t.Fatalf("ParsePreset() error = %v", err)
	}
	options := baseOptions()
	options.Provider = "argued"
	options.Catalog = providers.NewCatalog(providers.Document{Preset: preset, Layer: "user", Origin: "/test/argued.json", Raw: []byte(raw)})

	if _, err := plan(t, options, config.FileConfig{}); err == nil || !strings.Contains(err.Error(), "sets arguments that agent claude does not take") {
		t.Fatalf("NewPlan() error = %v, want the unread route", err)
	}
}
