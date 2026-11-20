package overlay

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

func TestEveryAgentHasADialect(t *testing.T) {
	registered := AgentTypes()
	declared := agents.GetAgentTypes()
	if len(registered) != len(declared) {
		t.Fatalf("dialects %v, agents %v", registered, declared)
	}
	for index, agentType := range declared {
		if registered[index] != agentType {
			t.Errorf("dialect %s does not match agent %s", registered[index], agentType)
		}
		if _, err := Resolve(agentType); err != nil {
			t.Errorf("Resolve(%s): %v", agentType, err)
		}
	}
}

func TestResolveRejectsAnUnknownAgent(t *testing.T) {
	if _, err := Resolve(types.AgentType("nope")); err == nil {
		t.Error("an unknown agent resolved")
	}
}

// importsOf returns the import paths of the non-test Go files in dir.
func importsOf(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var imports []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range file.Imports {
			path, _ := strconv.Unquote(spec.Path.Value)
			imports = append(imports, path)
		}
	}
	return imports
}

// TestLeavesAreImportedOnlyByTheRegistry keeps the axis orthogonal: a leaf knows
// the shared core and the artifact model, never a sibling leaf, and the shared
// core knows no leaf and no machine.
func TestLeavesAreImportedOnlyByTheRegistry(t *testing.T) {
	const module = "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/"
	leaves := []string{"claude", "codex", "copilot", "kiro", "opencode", "pi"}
	for _, leaf := range leaves {
		for _, path := range importsOf(t, leaf) {
			if strings.HasPrefix(path, module+"overlay/") && path != module+"overlay/shared" {
				t.Errorf("leaf %s imports %s", leaf, path)
			}
			if path == "os" || path == "os/exec" || strings.HasPrefix(path, "net") {
				t.Errorf("leaf %s imports %s", leaf, path)
			}
		}
	}
	for _, path := range importsOf(t, "shared") {
		for _, leaf := range leaves {
			if path == module+"overlay/"+leaf {
				t.Errorf("the shared core imports the leaf %s", leaf)
			}
		}
		if path == "os" || path == "os/exec" {
			t.Errorf("the shared core imports %s", path)
		}
	}
}

func TestCheckRefusesALoginAgentWhoseHomeTheOverlayReplaces(t *testing.T) {
	for _, agentType := range AgentTypes() {
		agent, err := agents.GetAgent(agentType)
		if err != nil {
			t.Fatal(err)
		}
		dialect, _ := Resolve(agentType)
		refused := dialect.Facts().ReplacesHome && agent.Access == types.AccessLogin
		if err := Check(agent); (err != nil) != refused {
			t.Errorf("Check(%s) = %v, want refused = %v", agentType, err, refused)
		}
	}
	kiro, _ := agents.GetAgent(types.AgentKiro)
	if err := Check(kiro); err == nil || !strings.Contains(err.Error(), "login") {
		t.Errorf("Check(kiro) = %v", err)
	}
	if err := Check(&types.AgentConfig{Type: "nope"}); err == nil {
		t.Error("an agent without a dialect passed")
	}
}
