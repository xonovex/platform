package colors

import "testing"

func TestSupported(t *testing.T) {
	tests := []struct {
		name     string
		term     string
		noColor  string
		expected bool
	}{
		{"xterm supports color", "xterm", "", true},
		{"xterm-256color supports color", "xterm-256color", "", true},
		{"dumb terminal no color", "dumb", "", false},
		{"NO_COLOR set", "xterm", "1", false},
		{"empty TERM", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Supported(tt.term, tt.noColor); got != tt.expected {
				t.Errorf("Supported(%q, %q) = %v, want %v", tt.term, tt.noColor, got, tt.expected)
			}
		})
	}
}

func TestColorizeEnabled(t *testing.T) {
	if got, want := Colorize("test", Red, true), Red+"test"+Reset; got != want {
		t.Errorf("Colorize() = %q, want %q", got, want)
	}
}

func TestColorizeDisabled(t *testing.T) {
	if got := Colorize("test", Red, false); got != "test" {
		t.Errorf("Colorize() disabled = %q, want %q", got, "test")
	}
}
