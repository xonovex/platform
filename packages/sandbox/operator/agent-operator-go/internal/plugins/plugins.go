// Package plugins is the operator plug-in registry: the only package that
// imports the concrete axis leaves (provision/nix, workspace/{git,jj},
// harness/{claude,opencode}). The composition root (internal/wiring) passes a
// Registry to the cores as their resolver and catalog ports, so no core names a
// leaf or this package.
package plugins

import (
	"fmt"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/harness/claude"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/harness/opencode"
	harnessshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/harness/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provision/nix"
	provshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/git"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/jj"
	wsshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/shared"
)

// Registry resolves the registered leaf of each axis. Its zero value is ready
// to use.
type Registry struct{}

// ResolveToolchain returns the Toolchain for a spec, or nil if the spec is nil or its
// type is not registered.
func (Registry) ResolveToolchain(tc *agentv1alpha1.ToolchainSpec) provshared.Toolchain {
	if tc == nil {
		return nil
	}
	switch tc.Type {
	case agentv1alpha1.ToolchainTypeNix:
		return nix.New(tc.Nix)
	default:
		return nil
	}
}

// ResolveHarnessCommand returns the command builder for the given agent type.
func (Registry) ResolveHarnessCommand(agent agentv1alpha1.AgentType) (harnessshared.CommandBuilder, error) {
	switch agent {
	case agentv1alpha1.AgentTypeClaude:
		return claude.CommandBuilder{}, nil
	case agentv1alpha1.AgentTypeOpencode:
		return opencode.CommandBuilder{}, nil
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agent)
	}
}

// ResolveVCSStrategy returns the VCS strategy for the given workspace type; the empty
// type is git.
func (Registry) ResolveVCSStrategy(wsType agentv1alpha1.WorkspaceType) (wsshared.VCSStrategy, error) {
	switch wsType {
	case "", agentv1alpha1.WorkspaceTypeGit:
		return &git.Strategy{}, nil
	case agentv1alpha1.WorkspaceTypeJujutsu:
		return &jj.Strategy{}, nil
	default:
		return nil, fmt.Errorf("unsupported workspace type: %s", wsType)
	}
}

// HasToolchain reports whether a toolchain type is registered.
func (r Registry) HasToolchain(tcType agentv1alpha1.ToolchainType) bool {
	return r.ResolveToolchain(&agentv1alpha1.ToolchainSpec{Type: tcType}) != nil
}

// HasHarness reports whether an agent type is registered.
func (r Registry) HasHarness(agent agentv1alpha1.AgentType) bool {
	_, err := r.ResolveHarnessCommand(agent)
	return err == nil
}

// HasWorkspace reports whether a workspace type is registered.
func (r Registry) HasWorkspace(wsType agentv1alpha1.WorkspaceType) bool {
	_, err := r.ResolveVCSStrategy(wsType)
	return err == nil
}
