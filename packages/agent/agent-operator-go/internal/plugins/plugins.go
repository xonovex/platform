// Package plugins is the operator composition root: it holds the per-axis
// registries and is the ONLY package that imports the concrete axis leaves
// (provision/nix, workspace/{git,jj}, harness/{claude,opencode}). Each axis's
// shared/ package holds only its leaf-free port, so the cores stay neutral and
// symmetric with the CLI's internal/sandbox/plugins composition root.
package plugins

import (
	"fmt"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/harness/claude"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/harness/opencode"
	harnessshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/harness/shared"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/provision/nix"
	provshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/workspace/git"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/workspace/jj"
	wsshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/workspace/shared"
)

// --- provision (toolchain) ---

var toolchainFactories = map[agentv1alpha1.ToolchainType]func(*agentv1alpha1.ToolchainSpec) provshared.Toolchain{
	agentv1alpha1.ToolchainTypeNix: func(tc *agentv1alpha1.ToolchainSpec) provshared.Toolchain { return nix.New(tc.Nix) },
}

// ResolveToolchain returns the Toolchain for a spec, or nil if the spec is nil or
// its type is not registered.
func ResolveToolchain(tc *agentv1alpha1.ToolchainSpec) provshared.Toolchain {
	if tc == nil {
		return nil
	}
	factory, ok := toolchainFactories[tc.Type]
	if !ok {
		return nil
	}
	return factory(tc)
}

// --- harness (command) ---

var harnessCommands = map[agentv1alpha1.AgentType]harnessshared.CommandBuilder{
	agentv1alpha1.AgentTypeClaude:   &claude.CommandBuilder{},
	agentv1alpha1.AgentTypeOpencode: &opencode.CommandBuilder{},
}

// GetHarnessCommand returns the command builder for the given agent type.
func GetHarnessCommand(agent agentv1alpha1.AgentType) (harnessshared.CommandBuilder, error) {
	b, ok := harnessCommands[agent]
	if !ok {
		return nil, fmt.Errorf("unsupported agent type: %s", agent)
	}
	return b, nil
}

// --- workspace (VCS strategy) ---

var vcsStrategies = map[agentv1alpha1.WorkspaceType]wsshared.VCSStrategy{
	"":                                 &git.Strategy{},
	agentv1alpha1.WorkspaceTypeGit:     &git.Strategy{},
	agentv1alpha1.WorkspaceTypeJujutsu: &jj.Strategy{},
}

// GetVCSStrategy returns the VCS strategy for the given workspace type.
func GetVCSStrategy(wsType agentv1alpha1.WorkspaceType) (wsshared.VCSStrategy, error) {
	s, ok := vcsStrategies[wsType]
	if !ok {
		return nil, fmt.Errorf("unsupported workspace type: %s", wsType)
	}
	return s, nil
}
