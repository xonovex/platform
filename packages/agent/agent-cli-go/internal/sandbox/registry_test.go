package sandbox

import (
	"os"
	"os/exec"
	"testing"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func TestGetExecutorNixFlake(t *testing.T) {
	executor, err := GetExecutor(types.SandboxNixFlake)
	if err != nil {
		t.Fatalf("GetExecutor(nixflake) error = %v, want nil", err)
	}
	if executor == nil {
		t.Fatal("GetExecutor(nixflake) returned nil executor")
	}
}

func TestGetExecutorUnknown(t *testing.T) {
	if _, err := GetExecutor(types.SandboxMethod("bogus")); err == nil {
		t.Fatal("GetExecutor(bogus) error = nil, want error")
	}
}

func TestSelectExecutor_RequirePinnedToolchain(t *testing.T) {
	deny := types.SandboxPolicy{RequirePinnedToolchain: true}
	cases := []struct {
		name    string
		method  types.SandboxMethod
		image   string
		policy  types.SandboxPolicy
		wantErr bool
	}{
		{"deny+bwrap rejects leak", types.SandboxBwrap, "", deny, true},
		{"deny+none rejects host exec", types.SandboxNone, "", deny, true},
		{"deny+docker without image", types.SandboxDocker, "", deny, true},
		{"deny+compose without image", types.SandboxCompose, "", deny, true},
		{"deny+docker with image", types.SandboxDocker, "alpine:3.20", deny, false},
		{"deny+nix ok", types.SandboxNix, "", deny, false},
		{"deny+nixflake ok", types.SandboxNixFlake, "", deny, false},
		{"policy off keeps bwrap", types.SandboxBwrap, "", types.SandboxPolicy{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exec, method, err := SelectExecutor(tc.method, tc.image, tc.policy)
			if (err != nil) != tc.wantErr {
				t.Fatalf("SelectExecutor(%s) err=%v, wantErr=%v", tc.method, err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if exec == nil {
				t.Fatalf("SelectExecutor(%s) returned nil executor", tc.method)
			}
			if method != tc.method {
				t.Fatalf("SelectExecutor(%s) method=%s, want %s", tc.method, method, tc.method)
			}
		})
	}
}

func TestTierIsolation_Classification(t *testing.T) {
	want := map[types.SandboxMethod]Isolation{
		types.SandboxNone:     IsolationHostToolsLeaked,
		types.SandboxBwrap:    IsolationHostToolsLeaked,
		types.SandboxDocker:   IsolationContainerPinned,
		types.SandboxCompose:  IsolationContainerPinned,
		types.SandboxNix:      IsolationHostToolsUnreachable,
		types.SandboxNixFlake: IsolationHostToolsUnreachable,
	}
	for m, w := range want {
		if got := tierIsolation(m); got != w {
			t.Errorf("tierIsolation(%s)=%d, want %d", m, got, w)
		}
	}
}

// TestRequirePinnedToolchain_NixflakeHostToolsUnreachable drives the policy seam
// end-to-end against the live nixflake tier: selection under the policy must
// mandate nixflake, and a probe run inside that sandbox must find host /usr/bin
// empty and a host-only binary unreachable. It is a real namespace check (the
// probe lists the live mount), so a regression that re-binds /usr fails it.
//
// Gated like nixflake's live smoke: a real `nix develop` needs nix, bwrap,
// /nix/store, and a flake fixture to enter.
//
//	NIXFLAKE_INTEGRATION=1 NIXFLAKE_INTEGRATION_FLAKE=/path/to/repo \
//	NIXFLAKE_INTEGRATION_SHELL=go go test ./internal/sandbox -run NixflakeHostToolsUnreachable -v
func TestRequirePinnedToolchain_NixflakeHostToolsUnreachable(t *testing.T) {
	for _, bin := range []string{"nix", "bwrap"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not available", bin)
		}
	}
	if _, err := os.Stat("/nix/store"); os.IsNotExist(err) {
		t.Skip("/nix/store not present")
	}
	flakeDir := os.Getenv("NIXFLAKE_INTEGRATION_FLAKE")
	if os.Getenv("NIXFLAKE_INTEGRATION") != "1" || flakeDir == "" {
		t.Skip("set NIXFLAKE_INTEGRATION=1 and NIXFLAKE_INTEGRATION_FLAKE=<flake dir> to run the live policy smoke")
	}

	// The policy seam under test: an unspecified method must resolve to nixflake.
	executor, method, err := SelectExecutor("", "", types.SandboxPolicy{RequirePinnedToolchain: true})
	if err != nil {
		t.Fatalf("select under policy: %v", err)
	}
	if method != types.SandboxNixFlake {
		t.Fatalf("policy forced %s, want nixflake", method)
	}

	// The probe is the "agent": opencode passes AgentArgs through verbatim, so the
	// nixflake executor runs `nix develop ... --command sh -c <probe>`. Under the
	// pinned tier the host /usr is never bound, so /usr/bin is empty inside the
	// namespace and a host-only binary does not resolve.
	probe := `test -z "$(ls -A /usr/bin 2>/dev/null)" || exit 1; ! command -v fdisk >/dev/null 2>&1 || exit 1`
	cfg := &types.SandboxConfig{
		Method:    method,
		Policy:    types.SandboxPolicy{RequirePinnedToolchain: true},
		WorkDir:   flakeDir,
		RepoDir:   flakeDir,
		Network:   true,
		Agent:     &types.AgentConfig{Type: types.AgentOpencode, Binary: "sh"},
		AgentArgs: []string{"-c", probe},
	}
	if shell := os.Getenv("NIXFLAKE_INTEGRATION_SHELL"); shell != "" {
		cfg.Image = "nixflake:shell=" + shell
	}
	if code, err := executor.Execute(cfg); err != nil || code != 0 {
		t.Fatalf("host tools reachable under pinned tier: code=%d err=%v", code, err)
	}
}
