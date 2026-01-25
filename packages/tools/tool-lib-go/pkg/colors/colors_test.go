package colors

import (
	"os"
	"testing"
)

func TestIsColorSupported(t *testing.T) {
	origTerm := os.Getenv("TERM")
	origNoColor := os.Getenv("NO_COLOR")
	defer func() {
		os.Setenv("TERM", origTerm)
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

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
			os.Setenv("TERM", tt.term)
			if tt.noColor == "" {
				os.Unsetenv("NO_COLOR")
			} else {
				os.Setenv("NO_COLOR", tt.noColor)
			}

			if got := IsColorSupported(); got != tt.expected {
				t.Errorf("IsColorSupported() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestColorize(t *testing.T) {
	origTerm := os.Getenv("TERM")
	origNoColor := os.Getenv("NO_COLOR")
	defer func() {
		os.Setenv("TERM", origTerm)
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	os.Setenv("TERM", "xterm")
	os.Unsetenv("NO_COLOR")

	result := Colorize("test", Red)
	expected := Red + "test" + Reset
	if result != expected {
		t.Errorf("Colorize() = %q, want %q", result, expected)
	}
}

func TestColorizeNoColor(t *testing.T) {
	origTerm := os.Getenv("TERM")
	origNoColor := os.Getenv("NO_COLOR")
	defer func() {
		os.Setenv("TERM", origTerm)
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	os.Setenv("TERM", "xterm")
	os.Setenv("NO_COLOR", "1")

	result := Colorize("test", Red)
	if result != "test" {
		t.Errorf("Colorize() with NO_COLOR = %q, want %q", result, "test")
	}
}

func TestColorFunctions(t *testing.T) {
	origTerm := os.Getenv("TERM")
	origNoColor := os.Getenv("NO_COLOR")
	defer func() {
		os.Setenv("TERM", origTerm)
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	os.Setenv("TERM", "xterm")
	os.Unsetenv("NO_COLOR")

	tests := []struct {
		name     string
		fn       func(string) string
		color    string
		expected string
	}{
		{"WithRed", WithRed, Red, Red + "test" + Reset},
		{"WithGreen", WithGreen, Green, Green + "test" + Reset},
		{"WithYellow", WithYellow, Yellow, Yellow + "test" + Reset},
		{"WithBlue", WithBlue, Blue, Blue + "test" + Reset},
		{"WithPurple", WithPurple, Purple, Purple + "test" + Reset},
		{"WithCyan", WithCyan, Cyan, Cyan + "test" + Reset},
		{"WithGray", WithGray, Gray, Gray + "test" + Reset},
		{"WithBold", WithBold, Bold, Bold + "test" + Reset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn("test")
			if result != tt.expected {
				t.Errorf("%s() = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}
