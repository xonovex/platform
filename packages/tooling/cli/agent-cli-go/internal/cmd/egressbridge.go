package cmd

import (
	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
)

// EgressBridge relays a loopback port to the allowlist proxy socket and runs a
// command beside it.
type EgressBridge interface {
	Run(socket, listen string, command []string) (int, error)
}

// newEgressBridgeCommand builds the hidden `agent-cli egress-bridge`, which the
// sandbox of a managed run executes first: this same binary is mounted into the
// sandbox, relays 127.0.0.1 to the proxy socket, points the proxy environment at
// itself, and runs the agent as its child. It is not part of the user interface.
func newEgressBridgeCommand(bridge EgressBridge) *cobra.Command {
	var socket, listen string
	command := &cobra.Command{
		Use:                   "egress-bridge --socket PATH --listen ADDRESS -- COMMAND [ARGS...]",
		Short:                 "Relay a loopback port to the egress proxy and run a command (used inside a sandbox)",
		Hidden:                true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := bridge.Run(socket, listen, args)
			if err != nil {
				return err
			}
			if code != 0 {
				cmd.SilenceErrors = true
				cmd.SilenceUsage = true
				return launch.ExitError{Code: code}
			}
			return nil
		},
	}
	command.Flags().StringVar(&socket, "socket", "", "Path of the proxy unix socket")
	command.Flags().StringVar(&listen, "listen", "", "Loopback address the bridge listens on")
	_ = command.MarkFlagRequired("socket")
	_ = command.MarkFlagRequired("listen")
	return command
}
