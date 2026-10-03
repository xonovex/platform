package tmux

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilterEnvironmentRejectsInvalidAndReservedNames(t *testing.T) {
	environment := []string{
		"PATH=/usr/bin",
		"SAFE_VALUE=content",
		"UID=1000",
		"9INVALID=value",
		"BAD;touch /tmp/injected=value",
		"missing-separator",
	}

	filtered := filterEnvironment(environment)

	if len(filtered) != 2 {
		t.Fatalf("filterEnvironment() returned %d variables, want 2", len(filtered))
	}
	if filtered[0].name != "PATH" || filtered[1].name != "SAFE_VALUE" {
		t.Fatalf("filterEnvironment() = %v, want PATH and SAFE_VALUE", filtered)
	}
}

func TestCreateLaunchScriptKeepsSecretsOutOfArgumentsAndExecutesEnvironment(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "environment.txt")
	secret := "token with '$HOME' and a newline\nvalue"
	command := []string{"sh", "-c", "printf '%s' \"$SAFE_VALUE\" > \"$1\"", "sh", outputPath}

	launchScript, err := createLaunchScript(command, []string{
		"SAFE_VALUE=" + secret,
		"BAD;touch " + filepath.Join(t.TempDir(), "injected") + "=value",
	})

	if err != nil {
		t.Fatalf("createLaunchScript() error = %v", err)
	}
	info, err := os.Stat(launchScript)
	if err != nil {
		t.Fatalf("stat launch script: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("launch script mode = %o, want 600", info.Mode().Perm())
	}
	if strings.Contains(launchScript, secret) {
		t.Fatal("launch script path contains secret")
	}

	if err := runLaunchScript(launchScript); err != nil {
		t.Fatalf("run launch script: %v", err)
	}
	result, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read command output: %v", err)
	}
	if string(result) != secret {
		t.Fatalf("command output = %q, want %q", result, secret)
	}
	if _, err := os.Stat(launchScript); !os.IsNotExist(err) {
		t.Fatalf("self-deleted launch script stat error = %v, want not exists", err)
	}
}

func TestCreateLaunchScriptRequiresCommand(t *testing.T) {
	if _, err := createLaunchScript(nil, nil); err == nil {
		t.Fatal("createLaunchScript(nil, nil) error = nil, want error")
	}
}

func runLaunchScript(path string) error {
	process, err := os.StartProcess("/bin/sh", []string{"sh", path}, &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})
	if err != nil {
		return err
	}
	state, err := process.Wait()
	if err != nil {
		return err
	}
	if !state.Success() {
		return &os.PathError{Op: "execute", Path: path, Err: os.ErrInvalid}
	}
	return nil
}
