package agents

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// codexAgent declares Codex, which speaks the Responses protocol and takes its
// provider through `-c` arguments. It reads the credential from the variable the
// provider arguments name, so it takes the provider environment too. Codex
// sandboxes its own commands, and that fails inside another sandbox, so an
// isolation turns it off.
func codexAgent() *types.AgentConfig {
	return &types.AgentConfig{
		Type:         types.AgentCodex,
		DisplayName:  "Codex",
		Binary:       "codex",
		NixPackage:   "codex",
		Protocol:     types.ProtocolResponses,
		Access:       types.AccessProvider,
		ProviderEnv:  true,
		ProviderArgs: true,
		SandboxArgs:  []string{"--sandbox", "danger-full-access"},
		VersionArgs:  []string{"--version"},
	}
}
