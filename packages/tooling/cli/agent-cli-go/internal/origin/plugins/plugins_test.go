package plugins

import (
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func TestDefaultRegistrySelectsADirectoryAndAGitRepository(t *testing.T) {
	host := systemtest.New().AddDir("/work/plugin")
	registry := DefaultRegistry(host)
	if len(registry) != 2 {
		t.Fatalf("registry = %v", registry)
	}
	if origin, err := registry.Select("plugin"); err != nil || origin != registry["directory"] {
		t.Errorf("directory: %v, %v", origin, err)
	}
	if origin, err := registry.Select("https://h/r.git"); err != nil || origin != registry["git"] {
		t.Errorf("git: %v, %v", origin, err)
	}
	if _, err := registry.Select("nothing"); err == nil {
		t.Error("an unknown source selected an origin")
	}
}
