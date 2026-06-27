package tmux

import (
	"strings"
)

// bashReservedVars contains variables that bash considers read-only
// and should not be exported to avoid errors
var bashReservedVars = map[string]bool{
	"UID":    true,
	"EUID":   true,
	"GID":    true,
	"GROUPS": true,
}

// FilterEnv filters out bash reserved variables from the environment
func FilterEnv(env []string) []string {
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		idx := strings.Index(e, "=")
		if idx <= 0 {
			continue
		}
		key := e[:idx]
		if bashReservedVars[key] {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}

// EscapeEnvValue escapes double quotes in an environment variable value
func EscapeEnvValue(value string) string {
	// Escape backslashes first, then double quotes
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	// Escape dollar signs to prevent variable expansion
	value = strings.ReplaceAll(value, "$", "\\$")
	// Escape backticks to prevent command substitution
	value = strings.ReplaceAll(value, "`", "\\`")
	return value
}

// BuildEnvExports builds a shell command to export environment variables
// Format: export KEY="VALUE"; export KEY2="VALUE2";
func BuildEnvExports(env []string) string {
	if len(env) == 0 {
		return ""
	}

	filtered := FilterEnv(env)
	if len(filtered) == 0 {
		return ""
	}

	var exports []string
	for _, e := range filtered {
		idx := strings.Index(e, "=")
		if idx <= 0 {
			continue
		}
		key := e[:idx]
		value := e[idx+1:]
		escapedValue := EscapeEnvValue(value)
		exports = append(exports, "export "+key+"=\""+escapedValue+"\"")
	}

	return strings.Join(exports, "; ") + "; "
}
