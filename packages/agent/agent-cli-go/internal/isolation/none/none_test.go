package none

import (
	"strings"
	"testing"

	isoshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func cfg(net netshared.Mode) isoshared.RunConfig {
	return isoshared.RunConfig{
		Agent:   &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"},
		WorkDir: "/tmp",
		Network: net,
	}
}

func TestNone_FailsClosedOnNetworkRestriction(t *testing.T) {
	for _, net := range []netshared.Mode{netshared.ModeNone, netshared.ModeProxy} {
		i := NewIsolator()
		c := provision.Contribution{}
		if code, err := i.Run(cfg(net), c); err == nil || code == 0 {
			t.Errorf("Run(network=%q) = (%d, %v), want fail-closed error (none cannot restrict egress)", net, code, err)
		}
		if _, err := i.Command(cfg(net), c); err == nil {
			t.Errorf("Command(network=%q) error = nil, want fail-closed error", net)
		}
		if _, _, err := i.TerminalCommand(cfg(net), c); err == nil {
			t.Errorf("TerminalCommand(network=%q) error = nil, want fail-closed error", net)
		}
	}
}

func TestNone_CommandNoPermissionBypass(t *testing.T) {
	// Host execution must NOT inject the sandbox permission bypass (Sandbox=false).
	cmd, err := NewIsolator().Command(cfg(netshared.ModeHost), provision.Contribution{})
	if err != nil {
		t.Fatalf("Command() error = %v", err)
	}
	if cmd[0] != "claude" {
		t.Fatalf("command[0] = %q, want claude", cmd[0])
	}
	if strings.Contains(strings.Join(cmd, " "), "bypassPermissions") {
		t.Error("host execution must not bypass permissions")
	}
}

func TestNone_AppliesContribution(t *testing.T) {
	c := provision.Contribution{
		PathEntries:  []string{"/nix/store/abc/bin"},
		Env:          map[string]string{"FOO": "bar"},
		InitCommands: []string{"echo hi"},
	}
	cmd, env, err := NewIsolator().hostCommand(cfg(netshared.ModeHost), c)
	if err != nil {
		t.Fatalf("hostCommand() error = %v", err)
	}

	if cmd[0] != "sh" || cmd[1] != "-c" {
		t.Fatalf("init command must wrap the agent in sh -c, got %v", cmd)
	}
	if !strings.Contains(strings.Join(cmd, " "), "echo hi") {
		t.Error("init command missing")
	}
	var foundPath, foundFoo bool
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=/nix/store/abc/bin:") {
			foundPath = true
		}
		if e == "FOO=bar" {
			foundFoo = true
		}
	}
	if !foundPath {
		t.Error("contribution PathEntries not prepended to PATH")
	}
	if !foundFoo {
		t.Error("contribution env not applied")
	}
}

func TestNone_CommandRejectsMissingProviderToken(t *testing.T) {
	cfg := cfg(netshared.ModeHost)
	cfg.Provider = &types.ModelProvider{AuthTokenEnv: "MISSING_NONE_TEST_TOKEN"}
	if _, err := NewIsolator().Command(cfg, provision.Contribution{}); err == nil {
		t.Error("Command() error = nil, want missing provider token error")
	}
}

func TestNone_Capabilities(t *testing.T) {
	i := NewIsolator()
	if i.HidesHost(false, "img") {
		t.Error("none always exposes the host")
	}
	if i.KernelIsolated("runsc") {
		t.Error("none has no kernel boundary")
	}
}
