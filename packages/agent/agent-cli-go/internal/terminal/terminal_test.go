package terminal

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	termshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/terminal/shared"
)

// withTmuxOnPath points PATH at a directory holding an executable named tmux, so
// availability is a property of the test rather than of the host.
func withTmuxOnPath(t *testing.T, installed bool) {
	t.Helper()
	directory := t.TempDir()
	if installed {
		script := []byte("#!/bin/sh\nexit 0\n")
		if err := os.WriteFile(filepath.Join(directory, "tmux"), script, 0o700); err != nil {
			t.Fatalf("write fake tmux: %v", err)
		}
	}
	t.Setenv("PATH", directory)
}

func TestGetExecutor(t *testing.T) {
	tests := []struct {
		name         string
		terminalType termshared.TerminalType
		expectNil    bool
	}{
		{"tmux returns executor", termshared.TerminalTmux, false},
		{"empty type returns nil", termshared.TerminalNone, true},
		{"unknown type returns nil", termshared.TerminalType("unknown"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetExecutor(tt.terminalType)
			if tt.expectNil && result != nil {
				t.Errorf("GetExecutor(%q) = %v, want nil", tt.terminalType, result)
			}
			if !tt.expectNil && result == nil {
				t.Errorf("GetExecutor(%q) = nil, want non-nil", tt.terminalType)
			}
		})
	}
}

func TestGetAvailableTypesReportsTmuxWhenItResolves(t *testing.T) {
	withTmuxOnPath(t, true)

	available := GetAvailableTypes()

	if !slices.Contains(available, termshared.TerminalTmux) {
		t.Errorf("GetAvailableTypes() = %v, want it to include %q", available, termshared.TerminalTmux)
	}
}

func TestGetAvailableTypesOmitsTmuxWhenItDoesNotResolve(t *testing.T) {
	withTmuxOnPath(t, false)

	available := GetAvailableTypes()

	if slices.Contains(available, termshared.TerminalTmux) {
		t.Errorf("GetAvailableTypes() = %v, want it to omit %q", available, termshared.TerminalTmux)
	}
}
