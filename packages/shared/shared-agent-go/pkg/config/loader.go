package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// FileConfig represents configuration loaded from file
type FileConfig struct {
	Method      string   `yaml:"method" toml:"method"`
	Provider    string   `yaml:"provider" toml:"provider"`
	HomeDir     string   `yaml:"homeDir" toml:"homeDir"`
	Network     *bool    `yaml:"network" toml:"network"`
	BindPaths   []string `yaml:"bindPaths" toml:"bindPaths"`
	RoBindPaths []string `yaml:"roBindPaths" toml:"roBindPaths"`
	CustomEnv   []string `yaml:"customEnv" toml:"customEnv"`
}

// GetDefaultConfigPath returns the default config file path
func GetDefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sandboxed-claude", "config")
}

// LoadConfigFile loads configuration from YAML or TOML file
// If path is empty, tries the default config path
func LoadConfigFile(path string) (*FileConfig, error) {
	if path == "" {
		// Try default config path
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
		// Try YAML first, then TOML, then key=value format
		if err = yaml.Unmarshal(data, config); err != nil {
			if err = toml.Unmarshal(data, config); err != nil {
				// Parse as simple key=value format
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

// parseKeyValueConfig parses a simple key=value config format
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
		case "method":
			config.Method = value
		case "homeDir", "SANDBOXHOMEDIR":
			config.HomeDir = value
		case "network", "ENABLE_NETWORK":
			boolVal := value == "true"
			config.Network = &boolVal
		case "provider":
			config.Provider = value
		}
	}

	return config
}

// LoadDefaultConfig loads config from the default path if it exists
func LoadDefaultConfig() *FileConfig {
	config, err := LoadConfigFile("")
	if err != nil {
		return &FileConfig{}
	}
	return config
}

// MergeConfig merges file config with CLI options
// CLI options take precedence over file config
func MergeConfig(fileConfig *FileConfig, method types.SandboxMethod, homeDir string, networkSet bool, network bool, bindPaths, roBindPaths, customEnv []string) *FileConfig {
	networkVal := network
	merged := &FileConfig{
		Method:      string(method),
		HomeDir:     homeDir,
		Network:     &networkVal,
		BindPaths:   bindPaths,
		RoBindPaths: roBindPaths,
		CustomEnv:   customEnv,
	}

	// File config provides defaults, CLI overrides
	if merged.Method == "" && fileConfig.Method != "" {
		merged.Method = fileConfig.Method
	}
	if merged.HomeDir == "" && fileConfig.HomeDir != "" {
		merged.HomeDir = fileConfig.HomeDir
	}
	// Network: use file config if CLI didn't explicitly set it
	if !networkSet && fileConfig.Network != nil {
		merged.Network = fileConfig.Network
	}
	// Combine bind paths from both sources
	if len(fileConfig.BindPaths) > 0 {
		merged.BindPaths = append(fileConfig.BindPaths, merged.BindPaths...)
	}
	if len(fileConfig.RoBindPaths) > 0 {
		merged.RoBindPaths = append(fileConfig.RoBindPaths, merged.RoBindPaths...)
	}
	if len(merged.CustomEnv) == 0 && len(fileConfig.CustomEnv) > 0 {
		merged.CustomEnv = fileConfig.CustomEnv
	}

	return merged
}

// GetNetwork returns the network value, defaulting to true if not set
func (c *FileConfig) GetNetwork() bool {
	if c.Network == nil {
		return true
	}
	return *c.Network
}
