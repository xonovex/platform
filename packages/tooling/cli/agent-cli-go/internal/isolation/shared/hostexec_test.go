package shared

import (
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func agentConfig() isoshared.RunConfig {
	return isoshared.RunConfig{
		Agent:   &types.AgentConfig{Type: "agent", Binary: "agent"},
		Process: isoshared.Process{Argv: []string{"agent", "--flag", "review"}, Dir: "/work"},
		HomeDir: "/home/user",
	}
}

func TestHostExecutionBuildsTheCommandAndEnvironment(t *testing.T) {
	host := systemtest.New()
	host.Env["PATH"] = "/usr/bin"
	host.Env["KEEP"] = "yes"
	host.Env["SHADOWED"] = "host"
	cfg := agentConfig()
	cfg.Process.Env = []string{"CUSTOM=1", "SHADOWED=custom"}
	cfg.ProviderEnv = map[string]string{"ROUTE": "provider", "SHADOWED": "provider"}
	cfg.Secrets = map[string]string{"TOKEN": "s3cret"}
	c := provision.Contribution{Env: map[string]string{"FROM_TOOL": "1"}, PathEntries: []string{"/nix/bin"}, InitCommands: []string{"prepare"}}

	command, env, err := HostExecution(host, cfg, c)
	if err != nil {
		t.Fatal(err)
	}
	if command[0] != "sh" || !strings.Contains(strings.Join(command, " "), "prepare && exec 'agent' '--flag' 'review'") {
		t.Errorf("command = %v, want init commands before the agent process", command)
	}
	for _, want := range []string{"KEEP=yes", "CUSTOM=1", "ROUTE=provider", "SHADOWED=custom", "TOKEN=s3cret", "FROM_TOOL=1", "PATH=/nix/bin:/usr/bin"} {
		if !slices.Contains(env, want) {
			t.Errorf("env lacks %q: %v", want, env)
		}
	}
}

func TestHostExecutionRunsTheProcessUnchanged(t *testing.T) {
	cfg := agentConfig()
	command, _, err := HostExecution(systemtest.New(), cfg, provision.Contribution{})
	if err != nil || !slices.Equal(command, cfg.Process.Argv) {
		t.Errorf("command = %v, %v, want the argument vector of the process", command, err)
	}
}

func TestHostExecutionPathWithoutHostPath(t *testing.T) {
	host := systemtest.New()
	_, env, err := HostExecution(host, agentConfig(), provision.Contribution{PathEntries: []string{"/nix/bin"}})
	if err != nil || !slices.Contains(env, "PATH=/nix/bin") {
		t.Errorf("env = %v, %v", env, err)
	}
}

func TestHostExecutionFailures(t *testing.T) {
	host := systemtest.New()
	cfg := agentConfig()
	cfg.Process.Argv = nil
	if _, _, err := HostExecution(host, cfg, provision.Contribution{}); err == nil || !strings.Contains(err.Error(), "no command") {
		t.Errorf("a process without a command must fail: %v", err)
	}
	cfg = agentConfig()
	cfg.Process.Env = []string{"no-equals"}
	if _, _, err := HostExecution(host, cfg, provision.Contribution{}); err == nil {
		t.Error("an invalid process variable must fail")
	}
}
