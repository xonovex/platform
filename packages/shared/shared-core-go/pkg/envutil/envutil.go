package envutil

import (
	"sort"
	"strings"
)

// ParseCustomEnv parses KEY=VALUE environment variable strings into a map.
// Entries without an "=" are ignored.
func ParseCustomEnv(customEnv []string) map[string]string {
	env := make(map[string]string)
	for _, e := range customEnv {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
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
