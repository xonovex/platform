package plugins

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/kube"
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

func TestHarnessResolvesEveryAgentOfTheSharedFacts(t *testing.T) {
	registry := New(agents.Embedded())
	for _, agent := range agents.Embedded().Types() {
		facts, err := registry.Harness(agentv1alpha1.AgentType(agent))
		if err != nil || facts.Type != agent {
			t.Errorf("Harness(%q) = %v, %v, want its facts", agent, facts, err)
		}
	}
	if got := len(agents.Embedded().Types()); got != 6 {
		t.Errorf("the shared facts hold %d agents, want 6", got)
	}
}

func TestHarnessRejectsAnUnknownAgent(t *testing.T) {
	if _, err := New(agents.Embedded()).Harness("bogus"); err == nil || !strings.Contains(err.Error(), "bogus") {
		t.Fatalf("Harness(unknown) err = %v, want an error naming the agent type", err)
	}
}

func TestSandboxSelectsKubeAndProvisionNone(t *testing.T) {
	box := Registry{}.Sandbox([]string{"gvisor"})

	if box.Isolation != kube.Method || box.Provision != provision.ProvisionNone {
		t.Fatalf("sandbox = %s/%s, want kube/none", box.Isolation, box.Provision)
	}
	confinement, err := box.Registry.Isolator(kube.Method)
	if err != nil {
		t.Fatalf("Isolator(kube) err = %v", err)
	}
	if !confinement.KernelIsolated("gvisor") || confinement.KernelIsolated("runc") {
		t.Error("the kube confinement must approve exactly the runtime classes of the policy")
	}
	provisioner, err := box.Registry.Provisioner(provision.ProvisionNone)
	if err != nil || provisioner.Pinned() {
		t.Errorf("Provisioner(none) = %v, %v, want the image provisioner", provisioner, err)
	}
	if methods := box.Registry.IsolationMethods(); len(methods) != 1 || methods[0] != kube.Method {
		t.Errorf("isolation methods = %v, want only kube", methods)
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
	registry := New(agents.Embedded())
	if !registry.HasToolchain(agentv1alpha1.ToolchainTypeNix) || registry.HasToolchain("bogus") {
		t.Error("HasToolchain must accept nix and reject an unregistered type")
	}
	if !registry.HasHarness(agentv1alpha1.AgentTypeOpencode) || !registry.HasHarness("codex") || registry.HasHarness("bogus") {
		t.Error("HasHarness must accept every agent of the shared facts and reject an unknown type")
	}
	if !registry.HasWorkspace(agentv1alpha1.WorkspaceTypeJujutsu) || registry.HasWorkspace("bogus") {
		t.Error("HasWorkspace must accept jj and reject an unregistered type")
	}
}
