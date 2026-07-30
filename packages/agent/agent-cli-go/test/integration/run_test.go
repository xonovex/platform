//go:build integration
// +build integration

package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func testBinary(t testing.TB) string {
	t.Helper()
	path := os.Getenv("AGENT_CLI_BINARY")
	if path == "" {
		t.Fatal("AGENT_CLI_BINARY is required")
	}
	return path
}

func testCommand(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	home := t.TempDir()
	cmd := exec.Command(testBinary(t), args...)
	cmd.Env = append(os.Environ(), "HOME="+home, "XDG_CONFIG_HOME="+home)
	return cmd
}

func TestRunCommand_Help(t *testing.T) {
	cmd := testCommand(t, "run", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run --help failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	expectedTexts := []string{
		"Run an AI coding agent",
		"--agent",
		"--provider",
		"--isolation",
		"--provision",
		"--network",
		"--isolation-docker-runtime",
		"--isolation-bwrap-passthrough",
		"--require-pinned-provision",
		"--require-host-tools-unreachable",
		"--require-egress-restricted",
		"--require-kernel-isolation",
		"--work-dir",
		"--worktree-branch",
		"--config",
	}

	for _, expected := range expectedTexts {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Help output missing expected text: %s", expected)
		}
	}
}

func TestVersion(t *testing.T) {
	cmd := testCommand(t, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version failed: %v\nOutput: %s", err, output)
	}

	expected := os.Getenv("AGENT_CLI_VERSION")
	if expected == "" {
		t.Fatal("AGENT_CLI_VERSION is required")
	}
	if outputStr := string(output); !strings.Contains(outputStr, expected) {
		t.Errorf("Version output = %q, want %q", outputStr, expected)
	}
}

func TestRunCommand_InvalidAgent(t *testing.T) {
	cmd := testCommand(t, "run", "-a", "invalid-agent-xyz")
	output, err := cmd.CombinedOutput()

	// Should fail with error
	if err == nil {
		t.Errorf("Expected error for invalid agent, got nil")
	}

	// Should mention available agents
	if !strings.Contains(string(output), "unknown agent type") {
		t.Errorf("Expected 'unknown agent type' error, got: %s", output)
	}
}

func TestRunCommand_InvalidIsolation(t *testing.T) {
	cmd := testCommand(t, "run", "--isolation", "invalid-isolation-xyz")
	output, err := cmd.CombinedOutput()

	// Should fail closed: the registry has no isolator for the unknown method.
	if err == nil {
		t.Errorf("Expected error for invalid isolation, got nil")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "invalid-isolation-xyz") {
		t.Errorf("Expected error mentioning the invalid isolation, got: %s", output)
	}
}

func TestRunCommand_Agents(t *testing.T) {
	agents := []string{"claude", "opencode"}

	for _, agent := range agents {
		t.Run(agent, func(t *testing.T) {
			cmd := testCommand(t, "run", "-a", agent, "--help")
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Agent %s help failed: %v\nOutput: %s", agent, err, output)
			}
		})
	}
}

func TestRunCommand_IsolationMethods(t *testing.T) {
	methods := []string{"none", "bwrap", "docker"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Test that the isolation axis is recognized (even if not available).
			cmd := testCommand(t, "run", "--isolation", method, "--help")
			_, err := cmd.CombinedOutput()

			// Help should always work
			if err != nil {
				t.Errorf("Isolation %s help failed: %v", method, err)
			}
		})
	}
}

func TestCompletionCommand(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			cmd := testCommand(t, "completion", shell)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("completion %s failed: %v\nOutput: %s", shell, err, output)
			}

			if len(output) == 0 {
				t.Errorf("completion %s returned empty output", shell)
			}
		})
	}
}

func TestWorktree_InvalidRepo(t *testing.T) {
	// Create temp non-git directory
	tmpDir := t.TempDir()

	cmd := testCommand(t, "run",
		"-w", tmpDir,
		"--worktree-branch", "test-branch",
		"-a", "claude")

	output, err := cmd.CombinedOutput()

	// Should fail because it's not a git repo
	if err == nil {
		t.Errorf("Expected error for non-git directory, got nil")
	}

	// Should mention git or branch-related error
	outputStr := string(output)
	if !strings.Contains(outputStr, "branch") && !strings.Contains(outputStr, "git") && !strings.Contains(outputStr, "128") {
		t.Errorf("Expected git-related error, got: %s", outputStr)
	}
}
