// Package shared is the harness axis core: the leaf-free command-builder port. The
// concrete leaves (claude, opencode) register in internal/plugins, and the
// composition root (internal/wiring) passes that registry to the cores.
package shared

import (
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

// CommandBuilder builds the command and args for an agent type.
type CommandBuilder interface {
	Command(run *agentv1alpha1.AgentRun, providerCliArgs []string) (command []string, args []string, err error)
}
