package schematest

import (
	"strings"
	"testing"
)

const schema = `{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object","required":["n"],"properties":{"n":{"type":"integer"}},"additionalProperties":false}`

func TestValidateAcceptsAConformingDocument(t *testing.T) {
	if err := Validate([]byte(schema), []byte(`{"n":1}`)); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestValidateRejectsANonConformingDocument(t *testing.T) {
	for name, document := range map[string]string{"wrong type": `{"n":"x"}`, "missing": `{}`, "extra": `{"n":1,"m":2}`} {
		if err := Validate([]byte(schema), []byte(document)); err == nil {
			t.Errorf("%s: Validate() error = nil, want a violation", name)
		}
	}
}

func TestValidateReportsABrokenInput(t *testing.T) {
	cases := map[string]struct{ schema, document, want string }{
		"schema not JSON":   {`{`, `{}`, "decode schema"},
		"schema not a type": {`{"type":"nonsense"}`, `{}`, "compile schema"},
		"schema scalar":     {`3`, `{}`, "compile schema"},
		"document not JSON": {schema, `{`, "decode document"},
	}
	for name, tc := range cases {
		if err := Validate([]byte(tc.schema), []byte(tc.document)); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: Validate() error = %v, want %q", name, err, tc.want)
		}
	}
}
