package cmd

import (
	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
)

// version is set at link time through -ldflags -X.
var version = "dev"

// NewRootCommand builds the agent-cli command tree; every run executes through
// deps, the plugin and marketplace commands through ports, and the hidden
// egress-bridge command through bridge.
func NewRootCommand(deps launch.Dependencies, ports PluginPorts, bridge EgressBridge) *cobra.Command {
	root := &cobra.Command{
		Use:   "agent-cli",
		Short: "Unified CLI for running AI coding agents",
		Long: `agent-cli is a unified CLI tool for running AI coding agents
with multiple model providers and sandbox options.`,
		Version: version,
	}
	root.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	root.AddCommand(newRunCommand(deps), newPluginCommand(ports), newMarketplaceCommand(ports), newCompletionCommand(), newEgressBridgeCommand(bridge))
	return root
}
