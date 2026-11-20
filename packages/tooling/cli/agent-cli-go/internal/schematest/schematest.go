// Package schematest validates a JSON document against a JSON Schema for the
// tests of this module. Production code never imports it, so the validator is a
// test dependency only.
package schematest

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Validate reports whether document satisfies schema (draft 2020-12). It reads
// both from memory and resolves no reference outside the schema.
func Validate(schema, document []byte) error {
	var schemaValue any
	if err := json.Unmarshal(schema, &schemaValue); err != nil {
		return fmt.Errorf("decode schema: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	if err := compiler.AddResource("memory:///schema.json", schemaValue); err != nil {
		return fmt.Errorf("add schema: %w", err)
	}
	compiled, err := compiler.Compile("memory:///schema.json")
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
	if err != nil {
		return fmt.Errorf("decode document: %w", err)
	}
	return compiled.Validate(instance)
}
