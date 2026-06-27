package sandboxutil

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/scriptlib"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/shell"
)

// SpawnSandbox spawns a sandbox process and waits for completion
func SpawnSandbox(command string, args []string, env []string, errorPrefix string, verbose bool) (int, error) {
	if verbose {
		scriptlib.LogDebug(verbose, fmt.Sprintf("Executing: %s %s", command, strings.Join(args, " ")))
	}

	cmd := exec.Command(command, args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, fmt.Errorf("%s: %w", errorPrefix, err)
	}

	return 0, nil
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

	// Build a shell command that chains init commands with &&, then runs the agent
	initChain := strings.Join(initCommands, " && ")
	mainCommand := buildShellCommand(command)
	fullCommand := initChain + " && exec " + mainCommand

	return []string{"sh", "-c", fullCommand}
}
