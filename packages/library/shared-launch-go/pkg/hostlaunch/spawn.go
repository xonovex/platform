package hostlaunch

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/shell"
)

// SpawnSandbox starts command in dir under env through host, waits for it, and
// returns its exit code. An empty dir inherits the working directory of this process.
func SpawnSandbox(host Host, log logging.Logger, command string, args, env []string, dir, errorPrefix string) (int, error) {
	log.Debug(fmt.Sprintf("Executing: %s %s", command, strings.Join(args, " ")))
	code, err := host.Run(command, args, env, dir)
	if err != nil {
		return 1, fmt.Errorf("%s: %w", errorPrefix, err)
	}
	return code, nil
}

// buildShellCommand builds a shell command string from an array of arguments.
func buildShellCommand(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shell.Quote(arg)
	}
	return strings.Join(quoted, " ")
}

// WrapWithInitCommands wraps a command with init commands that run before it.
// If no init commands are provided, returns the original command unchanged.
// Otherwise, returns a shell command that runs all init commands in sequence,
// stopping on first failure, then runs the main command.
func WrapWithInitCommands(command []string, initCommands []string) []string {
	if len(initCommands) == 0 {
		return command
	}

	initChain := strings.Join(initCommands, " && ")
	mainCommand := buildShellCommand(command)
	fullCommand := initChain + " && exec " + mainCommand

	return []string{"sh", "-c", fullCommand}
}
