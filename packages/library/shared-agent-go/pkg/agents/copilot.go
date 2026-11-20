package agents

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// copilotAgent declares GitHub Copilot CLI, whose model access is the login of
// the user with GitHub and takes no provider. Its command sandbox is off unless
// the user turns it on.
func copilotAgent() *types.AgentConfig {
	return &types.AgentConfig{
		Type:        types.AgentCopilot,
		DisplayName: "GitHub Copilot CLI",
		Binary:      "copilot",
		NixPackage:  "github-copilot-cli",
		Access:      types.AccessLogin,
		VersionArgs: []string{"--version"},
	}
}
