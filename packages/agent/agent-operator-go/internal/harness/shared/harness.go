// Package shared is the harness axis core: the leaf-free command-builder port. The
// agent-type registry that wires concrete leaves (claude, opencode) lives in the
// composition root (internal/plugins).
package shared

import (
	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// CommandBuilder builds the command and args for an agent type.
type CommandBuilder interface {
	Command(run *agentv1alpha1.AgentRun) (command []string, args []string)
}
