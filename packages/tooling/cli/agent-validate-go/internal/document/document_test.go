package document_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
)

func TestParseWithFrontmatter(t *testing.T) {
	doc, err := document.Parse("---\nname: a\ntags: [x]\n---\nbody\n")
	if err != nil {
		t.Fatal(err)
	}
	if !doc.HasFrontmatter || doc.Source != "name: a\ntags: [x]" || doc.Body != "body\n" || doc.BodyLine != 5 {
		t.Fatalf("doc = %+v", doc)
	}
	if name, ok := doc.String("name"); !ok || name != "a" {
		t.Fatal("String reads text fields")
	}
	if _, ok := doc.String("tags"); ok {
		t.Fatal("a list is not text")
	}
	if !reflect.DeepEqual(doc.Keys(), []string{"name", "tags"}) {
		t.Fatalf("keys = %v", doc.Keys())
	}
}

func TestParseWithoutFrontmatter(t *testing.T) {
	doc, err := document.Parse("# Title\n")
	if err != nil || doc.HasFrontmatter || doc.Body != "# Title\n" || doc.BodyLine != 1 || len(doc.Fields) != 0 {
		t.Fatalf("doc = %+v, %v", doc, err)
	}
	unterminated, err := document.Parse("---\nname: a\n---")
	if err != nil || unterminated.HasFrontmatter {
		t.Fatal("a closing fence needs a newline after it")
	}
}

func TestParseEmptyFrontmatter(t *testing.T) {
	doc, err := document.Parse("---\n\n---\nbody")
	if err != nil || !doc.HasFrontmatter || len(doc.Fields) != 0 {
		t.Fatalf("doc = %+v, %v", doc, err)
	}
}

func TestParseRejectsBrokenFrontmatter(t *testing.T) {
	if _, err := document.Parse("---\nname: [a\n---\nbody"); err == nil {
		t.Fatal("invalid YAML is an error")
	}
	if _, err := document.Parse("---\n- a\n- b\n---\nbody"); !errors.Is(err, document.ErrNotMapping) {
		t.Fatalf("err = %v", err)
	}
}

func TestKeyLine(t *testing.T) {
	source := "name: a\ndescription: b"
	if document.KeyLine(source, "description") != 3 || document.KeyLine(source, "missing") != 1 {
		t.Fatal("KeyLine counts the opening fence")
	}
}

func TestTypeName(t *testing.T) {
	cases := map[string]any{
		"null": nil, "bool": true, "int": 1, "float": 1.5, "str": "a", "list": []any{}, "dict": map[string]any{},
		"time.Duration": 3,
	}
	delete(cases, "time.Duration")
	for want, value := range cases {
		if got := document.TypeName(value); got != want {
			t.Errorf("TypeName(%v) = %s, want %s", value, got, want)
		}
	}
	if document.TypeName(int64(1)) != "int" || document.TypeName(uint64(1)) != "int" {
		t.Fatal("integers share a name")
	}
	if document.TypeName(struct{}{}) != "struct {}" {
		t.Fatal("other types use their Go name")
	}
}
