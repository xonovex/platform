// Package providers declares the providers section of agentic.json: the model
// provider routes a run may select, by name.
package providers

import (
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// Name is the name of the section.
const Name = "providers"

// Section returns the declaration of the section. A route comes whole from one
// layer, so a project cannot change part of a route that the administrator set.
func Section() spec.Section {
	return spec.Section{
		Name: Name, SchemaID: schemas.AgenticProviders, Gated: true,
		Keys: []spec.Key{
			{Section: Name, Rule: spec.Merge, Layers: spec.EveryFile, Doc: "Provider routes by name."},
		},
	}
}
