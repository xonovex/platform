package terminal

import (
	"os/exec"
	"testing"

	termshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/terminal/shared"
)

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

func TestGetAvailableTypes(t *testing.T) {
	available := GetAvailableTypes()

	_, err := exec.LookPath("tmux")
	tmuxInstalled := err == nil

	if tmuxInstalled {
		found := false
		for _, ty := range available {
			if ty == termshared.TerminalTmux {
				found = true
				break
			}
		}
		if !found {
			t.Error("GetAvailableTypes() should include tmux when tmux is installed")
		}
	} else {
		for _, ty := range available {
			if ty == termshared.TerminalTmux {
				t.Error("GetAvailableTypes() should not include tmux when tmux is not installed")
			}
		}
	}
}
