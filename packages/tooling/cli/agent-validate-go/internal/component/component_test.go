package component_test

import (
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/component"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

func TestFiles(t *testing.T) {
	tr := tree.Tree{FS: memfs.New(map[string]string{
		"a/commands/x.md":          "x",
		"a/commands/sub/y.md":      "y",
		"a/commands/notes.txt":     "z",
		"a/other/z.md":             "z",
		"skills/s/commands/ref.md": "r",
		"commands/top.md":          "t",
	}), Root: "."}
	got, err := component.Files(tr, []string{"."}, "commands")
	if err != nil || !reflect.DeepEqual(got, []string{"a/commands/sub/y.md", "a/commands/x.md", "commands/top.md"}) {
		t.Fatalf("Files = %v %v", got, err)
	}
}

func TestPredicates(t *testing.T) {
	cases := []struct {
		name  string
		valid func(any) bool
		yes   []any
		no    []any
	}{
		{"string", component.IsString, []any{"a", ""}, []any{1, nil, []any{}}},
		{"bool", component.IsBool, []any{true, false}, []any{"true", 1}},
		{"positive", component.IsPositiveInteger, []any{1, 7}, []any{0, -1, "1", 1.5}},
		{"mapping", component.IsMapping, []any{map[string]any{}}, []any{[]any{}, "a"}},
		{"list or mapping", component.IsListOrMapping, []any{map[string]any{}, []any{}}, []any{"a", 1}},
		{"strings", component.IsStrings, []any{"a, b", []any{"a", "b"}}, []any{"", "  ", []any{}, []any{""}, []any{1}, 1}},
		{"one of", component.OneOf("a", "b"), []any{"a", "b"}, []any{"c", 1}},
	}
	for _, tc := range cases {
		for _, v := range tc.yes {
			if !tc.valid(v) {
				t.Errorf("%s accepts %v", tc.name, v)
			}
		}
		for _, v := range tc.no {
			if tc.valid(v) {
				t.Errorf("%s rejects %v", tc.name, v)
			}
		}
	}
}

func TestWords(t *testing.T) {
	if got := component.Words("a, b ,c"); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("string = %v", got)
	}
	if got := component.Words([]any{" a ", 3}); !reflect.DeepEqual(got, []string{"a", ""}) {
		t.Fatalf("list = %v", got)
	}
	if component.Words(5) != nil {
		t.Fatal("another type has no words")
	}
}

var spec = component.Spec{
	Prefix:   "thing",
	Noun:     "a thing",
	Known:    []string{"name", "size"},
	Required: []string{"name"},
	Body:     "body text",
	Fields:   []component.Field{{Key: "size", Valid: component.IsPositiveInteger, Message: "size must be positive"}},
}

func codes(source string) ([]string, bool) {
	report := check.NewReporter("thing")
	_, ok := spec.Check(report, "t.md", source)
	var out []string
	for _, f := range report.Findings() {
		out = append(out, f.Code)
	}
	return out, ok
}

func TestSpecCheck(t *testing.T) {
	cases := map[string]struct {
		source string
		codes  []string
		ok     bool
	}{
		"valid":          {"---\nname: a\nsize: 2\n---\nbody\n", nil, true},
		"no frontmatter": {"just text", []string{"thing.frontmatter"}, false},
		"bad yaml":       {"---\na: [\n---\nbody", []string{"thing.frontmatter"}, false},
		"not mapping":    {"---\n- a\n---\nbody", []string{"thing.frontmatter"}, false},
		"unknown":        {"---\nname: a\nother: 1\n---\nbody", []string{"thing.unknown-field"}, true},
		"required":       {"---\nsize: 1\n---\nbody", []string{"thing.name"}, true},
		"blank required": {"---\nname: '  '\n---\nbody", []string{"thing.name"}, true},
		"field":          {"---\nname: a\nsize: 0\n---\nbody", []string{"thing.size"}, true},
		"body":           {"---\nname: a\n---\n\n", []string{"thing.body"}, true},
	}
	for name, tc := range cases {
		got, ok := codes(tc.source)
		if !reflect.DeepEqual(got, tc.codes) || ok != tc.ok {
			t.Errorf("%s: codes = %v ok = %v", name, got, ok)
		}
	}
}
