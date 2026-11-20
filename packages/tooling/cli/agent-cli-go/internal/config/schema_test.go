package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/schematest"
)

func TestSchemaDescribesEveryKey(t *testing.T) {
	data := Schema()
	var schema struct {
		ID         string                    `json:"$id"`
		Properties map[string]map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("the schema is not JSON: %v", err)
	}
	if schema.ID != SchemaID || len(schema.Properties) != len(Specs) {
		t.Fatalf("schema = %q with %d properties, want %q with %d", schema.ID, len(schema.Properties), SchemaID, len(Specs))
	}
	for _, spec := range Specs {
		property, ok := schema.Properties[string(spec.Key)]
		if !ok || property["x-rule"] != string(spec.Rule) {
			t.Errorf("key %s: property = %v, want x-rule %s", spec.Key, property, spec.Rule)
		}
	}
}

func TestSchemaValidatesConfigurationDocuments(t *testing.T) {
	schema := Schema()
	valid := `{"agent":"codex","bindPaths":["/data"],"requireKernelIsolation":true}`
	if err := schematest.Validate(schema, []byte(valid)); err != nil {
		t.Errorf("a valid document was rejected: %v", err)
	}
	invalid := map[string]string{
		"unknown key":        `{"colour":"red"}`,
		"wrong scalar type":  `{"agent":3}`,
		"wrong list type":    `{"bindPaths":"/data"}`,
		"empty list entry":   `{"bindPaths":[""]}`,
		"wrong require type": `{"requireKernelIsolation":"yes"}`,
	}
	for name, document := range invalid {
		if err := schematest.Validate(schema, []byte(document)); err == nil {
			t.Errorf("%s: the document was accepted", name)
		}
	}
	if !strings.HasSuffix(string(schema), "}\n") {
		t.Error("the schema must end with a newline")
	}
}
