package shared

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

var testLog = logging.New(io.Discard, io.Discard, false, false)

func TestSpawnSandboxRunsInDirectoryWithEnvironment(t *testing.T) {
	host := systemtest.New()
	host.Outcomes["sh -c true"] = systemtest.Outcome{}
	env := []string{"SANDBOX_VALUE=ready"}

	exitCode, err := SpawnSandbox(host, testLog, "sh", []string{"-c", "true"}, env, "/work", "spawn test")

	if err != nil || exitCode != 0 {
		t.Fatalf("SpawnSandbox() exitCode = %d, error = %v", exitCode, err)
	}
	if len(host.Calls) != 1 {
		t.Fatalf("calls = %+v, want one run", host.Calls)
	}
	call := host.Calls[0]
	if call.Kind != "run" || call.Dir != "/work" || !slices.Equal(call.Env, env) {
		t.Errorf("call = %+v, want a run in /work with the sandbox environment", call)
	}
}

func TestSpawnSandboxReturnsChildExitCode(t *testing.T) {
	host := systemtest.New()
	host.Outcomes["sh -c exit 7"] = systemtest.Outcome{ExitCode: 7}

	exitCode, err := SpawnSandbox(host, testLog, "sh", []string{"-c", "exit 7"}, nil, "", "spawn test")

	if err != nil || exitCode != 7 {
		t.Fatalf("SpawnSandbox() exitCode = %d, error = %v, want 7, nil", exitCode, err)
	}
}

func TestSpawnSandboxReportsStartFailure(t *testing.T) {
	host := systemtest.New()
	host.Outcomes["/missing/command"] = systemtest.Outcome{Err: errors.New("no such file")}

	exitCode, err := SpawnSandbox(host, testLog, "/missing/command", nil, nil, "", "start sandbox")

	if exitCode != 1 || err == nil || err.Error() != "start sandbox: no such file" {
		t.Fatalf("SpawnSandbox() exitCode = %d, error = %v, want 1 and the prefixed start error", exitCode, err)
	}
}

func TestSpawnSandboxLogsTheCommandAtDebugLevel(t *testing.T) {
	var debug bytes.Buffer
	host := systemtest.New()
	host.Outcomes["bwrap --die-with-parent"] = systemtest.Outcome{}

	if _, err := SpawnSandbox(host, logging.New(io.Discard, &debug, true, false), "bwrap", []string{"--die-with-parent"}, nil, "", "spawn test"); err != nil {
		t.Fatalf("SpawnSandbox() error = %v", err)
	}
	if !strings.Contains(debug.String(), "Executing: bwrap --die-with-parent") {
		t.Errorf("debug output = %q, want the command line", debug.String())
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
