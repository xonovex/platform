// Package audit declares the audit section of agentic.json, which agent-audit
// reads.
package audit

import (
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// Name is the name of the section.
const Name = "audit"

// Section returns the declaration of the section. The audit only reads files,
// so every file may tune every key and no digest approval applies.
func Section() spec.Section {
	key := func(pointer, doc string) spec.Key {
		return spec.Key{Section: Name, Pointer: pointer, Rule: spec.Replace, Layers: spec.EveryFile, Doc: doc}
	}
	return spec.Section{
		Name: Name, SchemaID: schemas.AgenticAudit,
		Keys: []spec.Key{
			key("/maxAge", "Days after which a review is overdue."),
			key("/versionMaxAge", "Days after which the review of a versioned source is overdue."),
			key("/coverDir", "Directory whose Markdown files each need a source, empty to turn the check off."),
		},
	}
}
