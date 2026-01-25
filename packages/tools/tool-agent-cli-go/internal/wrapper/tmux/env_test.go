package tmux

import (
	"strings"
	"testing"
)

func TestFilterEnv(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "filters bash reserved vars",
			input:    []string{"PATH=/usr/bin", "UID=1000", "HOME=/home/user", "EUID=1000", "GID=1000", "GROUPS=1000"},
			expected: []string{"PATH=/usr/bin", "HOME=/home/user"},
		},
		{
			name:     "keeps all non-reserved vars",
			input:    []string{"FOO=bar", "BAZ=qux", "CUSTOM_VAR=value"},
			expected: []string{"FOO=bar", "BAZ=qux", "CUSTOM_VAR=value"},
		},
		{
			name:     "handles empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "filters invalid entries",
			input:    []string{"VALID=value", "invalid", "=nokey", "ANOTHER=ok"},
			expected: []string{"VALID=value", "ANOTHER=ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterEnv(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("FilterEnv() returned %d items, want %d", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("FilterEnv()[%d] = %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestEscapeEnvValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain string unchanged",
			input:    "simple",
			expected: "simple",
		},
		{
			name:     "escapes double quotes",
			input:    `value with "quotes"`,
			expected: `value with \"quotes\"`,
		},
		{
			name:     "escapes backslashes",
			input:    `path\to\file`,
			expected: `path\\to\\file`,
		},
		{
			name:     "escapes dollar signs",
			input:    "value $VAR here",
			expected: `value \$VAR here`,
		},
		{
			name:     "escapes backticks",
			input:    "value `cmd` here",
			expected: "value \\`cmd\\` here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeEnvValue(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeEnvValue(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildEnvExports(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		contains []string
		isEmpty  bool
	}{
		{
			name:    "empty input returns empty string",
			input:   []string{},
			isEmpty: true,
		},
		{
			name:     "builds export statements",
			input:    []string{"FOO=bar", "BAZ=qux"},
			contains: []string{`export FOO="bar"`, `export BAZ="qux"`},
		},
		{
			name:     "escapes values in exports",
			input:    []string{`KEY=value with "quotes"`},
			contains: []string{`export KEY="value with \"quotes\""`},
		},
		{
			name:    "filters reserved vars",
			input:   []string{"UID=1000", "EUID=1000"},
			isEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildEnvExports(tt.input)
			if tt.isEmpty {
				if result != "" {
					t.Errorf("BuildEnvExports() = %q, want empty", result)
				}
				return
			}
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("BuildEnvExports() = %q, want to contain %q", result, substr)
				}
			}
			// Should end with "; "
			if !strings.HasSuffix(result, "; ") {
				t.Errorf("BuildEnvExports() = %q, should end with '; '", result)
			}
		})
	}
}
