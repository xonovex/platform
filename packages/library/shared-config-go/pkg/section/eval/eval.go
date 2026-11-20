// Package eval declares the eval section of agentic.json, which agent-eval
// reads.
package eval

import (
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// Name is the name of the section.
const Name = "eval"

// The budget ceilings that apply when no layer sets one. They keep an
// unattended invocation small: a larger run is resumed in further batches.
const (
	DefaultMaxModelRuns = 24
	DefaultMaxCostUSD   = 2.0
)

// Section returns the declaration of the section. An eval spends money and runs
// a harness, so the section of a project file takes effect only after the user
// approves its digest, and a project may only lower a budget.
func Section() spec.Section {
	key := func(pointer string, rule spec.Rule, doc string) spec.Key {
		return spec.Key{Section: Name, Pointer: pointer, Rule: rule, Layers: spec.EveryFile, Doc: doc}
	}
	maxModelRuns := key("/budgets/maxModelRuns", spec.Cap, "Model runs one invocation may start.")
	maxModelRuns.Ceiling = DefaultMaxModelRuns
	maxCostUSD := key("/budgets/maxCostUsd", spec.Cap, "Spend one invocation may reach, in US dollars.")
	maxCostUSD.Ceiling = DefaultMaxCostUSD
	return spec.Section{
		Name: Name, SchemaID: schemas.AgenticEval, Gated: true,
		Keys: []spec.Key{
			key("/pluginDirs", spec.Replace, "Plugin directories, relative to the file that sets them."),
			key("/instructionRoots", spec.Replace, "Instruction roots, relative to the file that sets them."),
			key("/defaultProfile", spec.Replace, "Profile that runs when the command line names none."),
			key("/profiles", spec.Merge, "Profiles by name: the harness, the model and its limits."),
			key("/protocols", spec.Merge, "Settings of each protocol by name."),
			maxModelRuns,
			maxCostUSD,
		},
	}
}
