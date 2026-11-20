// Package tekuri implements the jsonschema port on top of
// github.com/santhosh-tekuri/jsonschema/v6, the one Go validator that supports
// draft 2020-12 completely, asserts formats, and resolves every reference
// from documents given up front. The adapter holds every schema in memory and
// refuses to fetch any reference, so validation never touches the network or
// the disk.
package tekuri

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	tekuri "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
)

// Validator holds compiled schemas keyed by id.
type Validator struct {
	schemas map[string]*tekuri.Schema
}

var _ jsonschema.Validator = (*Validator)(nil)

// errOffline is the error of a reference that no given document satisfies.
var errOffline = errors.New("reference is not among the given schemas")

// offlineLoader refuses every URL the compiler does not already hold.
type offlineLoader struct{}

func (offlineLoader) Load(url string) (any, error) {
	return nil, fmt.Errorf("load %s: %w", url, errOffline)
}

// New compiles every document. The map key is the id of the schema, which is
// also the URL other schemas use to reference it. A document that is not valid
// JSON or does not compile makes New fail, so a broken schema never survives
// to the first validation.
func New(documents map[string][]byte) (*Validator, error) {
	compiler := tekuri.NewCompiler()
	compiler.DefaultDraft(tekuri.Draft2020)
	compiler.AssertFormat()
	compiler.UseLoader(offlineLoader{})

	ids := make([]string, 0, len(documents))
	for id := range documents {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		doc, err := tekuri.UnmarshalJSON(bytes.NewReader(documents[id]))
		if err != nil {
			return nil, fmt.Errorf("parse schema %s: %w", id, err)
		}
		if err := compiler.AddResource(id, doc); err != nil {
			return nil, fmt.Errorf("add schema %s: %w", id, err)
		}
	}

	compiled := make(map[string]*tekuri.Schema, len(ids))
	for _, id := range ids {
		schema, err := compiler.Compile(id)
		if err != nil {
			return nil, fmt.Errorf("compile schema %s: %w", id, err)
		}
		compiled[id] = schema
	}
	return &Validator{schemas: compiled}, nil
}

// IDs lists the schema ids the validator holds, sorted.
func (v *Validator) IDs() []string {
	ids := make([]string, 0, len(v.schemas))
	for id := range v.schemas {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Validate checks a document against the schema with the given id.
func (v *Validator) Validate(schemaID string, document []byte) ([]jsonschema.Violation, error) {
	schema, ok := v.schemas[schemaID]
	if !ok {
		return nil, fmt.Errorf("%w %q", jsonschema.ErrUnknownSchema, schemaID)
	}
	instance, err := tekuri.UnmarshalJSON(bytes.NewReader(document))
	if err != nil {
		return []jsonschema.Violation{{Message: "invalid JSON: " + err.Error()}}, nil
	}
	err = schema.Validate(instance)
	if err == nil {
		return nil, nil
	}
	var invalid *tekuri.ValidationError
	if !errors.As(err, &invalid) {
		return nil, fmt.Errorf("validate against %s: %w", schemaID, err)
	}
	return flatten(invalid), nil
}

// flatten collects the leaf errors of the error tree, sorted and without
// duplicates. A branch of the tree only says that something below it failed, so
// only the leaves state a rule.
//
// Version 6.0.3 of the library stores the live location slice of the validator
// in the error of a propertyNames or contentSchema keyword, and a later sibling
// check overwrites it, so the pointer of such an error is wrong. The contract
// schemas therefore use neither keyword; the schemas package test enforces it.
func flatten(invalid *tekuri.ValidationError) []jsonschema.Violation {
	seen := map[jsonschema.Violation]struct{}{}
	var out []jsonschema.Violation
	var walk func(unit tekuri.OutputUnit)
	walk = func(unit tekuri.OutputUnit) {
		if len(unit.Errors) == 0 && unit.Error != nil {
			violation := jsonschema.Violation{Pointer: unit.InstanceLocation, Message: unit.Error.String()}
			if _, dup := seen[violation]; !dup {
				seen[violation] = struct{}{}
				out = append(out, violation)
			}
		}
		for _, child := range unit.Errors {
			walk(child)
		}
	}
	walk(*invalid.DetailedOutput())
	sort.Slice(out, func(i, j int) bool {
		if out[i].Pointer != out[j].Pointer {
			return out[i].Pointer < out[j].Pointer
		}
		return out[i].Message < out[j].Message
	})
	return out
}
