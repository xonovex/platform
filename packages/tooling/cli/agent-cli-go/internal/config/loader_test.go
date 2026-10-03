package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfigFileLoadsLaunchFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("provider: gemini\nhomeDir: /tmp/home\nbindPaths:\n  - ./rw\nroBindPaths:\n  - ./ro\ncustomEnv:\n  - FEATURE=true\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	config, err := LoadConfigFile(path)
	if err != nil {
		t.Fatalf("LoadConfigFile() error = %v", err)
	}

	if config.Provider != "gemini" || config.HomeDir != "/tmp/home" {
		t.Errorf("scalar config = %#v, want provider and homeDir", config)
	}
	if !reflect.DeepEqual(config.BindPaths, []string{"./rw"}) || !reflect.DeepEqual(config.RoBindPaths, []string{"./ro"}) {
		t.Errorf("bind config = %#v, want rw and ro binds", config)
	}
	if !reflect.DeepEqual(config.CustomEnv, []string{"FEATURE=true"}) {
		t.Errorf("customEnv = %v, want FEATURE=true", config.CustomEnv)
	}
}

func TestLoadConfigFileRejectsMissingExplicitPath(t *testing.T) {
	if _, err := LoadConfigFile(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Error("LoadConfigFile() error = nil, want explicit missing-file error")
	}
}

func TestLoadConfigFileRejectsUnknownStructuredFields(t *testing.T) {
	tests := []struct {
		name    string
		content string
		ext     string
	}{
		{name: "yaml", content: "providre: gemini\n", ext: ".yaml"},
		{name: "toml", content: "providre = \"gemini\"\n", ext: ".toml"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config"+test.ext)
			if err := os.WriteFile(path, []byte(test.content), 0o600); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			if _, err := LoadConfigFile(path); err == nil {
				t.Error("LoadConfigFile() error = nil, want unknown-field error")
			}
		})
	}
}

func TestParseKeyValueConfigRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "malformed", content: "provider"},
		{name: "unknown", content: "providre=gemini"},
		{name: "duplicate", content: "provider=gemini\nprovider=glm"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseKeyValueConfig(test.content); err == nil {
				t.Error("parseKeyValueConfig() error = nil, want validation error")
			}
		})
	}
}

func TestParseKeyValueConfigLoadsSupportedKeys(t *testing.T) {
	config, err := parseKeyValueConfig("provider=gemini\nhomeDir=/tmp/home\n")
	if err != nil {
		t.Fatalf("parseKeyValueConfig() error = %v", err)
	}
	if config.Provider != "gemini" || config.HomeDir != "/tmp/home" {
		t.Errorf("parseKeyValueConfig() = %#v, want provider and homeDir", config)
	}
}
