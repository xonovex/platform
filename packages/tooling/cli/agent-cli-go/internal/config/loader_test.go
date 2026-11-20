package config

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func TestLoadReadsLaunchFields(t *testing.T) {
	files := systemtest.New().AddFile("/etc/agent/config.yaml", "provider: gemini\nhomeDir: /tmp/home\nbindPaths:\n  - ./rw\nroBindPaths:\n  - ./ro\ncustomEnv:\n  - FEATURE=true\n", 0o600)

	config, err := Load(files, "/etc/agent/config.yaml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
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

func TestLoadReadsTheDefaultPathUnderHome(t *testing.T) {
	files := systemtest.New().AddFile(DefaultPath("/home/user"), "provider=glm\n", 0o600)

	config, err := Load(files, "")
	if err != nil || config.Provider != "glm" {
		t.Fatalf("Load(default) = (%#v, %v), want provider glm", config, err)
	}
}

func TestLoadReturnsAnEmptyConfigWithoutTheDefaultFile(t *testing.T) {
	config, err := Load(systemtest.New(), "")
	if err != nil || !reflect.DeepEqual(config, FileConfig{}) {
		t.Fatalf("Load(absent default) = (%#v, %v), want an empty config", config, err)
	}
}

func TestLoadFailsWithoutAHomeForTheDefaultPath(t *testing.T) {
	files := systemtest.New()
	files.HomeErr = errors.New("no home")

	if _, err := Load(files, ""); err == nil || !strings.Contains(err.Error(), "resolve user home directory") {
		t.Fatalf("Load() error = %v, want home lookup error", err)
	}
}

func TestLoadRejectsMissingExplicitPath(t *testing.T) {
	if _, err := Load(systemtest.New(), "/etc/agent/missing.yaml"); err == nil {
		t.Error("Load() error = nil, want explicit missing-file error")
	}
}

func TestParseRejectsUnknownStructuredFields(t *testing.T) {
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
			if _, err := Parse([]byte(test.content), test.ext); err == nil {
				t.Error("Parse() error = nil, want unknown-field error")
			}
		})
	}
}

func TestParseFallsBackFromYAMLToTOMLToKeyValue(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "yaml", content: "provider: gemini\n"},
		{name: "toml", content: "provider = \"gemini\"\n"},
		{name: "key value", content: "provider=gemini\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := Parse([]byte(test.content), "")
			if err != nil || config.Provider != "gemini" {
				t.Errorf("Parse() = (%#v, %v), want provider gemini", config, err)
			}
		})
	}
	if _, err := Parse([]byte("not a config"), ""); err == nil {
		t.Error("Parse(invalid) error = nil, want an error")
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
		{name: "empty key", content: "=gemini"},
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
	config, err := parseKeyValueConfig("# comment\n\nprovider=gemini\nSANDBOXHOMEDIR=/tmp/home\n")
	if err != nil {
		t.Fatalf("parseKeyValueConfig() error = %v", err)
	}
	if config.Provider != "gemini" || config.HomeDir != "/tmp/home" {
		t.Errorf("parseKeyValueConfig() = %#v, want provider and homeDir", config)
	}
}
