package discover_test

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/discover"
)

func env(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func TestHostDirectories(t *testing.T) {
	cases := map[string]struct {
		host                 discover.Host
		managed, user, state string
	}{
		"linux defaults": {
			discover.Host{GOOS: "linux", Getenv: env(nil), Home: "/home/u"},
			"/etc/agentic", "/home/u/.config/agentic", "/home/u/.local/state/agentic",
		},
		"xdg variables": {
			discover.Host{GOOS: "linux", Getenv: env(map[string]string{"XDG_CONFIG_HOME": "/c", "XDG_STATE_HOME": "/s"}), Home: "/home/u"},
			"/etc/agentic", "/c/agentic", "/s/agentic",
		},
		"relative xdg is ignored": {
			discover.Host{GOOS: "linux", Getenv: env(map[string]string{"XDG_CONFIG_HOME": "c"}), Home: "/home/u"},
			"/etc/agentic", "/home/u/.config/agentic", "/home/u/.local/state/agentic",
		},
		"macos": {
			discover.Host{GOOS: "darwin", Getenv: env(nil), Home: "/Users/u"},
			"/Library/Application Support/agentic", "/Users/u/.config/agentic", "/Users/u/.local/state/agentic",
		},
		"windows": {
			discover.Host{GOOS: "windows", Getenv: env(map[string]string{"ProgramData": "/pd", "AppData": "/ad", "LocalAppData": "/lad"}), Home: "/h"},
			"/pd/agentic", "/ad/agentic", "/lad/agentic",
		},
		"windows without variables": {
			discover.Host{GOOS: "windows", Getenv: env(nil), Home: "/h"},
			"", "/h/.config/agentic", "/h/.local/state/agentic",
		},
	}
	for name, c := range cases {
		user, err := c.host.UserDir()
		if err != nil {
			t.Fatal(err)
		}
		state, err := c.host.StateDir()
		if err != nil {
			t.Fatal(err)
		}
		if managed := c.host.ManagedDir(); managed != c.managed || user != c.user || state != c.state {
			t.Errorf("%s: %q %q %q", name, managed, user, state)
		}
	}
}

func TestHostWithoutAHome(t *testing.T) {
	host := discover.Host{GOOS: "linux", Getenv: env(nil)}
	if _, err := host.UserDir(); !errors.Is(err, discover.ErrNoHome) {
		t.Fatalf("UserDir = %v", err)
	}
}
