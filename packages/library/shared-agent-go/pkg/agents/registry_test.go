package agents

import (
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

func TestGetAgent(t *testing.T) {
	agent, err := GetAgent(types.AgentClaude)
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}
	if agent.Type != types.AgentClaude || agent.Binary != "claude" {
		t.Fatalf("GetAgent() = %+v, want Claude agent", agent)
	}
}

func TestGetAgentReturnsIndependentConfigurations(t *testing.T) {
	first, err := GetAgent(types.AgentKiro)
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}
	first.Binary = "changed"
	first.LaunchArgs[0] = "changed"

	second, err := GetAgent(types.AgentKiro)
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}
	if second.Binary == "changed" || second.LaunchArgs[0] == "changed" {
		t.Fatal("GetAgent() returned shared mutable configuration")
	}
}

func TestGetAgentRejectsUnknownType(t *testing.T) {
	agent, err := GetAgent(types.AgentType("unknown"))
	if err == nil || agent != nil {
		t.Fatalf("GetAgent() = (%+v, %v), want nil agent and error", agent, err)
	}
}

func TestGetAgentTypesReturnsSortedTypes(t *testing.T) {
	got := GetAgentTypes()
	want := []types.AgentType{
		types.AgentClaude, types.AgentCodex, types.AgentCopilot,
		types.AgentKiro, types.AgentOpencode, types.AgentPi,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetAgentTypes() = %v, want %v", got, want)
	}
}

func TestEveryLeafDeclaresItsFacts(t *testing.T) {
	for _, agentType := range GetAgentTypes() {
		t.Run(string(agentType), func(t *testing.T) {
			agent, err := GetAgent(agentType)
			if err != nil {
				t.Fatalf("GetAgent() error = %v", err)
			}
			if agent.Type != agentType {
				t.Errorf("Type = %q, want %q", agent.Type, agentType)
			}
			if agent.DisplayName == "" || agent.Binary == "" || agent.NixPackage == "" {
				t.Errorf("agent %+v lacks a display name, binary or nix package", agent)
			}
			if len(agent.VersionArgs) == 0 {
				t.Errorf("agent %s declares no version probe", agentType)
			}
			if agent.Access != types.AccessProvider && agent.Access != types.AccessLogin {
				t.Errorf("Access = %q, want a declared model access", agent.Access)
			}
			if agent.Access == types.AccessLogin && (agent.ProviderEnv || agent.ProviderArgs || agent.Protocol != types.ProtocolNone) {
				t.Errorf("native-login agent %s declares a provider route or protocol: %+v", agentType, agent)
			}
		})
	}
}

func TestDeclaredFacts(t *testing.T) {
	tests := []struct {
		agent    types.AgentType
		protocol types.Protocol
		access   types.ModelAccess
		env      bool
		args     bool
		binary   string
		nix      string
	}{
		{types.AgentClaude, types.ProtocolMessages, types.AccessProvider, true, false, "claude", "claude-code"},
		{types.AgentCodex, types.ProtocolResponses, types.AccessProvider, true, true, "codex", "codex"},
		{types.AgentCopilot, types.ProtocolNone, types.AccessLogin, false, false, "copilot", "github-copilot-cli"},
		{types.AgentKiro, types.ProtocolNone, types.AccessLogin, false, false, "kiro-cli-chat", "kiro-cli-unwrapped"},
		{types.AgentOpencode, types.ProtocolNone, types.AccessProvider, false, true, "opencode", "opencode"},
		{types.AgentPi, types.ProtocolResponses, types.AccessProvider, true, true, "pi", "pi-coding-agent"},
	}
	for _, test := range tests {
		t.Run(string(test.agent), func(t *testing.T) {
			agent, err := GetAgent(test.agent)
			if err != nil {
				t.Fatalf("GetAgent() error = %v", err)
			}
			if agent.Protocol != test.protocol || agent.Access != test.access ||
				agent.ProviderEnv != test.env || agent.ProviderArgs != test.args ||
				agent.Binary != test.binary || agent.NixPackage != test.nix {
				t.Fatalf("GetAgent() = %+v, want protocol %q access %q env %v args %v binary %q nix %q",
					agent, test.protocol, test.access, test.env, test.args, test.binary, test.nix)
			}
		})
	}
}

func agentOf(t *testing.T, agentType types.AgentType) *types.AgentConfig {
	t.Helper()
	agent, err := GetAgent(agentType)
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}
	return agent
}

func TestBuildArgsOrdersLaunchSandboxProviderThenCaller(t *testing.T) {
	agent := &types.AgentConfig{
		LaunchArgs: []string{"chat"}, SandboxArgs: []string{"--open"}, ProviderArgs: true,
	}
	got := BuildArgs(agent, []string{"go"}, types.AgentExecOptions{SandboxArgs: agent.SandboxArgs, ProviderCliArgs: []string{"--model", "m"}})
	want := []string{"chat", "--open", "--model", "m", "go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildArgs() = %v, want %v", got, want)
	}
}

