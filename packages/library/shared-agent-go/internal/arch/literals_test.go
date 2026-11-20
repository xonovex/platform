//go:build arch

// Package arch holds the architecture fitness tests of the agent platform that
// read source trees, so they carry the arch build tag and run in the go-test-arch
// task, not in the unit run.
package arch

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// modules are the Go modules that ship a binary or a library the binaries use,
// relative to the packages directory. Production code of each holds no model id
// and no provider preset: both are data a user supplies.
var modules = []string{
	"library/shared-agent-go",
	"library/shared-launch-go",
	"tooling/cli/agent-cli-go",
	"tooling/cli/agent-eval-go",
	"sandbox/operator/agent-operator-go",
}

var (
	// modelID matches a string that looks like the id of a model of a known vendor
	// family: a vendor name, a version digit and an optional provider prefix.
	modelID = regexp.MustCompile(`^(?:[a-z0-9._-]+/)?(?:claude|gpt|gemini|glm|codex|llama|mistral|qwen|deepseek|grok|kimi)-?[a-z0-9._-]*[0-9][a-z0-9._-]*$|^o[0-9](?:-[a-z0-9.-]+)?$`)
	// presetName matches the name of a provider preset that an earlier release
	// shipped in the binary.
	presetName = regexp.MustCompile(`^(?:gemini|gemini-claude|glm|gpt|gpt5-codex)$`)
)

func packagesDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file is <packages>/library/shared-agent-go/internal/arch/literals_test.go.
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(file)))))
}

// productionFile reports whether path is Go code that ships. Test files and the
// packages that hold only test doubles do not ship behavior.
func productionFile(path string) bool {
	if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
		return false
	}
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part == "testdata" || part == "testutil" || strings.HasSuffix(part, "test") && part != "latest" {
			return false
		}
	}
	return true
}

type finding struct{ where, literal, kind string }

func scan(t *testing.T, root string) []finding {
	t.Helper()
	var found []finding
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name := entry.Name(); name == "node_modules" || name == "dist" || (strings.HasPrefix(name, ".") && path != root) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if !productionFile(rel) {
			return nil
		}
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
			where := fset.Position(literal.Pos()).String()
			switch {
			case presetName.MatchString(value):
				found = append(found, finding{where, value, "provider preset name"})
			case modelID.MatchString(value):
				found = append(found, finding{where, value, "model id"})
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("scan %s: %v", root, err)
	}
	return found
}

func TestProductionCodeHoldsNoModelIDAndNoProviderPreset(t *testing.T) {
	for _, module := range modules {
		for _, f := range scan(t, filepath.Join(packagesDir(t), filepath.FromSlash(module))) {
			t.Errorf("%s: the %s %q belongs in a provider file or a setting, not in code", f.where, f.kind, f.literal)
		}
	}
}

func TestTheScannerRecognizesWhatItForbids(t *testing.T) {
	for _, id := range []string{"claude-haiku-5-5", "gpt-6.2-sol", "gemini-3.8-flash", "glm-5.3", "google/gemini-4-carbon", "o3-mini", "gpt-5.3-codex"} {
		if !modelID.MatchString(id) {
			t.Errorf("modelID does not match %q", id)
		}
	}
	for _, name := range []string{"gemini", "gemini-claude", "glm", "gpt", "gpt5-codex"} {
		if !presetName.MatchString(name) {
			t.Errorf("presetName does not match %q", name)
		}
	}
	for _, ok := range []string{"claude", "codex", "claude-plugin", "<model-id>", "model-main", "opencode", "provider/name"} {
		if modelID.MatchString(ok) || presetName.MatchString(ok) {
			t.Errorf("the scanner flags %q, which names no model", ok)
		}
	}
	for path, want := range map[string]bool{
		"pkg/a/a.go": true, "pkg/a/a_test.go": false, "internal/system/systemtest/fake.go": false,
		"test/testutil/x.go": false, "testdata/x.go": false, "readme.md": false, "internal/latest/a.go": true,
	} {
		if got := productionFile(path); got != want {
			t.Errorf("productionFile(%q) = %v, want %v", path, got, want)
		}
	}
}
