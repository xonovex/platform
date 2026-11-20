// Package jsonschema is the port that checks a JSON document against a named
// schema. Core packages depend on this port. The tekuri adapter implements it
// for JSON Schema draft 2020-12 with format assertions, offline.
package jsonschema

import "errors"

// ErrUnknownSchema reports a schema id that the validator does not hold.
var ErrUnknownSchema = errors.New("unknown schema")

// Violation is one way a document breaks a schema.
type Violation struct {
	// Pointer is the JSON Pointer of the offending value, empty for the root.
	Pointer string
	// Message states the broken rule.
	Message string
}

// Validator checks documents against the schemas it holds.
type Validator interface {
	// Validate checks document against the schema with the given id. A
	// document that is not valid JSON, or that breaks the schema, yields
	// violations and a nil error. The error is for infrastructure faults: an
	// unknown schema id wraps ErrUnknownSchema.
	Validate(schemaID string, document []byte) ([]Violation, error)
}
