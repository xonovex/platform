package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/legacy"
)

func TestConfigShowPrintsTheMergedSections(t *testing.T) {
	r := newTestRoot()
	r.host.AddFile("/home/user/.config/agentic/agentic.json", `{"version":1,"run":{"agent":"codex","bindPaths":["/data"]},"providers":{"alpha":{"agent":"claude"}}}`, 0o600)

	if err := r.execute("config", "show"); err != nil {
		t.Fatalf("config show error = %v", err)
	}

	want := "{\n  \"providers\": {\n    \"alpha\": {\n      \"agent\": \"claude\"\n    }\n  },\n  \"run\": {\n    \"agent\": \"codex\",\n    \"bindPaths\": [\n      \"/data\"\n    ]\n  }\n}\n"
	if r.out.String() != want {
		t.Errorf("output = %q, want %q", r.out.String(), want)
	}
}

func TestConfigShowNamesTheSourceOfEachValue(t *testing.T) {
	r := newTestRoot()
	r.host.AddFile("/home/user/.config/agentic/agentic.json", `{"version":1,"run":{"agent":"codex"}}`, 0o600)
	r.host.AddFile("/etc/agentic/agentic.json", `{"version":1,"run":{"network":"none"}}`, 0o644)

	if err := r.execute("config", "show", "--sources"); err != nil {
		t.Fatalf("config show --sources error = %v", err)
	}

	for _, want := range []string{"run.agent", "codex", "/home/user/.config/agentic/agentic.json", "run.network", "/etc/agentic/agentic.json"} {
		if !strings.Contains(r.out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, r.out.String())
		}
	}
}

func TestConfigShowPassesTheProjectThroughTheTrustGate(t *testing.T) {
	r := projectRoot()
	if err := r.execute("config", "show"); err == nil || !strings.Contains(err.Error(), "--approve-config-digest "+projectDigest) {
		t.Fatalf("config show error = %v, want the digest to approve", err)
	}
	if err := r.execute("config", "show", "--approve-config-digest", projectDigest); err != nil {
		t.Errorf("config show with the approval error = %v", err)
	}
	r.deps.Out = brokenWriter{}
	if err := r.execute("config", "show"); err == nil {
		t.Error("config show ignored a write failure")
	}
}

func TestConfigMigrateWritesTheLegacyFilesAsAgenticJSON(t *testing.T) {
	r := newTestRoot()
	r.deps.Legacy = legacy.Adapter{}
	r.host.AddFile("/home/user/.config/sandboxed-claude/config", "provider=alpha\n", 0o600)
	r.host.AddFile("/work/sub/.agent-cli.yaml", "agent: codex\n", 0o644)

	if err := r.execute("config", "migrate", "-w", "sub"); err != nil {
		t.Fatalf("config migrate error = %v", err)
	}

	want := "Wrote /home/user/.config/agentic/agentic.json from /home/user/.config/sandboxed-claude/config\n" +
		"Wrote /work/sub/agentic.json from /work/sub/.agent-cli.yaml\n" +
		"Check the new files, then remove the legacy files.\n"
	if r.out.String() != want {
		t.Errorf("output = %q, want %q", r.out.String(), want)
	}
	if data, _ := r.host.File("/work/sub/agentic.json"); !strings.Contains(data, `"agent": "codex"`) {
		t.Errorf("project agentic.json = %s", data)
	}
}

func TestConfigMigrateReportsEachOutcome(t *testing.T) {
	r := newTestRoot()
	if err := r.execute("config", "migrate"); err == nil || !strings.Contains(err.Error(), "reads no legacy configuration") {
		t.Errorf("config migrate without the port error = %v", err)
	}

	r.deps.Legacy = legacy.Adapter{}
	if err := r.execute("config", "migrate", "-w", "/work"); err != nil || r.out.String() != "No legacy configuration file was found.\n" {
		t.Errorf("config migrate of nothing = %q, %v", r.out.String(), err)
	}

	r.host.AddFile("/etc/agent-cli/managed.toml", "network = \"none\"\n", 0o644)
	r.host.AddFile("/home/user/.config/sandboxed-claude/config", "provider=alpha\n", 0o600)
	r.host.AddFile("/home/user/.config/agentic/agentic.json", "{}", 0o600)
	if err := r.execute("config", "migrate", "--managed"); err == nil || !strings.Contains(err.Error(), "/home/user/.config/agentic/agentic.json exists") {
		t.Errorf("config migrate over a file error = %v", err)
	}
	if r.host.Exists("/etc/agentic/agentic.json") {
		t.Error("config migrate wrote the managed file before it refused")
	}

	r.deps.Out = brokenWriter{}
	r.host.AddFile("/home/user/.config/agentic/agentic.json", "", 0o600)
	if err := r.execute("config", "migrate"); err == nil {
		t.Error("config migrate ignored a failed refusal")
	}
}

func TestConfigMigrateReportsAHostFault(t *testing.T) {
	r := newTestRoot()
	r.deps.Legacy = legacy.Adapter{}
	r.deps.Host = noWorkingDir{r.host}
	if err := r.execute("config", "migrate"); err == nil || !strings.Contains(err.Error(), "find the working directory") {
		t.Errorf("config migrate without a working directory error = %v", err)
	}

	r = newTestRoot()
	r.deps.Legacy = legacy.Adapter{}
	r.host.AddFile("/home/user/.config/sandboxed-claude/config", "provider=alpha\n", 0o600)
	r.deps.Out = brokenWriter{}
	if err := r.execute("config", "migrate"); err == nil {
		t.Error("config migrate ignored a write failure")
	}
}

// noWorkingDir is a host without a working directory.
type noWorkingDir struct{ *oshosttest.Fake }

func (noWorkingDir) Getwd() (string, error) { return "", errors.New("no working directory") }
