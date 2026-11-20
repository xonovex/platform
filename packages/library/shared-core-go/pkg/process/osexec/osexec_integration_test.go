//go:build integration

package osexec_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process/osexec"
)

func sh(script string) process.Request {
	return process.Request{Name: "sh", Args: []string{"-c", script}}
}

func TestCaptureKeepsTheExactOutputAndTheExitStatus(t *testing.T) {
	request := sh(`printf 'a\r\n\nlast'; printf 'oops' >&2; exit 3`)
	request.Dir = t.TempDir()
	got, err := process.Capture(context.Background(), osexec.New(), request)
	if err != nil || string(got.Stdout) != "a\r\n\nlast" || got.Stderr != "oops" || got.ExitCode != 3 || got.Succeeded() {
		t.Fatalf("Capture = %+v, %v", got, err)
	}
}

func TestRunPassesTheDirectoryAndTheEnvironment(t *testing.T) {
	request := sh(`printf '%s %s' "$PWD" "$PROCESS_PROBE"`)
	request.Dir = t.TempDir()
	request.Env = map[string]string{"PROCESS_PROBE": "set"}
	got, err := process.Capture(context.Background(), osexec.New(), request)
	if err != nil || !strings.HasSuffix(string(got.Stdout), " set") || !got.Succeeded() {
		t.Fatalf("Capture = %+v, %v", got, err)
	}
}

func TestStopEndsTheProgramEarly(t *testing.T) {
	// The timeout fires only when the stop fails to end the program.
	request := sh(`echo first; echo second; sleep 30`)
	request.Timeout = 10 * time.Second
	var seen []string
	got, err := osexec.New().Run(context.Background(), request, process.Lines(func(line string) bool {
		seen = append(seen, line)
		return line == "second"
	}))
	if err != nil || !got.Stopped || got.TimedOut || len(seen) != 2 {
		t.Fatalf("Run = %+v, %v, %q", got, err, seen)
	}
}

func TestTimeoutEndsTheProgram(t *testing.T) {
	request := sh(`sleep 30`)
	request.Timeout = 200 * time.Millisecond
	got, err := osexec.New().Run(context.Background(), request, nil)
	if err != nil || !got.TimedOut || got.Succeeded() {
		t.Fatalf("Run = %+v, %v", got, err)
	}
}

func TestFaultsAreErrors(t *testing.T) {
	if _, err := osexec.New().Run(context.Background(), process.Request{Name: "shared-core-no-such-program"}, nil); err == nil {
		t.Error("a program that cannot start must be an error")
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := osexec.New().Run(cancelled, sh(`true`), nil); err == nil {
		t.Error("a cancelled context must be an error")
	}
}
