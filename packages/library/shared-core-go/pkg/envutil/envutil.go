package envutil

import (
	"fmt"
	"sort"
	"strings"
)

// ParseEnv parses environment entries received from the operating system.
func ParseEnv(entries []string) map[string]string {
	env := make(map[string]string)
	for _, e := range entries {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}

// ParseCustomEnv validates and parses caller-supplied KEY=VALUE entries.
func ParseCustomEnv(entries []string) (map[string]string, error) {
	env := make(map[string]string, len(entries))
	for _, entry := range entries {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("environment entry %q must use KEY=VALUE", entry)
		}
		if !validEnvName(parts[0]) {
			return nil, fmt.Errorf("environment variable name %q is invalid", parts[0])
		}
		env[parts[0]] = parts[1]
	}
	return env, nil
}

func validEnvName(name string) bool {
	if name == "" || !isEnvNameStart(name[0]) {
		return false
	}
	for index := 1; index < len(name); index++ {
		character := name[index]
		if !isEnvNameStart(character) && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}

func isEnvNameStart(character byte) bool {
	return character == '_' || character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z'
}

// MergeEnvMaps merges multiple environment maps; later maps override earlier.
func MergeEnvMaps(envMaps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, envMap := range envMaps {
		for k, v := range envMap {
			result[k] = v
		}
	}
	return result
}

// EnvMapToSlice converts a map of environment variables to a KEY=VALUE slice.
func EnvMapToSlice(envMap map[string]string) []string {
	keys := make([]string, 0, len(envMap))
	for key := range envMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	env := make([]string, 0, len(keys))
	for _, key := range keys {
		env = append(env, key+"="+envMap[key])
	}
	return env
}
