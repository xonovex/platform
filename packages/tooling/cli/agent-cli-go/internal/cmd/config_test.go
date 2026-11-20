package cmd

import (
	"strings"
	"testing"
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
