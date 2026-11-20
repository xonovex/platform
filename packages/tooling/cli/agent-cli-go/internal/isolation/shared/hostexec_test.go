package shared

import (
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func claudeConfig(t *testing.T) RunConfig {
	t.Helper()
	agent, err := agents.GetAgent(types.AgentClaude)
	if err != nil {
		t.Fatal(err)
	}
	return RunConfig{Agent: agent, WorkDir: "/work", HomeDir: "/home/user"}
}

func TestHostExecutionBuildsTheCommandAndEnvironment(t *testing.T) {
	host := systemtest.New()
	host.Env["PATH"] = "/usr/bin"
	host.Env["KEEP"] = "yes"
	cfg := claudeConfig(t)
	cfg.CustomEnv = []string{"CUSTOM=1"}
	cfg.Secrets = map[string]string{"TOKEN": "s3cret"}
	c := provision.Contribution{Env: map[string]string{"FROM_TOOL": "1"}, PathEntries: []string{"/nix/bin"}, InitCommands: []string{"prepare"}}

	command, env, err := HostExecution(host, cfg, c, true)
	if err != nil {
		t.Fatal(err)
	}
	if command[0] != "sh" || !strings.Contains(strings.Join(command, " "), "prepare") {
		t.Errorf("command = %v, want init commands before the agent", command)
	}
	for _, want := range []string{"KEEP=yes", "CUSTOM=1", "TOKEN=s3cret", "FROM_TOOL=1", "PATH=/nix/bin:/usr/bin"} {
		if !slices.Contains(env, want) {
			t.Errorf("env lacks %q: %v", want, env)
		}
	}
}

func TestHostExecutionSandboxFlagReachesTheAgentArguments(t *testing.T) {
	host := systemtest.New()
	plain, _, err := HostExecution(host, claudeConfig(t), provision.Contribution{}, false)
	if err != nil {
		t.Fatal(err)
	}
	boxed, _, err := HostExecution(host, claudeConfig(t), provision.Contribution{}, true)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Equal(plain, boxed) {
		t.Errorf("the sandbox flag must change the agent arguments: %v", plain)
	}
}

func TestHostExecutionPathWithoutHostPath(t *testing.T) {
	host := systemtest.New()
	_, env, err := HostExecution(host, claudeConfig(t), provision.Contribution{PathEntries: []string{"/nix/bin"}}, false)
	if err != nil || !slices.Contains(env, "PATH=/nix/bin") {
		t.Errorf("env = %v, %v", env, err)
	}
}

func TestHostExecutionFailures(t *testing.T) {
	host := systemtest.New()
	cfg := claudeConfig(t)
	cfg.Agent = &types.AgentConfig{Type: "nope", Binary: "nope"}
	if _, _, err := HostExecution(host, cfg, provision.Contribution{}, false); err == nil {
		t.Error("an unknown agent must fail")
	}
	cfg = claudeConfig(t)
	cfg.CustomEnv = []string{"no-equals"}
	if _, _, err := HostExecution(host, cfg, provision.Contribution{}, false); err == nil {
		t.Error("an invalid custom variable must fail")
	}
	cfg = claudeConfig(t)
	cfg.Provider = &types.ModelProvider{Name: "route", AgentType: types.AgentClaude, CredentialSourceEnv: "ROUTE_KEY", CredentialTargetEnv: "AGENT_KEY"}
	if _, _, err := HostExecution(host, cfg, provision.Contribution{}, false); err == nil {
		t.Error("a provider whose credential is not set must fail")
	}
}
