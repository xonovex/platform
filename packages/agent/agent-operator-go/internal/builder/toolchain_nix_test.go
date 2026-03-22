package builder

import (
	"strings"
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestNixToolchain_InstallScript_PlainPackages(t *testing.T) {
	tc := NewNixToolchain(&agentv1alpha1.NixSpec{
		Packages: []string{"git", "curl"},
	})
	script := tc.installScript()
	if !strings.Contains(script, "nixpkgs#git") {
		t.Error("expected nixpkgs#git in script")
	}
	if !strings.Contains(script, "nixpkgs#curl") {
		t.Error("expected nixpkgs#curl in script")
	}
}

func TestNixToolchain_InstallScript_PackageSetExpansion(t *testing.T) {
	tc := NewNixToolchain(&agentv1alpha1.NixSpec{
		Packages: []string{"python"},
	})
	script := tc.installScript()
	if !strings.Contains(script, "nixpkgs#python312") {
		t.Error("expected nixpkgs#python312 from python set expansion")
	}
	if !strings.Contains(script, "nixpkgs#python312Packages.pip") {
		t.Error("expected nixpkgs#python312Packages.pip from python set expansion")
	}
	// The set name itself should not appear as a package ref
	if strings.Contains(script, "nixpkgs#python ") || strings.HasSuffix(script, "nixpkgs#python\n") {
		t.Error("set name 'python' should be expanded, not passed through")
	}
}

func TestNixToolchain_InstallScript_UnknownPreserved(t *testing.T) {
	tc := NewNixToolchain(&agentv1alpha1.NixSpec{
		Packages: []string{"my-custom-pkg"},
	})
	script := tc.installScript()
	if !strings.Contains(script, "nixpkgs#my-custom-pkg") {
		t.Error("expected unknown package name preserved as nixpkgs#my-custom-pkg")
	}
}

func TestNixToolchain_InstallScript_Deduplication(t *testing.T) {
	tc := NewNixToolchain(&agentv1alpha1.NixSpec{
		Packages: []string{"python312", "python"},
	})
	script := tc.installScript()
	// python312 appears in both explicit and python set — should be deduplicated
	count := strings.Count(script, "nixpkgs#python312 ")
	// Also check end of package list (last package doesn't have trailing space)
	if count > 1 {
		t.Errorf("expected python312 once, found %d occurrences", count)
	}
}

func TestNixToolchain_InstallScript_NodejsSet(t *testing.T) {
	tc := NewNixToolchain(&agentv1alpha1.NixSpec{
		Packages: []string{"nodejs"},
	})
	script := tc.installScript()
	for _, expected := range []string{"nodejs_24", "python312", "gnumake", "gcc"} {
		if !strings.Contains(script, "nixpkgs#"+expected) {
			t.Errorf("expected nixpkgs#%s from nodejs set expansion", expected)
		}
	}
}
