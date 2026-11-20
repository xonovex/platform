package registry_test

import (
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/registry"
)

func TestAllNamesEachCheckOnce(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range registry.All() {
		name := c.Name()
		if name == "" || c.Summary() == "" || len(c.DefaultPaths()) == 0 {
			t.Errorf("check %q lacks a name, a summary or a default path", name)
		}
		if seen[name] {
			t.Errorf("check %q is registered twice", name)
		}
		seen[name] = true
	}
	for _, want := range []string{
		"skill", "sources", "drift", "links", "routing", "plugin", "marketplace", "instructions", "command",
		"agent", "hook", "output-style", "kit", "llms", "schema", "repository", "workspace", "text",
	} {
		if !seen[want] {
			t.Errorf("check %q is not registered", want)
		}
	}
}
