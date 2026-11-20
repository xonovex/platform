package agents

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// kiroAgent declares Kiro CLI, whose model access is the login of the user with
// the vendor and takes no provider. It starts the chat program directly, because
// the `kiro-cli` router starts a program at a path under the real home
// directory, and the nix package that holds the chat program is the unwrapped one.
func kiroAgent() *types.AgentConfig {
	return &types.AgentConfig{
		Type:        types.AgentKiro,
		DisplayName: "Kiro CLI",
		Binary:      "kiro-cli-chat",
		NixPackage:  "kiro-cli-unwrapped",
		Access:      types.AccessLogin,
		LaunchArgs:  []string{"chat", "--agent-engine", "v3"},
		VersionArgs: []string{"--version"},
	}
}
