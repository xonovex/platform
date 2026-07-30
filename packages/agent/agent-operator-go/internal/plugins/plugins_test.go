package plugins

import (
	"strings"
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestResolveToolchain(t *testing.T) {
	tc := &agentv1alpha1.ToolchainSpec{
		Type: agentv1alpha1.ToolchainTypeNix,
		Nix:  &agentv1alpha1.NixSpec{Image: "ghcr.io/xonovex/agent@sha256:abc"},
	}
	tcl := ResolveToolchain(tc)
	if tcl == nil || tcl.Image() != "ghcr.io/xonovex/agent@sha256:abc" {
		t.Fatalf("ResolveToolchain(nix).Image() = %v, want the pre-built image", tcl)
	}
	if !tcl.Pinned() {
		t.Error("nix toolchain must report Pinned()=true")
	}
	if ResolveToolchain(nil) != nil {
		t.Error("ResolveToolchain(nil) must be nil")
	}
	if ResolveToolchain(&agentv1alpha1.ToolchainSpec{Type: "bogus"}) != nil {
		t.Error("ResolveToolchain(unknown type) must be nil")
	}
}

func TestGetHarnessCommandResolvesEveryRegisteredAgent(t *testing.T) {
	for _, agent := range []agentv1alpha1.AgentType{
		agentv1alpha1.AgentTypeClaude,
		agentv1alpha1.AgentTypeOpencode,
	} {
		builder, err := GetHarnessCommand(agent)
		if err != nil {
			t.Errorf("GetHarnessCommand(%q) returned error %v, want a builder", agent, err)
			continue
		}
		if builder == nil {
			t.Errorf("GetHarnessCommand(%q) returned a nil builder", agent)
		}
	}
}

func TestGetHarnessCommandRejectsUnregisteredAgent(t *testing.T) {
	builder, err := GetHarnessCommand("bogus")
	if err == nil {
		t.Fatal("GetHarnessCommand(unregistered) must return an error")
	}
	if builder != nil {
		t.Errorf("GetHarnessCommand(unregistered) returned builder %v, want nil", builder)
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error %q must name the unsupported agent type", err)
	}
}

// An empty workspace type means git, so a spec that omits it still resolves.
func TestGetVCSStrategyResolvesEveryRegisteredWorkspace(t *testing.T) {
	for _, wsType := range []agentv1alpha1.WorkspaceType{
		"",
		agentv1alpha1.WorkspaceTypeGit,
		agentv1alpha1.WorkspaceTypeJujutsu,
	} {
		strategy, err := GetVCSStrategy(wsType)
		if err != nil {
			t.Errorf("GetVCSStrategy(%q) returned error %v, want a strategy", wsType, err)
			continue
		}
		if strategy == nil {
			t.Errorf("GetVCSStrategy(%q) returned a nil strategy", wsType)
			continue
		}
		if strategy.InitContainerName() == "" {
			t.Errorf("GetVCSStrategy(%q) returned a strategy with no init container name", wsType)
		}
	}
}

func TestGetVCSStrategyDefaultsEmptyTypeToGit(t *testing.T) {
	defaulted, err := GetVCSStrategy("")
	if err != nil {
		t.Fatalf("GetVCSStrategy(\"\") returned error %v", err)
	}
	explicit, err := GetVCSStrategy(agentv1alpha1.WorkspaceTypeGit)
	if err != nil {
		t.Fatalf("GetVCSStrategy(git) returned error %v", err)
	}
	if defaulted.InitContainerName() != explicit.InitContainerName() {
		t.Errorf(
			"empty workspace type resolved to %q, want git's %q",
			defaulted.InitContainerName(),
			explicit.InitContainerName(),
		)
	}
}

func TestGetVCSStrategyRejectsUnregisteredWorkspace(t *testing.T) {
	strategy, err := GetVCSStrategy("bogus")
	if err == nil {
		t.Fatal("GetVCSStrategy(unregistered) must return an error")
	}
	if strategy != nil {
		t.Errorf("GetVCSStrategy(unregistered) returned strategy %v, want nil", strategy)
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error %q must name the unsupported workspace type", err)
	}
}
