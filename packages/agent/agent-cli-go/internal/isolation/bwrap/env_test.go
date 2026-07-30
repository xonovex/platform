package bwrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	netshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
)

func envValue(env []string, key string) (string, bool) {
	prefix := key + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix), true
		}
	}
	return "", false
}

// processEnv is the environment bwrap itself is launched with, so it carries the
// host environment merged with the sandbox allowlist.
func TestProcessEnv_MergesTheSandboxAllowlistOverTheHostEnvironment(t *testing.T) {
	t.Setenv("XONOVEX_PROCESS_ENV_PROBE", "host-value")
	workDir := t.TempDir()
	homeDir := t.TempDir()
	cfg := claudeCfg(netshared.ModeNone, false, workDir)
	cfg.HomeDir = homeDir

	env, err := NewIsolator().processEnv(cfg, provision.Contribution{})

	if err != nil {
		t.Fatalf("processEnv() error = %v", err)
	}
	if got, ok := envValue(env, "XONOVEX_PROCESS_ENV_PROBE"); !ok || got != "host-value" {
		t.Errorf("host variable = %q (present %v), want it carried through", got, ok)
	}
	if got, _ := envValue(env, "HOME"); got != homeDir {
		t.Errorf("HOME = %q, want the resolved sandbox home %q", got, homeDir)
	}
	if got, _ := envValue(env, "TMPDIR"); got != "/tmp" {
		t.Errorf("TMPDIR = %q, want %q", got, "/tmp")
	}
}

func TestProcessEnv_ReportsAnUnresolvableWorkDir(t *testing.T) {
	cfg := claudeCfg(netshared.ModeNone, false, t.TempDir())
	cfg.HomeDir = filepath.Join(t.TempDir(), "absent")

	if _, err := NewIsolator().processEnv(cfg, provision.Contribution{}); err == nil {
		t.Fatal("processEnv() must fail when the home directory cannot be resolved")
	}
}

// TerminalCommand hands a terminal wrapper both the command and the environment
// it must be launched with, so the two must stay consistent.
func TestTerminalCommand_ReturnsTheCommandAndItsEnvironment(t *testing.T) {
	workDir := t.TempDir()
	homeDir := t.TempDir()
	cfg := claudeCfg(netshared.ModeNone, false, workDir)
	cfg.HomeDir = homeDir

	command, env, err := NewIsolator().TerminalCommand(cfg, provision.Contribution{})

	if err != nil {
		t.Fatalf("TerminalCommand() error = %v", err)
	}
	if len(command) == 0 || command[0] != "bwrap" {
		t.Fatalf("command = %v, want it to start with bwrap", command)
	}
	expected := bwrapCommand(t, cfg, provision.Contribution{})
	if strings.Join(command, " ") != strings.Join(expected, " ") {
		t.Errorf("command = %v, want the same command Command() returns", command)
	}
	if got, _ := envValue(env, "HOME"); got != homeDir {
		t.Errorf("HOME = %q, want the resolved sandbox home %q", got, homeDir)
	}
}

func TestTerminalCommand_ReportsAnUnresolvableDirectory(t *testing.T) {
	cfg := claudeCfg(netshared.ModeNone, false, filepath.Join(t.TempDir(), "absent"))

	if _, _, err := NewIsolator().TerminalCommand(
		cfg, provision.Contribution{},
	); err == nil {
		t.Fatal("TerminalCommand() must fail when the work directory cannot be resolved")
	}
}

// bwrap contributes no tools of its own, so immutability is exactly the
// provisioner's.
func TestPinnedProvision_MirrorsTheProvisioner(t *testing.T) {
	isolator := NewIsolator()
	for _, pinned := range []bool{true, false} {
		if got := isolator.PinnedProvision(provision.ProvisionNix, pinned, ""); got != pinned {
			t.Errorf("PinnedProvision(pinned=%v) = %v, want %v", pinned, got, pinned)
		}
	}
}

func TestAvailable_TracksTheBwrapBinaryOnPath(t *testing.T) {
	isolator := NewIsolator()

	t.Setenv("PATH", t.TempDir())
	absent, err := isolator.Available()
	if err != nil {
		t.Fatalf("Available() error = %v", err)
	}
	if absent {
		t.Error("Available() = true with an empty PATH, want false")
	}

	binDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(binDir, "bwrap"), []byte("#!/bin/sh\n"), 0o755,
	); err != nil {
		t.Fatalf("write bwrap stub: %v", err)
	}
	t.Setenv("PATH", binDir)
	present, err := isolator.Available()
	if err != nil {
		t.Fatalf("Available() error = %v", err)
	}
	if !present {
		t.Error("Available() = false with bwrap on PATH, want true")
	}
}
