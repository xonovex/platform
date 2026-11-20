// Package plugins is the operator plug-in registry: the only package that
// imports the concrete axis leaves (isolation/kube, provision/{none,nix},
// workspace/{git,jj}). The harnesses are the shared agent facts. The
// composition root (internal/wiring) passes a Registry to the cores as their
// resolver and catalog ports, so no core names a leaf or this package.
package plugins

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	launchprov "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/kube"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/launch"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provision/nix"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provision/none"
	provshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/git"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/jj"
	wsshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/shared"
)

// Registry resolves the registered leaf of each axis and the facts of each
// harness.
type Registry struct {
	harnesses agents.Catalog
}

// New returns the Registry whose harnesses are the facts of harnesses.
func New(harnesses agents.Catalog) Registry {
	return Registry{harnesses: harnesses}
}

// Sandbox returns the kube isolation, which approves runtimeClasses as
// sandboxed runtimes, and provision=none: the pod image, a nix toolchain image
// included, supplies every tool.
func (Registry) Sandbox(runtimeClasses []string) launch.Sandbox {
	confinement := kube.New(runtimeClasses)
	registry := sandbox.NewRegistry[isoshared.Confinement]().
		RegisterIsolator(kube.Method, func() isoshared.Confinement { return confinement }).
		RegisterProvisioner(provision.ProvisionNone, func() launchprov.Provisioner { return none.Provisioner{} })
	return launch.Sandbox{Registry: registry, Isolation: kube.Method, Provision: provision.ProvisionNone}
}

// Harness returns the facts of an agent type.
func (r Registry) Harness(agent agentv1alpha1.AgentType) (*types.AgentConfig, error) {
	return r.harnesses.Get(types.AgentType(agent))
}

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
	_, err := r.Harness(agent)
	return err == nil
}

// HasWorkspace reports whether a workspace type is registered.
func (r Registry) HasWorkspace(wsType agentv1alpha1.WorkspaceType) bool {
	_, err := r.ResolveVCSStrategy(wsType)
	return err == nil
}
