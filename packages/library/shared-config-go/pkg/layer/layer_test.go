package layer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

func TestDigestIsTheSHA256OfTheBytes(t *testing.T) {
	const want = "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got := layer.Digest(nil); got != want {
		t.Fatalf("Digest = %s", got)
	}
}

func TestParseSplitsTheSections(t *testing.T) {
	data := []byte(`{"$schema":"x","version":1,"validate":{"skip":["a"]},"eval":{}}`)
	doc, err := layer.Parse(spec.Project, "/r/agentic.json", data)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Layer != spec.Project || doc.Path != "/r/agentic.json" || doc.Digest != layer.Digest(data) || string(doc.Data) != string(data) {
		t.Fatalf("doc = %+v", doc)
	}
	if len(doc.Sections) != 2 || string(doc.Sections["validate"]) != `{"skip":["a"]}` {
		t.Fatalf("sections = %v", doc.Sections)
	}
}

func TestParseRejectsABrokenEnvelope(t *testing.T) {
	cases := map[string]struct{ data, want string }{
		"not json":         {`{`, "not a JSON object"},
		"array":            {`[]`, "not a JSON object"},
		"null":             {`null`, "not a JSON object"},
		"no version":       {`{"validate":{}}`, "version is missing"},
		"wrong version":    {`{"version":2}`, "version is 2"},
		"text version":     {`{"version":"1"}`, `version is "1"`},
		"schema number":    {`{"version":1,"$schema":1}`, "$schema is not a string"},
		"unknown member":   {`{"version":1,"lint":{}}`, `unknown member "lint"`},
		"section an array": {`{"version":1,"eval":[]}`, "the eval section is not an object"},
		"section a string": {`{"version":1,"audit":"{}"}`, "the audit section is not an object"},
	}
	for name, c := range cases {
		if _, err := layer.Parse(spec.User, "/u/agentic.json", []byte(c.data)); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want %q", name, err, c.want)
		}
	}
}

type validator struct {
	violations []jsonschema.Violation
	err        error
	checked    []string
}

func (v *validator) Validate(id string, _ []byte) ([]jsonschema.Violation, error) {
	v.checked = append(v.checked, id)
	return v.violations, v.err
}

var sections = []spec.Section{{Name: "validate", SchemaID: "v"}, {Name: "audit", SchemaID: "a"}}

func TestCheckValidatesOnlyTheOwnSectionsTheFileSets(t *testing.T) {
	doc, err := layer.Parse(spec.Project, "/p", []byte(`{"version":1,"validate":{},"eval":{"anything":1}}`))
	if err != nil {
		t.Fatal(err)
	}
	v := &validator{}
	if err := doc.Check(v, sections); err != nil {
		t.Fatal(err)
	}
	if len(v.checked) != 1 || v.checked[0] != "v" {
		t.Fatalf("checked = %v", v.checked)
	}
}

func TestCheckListsTheViolations(t *testing.T) {
	doc, _ := layer.Parse(spec.Project, "/p", []byte(`{"version":1,"validate":{}}`))
	v := &validator{violations: []jsonschema.Violation{{Pointer: "/skip", Message: "bad"}, {Pointer: "", Message: "worse"}}}
	err := doc.Check(v, sections)
	if err == nil || !strings.Contains(err.Error(), "/p: invalid validate section: /validate/skip: bad; /validate: worse") {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckReturnsTheFault(t *testing.T) {
	doc, _ := layer.Parse(spec.Project, "/p", []byte(`{"version":1,"audit":{}}`))
	errBoom := errors.New("boom")
	if err := doc.Check(&validator{err: errBoom}, sections); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}
