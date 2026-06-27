// Package config loads the CLI's optional file configuration (YAML/TOML/key=value).
// It was formerly shared pkg/config; it has a single consumer (this CLI), so it
// and its go-toml/yaml dependencies live in the CLI module, not the shared one.
package config

import (
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
func GetDefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sandboxed-claude", "config")
}

// LoadConfigFile loads configuration from a YAML or TOML file. If path is empty,
// it tries the default config path.
func LoadConfigFile(path string) (*FileConfig, error) {
	if path == "" {
		defaultPath := GetDefaultConfigPath()
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			return &FileConfig{}, nil
		}
		path = defaultPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &FileConfig{}, nil
		}
		return nil, err
	}

	config := &FileConfig{}
	ext := filepath.Ext(path)

	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, config)
	case ".toml":
		err = toml.Unmarshal(data, config)
	default:
		// Try YAML first, then TOML, then key=value format.
		if err = yaml.Unmarshal(data, config); err != nil {
			if err = toml.Unmarshal(data, config); err != nil {
				config = parseKeyValueConfig(string(data))
				err = nil
			}
		}
	}

	if err != nil {
		return nil, err
	}

	return config, nil
}

// parseKeyValueConfig parses a simple key=value config format.
func parseKeyValueConfig(content string) *FileConfig {
	config := &FileConfig{}

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "homeDir", "SANDBOXHOMEDIR":
			config.HomeDir = value
		case "provider":
			config.Provider = value
		}
	}

	return config
}

// LoadDefaultConfig loads config from the default path if it exists.
func LoadDefaultConfig() *FileConfig {
	config, err := LoadConfigFile("")
	if err != nil {
		return &FileConfig{}
	}
	return config
}
