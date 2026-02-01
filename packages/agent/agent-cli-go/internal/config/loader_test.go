package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/types"
)

func TestLoadConfigFile_YAML(t *testing.T) {
	content := `method: docker
homeDir: /custom/home
network: true
bindPaths:
  - /path/one
  - /path/two
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	config, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFile() error = %v", err)
	}

	if config.Method != "docker" {
		t.Errorf("Method = %v, want docker", config.Method)
	}
	if config.HomeDir != "/custom/home" {
		t.Errorf("HomeDir = %v, want /custom/home", config.HomeDir)
	}
	if config.Network == nil || !*config.Network {
		t.Errorf("Network = %v, want true", config.Network)
	}
	if len(config.BindPaths) != 2 {
		t.Errorf("len(BindPaths) = %v, want 2", len(config.BindPaths))
	}
}

func TestLoadConfigFile_TOML(t *testing.T) {
	content := `method = "bwrap"
homeDir = "/toml/home"
network = false
bindPaths = ["/toml/path"]
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	config, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFile() error = %v", err)
	}

	if config.Method != "bwrap" {
		t.Errorf("Method = %v, want bwrap", config.Method)
	}
	if config.HomeDir != "/toml/home" {
		t.Errorf("HomeDir = %v, want /toml/home", config.HomeDir)
	}
	if config.Network == nil || *config.Network {
		t.Errorf("Network = %v, want false", config.Network)
	}
	if len(config.BindPaths) != 1 {
		t.Errorf("len(BindPaths) = %v, want 1", len(config.BindPaths))
	}
}

func TestLoadConfigFile_Empty(t *testing.T) {
	config, err := LoadConfigFile("")
	if err != nil {
		t.Fatalf("LoadConfigFile() error = %v", err)
	}

	if config.Method != "" {
		t.Errorf("Method = %v, want empty", config.Method)
	}
}

func TestLoadConfigFile_NotFound(t *testing.T) {
	_, err := LoadConfigFile("/nonexistent/config.yaml")
	if err != nil {
		t.Error("LoadConfigFile() should return empty config for nonexistent file")
	}
}

func TestMergeConfig_CLIOverridesFile(t *testing.T) {
	networkTrue := true
	fileConfig := &FileConfig{
		Method:      "docker",
		HomeDir:     "/file/home",
		Network:     &networkTrue,
		BindPaths:   []string{"/file/path"},
		RoBindPaths: []string{"/file/ro"},
		CustomEnv:   []string{"FILE_VAR=1"},
	}

	// CLI explicitly sets network to false (networkSet=true, network=false)
	merged := MergeConfig(fileConfig, types.SandboxBwrap, "/cli/home", true, false, []string{"/cli/path"}, nil, nil)

	if merged.Method != "bwrap" {
		t.Errorf("Method = %v, want bwrap", merged.Method)
	}
	if merged.HomeDir != "/cli/home" {
		t.Errorf("HomeDir = %v, want /cli/home", merged.HomeDir)
	}
	if merged.GetNetwork() {
		t.Errorf("Network = %v, want false", merged.GetNetwork())
	}
	// BindPaths should combine file and CLI
	if len(merged.BindPaths) != 2 {
		t.Errorf("BindPaths = %v, want 2 items", merged.BindPaths)
	}
	// RoBindPaths should combine file and CLI (CLI has none)
	if len(merged.RoBindPaths) != 1 || merged.RoBindPaths[0] != "/file/ro" {
		t.Errorf("RoBindPaths = %v, want [/file/ro]", merged.RoBindPaths)
	}
	// CustomEnv should fall back to file config
	if len(merged.CustomEnv) != 1 || merged.CustomEnv[0] != "FILE_VAR=1" {
		t.Errorf("CustomEnv = %v, want [FILE_VAR=1]", merged.CustomEnv)
	}
}

func TestMergeConfig_EmptyFallsBackToFile(t *testing.T) {
	fileConfig := &FileConfig{
		Method:  "docker",
		HomeDir: "/file/home",
	}

	// networkSet=false means we should use file config for network (which is nil, so default to true)
	merged := MergeConfig(fileConfig, "", "", false, false, nil, nil, nil)

	if merged.Method != "docker" {
		t.Errorf("Method = %v, want docker", merged.Method)
	}
	if merged.HomeDir != "/file/home" {
		t.Errorf("HomeDir = %v, want /file/home", merged.HomeDir)
	}
}

func TestLoadConfigFile_KeyValue(t *testing.T) {
	content := `# Comment line
method=nix
homeDir=/kv/home
SANDBOXHOMEDIR=/legacy/home
network=true
ENABLE_NETWORK=false
provider=glm
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	config, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFile() error = %v", err)
	}

	if config.Method != "nix" {
		t.Errorf("Method = %v, want nix", config.Method)
	}
	// ENABLE_NETWORK comes after network=true, so should override
	if config.Network == nil || *config.Network {
		t.Errorf("Network = %v, want false (ENABLE_NETWORK=false)", config.Network)
	}
	if config.Provider != "glm" {
		t.Errorf("Provider = %v, want glm", config.Provider)
	}
}
