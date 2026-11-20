package shared

import (
	"errors"
	"testing"

	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
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

func TestExecGit_ReturnsCapturedStdoutFromTheDirectory(t *testing.T) {
	host := systemtest.New()
	host.Outcomes[systemtest.InDir("/repo", "git rev-parse HEAD")] = systemtest.Outcome{Stdout: "abc123\n"}

	got, err := ExecGit(host, []string{"rev-parse", "HEAD"}, "/repo")
	if err != nil {
		t.Fatalf("ExecGit() error = %v", err)
	}
	if got != "abc123" {
		t.Errorf("ExecGit() = %q, want %q", got, "abc123")
	}
}

func TestExecGit_PropagatesTheCommandError(t *testing.T) {
	host := systemtest.New()
	host.Outcomes["git rev-parse HEAD"] = systemtest.Outcome{Err: errors.New("boom")}

	if _, err := ExecGit(host, []string{"rev-parse", "HEAD"}, "/repo"); err == nil {
		t.Fatal("ExecGit() must surface the command error")
	}
}

func TestGetCurrentBranchSync(t *testing.T) {
	tests := []struct {
		name    string
		outcome systemtest.Outcome
		want    string
	}{
		{"branch", systemtest.Outcome{Stdout: "feature/x\n"}, "feature/x"},
		// A detached HEAD makes git echo the literal "HEAD", which is not a branch.
		{"detached", systemtest.Outcome{Stdout: "HEAD\n"}, ""},
		{"outside a repository", systemtest.Outcome{Err: errors.New("not a repository")}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host := systemtest.New()
			host.Outcomes["git rev-parse --abbrev-ref HEAD"] = tt.outcome
			if got := GetCurrentBranchSync(host, "/repo"); got != tt.want {
				t.Errorf("GetCurrentBranchSync() = %q, want %q", got, tt.want)
			}
		})
	}
}
