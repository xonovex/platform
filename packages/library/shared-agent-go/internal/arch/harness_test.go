//go:build arch

package arch

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// harnessModules are the Go modules whose production code reads the facts of a
// harness, relative to the packages directory.
var harnessModules = []string{
	"library/shared-agent-go",
	"library/shared-launch-go",
	"tooling/cli/agent-cli-go",
	"tooling/cli/agent-eval-go",
	"tooling/cli/agent-validate-go",
	"sandbox/operator/agent-operator-go",
}

// layoutCode holds the overlay dialects, whose directory layouts stay code by
// design: they name the files of a harness because they write them.
const layoutCode = "library/shared-agent-go/pkg/overlay"

// harnessCode lists the packages that turn the facts of a harness into its
// command line. They hold no fact value at all, not even a generic one.
var harnessCode = []string{
	"library/shared-agent-go/pkg/agents",
	"tooling/cli/agent-eval-go/internal/harness",
}

// fact is one value of the shipped facts.
type fact struct {
	// what names the harness and the fact.
	what string
	// distinctive is true for a value no other tool shares: a path, a program, a
	// package, a variable or a flag of more than one word. Other code holds no
	// distinctive value; a generic one, such as -p or json, is forbidden only in
	// harness code. The instruction files are also file names of the repository
	// conventions, so only harness code is checked for them.
	distinctive bool
}

// factValues returns every string of the shipped facts. The name of a harness
// is its agent type, which code names to select it, so a binary that equals the
// name is no finding. A value holding a placeholder is a template, never a
// literal of its own.
func factValues(t *testing.T) map[string]fact {
	t.Helper()
	values := map[string]fact{}
	add := func(agent types.AgentType, what string, named bool, items ...string) {
		for _, item := range items {
			if item == "" || item == string(agent) || strings.Contains(item, types.OptionValue) {
				continue
			}
			flag := strings.HasPrefix(item, "--") && strings.Contains(item[2:], "-")
			values[item] = fact{what: string(agent) + " " + what, distinctive: named || flag}
			// A path joined from its segments holds its first segment as a
			// literal of its own; the segment is distinctive when it names the
			// harness, as .claude-plugin does.
			if first, _, nested := strings.Cut(item, "/"); named && nested && strings.Contains(first, string(agent)) {
				values[first] = fact{what: string(agent) + " " + what, distinctive: true}
			}
		}
	}
	catalog := agents.Embedded()
	for _, agentType := range catalog.Types() {
		agent, err := catalog.Get(agentType)
		if err != nil {
			t.Fatal(err)
		}
		add(agentType, "binary", true, agent.Binary)
		add(agentType, "nix package", true, agent.NixPackage)
		add(agentType, "user configuration path", true, agent.UserConfigPaths...)
		add(agentType, "skills directory", true, agent.SkillsDir)
		add(agentType, "plugin layout", true, agent.Plugin.Manifest, agent.Plugin.Marketplace)
		if agent.ConfigHome != nil {
			add(agentType, "configuration home", true, agent.ConfigHome.Env, agent.ConfigHome.Dir)
		}
		add(agentType, "instruction file", false, agent.InstructionFile)
		add(agentType, "launch argument", false, agent.LaunchArgs...)
		add(agentType, "sandbox argument", false, agent.SandboxArgs...)
		add(agentType, "version probe", false, agent.VersionArgs...)
		add(agentType, "headless argument", false, agent.Headless.Args...)
		add(agentType, "hermetic argument", false, agent.Headless.HermeticArgs...)
		for _, args := range agent.Headless.Output {
			add(agentType, "output argument", false, args...)
		}
		for _, args := range agent.Headless.Options {
			add(agentType, "option", false, args...)
		}
	}
	// The names of the output formats and of the options are the vocabulary
	// code selects a fact by, so a value that equals one is no finding.
	for _, agentType := range catalog.Types() {
		agent, _ := catalog.Get(agentType)
		for name := range agent.Headless.Output {
			delete(values, name)
		}
		for name := range agent.Headless.Options {
			delete(values, name)
		}
	}
	return values
}

func inHarnessCode(rel string) bool {
	for _, dir := range harnessCode {
		if rel == dir || strings.HasPrefix(rel, dir+"/") {
			return true
		}
	}
	return false
}

func TestProductionCodeHoldsNoHarnessFact(t *testing.T) {
	values := factValues(t)
	root := packagesDir(t)
	fset := token.NewFileSet()
	for _, module := range harnessModules {
		dir := filepath.Join(root, filepath.FromSlash(module))
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(root, path)
			if entry.IsDir() {
				if name := entry.Name(); name == "node_modules" || name == "dist" || filepath.ToSlash(rel) == layoutCode || (strings.HasPrefix(name, ".") && path != dir) {
					return filepath.SkipDir
				}
				return nil
			}
			if !productionFile(rel) {
				return nil
			}
			strict := inHarnessCode(filepath.ToSlash(filepath.Dir(rel)))
			parsed, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
			if err != nil {
				return err
			}
			ast.Inspect(parsed, func(node ast.Node) bool {
				literal, ok := node.(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					return true
				}
				value, err := strconv.Unquote(literal.Value)
				if err != nil {
					return true
				}
				if fact, found := values[value]; found && (strict || fact.distinctive) {
					t.Errorf("%s: %q is the %s; read it from the harness facts", fset.Position(literal.Pos()), value, fact.what)
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s: %v", module, err)
		}
	}
}

func TestTheHarnessScannerKnowsTheFacts(t *testing.T) {
	values := factValues(t)
	for value, distinctive := range map[string]bool{
		"claude-code": true, ".codex-plugin/plugin.json": true, ".claude-plugin": true, "CODEX_HOME": true, "--permission-mode": true,
		"kiro-cli-chat": true, "--no-chrome": true, "CLAUDE.md": false, "-p": false, "--json": false,
	} {
		if got, ok := values[value]; !ok || got.distinctive != distinctive {
			t.Errorf("factValues()[%q] = %+v, %v, want distinctive %v", value, got, ok, distinctive)
		}
	}
	for _, skip := range []string{"claude", "codex", "", "{value}", "--model={value}", ".agents", ".config", "json"} {
		if _, ok := values[skip]; ok {
			t.Errorf("factValues() holds %q, which is no literal fact", skip)
		}
	}
	for rel, want := range map[string]bool{
		"tooling/cli/agent-eval-go/internal/harness": true, "tooling/cli/agent-eval-go/internal/harness/claude": true,
		"tooling/cli/agent-eval-go/internal/harnessx": false, "tooling/cli/agent-eval-go/internal/catalog": false,
	} {
		if got := inHarnessCode(rel); got != want {
			t.Errorf("inHarnessCode(%q) = %v, want %v", rel, got, want)
		}
	}
}
