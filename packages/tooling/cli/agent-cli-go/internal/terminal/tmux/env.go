package tmux

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/shell"
)

type environmentVariable struct {
	name  string
	value string
}

// filterEnvironment returns shell-safe, writable environment variables.
func filterEnvironment(env []string) []environmentVariable {
	filtered := make([]environmentVariable, 0, len(env))
	for _, entry := range env {
		name, value, found := strings.Cut(entry, "=")
		if !found || !isEnvironmentName(name) || isShellReservedVariable(name) {
			continue
		}
		filtered = append(filtered, environmentVariable{name: name, value: value})
	}
	return filtered
}

func isEnvironmentName(name string) bool {
	if name == "" || !isEnvironmentNameStart(name[0]) {
		return false
	}
	for index := 1; index < len(name); index++ {
		character := name[index]
		if !isEnvironmentNameStart(character) && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}

func isEnvironmentNameStart(character byte) bool {
	return character == '_' || character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z'
}

func isShellReservedVariable(name string) bool {
	switch name {
	case "UID", "EUID", "GID", "GROUPS":
		return true
	default:
		return false
	}
}

// createLaunchScript writes a private, self-deleting script that installs the
// requested environment before replacing itself with the agent command.
func createLaunchScript(host Host, command []string, env []string) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("tmux command is required")
	}

	path, err := host.CreateTemp("xonovex-agent-tmux-*.sh")
	if err != nil {
		return "", fmt.Errorf("create tmux launch script: %w", err)
	}

	var script strings.Builder
	script.WriteString("#!/bin/sh\n")
	script.WriteString("rm -f -- ")
	script.WriteString(shell.Quote(path))
	script.WriteByte('\n')
	for _, variable := range filterEnvironment(env) {
		script.WriteString("export ")
		script.WriteString(variable.name)
		script.WriteByte('=')
		script.WriteString(shell.Quote(variable.value))
		script.WriteByte('\n')
	}
	script.WriteString("exec ")
	script.WriteString(buildShellCommand(command))
	script.WriteByte('\n')

	if err := host.WriteFile(path, []byte(script.String()), 0o600); err != nil {
		_ = host.Remove(path)
		return "", fmt.Errorf("write tmux launch script: %w", err)
	}
	return path, nil
}
