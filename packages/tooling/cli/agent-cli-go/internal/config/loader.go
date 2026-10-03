// Package config loads the CLI's optional file configuration (YAML/TOML/key=value).
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// FileConfig represents configuration loaded from file.
type FileConfig struct {
	Provider    string   `yaml:"provider" toml:"provider"`
	HomeDir     string   `yaml:"homeDir" toml:"homeDir"`
	BindPaths   []string `yaml:"bindPaths" toml:"bindPaths"`
	RoBindPaths []string `yaml:"roBindPaths" toml:"roBindPaths"`
	CustomEnv   []string `yaml:"customEnv" toml:"customEnv"`
}

// GetDefaultConfigPath returns the default config file path.
func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(home, ".config", "sandboxed-claude", "config"), nil
}

// LoadConfigFile loads configuration from a YAML or TOML file. If path is empty,
// it tries the default config path.
func LoadConfigFile(path string) (*FileConfig, error) {
	explicit := path != ""
	if path == "" {
		defaultPath, err := GetDefaultConfigPath()
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			return &FileConfig{}, nil
		} else if err != nil {
			return nil, fmt.Errorf("inspect default config %q: %w", defaultPath, err)
		}
		path = defaultPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && !explicit {
			return &FileConfig{}, nil
		}
		return nil, err
	}

	config := &FileConfig{}
	ext := filepath.Ext(path)

	switch ext {
	case ".yaml", ".yml":
		err = decodeYAML(data, config)
	case ".toml":
		err = decodeTOML(data, config)
	default:
		if err = decodeYAML(data, config); err != nil {
			if err = decodeTOML(data, config); err != nil {
				config, err = parseKeyValueConfig(string(data))
			}
		}
	}

	if err != nil {
		return nil, err
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
func parseKeyValueConfig(content string) (*FileConfig, error) {
	config := &FileConfig{}
	seen := make(map[string]struct{})

	for lineNumber, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: expected key=value", lineNumber+1)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("line %d: configuration key is empty", lineNumber+1)
		}
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("line %d: duplicate configuration key %q", lineNumber+1, key)
		}
		seen[key] = struct{}{}

		switch key {
		case "homeDir", "SANDBOXHOMEDIR":
			config.HomeDir = value
		case "provider":
			config.Provider = value
		default:
			return nil, fmt.Errorf("line %d: unknown configuration key %q", lineNumber+1, key)
		}
	}

	return config, nil
}
