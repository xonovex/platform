package cmd

import (
	"github.com/spf13/cobra"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "agent-cli",
	Short: "Unified CLI for running AI coding agents",
	Long: `agent-cli is a unified CLI tool for running AI coding agents
with multiple model providers and sandbox options.`,
	Version: version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")
}
