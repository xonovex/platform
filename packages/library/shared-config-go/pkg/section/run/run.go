// Package run declares the run section of agentic.json, which agent-cli reads
// to plan a sandboxed run.
package run

import (
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// Name is the name of the section.
const Name = "run"

// Section returns the declaration of the section. A project file may set only
// the keys that select the agent and the provider or demand a guarantee; the keys
// that widen what the sandbox reaches belong to the user and the administrator.
func Section() spec.Section {
	key := func(pointer string, rule spec.Rule, layers []spec.Layer, doc string) spec.Key {
		return spec.Key{Section: Name, Pointer: pointer, Rule: rule, Layers: layers, Doc: doc}
	}
	return spec.Section{
		Name: Name, SchemaID: schemas.AgenticRun, Gated: true,
		Keys: []spec.Key{
			key("/agent", spec.Replace, spec.EveryFile, "Agent to run."),
			key("/provider", spec.Replace, spec.EveryFile, "Model provider to use, by name."),
			key("/models", spec.Append, spec.EveryFile, "Model slot overrides of the provider, as slot=id."),
			key("/isolation", spec.Replace, spec.TrustedFiles, "Isolation axis."),
			key("/provision", spec.Replace, spec.TrustedFiles, "Provision axis."),
			key("/network", spec.Replace, spec.TrustedFiles, "Network egress axis."),
			key("/image", spec.Replace, spec.TrustedFiles, "Container image of a docker or openshell run."),
			key("/homeDir", spec.Replace, spec.TrustedFiles, "Home directory of the sandbox."),
			key("/bindPaths", spec.Append, spec.TrustedFiles, "Read-write bind mounts."),
			key("/roBindPaths", spec.Append, spec.TrustedFiles, "Read-only bind mounts."),
			key("/customEnv", spec.Append, spec.TrustedFiles, "Environment variables, as KEY=VALUE."),
			key("/allowHosts", spec.Append, spec.TrustedFiles, "Hosts a managed network run may reach."),
			key("/credentials", spec.Append, spec.TrustedFiles, "Credential bindings, as NAME=scheme:locator."),
			key("/requirePinnedProvision", spec.Require, spec.EveryFile, "Require provisioning from a pinned source."),
			key("/requireHostToolsUnreachable", spec.Require, spec.EveryFile, "Require host tools to be unreachable from the sandbox."),
			key("/requireEgressRestricted", spec.Require, spec.EveryFile, "Require network egress to be disabled or enforced."),
			key("/requireKernelIsolation", spec.Require, spec.EveryFile, "Require a kernel-isolating runtime."),
		},
	}
}
