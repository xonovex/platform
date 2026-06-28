//go:build integration
// +build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func init() {
	// Resolve binary path relative to test file location
	wd, _ := os.Getwd()
	binaryPath = filepath.Join(wd, "..", "..", "dist", "agent-cli")
}

func TestRunCommand_Help(t *testing.T) {
	cmd := exec.Command(binaryPath, "run", "--help")
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
		"--network-proxy-egress-allow",
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
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "0.1.0") {
		t.Errorf("Version output incorrect: %s", output)
	}
}

func TestRunCommand_InvalidAgent(t *testing.T) {
	cmd := exec.Command(binaryPath, "run", "-a", "invalid-agent-xyz")
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
	cmd := exec.Command(binaryPath, "run", "--isolation", "invalid-isolation-xyz")
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
			cmd := exec.Command(binaryPath, "run", "-a", agent, "--help")
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
			cmd := exec.Command(binaryPath, "run", "--isolation", method, "--help")
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
			cmd := exec.Command(binaryPath, "completion", shell)
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

	cmd := exec.Command(binaryPath, "run",
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
