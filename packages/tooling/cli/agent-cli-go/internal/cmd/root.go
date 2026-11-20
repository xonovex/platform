package cmd

import (
	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
)

// version is set at link time through -ldflags -X.
var version = "dev"

// NewRootCommand builds the agent-cli command tree over the given logger.
func NewRootCommand(log logging.Logger) *cobra.Command {
	root := &cobra.Command{
		Use:   "agent-cli",
		Short: "Unified CLI for running AI coding agents",
		Long: `agent-cli is a unified CLI tool for running AI coding agents
with multiple model providers and sandbox options.`,
		Version: version,
	}
	root.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	root.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")
	root.AddCommand(newRunCommand(log), newCompletionCommand())
	return root
}
