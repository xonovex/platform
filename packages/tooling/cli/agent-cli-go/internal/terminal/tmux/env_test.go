package tmux

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

// failingHost makes the launch script operations fail on demand.
type failingHost struct {
	*systemtest.Fake
	createErr error
	writeErr  error
}

func (h failingHost) CreateTemp(pattern string) (string, error) {
	if h.createErr != nil {
		return "", h.createErr
	}
	return h.Fake.CreateTemp(pattern)
}

func (h failingHost) WriteFile(path string, data []byte, perm fs.FileMode) error {
	if h.writeErr != nil {
		return h.writeErr
	}
	return h.Fake.WriteFile(path, data, perm)
}

func TestFilterEnvironmentRejectsInvalidAndReservedNames(t *testing.T) {
	environment := []string{
		"PATH=/usr/bin",
		"SAFE_VALUE=content",
		"UID=1000",
		"9INVALID=value",
		"BAD;touch /tmp/injected=value",
		"missing-separator",
		"=empty",
	}

	filtered := filterEnvironment(environment)

	if len(filtered) != 2 {
		t.Fatalf("filterEnvironment() returned %d variables, want 2", len(filtered))
	}
	if filtered[0].name != "PATH" || filtered[1].name != "SAFE_VALUE" {
		t.Fatalf("filterEnvironment() = %v, want PATH and SAFE_VALUE", filtered)
	}
}

func TestCreateLaunchScriptWritesAPrivateSelfDeletingScript(t *testing.T) {
	host := systemtest.New()
	secret := "token with '$HOME' and a newline\nvalue"

	path, err := createLaunchScript(host, []string{"claude", "--model", "opus"}, []string{
		"SAFE_VALUE=" + secret,
		"BAD;touch /tmp/injected=value",
	})

	if err != nil {
		t.Fatalf("createLaunchScript() error = %v", err)
	}
	if !strings.HasPrefix(path, "/tmp/xonovex-agent-tmux-") || !strings.HasSuffix(path, ".sh") {
		t.Errorf("launch script path = %q, want the xonovex-agent-tmux-*.sh pattern", path)
	}
	if strings.Contains(path, secret) {
		t.Fatal("launch script path contains secret")
	}
	info, err := host.Stat(path)
	if err != nil {
		t.Fatalf("stat launch script: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("launch script mode = %o, want 600", info.Mode().Perm())
	}
	content, _ := host.File(path)
	want := "#!/bin/sh\n" +
		"rm -f -- '" + path + "'\n" +
		"export SAFE_VALUE='token with '\\''$HOME'\\'' and a newline\nvalue'\n" +
		"exec 'claude' '--model' 'opus'\n"
	if content != want {
		t.Errorf("launch script = %q, want %q", content, want)
	}
	if strings.Contains(content, "injected") {
		t.Error("launch script exports an invalid variable name")
	}
}

func TestCreateLaunchScriptRequiresCommand(t *testing.T) {
	host := systemtest.New()
	if _, err := createLaunchScript(host, nil, nil); err == nil {
		t.Fatal("createLaunchScript(nil, nil) error = nil, want error")
	}
}

func TestCreateLaunchScriptReportsCreateFailure(t *testing.T) {
	host := failingHost{Fake: systemtest.New(), createErr: errors.New("disk full")}

	_, err := createLaunchScript(host, []string{"claude"}, nil)

	if err == nil || !strings.Contains(err.Error(), "create tmux launch script") {
		t.Fatalf("createLaunchScript() error = %v, want create failure", err)
	}
}

func TestCreateLaunchScriptRemovesTheFileOnWriteFailure(t *testing.T) {
	host := failingHost{Fake: systemtest.New(), writeErr: errors.New("disk full")}

	_, err := createLaunchScript(host, []string{"claude"}, nil)

	if err == nil || !strings.Contains(err.Error(), "write tmux launch script") {
		t.Fatalf("createLaunchScript() error = %v, want write failure", err)
	}
	if host.Exists("/tmp/xonovex-agent-tmux-1.sh") {
		t.Error("launch script remains after a write failure")
	}
}
