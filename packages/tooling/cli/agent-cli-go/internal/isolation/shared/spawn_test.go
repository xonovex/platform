package shared

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestSpawnSandboxInDirUsesDirectoryAndEnvironment(t *testing.T) {
	workDir := t.TempDir()
	outputPath := filepath.Join(t.TempDir(), "result")
	command := []string{"-c", "printf '%s:%s' \"$PWD\" \"$SANDBOX_VALUE\" > \"$1\"", "sh", outputPath}

	exitCode, err := SpawnSandboxInDir("sh", command, append(os.Environ(), "SANDBOX_VALUE=ready"), workDir, "spawn test", false)

	if err != nil || exitCode != 0 {
		t.Fatalf("SpawnSandboxInDir() exitCode = %d, error = %v", exitCode, err)
	}
	result, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read command output: %v", err)
	}
	if string(result) != workDir+":ready" {
		t.Fatalf("command output = %q, want %q", result, workDir+":ready")
	}
}

func TestSpawnSandboxReturnsChildExitCode(t *testing.T) {
	exitCode, err := SpawnSandbox("sh", []string{"-c", "exit 7"}, os.Environ(), "spawn test", false)

	if err != nil || exitCode != 7 {
		t.Fatalf("SpawnSandbox() exitCode = %d, error = %v, want 7, nil", exitCode, err)
	}
}

func TestSpawnSandboxReportsStartFailure(t *testing.T) {
	exitCode, err := SpawnSandbox("/missing/xonovex-command", nil, os.Environ(), "start sandbox", true)

	if exitCode != 1 || err == nil || !strings.Contains(err.Error(), "start sandbox") {
		t.Fatalf("SpawnSandbox() exitCode = %d, error = %v, want prefixed start error", exitCode, err)
	}
}

func TestWrapWithInitCommands(t *testing.T) {
	command := []string{"agent", "argument with spaces"}

	unwrapped := WrapWithInitCommands(command, nil)
	wrapped := WrapWithInitCommands(command, []string{"prepare", "configure"})

	if !slices.Equal(unwrapped, command) {
		t.Fatalf("WrapWithInitCommands() without init = %v, want %v", unwrapped, command)
	}
	want := []string{"sh", "-c", "prepare && configure && exec 'agent' 'argument with spaces'"}
	if !slices.Equal(wrapped, want) {
		t.Fatalf("WrapWithInitCommands() = %v, want %v", wrapped, want)
	}
}
