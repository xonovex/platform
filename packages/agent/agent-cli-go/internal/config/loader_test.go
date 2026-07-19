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
