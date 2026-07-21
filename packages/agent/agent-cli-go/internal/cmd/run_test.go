package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	isoshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"
)

func TestParseNetwork(t *testing.T) {
	if _, err := netshared.ParseMode("host"); err != nil {
		t.Errorf("ParseMode(host) = %v", err)
	}
	if _, err := netshared.ParseMode("proxy"); !errors.Is(err, netshared.ErrProxyEnforcementUnavailable) {
		t.Errorf("ParseMode(proxy) error = %v, want %v", err, netshared.ErrProxyEnforcementUnavailable)
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

func TestPrepareWorkspaceWithoutWorktreeUsesSourceDirectory(t *testing.T) {
	workDir := t.TempDir()

	workspace, err := prepareWorkspace(runOptions{}, workDir, false)

	if err != nil {
		t.Fatalf("prepareWorkspace() error = %v", err)
	}
	if workspace.sourceRepoDir != workDir || workspace.executionDir != workDir || workspace.displayDir != workDir {
		t.Errorf("workspace = %+v, want every directory to be %q", workspace, workDir)
	}
}

func TestRunAgentDryRunBuildsHostCommand(t *testing.T) {
	workDir := t.TempDir()
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "claude"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("create agent executable: %v", err)
	}
	t.Setenv("PATH", binDir)
	configPath := filepath.Join(workDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	command := &cobra.Command{}
	command.Flags().Bool("verbose", false, "")
	command.Flags().String("isolation", "none", "")
	command.Flags().String("provision", "none", "")
	options := runOptions{
		agent:     "claude",
		isolation: "none",
		provision: "none",
		network:   "host",
		workDir:   workDir,
		config:    configPath,
		vcs:       "git",
		dryRun:    true,
	}

	err := runAgent(command, []string{"review"}, options)

	if err != nil {
		t.Fatalf("runAgent() error = %v", err)
	}
}

func TestRunAgentRejectsMalformedCustomEnvironment(t *testing.T) {
	workDir := t.TempDir()
	configPath := filepath.Join(workDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	command := &cobra.Command{}
	command.Flags().Bool("verbose", false, "")
	command.Flags().String("isolation", "none", "")
	command.Flags().String("provision", "none", "")
	options := runOptions{
		agent:     "claude",
		isolation: "none",
		provision: "none",
		network:   "host",
		workDir:   workDir,
		config:    configPath,
		vcs:       "git",
		dryRun:    true,
		customEnv: []string{"MISSING_EQUALS"},
	}

	err := runAgent(command, nil, options)

	if err == nil || !strings.Contains(err.Error(), "must use KEY=VALUE") {
		t.Fatalf("runAgent() error = %v, want custom environment validation error", err)
	}
}

func TestRunAgentRejectsUnknownAgent(t *testing.T) {
	command := &cobra.Command{}
	command.Flags().Bool("verbose", false, "")

	err := runAgent(command, nil, runOptions{agent: "unknown"})

	if err == nil {
		t.Fatal("runAgent() error = nil, want unknown-agent error")
	}
}

func TestResolveProviderPrefersCommandOption(t *testing.T) {
	provider, err := resolveProvider(types.AgentClaude, "gemini", "glm")

	if err != nil {
		t.Fatalf("resolveProvider() error = %v", err)
	}
	if provider == nil || provider.Name != "gemini" {
		t.Fatalf("resolveProvider() = %+v, want gemini", provider)
	}
}

func TestResolveProviderRejectsUnknownProvider(t *testing.T) {
	provider, err := resolveProvider(types.AgentClaude, "unknown", "")

	if err == nil || provider != nil {
		t.Fatalf("resolveProvider() = (%+v, %v), want nil provider and error", provider, err)
	}
}

func TestProvisionInputPreservesInitCommandsWithoutNix(t *testing.T) {
	options := runOptions{initCommands: []string{"npm install"}}
	agent := &types.AgentConfig{Binary: "claude", NixPackage: "claude-code"}

	input, err := provisionInput(provision.ProvisionNone, options, agent, t.TempDir(), t.TempDir())

	if err != nil {
		t.Fatalf("provisionInput() error = %v", err)
	}
	if len(input.InitCommands) != 1 || input.InitCommands[0] != "npm install" {
		t.Fatalf("InitCommands = %v, want npm install", input.InitCommands)
	}
}

func TestProvisionInputRejectsUnknownNixSource(t *testing.T) {
	options := runOptions{nixSource: "unknown"}
	agent := &types.AgentConfig{Binary: "claude", NixPackage: "claude-code"}

	_, err := provisionInput(provision.ProvisionNix, options, agent, t.TempDir(), t.TempDir())

	if err == nil {
		t.Fatal("provisionInput() error = nil, want invalid-source error")
	}
}

func TestProvisionInputAddsSelectedAgentToNixPackages(t *testing.T) {
	options := runOptions{
		nixSource:   "packages",
		nixRev:      strings.Repeat("a", 40),
		nixPackages: []string{"ripgrep"},
	}
	agent := &types.AgentConfig{Binary: "opencode", NixPackage: "opencode"}

	input, err := provisionInput(provision.ProvisionNix, options, agent, t.TempDir(), t.TempDir())

	if err != nil {
		t.Fatalf("provisionInput() error = %v", err)
	}
	if !slices.Equal(input.NixSource.Packages, []string{"ripgrep", "opencode"}) {
		t.Errorf("Nix packages = %v, want ripgrep and opencode", input.NixSource.Packages)
	}
}

func TestProvisionInputDoesNotDuplicateSelectedAgentPackage(t *testing.T) {
	options := runOptions{
		nixSource:   "packages",
		nixRev:      strings.Repeat("a", 40),
		nixPackages: []string{"claude-code"},
	}
	agent := &types.AgentConfig{Binary: "claude", NixPackage: "claude-code"}

	input, err := provisionInput(provision.ProvisionNix, options, agent, t.TempDir(), t.TempDir())

	if err != nil {
		t.Fatalf("provisionInput() error = %v", err)
	}
	if !slices.Equal(input.NixSource.Packages, []string{"claude-code"}) {
		t.Errorf("Nix packages = %v, want one claude-code", input.NixSource.Packages)
	}
}

func TestValidateAgentExecutableRejectsUnprovisionedDefaultDockerImage(t *testing.T) {
	agent := &types.AgentConfig{Binary: "claude", NixPackage: "claude-code"}
	axes := resolvedAxes{IsolationName: isolation.IsolationDocker, ProvisionName: provision.ProvisionNone}

	err := validateAgentExecutable(axes, agent, provision.Contribution{})

	if err == nil || !strings.Contains(err.Error(), "default docker image") {
		t.Fatalf("validateAgentExecutable() error = %v, want default-image error", err)
	}
}

func TestValidateAgentExecutableRejectsUnprovisionedBwrapWithoutPassthrough(t *testing.T) {
	agent := &types.AgentConfig{Binary: "claude", NixPackage: "claude-code"}
	axes := resolvedAxes{IsolationName: isolation.IsolationBwrap, ProvisionName: provision.ProvisionNone}

	err := validateAgentExecutable(axes, agent, provision.Contribution{})

	if err == nil || !strings.Contains(err.Error(), "host passthrough is disabled") {
		t.Fatalf("validateAgentExecutable() error = %v, want passthrough error", err)
	}
}

func TestValidateAgentExecutableAcceptsProvisionedNixBinary(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("create bin directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "opencode"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("create agent executable: %v", err)
	}
	agent := &types.AgentConfig{Binary: "opencode", NixPackage: "opencode"}
	axes := resolvedAxes{IsolationName: isolation.IsolationBwrap, ProvisionName: provision.ProvisionNix}
	contribution := provision.Contribution{PathEntries: []string{binDir}}

	err := validateAgentExecutable(axes, agent, contribution)

	if err != nil {
		t.Fatalf("validateAgentExecutable() error = %v", err)
	}
}

func TestValidateAgentExecutableRejectsNixClosureWithoutAgent(t *testing.T) {
	agent := &types.AgentConfig{Binary: "opencode", NixPackage: "opencode"}
	axes := resolvedAxes{IsolationName: isolation.IsolationBwrap, ProvisionName: provision.ProvisionNix}
	contribution := provision.Contribution{PathEntries: []string{t.TempDir()}}

	err := validateAgentExecutable(axes, agent, contribution)

	if err == nil || !strings.Contains(err.Error(), "did not supply") {
		t.Fatalf("validateAgentExecutable() error = %v, want missing-agent error", err)
	}
}

func TestExecuteWithTerminalRejectsUnknownType(t *testing.T) {
	options := runOptions{terminal: "unknown"}

	err := executeWithTerminal(resolvedAxes{}, isoshared.RunConfig{}, provision.Contribution{}, t.TempDir(), false, options)

	if err == nil {
		t.Fatal("executeWithTerminal() error = nil, want unknown-terminal error")
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
