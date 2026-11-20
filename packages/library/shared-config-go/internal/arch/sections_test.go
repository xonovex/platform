//go:build arch

package arch

import (
	"encoding/json"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/archtest"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/audit"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/eval"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/providers"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/run"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/validate"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// sections are the declarations, by the directory of their package.
func sections() map[string]spec.Section {
	return map[string]spec.Section{
		"pkg/section/run":       run.Section(),
		"pkg/section/providers": providers.Section(),
		"pkg/section/validate":  validate.Section(),
		"pkg/section/audit":     audit.Section(),
		"pkg/section/eval":      eval.Section(),
	}
}

type schemaNode struct {
	Ref                  string                `json:"$ref"`
	Properties           map[string]schemaNode `json:"properties"`
	AdditionalProperties json.RawMessage       `json:"additionalProperties"`
	PatternProperties    json.RawMessage       `json:"patternProperties"`
	Defs                 map[string]schemaNode `json:"$defs"`
}

func schemaDocuments(t *testing.T) map[string]schemaNode {
	t.Helper()
	documents, err := schemas.Documents()
	if err != nil {
		t.Fatal(err)
	}
	nodes := map[string]schemaNode{}
	for id, data := range documents {
		var node schemaNode
		if err := json.Unmarshal(data, &node); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		nodes[id] = node
	}
	return nodes
}

// keyPointers walks the properties of a section schema down to the pointers
// that keys declare. A property with no key and no properties of its own is a
// schema member that no key merges.
func keyPointers(t *testing.T, root schemaNode, node schemaNode, prefix string, declared map[string]bool, found map[string]bool) {
	t.Helper()
	if declared[prefix] {
		found[prefix] = true
		return
	}
	if strings.HasPrefix(node.Ref, "#/$defs/") {
		node = root.Defs[strings.TrimPrefix(node.Ref, "#/$defs/")]
	}
	if len(node.Properties) == 0 {
		t.Errorf("the schema member %q has no key and no properties", prefix)
		return
	}
	for name, child := range node.Properties {
		keyPointers(t, root, child, prefix+"/"+name, declared, found)
	}
}

// TestTheKeysAndTheSchemasAgree keeps a key and its schema from drifting apart:
// every member of a section schema is a key or holds keys, and every key is a
// member of the schema.
func TestTheKeysAndTheSchemasAgree(t *testing.T) {
	nodes := schemaDocuments(t)
	for dir, section := range sections() {
		schema, ok := nodes[section.SchemaID]
		if !ok {
			t.Errorf("%s names the schema %s, which the contract does not hold", dir, section.SchemaID)
			continue
		}
		declared := map[string]bool{}
		for _, key := range section.Keys {
			if strings.Contains(key.Pointer, "~") {
				t.Errorf("%s: %s needs JSON Pointer escaping, which the merge does not do", dir, key.Name())
			}
			declared[key.Pointer] = true
		}
		found := map[string]bool{}
		keyPointers(t, schema, schema, "", declared, found)
		for pointer := range declared {
			if !found[pointer] {
				t.Errorf("%s declares %s%s, which the schema %s does not hold", dir, section.Name, pointer, section.SchemaID)
			}
		}
	}
}

// TestTheFileSchemaHoldsEverySection keeps the file schema, the envelope check
// and the declarations to one list of sections.
func TestTheFileSchemaHoldsEverySection(t *testing.T) {
	file := schemaDocuments(t)[schemas.Agentic]
	var members []string
	for name, node := range file.Properties {
		if name == "$schema" || name == "version" {
			continue
		}
		members = append(members, name)
		var section spec.Section
		for _, s := range sections() {
			if s.Name == name {
				section = s
			}
		}
		if section.Name == "" {
			t.Errorf("the file schema holds the section %s, which no package declares", name)
			continue
		}
		if schemas.Namespace+node.Ref != section.SchemaID {
			t.Errorf("the file schema refers the %s section to %s, but its declaration names %s", name, node.Ref, section.SchemaID)
		}
	}
	sort.Strings(members)
	want := layer.SectionNames()
	sort.Strings(want)
	if !slices.Equal(members, want) {
		t.Errorf("the file schema holds %v; the envelope check knows %v", members, want)
	}
	if len(sections()) != len(want) {
		t.Errorf("%d section packages for %d sections", len(sections()), len(want))
	}
}

// TestNoSectionNamesAnotherSection keeps each tool to its own section: a
// section package spells no other section name and declares keys only in its
// own section.
func TestNoSectionNamesAnotherSection(t *testing.T) {
	declared := sections()
	for _, f := range archtest.Production(moduleFiles(t)) {
		own, ok := declared[f.Dir]
		if !ok {
			continue
		}
		ast.Inspect(f.Syntax, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(literal.Value)
			if err == nil && value != own.Name && slices.Contains(layer.SectionNames(), value) {
				t.Errorf("%s names the %s section", f.Path, value)
			}
			return true
		})
	}
	for dir, section := range declared {
		if filepath.Base(dir) != section.Name {
			t.Errorf("%s declares the %s section", dir, section.Name)
		}
		for _, key := range section.Keys {
			if key.Section != section.Name {
				t.Errorf("%s declares %s in the %s section", dir, key.Name(), key.Section)
			}
		}
	}
}

// fixture renders a project file that sets the key.
func fixture(key spec.Key) []byte {
	var value any = "x"
	tokens := strings.Split(strings.TrimPrefix(key.Pointer, "/"), "/")
	if key.Pointer == "" {
		tokens = nil
		value = map[string]any{"x": map[string]any{}}
	}
	for i := len(tokens) - 1; i >= 0; i-- {
		value = map[string]any{tokens[i]: value}
	}
	data, _ := json.Marshal(map[string]any{"version": 1, key.Section: value})
	return data
}

// TestAProjectFileCannotSetATrustedKey proves, for every key that a project may
// not set, that the merge rejects a project file that sets it.
func TestAProjectFileCannotSetATrustedKey(t *testing.T) {
	checked := 0
	for _, section := range sections() {
		for _, key := range section.Keys {
			document, err := layer.Parse(spec.Project, "/repo/agentic.json", fixture(key))
			if err != nil {
				t.Fatalf("%s: %v", key.Name(), err)
			}
			_, err = merge.Merge([]spec.Section{section}, []layer.Document{document}, nil)
			if key.Allows(spec.Project) {
				continue
			}
			checked++
			if err == nil || !strings.Contains(err.Error(), "the project file sets "+key.Name()) {
				t.Errorf("a project file sets %s: err = %v", key.Name(), err)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no key is closed to a project file")
	}
}

// TestTheModuleDependsOnlyOnTheContractAndTheCore keeps the library graph a
// DAG: the configuration sits above the contract and the core, and below the
// tools.
func TestTheModuleDependsOnlyOnTheContractAndTheCore(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(moduleRoot(t), "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	allowedModules := []string{contract, "github.com/xonovex/platform/packages/library/shared-core-go"}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "require "))
		if len(fields) == 0 || !strings.HasPrefix(fields[0], "github.com/xonovex/platform/") {
			continue
		}
		if !slices.Contains(allowedModules, fields[0]) {
			t.Errorf("go.mod requires %s; the configuration may depend only on %v", fields[0], allowedModules)
		}
	}
}
