package plugins

import (
	"strings"
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

func TestResolveToolchain(t *testing.T) {
	tc := &agentv1alpha1.ToolchainSpec{
		Type: agentv1alpha1.ToolchainTypeNix,
		Nix:  &agentv1alpha1.NixSpec{Image: "ghcr.io/xonovex/agent@sha256:abc"},
	}
	tcl := Registry{}.ResolveToolchain(tc)
	if tcl == nil || tcl.Image() != "ghcr.io/xonovex/agent@sha256:abc" {
		t.Fatalf("Registry{}.ResolveToolchain(nix).Image() = %v, want the pre-built image", tcl)
	}
	if !tcl.Pinned() {
		t.Error("nix toolchain must report Pinned()=true")
	}
	if (Registry{}).ResolveToolchain(nil) != nil {
		t.Error("Registry{}.ResolveToolchain(nil) must be nil")
	}
	if (Registry{}).ResolveToolchain(&agentv1alpha1.ToolchainSpec{Type: "bogus"}) != nil {
		t.Error("Registry{}.ResolveToolchain(unknown type) must be nil")
	}
}

func TestResolveHarnessCommandResolvesEveryRegisteredAgent(t *testing.T) {
	for _, agent := range []agentv1alpha1.AgentType{
		agentv1alpha1.AgentTypeClaude,
		agentv1alpha1.AgentTypeOpencode,
	} {
		builder, err := Registry{}.ResolveHarnessCommand(agent)
		if err != nil {
			t.Errorf("ResolveHarnessCommand(%q) returned error %v, want a builder", agent, err)
			continue
		}
		if builder == nil {
			t.Errorf("ResolveHarnessCommand(%q) returned a nil builder", agent)
		}
	}
}

func TestResolveHarnessCommandRejectsUnregisteredAgent(t *testing.T) {
	builder, err := Registry{}.ResolveHarnessCommand("bogus")
	if err == nil {
		t.Fatal("ResolveHarnessCommand(unregistered) must return an error")
	}
	if builder != nil {
		t.Errorf("ResolveHarnessCommand(unregistered) returned builder %v, want nil", builder)
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error %q must name the unsupported agent type", err)
	}
}

// An empty workspace type means git, so a spec that omits it still resolves.
func TestResolveVCSStrategyResolvesEveryRegisteredWorkspace(t *testing.T) {
	for _, wsType := range []agentv1alpha1.WorkspaceType{
		"",
		agentv1alpha1.WorkspaceTypeGit,
		agentv1alpha1.WorkspaceTypeJujutsu,
	} {
		strategy, err := Registry{}.ResolveVCSStrategy(wsType)
		if err != nil {
			t.Errorf("ResolveVCSStrategy(%q) returned error %v, want a strategy", wsType, err)
			continue
		}
		if strategy == nil {
			t.Errorf("ResolveVCSStrategy(%q) returned a nil strategy", wsType)
			continue
		}
		if strategy.InitContainerName() == "" {
			t.Errorf("ResolveVCSStrategy(%q) returned a strategy with no init container name", wsType)
		}
	}
}

func TestResolveVCSStrategyDefaultsEmptyTypeToGit(t *testing.T) {
	defaulted, err := Registry{}.ResolveVCSStrategy("")
	if err != nil {
		t.Fatalf("ResolveVCSStrategy(\"\") returned error %v", err)
	}
	explicit, err := Registry{}.ResolveVCSStrategy(agentv1alpha1.WorkspaceTypeGit)
	if err != nil {
		t.Fatalf("ResolveVCSStrategy(git) returned error %v", err)
	}
	if defaulted.InitContainerName() != explicit.InitContainerName() {
		t.Errorf(
			"empty workspace type resolved to %q, want git's %q",
			defaulted.InitContainerName(),
			explicit.InitContainerName(),
		)
	}
}

func TestResolveVCSStrategyRejectsUnregisteredWorkspace(t *testing.T) {
	strategy, err := Registry{}.ResolveVCSStrategy("bogus")
	if err == nil {
		t.Fatal("ResolveVCSStrategy(unregistered) must return an error")
	}
	if strategy != nil {
		t.Errorf("ResolveVCSStrategy(unregistered) returned strategy %v, want nil", strategy)
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error %q must name the unsupported workspace type", err)
	}
}

func TestRegistryReportsRegisteredTypes(t *testing.T) {
	registry := Registry{}
	if !registry.HasToolchain(agentv1alpha1.ToolchainTypeNix) || registry.HasToolchain("bogus") {
		t.Error("HasToolchain must accept nix and reject an unregistered type")
	}
	if !registry.HasHarness(agentv1alpha1.AgentTypeOpencode) || registry.HasHarness("bogus") {
		t.Error("HasHarness must accept opencode and reject an unregistered type")
	}
	if !registry.HasWorkspace(agentv1alpha1.WorkspaceTypeJujutsu) || registry.HasWorkspace("bogus") {
		t.Error("HasWorkspace must accept jj and reject an unregistered type")
	}
}
