package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/policy"
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

func TestResolveAxes_ProxyFailsClosed(t *testing.T) {
	if _, err := resolveAxes(flags{network: "proxy"}); !errors.Is(err, netshared.ErrProxyEnforcementUnavailable) {
		t.Fatalf("resolveAxes(proxy) error = %v, want %v", err, netshared.ErrProxyEnforcementUnavailable)
	}
}

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

func TestResolveAxes_PinnedComboDefault(t *testing.T) {
	axes, err := resolveAxes(flags{requirePinnedProvision: true})
	if err != nil {
		t.Fatalf("resolveAxes = %v", err)
	}
	if axes.IsolationName != "bwrap" || axes.ProvisionName != "nix" {
		t.Errorf("pinned default = (%s, %s), want (bwrap, nix)", axes.IsolationName, axes.ProvisionName)
	}
}

func TestPrepareWorkspaceDryRunDoesNotCreateDirectory(t *testing.T) {
	repoDir := t.TempDir()
	target := filepath.Join(repoDir, "worktree")
	workspace, err := prepareWorkspace(runOptions{
		worktreeBranch: "feature/dry-run",
		worktreeDir:    target,
		vcs:            "git",
		dryRun:         true,
	}, repoDir, false)
	if err != nil {
		t.Fatalf("prepareWorkspace() error = %v", err)
	}
	if workspace.executionDir != repoDir {
		t.Errorf("executionDir = %q, want source repo %q", workspace.executionDir, repoDir)
	}
	if workspace.displayDir != target {
		t.Errorf("displayDir = %q, want target %q", workspace.displayDir, target)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("dry run created worktree directory %q", target)
	}
}

func TestResolveAxes_DockerPinnedPolicyRequiresDigest(t *testing.T) {
	digest := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	base := flags{
		isolation:              "docker",
		provision:              "none",
		requirePinnedProvision: true,
		isolationChanged:       true,
		provisionChanged:       true,
	}

	pinned := base
	pinned.image = "ghcr.io/xonovex/agent@sha256:" + digest
	if _, err := resolveAxes(pinned); err != nil {
		t.Fatalf("resolveAxes(digest image) error = %v", err)
	}

	mutable := base
	mutable.image = "ghcr.io/xonovex/agent:latest"
	if _, err := resolveAxes(mutable); !errors.Is(err, policy.ErrPinnedProvisionUnmet) {
		t.Fatalf("resolveAxes(mutable image) error = %v, want %v", err, policy.ErrPinnedProvisionUnmet)
	}
}

func TestFlagsPolicyMapsGuaranteesIndependently(t *testing.T) {
	tests := []struct {
		name  string
		flags flags
		want  policy.SandboxPolicy
	}{
		{"pinned", flags{requirePinnedProvision: true}, policy.SandboxPolicy{RequirePinnedProvision: true}},
		{"host tools", flags{requireHostToolsUnreachable: true}, policy.SandboxPolicy{RequireHostToolsUnreachable: true}},
		{"egress", flags{requireEgressRestricted: true}, policy.SandboxPolicy{RequireEgressRestricted: true}},
		{"kernel", flags{requireKernelIsolation: true}, policy.SandboxPolicy{RequireKernelIsolation: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flags.policy(); got != tt.want {
				t.Errorf("policy() = %+v, want %+v", got, tt.want)
			}
		})
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
