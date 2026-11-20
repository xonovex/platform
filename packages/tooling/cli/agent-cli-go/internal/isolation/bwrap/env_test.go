package bwrap

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
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
	host := newHost()
	host.Env["XONOVEX_PROCESS_ENV_PROBE"] = "host-value"

	env, err := NewIsolator(host, testLog).processEnv(claudeCfg(network.NetworkNone, false), provision.Contribution{})

	if err != nil {
		t.Fatalf("processEnv() error = %v", err)
	}
	if got, ok := envValue(env, "XONOVEX_PROCESS_ENV_PROBE"); !ok || got != "host-value" {
		t.Errorf("host variable = %q (present %v), want it carried through", got, ok)
	}
	if got, _ := envValue(env, "HOME"); got != "/home/user" {
		t.Errorf("HOME = %q, want the resolved sandbox home /home/user", got)
	}
	if got, _ := envValue(env, "TMPDIR"); got != "/tmp" {
		t.Errorf("TMPDIR = %q, want %q", got, "/tmp")
	}
}

func TestProcessEnv_ReportsAnUnresolvableHomeDir(t *testing.T) {
	cfg := claudeCfg(network.NetworkNone, false)
	cfg.HomeDir = "/home/absent"

	if _, err := NewIsolator(newHost(), testLog).processEnv(cfg, provision.Contribution{}); err == nil {
		t.Fatal("processEnv() must fail when the home directory cannot be resolved")
	}
}

func TestProcessEnv_ReportsAMalformedCustomEnvironment(t *testing.T) {
	cfg := claudeCfg(network.NetworkNone, false)
	cfg.CustomEnv = []string{"MISSING_EQUALS"}

	if _, err := NewIsolator(newHost(), testLog).processEnv(cfg, provision.Contribution{}); err == nil {
		t.Fatal("processEnv() must fail on a malformed custom environment")
	}
}

// TerminalCommand hands a terminal wrapper both the command and the environment
// it must be launched with, so the two must stay consistent.
func TestTerminalCommand_ReturnsTheCommandAndItsEnvironment(t *testing.T) {
	host := newHost()
	cfg := claudeCfg(network.NetworkNone, false)

	command, env, err := NewIsolator(host, testLog).TerminalCommand(cfg, provision.Contribution{})

	if err != nil {
		t.Fatalf("TerminalCommand() error = %v", err)
	}
	if len(command) == 0 || command[0] != "bwrap" {
		t.Fatalf("command = %v, want it to start with bwrap", command)
	}
	expected := bwrapCommand(t, host, cfg, provision.Contribution{})
	if strings.Join(command, " ") != strings.Join(expected, " ") {
		t.Errorf("command = %v, want the same command Command() returns", command)
	}
	if got, _ := envValue(env, "HOME"); got != "/home/user" {
		t.Errorf("HOME = %q, want the resolved sandbox home /home/user", got)
	}
}

func TestTerminalCommand_ReportsAnUnresolvableDirectory(t *testing.T) {
	cfg := claudeCfg(network.NetworkNone, false)
	cfg.WorkDir = "/work/absent"

	if _, _, err := NewIsolator(newHost(), testLog).TerminalCommand(cfg, provision.Contribution{}); err == nil {
		t.Fatal("TerminalCommand() must fail when the work directory cannot be resolved")
	}
}

// bwrap contributes no tools of its own, so immutability is exactly the
// provisioner's.
func TestPinnedProvision_MirrorsTheProvisioner(t *testing.T) {
	isolator := NewIsolator(newHost(), testLog)
	for _, pinned := range []bool{true, false} {
		if got := isolator.PinnedProvision(provision.ProvisionNix, pinned, ""); got != pinned {
			t.Errorf("PinnedProvision(pinned=%v) = %v, want %v", pinned, got, pinned)
		}
	}
}

func TestAvailable_TracksTheBwrapBinaryOnPath(t *testing.T) {
	host := newHost()
	isolator := NewIsolator(host, testLog)

	absent, err := isolator.Available()
	if err != nil {
		t.Fatalf("Available() error = %v", err)
	}
	if absent {
		t.Error("Available() = true without bwrap on PATH, want false")
	}

	host.AddExecutable("/usr/bin/bwrap")
	present, err := isolator.Available()
	if err != nil {
		t.Fatalf("Available() error = %v", err)
	}
	if !present {
		t.Error("Available() = false with bwrap on PATH, want true")
	}
}
