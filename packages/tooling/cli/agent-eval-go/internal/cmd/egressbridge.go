package cmd

import "github.com/spf13/cobra"

// EgressBridge relays a loopback port to the allowlist proxy socket and runs a
// command beside it.
type EgressBridge interface {
	Run(socket, listen string, command []string) (int, error)
}

// newEgressBridgeCommand builds the hidden `agent-eval egress-bridge`, which a
// sandbox with network managed executes first: the isolator mounts this binary
// into the sandbox, and it relays 127.0.0.1 to the proxy socket, points the
// proxy environment at itself and runs the harness as its child. It is not part
// of the user interface.
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
				return ExitError{Code: code}
			}
			return nil
		},
	}
	command.Flags().StringVar(&socket, "socket", "", "path of the proxy unix socket")
	command.Flags().StringVar(&listen, "listen", "", "loopback address the bridge listens on")
	_ = command.MarkFlagRequired("socket")
	_ = command.MarkFlagRequired("listen")
	return command
}
