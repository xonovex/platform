package legacy

import (
	"strings"
	"testing"
)

func TestParseReadsEachFormat(t *testing.T) {
	want := fileConfig{Provider: "mine", HomeDir: "/sandbox"}
	for name, c := range map[string]struct {
		data, ext string
	}{
		"yaml extension": {"provider: mine\nhomeDir: /sandbox\n", ".yaml"},
		"yml extension":  {"provider: mine\nhomeDir: /sandbox\n", ".yml"},
		"toml extension": {"provider = \"mine\"\nhomeDir = \"/sandbox\"\n", ".toml"},
		"yaml content":   {"provider: mine\nhomeDir: /sandbox\n", ""},
		"toml content":   {"provider = \"mine\"\nhomeDir = \"/sandbox\"\n", ""},
		"key=value":      {"# home\n\nprovider=mine\nhomeDir = /sandbox\n", ""},
		"old home key":   {"provider=mine\nSANDBOXHOMEDIR=/sandbox\n", ""},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := parse([]byte(c.data), c.ext)
			if err != nil || got.Provider != want.Provider || got.HomeDir != want.HomeDir {
				t.Errorf("parse() = %+v, %v, want %+v", got, err, want)
			}
		})
	}
}

func TestParseRefusesWhatTheFormatCannotHold(t *testing.T) {
	for name, c := range map[string]struct {
		data, ext, want string
	}{
		"unknown yaml key":  {"colour: red\n", ".yaml", "colour"},
		"unknown toml key":  {"colour = \"red\"\n", ".toml", "colour"},
		"no equals sign":    {"provider\n", "", "line 1: expected key=value"},
		"empty key":         {"=x\n", "", "line 1: configuration key is empty"},
		"duplicate key":     {"provider=a\nprovider=b\n", "", `line 2: duplicate configuration key "provider"`},
		"unknown plain key": {"colour=red\n", "", `line 1: unknown configuration key "colour"`},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parse([]byte(c.data), c.ext); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("parse() error = %v, want %q", err, c.want)
			}
		})
	}
}

func TestAgenticSetsOnlyWhatTheFileSets(t *testing.T) {
	full := fileConfig{
		Agent: "claude", Provider: "p", Isolation: "bwrap", Provision: "nix", Network: "none", HomeDir: "/h",
		Models: []string{"big=x"}, BindPaths: []string{"/b"}, RoBindPaths: []string{"/r"}, CustomEnv: []string{"A=1"},
		AllowHosts: []string{"h.test"}, Credentials: []string{"c"},
		RequirePinnedProvision: true, RequireHostToolsUnreachable: true, RequireEgressRestricted: true, RequireKernelIsolation: true,
	}
	got := string(full.agentic())
	for _, want := range []string{
		`"$schema": "https://schemas.xonovex.com/agent/agentic.schema.json"`, `"version": 1`,
		`"agent": "claude"`, `"isolation": "bwrap"`, `"provision": "nix"`, `"network": "none"`, `"homeDir": "/h"`,
		`"bindPaths": [`, `"roBindPaths": [`, `"customEnv": [`, `"allowHosts": [`, `"credentials": [`, `"models": [`,
		`"requirePinnedProvision": true`, `"requireHostToolsUnreachable": true`, `"requireEgressRestricted": true`, `"requireKernelIsolation": true`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("agentic() lacks %s:\n%s", want, got)
		}
	}
	if got := string(fileConfig{}.agentic()); strings.Contains(got, "run") || !strings.HasSuffix(got, "}\n") {
		t.Errorf("agentic() of an empty file = %s", got)
	}
}
