// Package validate declares the validate section of agentic.json, which
// agent-validate reads.
package validate

import (
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// Name is the name of the section.
const Name = "validate"

// Section returns the declaration of the section. A check only reads files, so
// every file may tune every key and no digest approval applies.
func Section() spec.Section {
	return spec.Section{
		Name: Name, SchemaID: schemas.AgenticValidate,
		Keys: []spec.Key{
			{Section: Name, Pointer: "/skip", Rule: spec.Append, Layers: spec.EveryFile, Doc: "Directory names that no check descends into."},
			{Section: Name, Pointer: "/checks", Rule: spec.Merge, Layers: spec.EveryFile, Doc: "The settings of each check by name."},
		},
	}
}
