package launch

import (
	"fmt"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
)

// providerFile is a provider document with one model slot. The claude agent reads
// its route from the environment and the codex agent from its arguments.
func providerFile(name, agent, display string) string {
	route := `"environment":{"ROUTE_MODEL":"{model:main}"}`
	if agent == "codex" {
		route = `"args":["--model","{model:main}"]`
	}
	return fmt.Sprintf(`{"name":%q,"agent":%q,"displayName":%q,"models":{"main":{"id":"model-main"}},%s}`, name, agent, display, route)
}

// testCatalog defines the providers the plan tests select: alpha and beta for
// the claude agent, gamma for the codex agent.
func testCatalog(t *testing.T) providers.Catalog {
	t.Helper()
	var documents []providers.Document
	for _, p := range []struct{ name, agent string }{{"alpha", "claude"}, {"beta", "claude"}, {"gamma", "codex"}} {
		raw := providerFile(p.name, p.agent, "Display "+p.name)
		preset, err := providers.ParsePreset([]byte(raw))
		if err != nil {
			t.Fatalf("ParsePreset(%s) error = %v", raw, err)
		}
		documents = append(documents, providers.Document{Preset: preset, Layer: "user", Origin: "/test/" + p.name + ".json", Raw: []byte(raw)})
	}
	return providers.NewCatalog(documents...)
}
