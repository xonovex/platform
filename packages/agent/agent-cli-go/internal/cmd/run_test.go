package cmd

import (
	"testing"

	netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"
)

func TestParseNetwork(t *testing.T) {
	if _, err := netshared.ParseMode("proxy"); err != nil {
		t.Errorf("ParseMode(proxy) = %v", err)
	}
	if _, err := netshared.ParseMode("bogus"); err == nil {
		t.Error("ParseMode(bogus) = nil, want error")
	}
}

// TestProxyEnv_GatedByMode locks the fix that a host-side AGENT_SANDBOX_PROXY is
// only injected for --network proxy, never for host/none.
func TestProxyEnv_GatedByMode(t *testing.T) {
	t.Setenv("AGENT_SANDBOX_PROXY", "http://127.0.0.1:3128")
	if env := proxyEnv(netshared.ModeHost); env != nil {
		t.Errorf("proxyEnv(host) = %v, want nil even with AGENT_SANDBOX_PROXY set", env)
	}
	if env := proxyEnv(netshared.ModeNone); env != nil {
		t.Errorf("proxyEnv(none) = %v, want nil", env)
	}
	if env := proxyEnv(netshared.ModeProxy); env == nil || env["HTTPS_PROXY"] != "http://127.0.0.1:3128" {
		t.Errorf("proxyEnv(proxy) = %v, want the proxy env", env)
	}
}

// TestResolveAxes_DockerRuntimeWiresKernelIsolation confirms the new
// --isolation-docker-runtime flag makes the kernel-isolation capability reachable
// (previously the hardcoded empty runtime left it dead).
func TestResolveAxes_DockerRuntimeWiresKernelIsolation(t *testing.T) {
	axes, err := resolveAxes(flags{isolation: "docker", isolationDockerRuntime: "runsc"})
	if err != nil {
		t.Fatalf("resolveAxes = %v", err)
	}
	if !axes.Isolation.KernelIsolated("runsc") {
		t.Error("docker + --isolation-docker-runtime runsc must be kernel-isolated")
	}
	if axes.Runtime != "runsc" {
		t.Errorf("axes.Runtime = %q, want runsc", axes.Runtime)
	}
}

// TestResolveAxes_PinnedComboDefault confirms that requiring a pinned toolchain
// with no explicit cell selects the bwrap × nix combo.
func TestResolveAxes_PinnedComboDefault(t *testing.T) {
	axes, err := resolveAxes(flags{requirePinned: true})
	if err != nil {
		t.Fatalf("resolveAxes = %v", err)
	}
	if axes.IsolationName != "bwrap" || axes.ProvisionName != "nix" {
		t.Errorf("pinned default = (%s, %s), want (bwrap, nix)", axes.IsolationName, axes.ProvisionName)
	}
}

func TestWorktreeBranchValidation(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		wantErr bool
	}{
		{"valid simple", "feature/my-work", false},
		{"valid main", "main", false},
		{"valid release", "release-1.0", false},
		{"invalid semicolon", "branch;rm -rf /", true},
		{"invalid pipe", "branch|evil", true},
		{"invalid dollar", "branch$(whoami)", true},
		{"invalid backtick", "branch`id`", true},
		{"invalid spaces", "branch name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateBranch(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBranch(%q) error = %v, wantErr %v", tt.branch, err, tt.wantErr)
			}
		})
	}
}
