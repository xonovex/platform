package wrapper

import (
	"os/exec"
	"testing"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func TestGetExecutor(t *testing.T) {
	tests := []struct {
		name         string
		terminalType types.TerminalType
		expectNil    bool
	}{
		{
			name:         "tmux returns executor",
			terminalType: types.TerminalTmux,
			expectNil:    false,
		},
		{
			name:         "empty type returns nil",
			terminalType: types.TerminalNone,
			expectNil:    true,
		},
		{
			name:         "unknown type returns nil",
			terminalType: types.TerminalType("unknown"),
			expectNil:    true,
		},
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

	// Check if tmux is installed
	_, err := exec.LookPath("tmux")
	tmuxInstalled := err == nil

	if tmuxInstalled {
		found := false
		for _, t := range available {
			if t == types.TerminalTmux {
				found = true
				break
			}
		}
		if !found {
			t.Error("GetAvailableTypes() should include tmux when tmux is installed")
		}
	} else {
		for _, tt := range available {
			if tt == types.TerminalTmux {
				t.Error("GetAvailableTypes() should not include tmux when tmux is not installed")
			}
		}
	}
}
