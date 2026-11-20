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
// approves its digest, and a project may only lower a budget. The sandbox keys
// follow the confinement keys of the run section: a project may demand a
// guarantee, and only the user and the administrator choose the confinement and
// what it may reach.
func Section() spec.Section {
	key := func(pointer string, rule spec.Rule, doc string) spec.Key {
		return spec.Key{Section: Name, Pointer: pointer, Rule: rule, Layers: spec.EveryFile, Doc: doc}
	}
	trusted := func(pointer string, rule spec.Rule, doc string) spec.Key {
		return spec.Key{Section: Name, Pointer: pointer, Rule: rule, Layers: spec.TrustedFiles, Doc: doc}
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
			trusted("/sandbox/isolation", spec.Replace, "Isolation axis of every harness run."),
			trusted("/sandbox/provision", spec.Replace, "Provision axis of every harness run."),
			trusted("/sandbox/network", spec.Replace, "Network egress axis of every harness run."),
			trusted("/sandbox/image", spec.Replace, "Container image of a docker run."),
			trusted("/sandbox/runtime", spec.Replace, "Container runtime of a docker run."),
			trusted("/sandbox/hostPassthrough", spec.Replace, "Expose the host tools to a bwrap run."),
			trusted("/sandbox/initCommands", spec.Append, "Commands a run with provision command executes first."),
			trusted("/sandbox/nixPackages", spec.Append, "Packages a run with provision nix adds."),
			trusted("/sandbox/nixRev", spec.Replace, "Pinned nixpkgs revision of a run with provision nix."),
			trusted("/sandbox/allowHosts", spec.Append, "Hosts a run with network managed may reach."),
			trusted("/sandbox/credentials", spec.Append, "Credential bindings of the harness, as NAME=scheme:locator."),
			trusted("/sandbox/customEnv", spec.Append, "Environment variables of the harness, as KEY=VALUE."),
			key("/sandbox/requirePinnedProvision", spec.Require, "Require provisioning from a pinned source."),
			key("/sandbox/requireHostToolsUnreachable", spec.Require, "Require host tools to be unreachable from the harness."),
			key("/sandbox/requireEgressRestricted", spec.Require, "Require network egress of the harness to be disabled or enforced."),
			key("/sandbox/requireKernelIsolation", spec.Require, "Require a kernel-isolating runtime."),
		},
	}
}
