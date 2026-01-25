package tmux

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "alphanumeric unchanged",
			input:    "myproject",
			expected: "myproject",
		},
		{
			name:     "dots replaced with hyphens",
			input:    "my.project.name",
			expected: "my-project-name",
		},
		{
			name:     "spaces replaced with hyphens",
			input:    "my project",
			expected: "my-project",
		},
		{
			name:     "special chars replaced",
			input:    "project@v1.0!test",
			expected: "project-v1-0-test",
		},
		{
			name:     "multiple hyphens collapsed",
			input:    "my...project",
			expected: "my-project",
		},
		{
			name:     "leading/trailing hyphens removed",
			input:    ".project.",
			expected: "project",
		},
		{
			name:     "underscores preserved",
			input:    "my_project_name",
			expected: "my_project_name",
		},
		{
			name:     "empty string returns agent",
			input:    "",
			expected: "agent",
		},
		{
			name:     "only special chars returns agent",
			input:    "...",
			expected: "agent",
		},
		{
			name:     "long names truncated",
			input:    "this-is-a-very-long-project-name-that-exceeds-thirty-characters",
			expected: "this-is-a-very-long-project-na",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateSessionName(t *testing.T) {
	// Test with a non-git directory (fallback behavior)
	t.Run("non-git directory falls back to agent prefix", func(t *testing.T) {
		result := generateSessionName("/tmp")
		// Should use fallback naming since /tmp is not a git repo
		if !strings.HasPrefix(result, "agent-") && !strings.Contains(result, "/") {
			t.Errorf("generateSessionName(/tmp) = %q, want either 'agent-*' prefix or '<repo>/<branch>' format", result)
		}
	})

	// Test with current directory (which is a git repo)
	t.Run("git directory uses repo/branch format", func(t *testing.T) {
		cwd, _ := os.Getwd()
		result := generateSessionName(cwd)
		// Should contain a slash for repo/branch format
		if !strings.Contains(result, "/") {
			t.Errorf("generateSessionName(%q) = %q, want '<repo>/<branch>' format with slash", cwd, result)
		}
	})
}

func TestGenerateWindowName(t *testing.T) {
	// Test with a non-git directory (fallback behavior)
	t.Run("non-git directory uses basename", func(t *testing.T) {
		result := generateWindowName("/tmp")
		// Should use fallback naming since /tmp is not a git repo
		if result != "tmp" && !strings.Contains(result, "/") {
			t.Errorf("generateWindowName(/tmp) = %q, want 'tmp' or '<branch>/<commit>' format", result)
		}
	})

	// Test with current directory (which is a git repo)
	t.Run("git directory uses branch/commit format", func(t *testing.T) {
		cwd, _ := os.Getwd()
		result := generateWindowName(cwd)
		// Should contain a slash for branch/commit format
		if !strings.Contains(result, "/") {
			t.Errorf("generateWindowName(%q) = %q, want '<branch>/<commit>' format with slash", cwd, result)
		}
	})
}

func TestGetGitInfo(t *testing.T) {
	t.Run("returns nil for non-git directory", func(t *testing.T) {
		info := getGitInfo("/tmp")
		if info != nil {
			t.Errorf("getGitInfo(/tmp) = %+v, want nil", info)
		}
	})

	t.Run("returns info for git directory", func(t *testing.T) {
		cwd, _ := os.Getwd()
		info := getGitInfo(cwd)
		if info == nil {
			t.Fatal("getGitInfo(cwd) = nil, want non-nil")
		}
		if info.parentDir == "" {
			t.Error("getGitInfo(cwd).parentDir is empty")
		}
		if info.repoName == "" {
			t.Error("getGitInfo(cwd).repoName is empty")
		}
		if info.branchName == "" {
			t.Error("getGitInfo(cwd).branchName is empty")
		}
		if info.shortCommit == "" {
			t.Error("getGitInfo(cwd).shortCommit is empty")
		}
	})
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string unchanged",
			input:    "simple",
			expected: "simple",
		},
		{
			name:     "path unchanged",
			input:    "/usr/bin/claude",
			expected: "/usr/bin/claude",
		},
		{
			name:     "equals sign unchanged",
			input:    "KEY=value",
			expected: "KEY=value",
		},
		{
			name:     "spaces quoted",
			input:    "hello world",
			expected: "'hello world'",
		},
		{
			name:     "single quotes escaped",
			input:    "it's",
			expected: `'it'"'"'s'`,
		},
		{
			name:     "special chars quoted",
			input:    "echo $VAR",
			expected: "'echo $VAR'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shellQuote(tt.input)
			if result != tt.expected {
				t.Errorf("shellQuote(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildShellCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple command",
			args:     []string{"echo", "hello"},
			expected: "echo hello",
		},
		{
			name:     "command with path",
			args:     []string{"/usr/bin/claude", "--model", "opus"},
			expected: "/usr/bin/claude --model opus",
		},
		{
			name:     "command with spaces in arg",
			args:     []string{"echo", "hello world"},
			expected: "echo 'hello world'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildShellCommand(tt.args)
			if result != tt.expected {
				t.Errorf("buildShellCommand(%v) = %q, want %q", tt.args, result, tt.expected)
			}
		})
	}
}

func TestExecutor_IsAvailable(t *testing.T) {
	e := NewExecutor()
	available := e.IsAvailable()

	// Check if tmux is actually installed
	_, err := exec.LookPath("tmux")
	expected := err == nil

	if available != expected {
		t.Errorf("IsAvailable() = %v, want %v", available, expected)
	}
}

func TestExecutor_IsInside(t *testing.T) {
	e := NewExecutor()
	// This test just verifies the method runs without error
	// The actual result depends on the test environment
	_ = e.IsInside()
}