// TestBuildArgsSkipsWhatTheLaunchDoesNotUse also proves that BuildArgs never
// adds the declared sandbox arguments on its own: whether they apply is the
// decision of the launch.
func TestBuildArgsSkipsWhatTheLaunchDoesNotUse(t *testing.T) {
	agent := &types.AgentConfig{SandboxArgs: []string{"--open"}, ProviderArgs: false}
	got := BuildArgs(agent, []string{"go"}, types.AgentExecOptions{ProviderCliArgs: []string{"--model", "m"}})
	if want := []string{"go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildArgs() = %v, want %v", got, want)
	}
}

func TestBuildArgsDoesNotMutateLaunchArgs(t *testing.T) {
	agent := agentOf(t, types.AgentKiro)
	BuildArgs(agent, []string{"x"}, types.AgentExecOptions{})
	if len(agent.LaunchArgs) != 3 {
		t.Fatalf("LaunchArgs = %v, want it unchanged", agent.LaunchArgs)
	}
}

func TestBuildClaudeArgsTurnOffThePromptsInsideAnIsolation(t *testing.T) {
	agent := agentOf(t, types.AgentClaude)
	got := BuildArgs(agent, []string{"review"}, types.AgentExecOptions{SandboxArgs: agent.SandboxArgs, ProviderCliArgs: []string{"ignored"}})
	want := []string{"--permission-mode", "bypassPermissions", "review"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildArgs() = %v, want %v", got, want)
	}
}

func TestBuildOpencodeArgsTakeTheProviderArguments(t *testing.T) {
	got := BuildArgs(agentOf(t, types.AgentOpencode), []string{"review"}, types.AgentExecOptions{ProviderCliArgs: []string{"--model", "gemini"}})
	want := []string{"--model", "gemini", "review"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildArgs() = %v, want %v", got, want)
	}
}

func TestBuildCodexArgsTurnOffTheCommandSandboxInsideAnIsolation(t *testing.T) {
	agent := agentOf(t, types.AgentCodex)
	inside := BuildArgs(agent, []string{"go"}, types.AgentExecOptions{SandboxArgs: agent.SandboxArgs})
	if want := []string{"--sandbox", "danger-full-access", "go"}; !reflect.DeepEqual(inside, want) {
		t.Fatalf("BuildArgs() inside = %v, want %v", inside, want)
	}
	outside := BuildArgs(agent, []string{"go"}, types.AgentExecOptions{})
	if want := []string{"go"}; !reflect.DeepEqual(outside, want) {
		t.Fatalf("BuildArgs() outside = %v, want %v", outside, want)
	}
}

func TestBuildEnvReturnsOnlyWhatTheAgentReads(t *testing.T) {
	providerEnv := map[string]string{"TOKEN": "secret"}

	env := BuildEnv(agentOf(t, types.AgentClaude), providerEnv)
	if env["TOKEN"] != "secret" {
		t.Fatalf("BuildEnv() = %v, want the provider environment", env)
	}
	env["TOKEN"] = "changed"
	if providerEnv["TOKEN"] != "secret" {
		t.Fatal("BuildEnv() mutated provider environment")
	}

	if got := BuildEnv(agentOf(t, types.AgentOpencode), providerEnv); len(got) != 0 {
		t.Fatalf("BuildEnv() = %v, want empty for an agent that reads no provider environment", got)
	}
}

func TestTakesProvider(t *testing.T) {
	for agentType, want := range map[types.AgentType]bool{
		types.AgentClaude: true, types.AgentCodex: true, types.AgentOpencode: true, types.AgentPi: true,
		types.AgentCopilot: false, types.AgentKiro: false,
	} {
		if got := TakesProvider(agentOf(t, agentType)); got != want {
			t.Errorf("TakesProvider(%s) = %v, want %v", agentType, got, want)
		}
	}
}

func TestCheckRoutes(t *testing.T) {
	envProvider := &types.ModelProvider{Name: "e", Environment: map[string]string{"A": "1"}}
	credentialProvider := &types.ModelProvider{Name: "c", CredentialTargetEnv: "KEY"}
	argsProvider := &types.ModelProvider{Name: "a", CliArgs: []string{"--model", "m"}}
	tests := []struct {
		name     string
		agent    types.AgentType
		provider *types.ModelProvider
		wantErr  string
	}{
		{"no provider", types.AgentKiro, nil, ""},
		{"environment into a reading agent", types.AgentClaude, envProvider, ""},
		{"arguments into a taking agent", types.AgentOpencode, argsProvider, ""},
		{"both routes", types.AgentCodex, argsProvider, ""},
		{"environment into an agent that ignores it", types.AgentOpencode, envProvider, "environment"},
		{"credential into an agent that ignores it", types.AgentOpencode, credentialProvider, "environment"},
		{"arguments into an agent that ignores them", types.AgentClaude, argsProvider, "arguments"},
		{"provider into a login agent", types.AgentCopilot, envProvider, "own login"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := CheckRoutes(agentOf(t, test.agent), test.provider)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("CheckRoutes() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("CheckRoutes() error = %v, want it to mention %q", err, test.wantErr)
			}
		})
	}
}
