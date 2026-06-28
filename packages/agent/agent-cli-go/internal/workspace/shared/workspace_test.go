package shared

import (
	"testing"

	wsp "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/workspace"
)

func TestSanitizeBranchName(t *testing.T) {
	cases := map[string]string{
		"feature/foo":  "feature-foo",
		"a//b":         "a-b",
		"-trim-":       "trim",
		"weird@chars!": "weird-chars",
	}
	for in, want := range cases {
		if got := SanitizeBranchName(in); got != want {
			t.Errorf("SanitizeBranchName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGetDefaultDir(t *testing.T) {
	if got := GetDefaultDir("feature/x", "myrepo"); got != "../myrepo-feature-x" {
		t.Errorf("GetDefaultDir = %q, want ../myrepo-feature-x", got)
	}
}

func TestBuildBindPaths(t *testing.T) {
	if got := BuildBindPaths([]string{"/a"}, ""); len(got) != 1 {
		t.Errorf("BuildBindPaths with empty repo = %v, want unchanged", got)
	}
	got := BuildBindPaths([]string{"/a"}, "/repo")
	if len(got) != 2 || got[1] != "/repo" {
		t.Errorf("BuildBindPaths = %v, want source repo appended", got)
	}
}

// TestVCSType_ConsumedFromShared confirms the CLI consumes the VCSType enum from
// the shared module (no local redefinition).
func TestVCSType_ConsumedFromShared(t *testing.T) {
	cases := []struct {
		vcs   wsp.VCSType
		valid bool
	}{
		{wsp.VCSGit, true},
		{wsp.VCSJujutsu, true},
		{"unknown", false},
		{"", false},
	}
	for _, tt := range cases {
		if got := tt.vcs.IsValid(); got != tt.valid {
			t.Errorf("VCSType(%q).IsValid() = %v, want %v", tt.vcs, got, tt.valid)
		}
	}
}
