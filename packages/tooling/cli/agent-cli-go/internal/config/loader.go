// Package config loads the configuration of the CLI from its layers and merges
// them. A layer is a file (YAML, TOML or key=value) or the command line; each key
// declares its own merge rule as data in Specs, so no layer and no key is named
// by the merge itself.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// FileConfig is the content of one configuration file. Every field is one key of
// Specs; a field left empty means the file does not set the key.
type FileConfig struct {
	Agent     string `yaml:"agent" toml:"agent"`
	Provider  string `yaml:"provider" toml:"provider"`
	Isolation string `yaml:"isolation" toml:"isolation"`
	Provision string `yaml:"provision" toml:"provision"`
	Network   string `yaml:"network" toml:"network"`
	HomeDir   string `yaml:"homeDir" toml:"homeDir"`

	// Models override model slots of the provider, as "slot=id".
	Models []string `yaml:"models" toml:"models"`

	BindPaths   []string `yaml:"bindPaths" toml:"bindPaths"`
	RoBindPaths []string `yaml:"roBindPaths" toml:"roBindPaths"`
	CustomEnv   []string `yaml:"customEnv" toml:"customEnv"`
	// AllowHosts are the hosts of a managed network run, as "host", "host:port" or
	// "*.domain[:port]".
	AllowHosts []string `yaml:"allowHosts" toml:"allowHosts"`
	// Credentials bind an environment variable to a credential store entry, as
	// "NAME=scheme:locator". The file holds where a credential is, never its value.
	Credentials []string `yaml:"credentials" toml:"credentials"`

	RequirePinnedProvision      bool `yaml:"requirePinnedProvision" toml:"requirePinnedProvision"`
	RequireHostToolsUnreachable bool `yaml:"requireHostToolsUnreachable" toml:"requireHostToolsUnreachable"`
	RequireEgressRestricted     bool `yaml:"requireEgressRestricted" toml:"requireEgressRestricted"`
	RequireKernelIsolation      bool `yaml:"requireKernelIsolation" toml:"requireKernelIsolation"`
}

// Entries returns the keys the file sets.
func (f FileConfig) Entries() Entries {
	entries := Entries{}
	for key, text := range map[Key]string{
		KeyAgent: f.Agent, KeyProvider: f.Provider, KeyIsolation: f.Isolation,
		KeyProvision: f.Provision, KeyNetwork: f.Network, KeyHomeDir: f.HomeDir,
	} {
		if text != "" {
			entries[key] = Value{Text: text}
		}
	}
	for key, items := range map[Key][]string{
		KeyModels: f.Models, KeyBindPaths: f.BindPaths, KeyRoBindPaths: f.RoBindPaths, KeyCustomEnv: f.CustomEnv,
		KeyAllowHosts: f.AllowHosts, KeyCredentials: f.Credentials,
	} {
		if len(items) > 0 {
			entries[key] = Value{Items: items}
		}
	}
	for key, on := range map[Key]bool{
		KeyRequirePinnedProvision: f.RequirePinnedProvision, KeyRequireHostToolsUnreachable: f.RequireHostToolsUnreachable,
		KeyRequireEgressRestricted: f.RequireEgressRestricted, KeyRequireKernelIsolation: f.RequireKernelIsolation,
	} {
		if on {
			entries[key] = Value{On: true}
		}
	}
	return entries
}

// Files is the port the loader reads the configuration through.
type Files interface {
	ReadFile(path string) ([]byte, error)
	Stat(path string) (fs.FileInfo, error)
	UserHomeDir() (string, error)
	// ManagedConfigDir is the directory of the system-wide configuration.
	ManagedConfigDir() string
}

// DefaultPath returns the default configuration path under home.
func DefaultPath(home string) string {
	return filepath.Join(home, ".config", "sandboxed-claude", "config")
}

// Load reads the configuration at path, which must exist. An empty path reads
// the default configuration, and an absent default yields an empty
// configuration.
func Load(files Files, path string) (FileConfig, error) {
	explicit := path != ""
	if !explicit {
		home, err := files.UserHomeDir()
		if err != nil {
			return FileConfig{}, fmt.Errorf("resolve user home directory: %w", err)
		}
		path = DefaultPath(home)
	}

	data, err := files.ReadFile(path)
	if err != nil {
		if !explicit && errors.Is(err, fs.ErrNotExist) {
			return FileConfig{}, nil
		}
		return FileConfig{}, err
	}
	return Parse(data, filepath.Ext(path))
}

// Parse decodes configuration data. A ".yaml", ".yml" or ".toml" extension
// selects the format; any other extension tries YAML, then TOML, then key=value.
func Parse(data []byte, ext string) (FileConfig, error) {
	var config FileConfig
	var err error
	switch ext {
	case ".yaml", ".yml":
		err = decodeYAML(data, &config)
	case ".toml":
		err = decodeTOML(data, &config)
	default:
		if err = decodeYAML(data, &config); err != nil {
			config = FileConfig{}
			if err = decodeTOML(data, &config); err != nil {
				config, err = parseKeyValueConfig(string(data))
			}
		}
	}
	if err != nil {
		return FileConfig{}, err
	}
	return config, nil
}

func decodeYAML(data []byte, config *FileConfig) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	return decoder.Decode(config)
}

func decodeTOML(data []byte, config *FileConfig) error {
	return toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields().Decode(config)
}

// parseKeyValueConfig parses a simple key=value config format.
func parseKeyValueConfig(content string) (FileConfig, error) {
	var config FileConfig
	seen := make(map[string]struct{})

	for lineNumber, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return FileConfig{}, fmt.Errorf("line %d: expected key=value", lineNumber+1)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return FileConfig{}, fmt.Errorf("line %d: configuration key is empty", lineNumber+1)
		}
		if _, exists := seen[key]; exists {
			return FileConfig{}, fmt.Errorf("line %d: duplicate configuration key %q", lineNumber+1, key)
		}
		seen[key] = struct{}{}

		switch key {
		case "homeDir", "SANDBOXHOMEDIR":
			config.HomeDir = value
		case "provider":
			config.Provider = value
		default:
			return FileConfig{}, fmt.Errorf("line %d: unknown configuration key %q", lineNumber+1, key)
		}
	}

	return config, nil
}
