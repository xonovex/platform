package shared

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
)

// failingHost makes the launch script operations fail on demand.
type failingHost struct {
	*oshosttest.Fake
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
	host := oshosttest.New()
	secret := "token with '$HOME' and a newline\nvalue"

	path, err := LaunchScript(host, []string{"claude", "--model", "opus"}, []string{
		"SAFE_VALUE=" + secret,
		"BAD;touch /tmp/injected=value",
	})

	if err != nil {
		t.Fatalf("LaunchScript() error = %v", err)
	}
	if !strings.HasPrefix(path, "/tmp/xonovex-agent-terminal-") || !strings.HasSuffix(path, ".sh") {
		t.Errorf("launch script path = %q, want the xonovex-agent-terminal-*.sh pattern", path)
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
	host := oshosttest.New()
	if _, err := LaunchScript(host, nil, nil); err == nil {
		t.Fatal("LaunchScript(nil, nil) error = nil, want error")
	}
}

func TestCreateLaunchScriptReportsCreateFailure(t *testing.T) {
	host := failingHost{Fake: oshosttest.New(), createErr: errors.New("disk full")}

	_, err := LaunchScript(host, []string{"claude"}, nil)

	if err == nil || !strings.Contains(err.Error(), "create terminal launch script") {
		t.Fatalf("LaunchScript() error = %v, want create failure", err)
	}
}

func TestCreateLaunchScriptRemovesTheFileOnWriteFailure(t *testing.T) {
	host := failingHost{Fake: oshosttest.New(), writeErr: errors.New("disk full")}

	_, err := LaunchScript(host, []string{"claude"}, nil)

	if err == nil || !strings.Contains(err.Error(), "write terminal launch script") {
		t.Fatalf("LaunchScript() error = %v, want write failure", err)
	}
	if host.Exists("/tmp/xonovex-agent-terminal-1.sh") {
		t.Error("launch script remains after a write failure")
	}
}

func TestBuildShellCommand_QuotesAllArgs(t *testing.T) {
	result := ShellCommand([]string{"echo", "hello"})
	if result != "'echo' 'hello'" {
		t.Errorf("ShellCommand quoted = %q, want %q", result, "'echo' 'hello'")
	}
}

func TestBuildShellCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple command",
			args:     []string{"echo", "hello"},
			expected: "'echo' 'hello'",
		},
		{
			name:     "command with path",
			args:     []string{"/usr/bin/claude", "--model", "opus"},
			expected: "'/usr/bin/claude' '--model' 'opus'",
		},
		{
			name:     "command with spaces in arg",
			args:     []string{"echo", "hello world"},
			expected: "'echo' 'hello world'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShellCommand(tt.args)
			if result != tt.expected {
				t.Errorf("ShellCommand(%v) = %q, want %q", tt.args, result, tt.expected)
			}
		})
	}
}
