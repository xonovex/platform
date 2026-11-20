package legacy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
)

// fileConfig is the content of one legacy configuration file. A field left
// empty means the file does not set the key.
type fileConfig struct {
	Agent     string `yaml:"agent" toml:"agent"`
	Provider  string `yaml:"provider" toml:"provider"`
	Isolation string `yaml:"isolation" toml:"isolation"`
	Provision string `yaml:"provision" toml:"provision"`
	Network   string `yaml:"network" toml:"network"`
	HomeDir   string `yaml:"homeDir" toml:"homeDir"`

	Models      []string `yaml:"models" toml:"models"`
	BindPaths   []string `yaml:"bindPaths" toml:"bindPaths"`
	RoBindPaths []string `yaml:"roBindPaths" toml:"roBindPaths"`
	CustomEnv   []string `yaml:"customEnv" toml:"customEnv"`
	AllowHosts  []string `yaml:"allowHosts" toml:"allowHosts"`
	Credentials []string `yaml:"credentials" toml:"credentials"`

	RequirePinnedProvision      bool `yaml:"requirePinnedProvision" toml:"requirePinnedProvision"`
	RequireHostToolsUnreachable bool `yaml:"requireHostToolsUnreachable" toml:"requireHostToolsUnreachable"`
	RequireEgressRestricted     bool `yaml:"requireEgressRestricted" toml:"requireEgressRestricted"`
	RequireKernelIsolation      bool `yaml:"requireKernelIsolation" toml:"requireKernelIsolation"`
}

// run returns the run section the file sets, keyed by the member names of
// agentic.json, which are the keys of the legacy file.
func (f fileConfig) run() map[string]any {
	section := map[string]any{}
	for key, text := range map[string]string{
		"agent": f.Agent, "provider": f.Provider, "isolation": f.Isolation,
		"provision": f.Provision, "network": f.Network, "homeDir": f.HomeDir,
	} {
		if text != "" {
			section[key] = text
		}
	}
	for key, items := range map[string][]string{
		"models": f.Models, "bindPaths": f.BindPaths, "roBindPaths": f.RoBindPaths,
		"customEnv": f.CustomEnv, "allowHosts": f.AllowHosts, "credentials": f.Credentials,
	} {
		if len(items) > 0 {
			section[key] = items
		}
	}
	for key, on := range map[string]bool{
		"requirePinnedProvision": f.RequirePinnedProvision, "requireHostToolsUnreachable": f.RequireHostToolsUnreachable,
		"requireEgressRestricted": f.RequireEgressRestricted, "requireKernelIsolation": f.RequireKernelIsolation,
	} {
		if on {
			section[key] = true
		}
	}
	return section
}

// agentic returns the agentic.json document that sets the run section of the
// file.
func (f fileConfig) agentic() []byte {
	document := map[string]any{"$schema": schemas.Agentic, "version": 1}
	if run := f.run(); len(run) > 0 {
		document["run"] = run
	}
	// Strings, string lists and booleans always encode.
	data, _ := json.MarshalIndent(document, "", "  ")
	return append(data, '\n')
}

// parse decodes a legacy file. A ".yaml", ".yml" or ".toml" extension selects
// the format; any other extension tries YAML, then TOML, then key=value.
func parse(data []byte, ext string) (fileConfig, error) {
	switch ext {
	case ".yaml", ".yml":
		return decodeYAML(data)
	case ".toml":
		return decodeTOML(data)
	}
	if config, err := decodeYAML(data); err == nil {
		return config, nil
	}
	if config, err := decodeTOML(data); err == nil {
		return config, nil
	}
	return parseKeyValue(string(data))
}

func decodeYAML(data []byte) (fileConfig, error) {
	var config fileConfig
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	err := decoder.Decode(&config)
	return config, err
}

func decodeTOML(data []byte) (fileConfig, error) {
	var config fileConfig
	err := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields().Decode(&config)
	// The strict error alone does not name the key; its String shows the line.
	var strict *toml.StrictMissingError
	if errors.As(err, &strict) {
		return fileConfig{}, fmt.Errorf("unknown configuration key:\n%s", strict.String())
	}
	return config, err
}

// parseKeyValue parses the key=value format, which sets only the sandbox
// home directory and the provider.
func parseKeyValue(content string) (fileConfig, error) {
	var config fileConfig
	seen := map[string]bool{}
	for number, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			return fileConfig{}, fmt.Errorf("line %d: expected key=value", number+1)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if key == "" {
			return fileConfig{}, fmt.Errorf("line %d: configuration key is empty", number+1)
		}
		if seen[key] {
			return fileConfig{}, fmt.Errorf("line %d: duplicate configuration key %q", number+1, key)
		}
		seen[key] = true
		switch key {
		case "homeDir", "SANDBOXHOMEDIR":
			config.HomeDir = value
		case "provider":
			config.Provider = value
		default:
			return fileConfig{}, fmt.Errorf("line %d: unknown configuration key %q", number+1, key)
		}
	}
	return config, nil
}
