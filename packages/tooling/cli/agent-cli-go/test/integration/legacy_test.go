//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestALegacyConfigurationRunsUntilItIsMigrated(t *testing.T) {
	m := newMachine(t, "0")
	legacyDir := filepath.Join(m.home, ".config", "sandboxed-claude")
	writeFile(t, filepath.Join(legacyDir, "config"), "provider=route\n", 0o600)
	writeFile(t, filepath.Join(legacyDir, "providers", "route.json"), routeProvider, 0o644)

	cmd, stdout, stderr := m.command("run", "--dry-run")
	if err := cmd.Run(); err != nil {
		t.Fatalf("dry run of the legacy files: %v\n%s", err, stderr)
	}
	output := stdout.String() + stderr.String()
	for _, want := range []string{"sandboxed-claude/config (legacy)", "is a legacy user configuration file", "agent-cli config migrate"} {
		if !strings.Contains(output, want) {
			t.Errorf("dry run lacks %q:\n%s", want, output)
		}
	}

	cmd, stdout, stderr = m.command("config", "migrate")
	if err := cmd.Run(); err != nil {
		t.Fatalf("config migrate: %v\n%s", err, stderr)
	}
	for _, want := range []string{filepath.Join(m.home, "agentic", "agentic.json"), filepath.Join(m.userProviders(), "route.json")} {
		if !strings.Contains(stdout.String(), "Wrote "+want) {
			t.Errorf("config migrate output lacks %s:\n%s", want, stdout)
		}
	}

	cmd, _, stderr = m.command("run", "--dry-run")
	if got := exitCode(cmd.Run()); got != 1 || !strings.Contains(stderr.String(), "the user layer has") {
		t.Errorf("dry run with both files = exit %d, %s; want the conflict named", got, stderr)
	}

	if err := os.RemoveAll(legacyDir); err != nil {
		t.Fatal(err)
	}
	cmd, stdout, stderr = m.command("run", "--dry-run")
	if err := cmd.Run(); err != nil {
		t.Fatalf("dry run of the migrated files: %v\n%s", err, stderr)
	}
	if output := stdout.String() + stderr.String(); !strings.Contains(output, "provider: route (user "+filepath.Join(m.home, "agentic", "agentic.json")) || strings.Contains(output, "legacy") {
		t.Errorf("dry run of the migrated files:\n%s", output)
	}
}
