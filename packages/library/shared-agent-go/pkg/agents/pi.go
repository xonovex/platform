package agents

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// piAgent declares Pi, which speaks the Responses protocol through its Azure
// provider and reads the endpoint and key from the provider environment. Pi has
// no command sandbox of its own.
func piAgent() *types.AgentConfig {
	return &types.AgentConfig{
		Type:         types.AgentPi,
		DisplayName:  "Pi",
		Binary:       "pi",
		NixPackage:   "pi-coding-agent",
		Protocol:     types.ProtocolResponses,
		Access:       types.AccessProvider,
		ProviderEnv:  true,
		ProviderArgs: true,
		VersionArgs:  []string{"--version"},
	}
}
