package logging

import (
	"bytes"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/colors"
)

func newTestLogger(debug, color bool) (Logger, *bytes.Buffer, *bytes.Buffer) {
	var out, errOut bytes.Buffer
	return New(&out, &errOut, debug, color), &out, &errOut
}

func TestLevelsWriteToErrorSink(t *testing.T) {
	tests := []struct {
		name string
		log  func(Logger)
		want string
	}{
		{"info", func(l Logger) { l.Info("test message") }, "[INFO] test message\n"},
		{"success", func(l Logger) { l.Success("test message") }, "[SUCCESS] test message\n"},
		{"warning", func(l Logger) { l.Warning("test message") }, "[WARNING] test message\n"},
		{"error", func(l Logger) { l.Error("test message") }, "[ERROR] test message\n"},
		{"debug", func(l Logger) { l.Debug("test message") }, "[DEBUG] test message\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, out, errOut := newTestLogger(true, false)
			tt.log(l)
			if got := errOut.String(); got != tt.want {
				t.Errorf("error sink = %q, want %q", got, tt.want)
			}
			if out.Len() != 0 {
				t.Errorf("out sink = %q, want empty", out.String())
			}
		})
	}
}

func TestDebugIsDroppedWhenGateIsClosed(t *testing.T) {
	l, _, errOut := newTestLogger(false, false)
	l.Debug("should not appear")
	if errOut.Len() != 0 {
		t.Errorf("Debug() with closed gate wrote %q", errOut.String())
	}
}

func TestMultipleArgsAreJoinedBySpaces(t *testing.T) {
	l, _, errOut := newTestLogger(false, false)
	l.Info("multiple", "args", 123)
	if got, want := errOut.String(), "[INFO] multiple args 123\n"; got != want {
		t.Errorf("Info() = %q, want %q", got, want)
	}
}

func TestColorWrapsTheTagOnly(t *testing.T) {
	l, _, errOut := newTestLogger(false, true)
	l.Info("hello")
	if got, want := errOut.String(), colors.Blue+"[INFO]"+colors.Reset+" hello\n"; got != want {
		t.Errorf("Info() = %q, want %q", got, want)
	}
}

func TestSection(t *testing.T) {
	l, out, _ := newTestLogger(false, false)
	l.Section("Title", "content here")
	if got, want := out.String(), "\nTitle\n=====\ncontent here\n"; got != want {
		t.Errorf("Section() = %q, want %q", got, want)
	}
}

func TestSectionWithoutContent(t *testing.T) {
	l, out, _ := newTestLogger(false, false)
	l.Section("Title", "")
	if got, want := out.String(), "\nTitle\n=====\n"; got != want {
		t.Errorf("Section() = %q, want %q", got, want)
	}
}

func TestSubsection(t *testing.T) {
	l, out, _ := newTestLogger(false, false)
	l.Subsection("Sub")
	if got, want := out.String(), "\nSub\n---\n"; got != want {
		t.Errorf("Subsection() = %q, want %q", got, want)
	}
}

func TestCheckResultSymbols(t *testing.T) {
	tests := []struct {
		name   string
		status CheckStatus
		want   string
	}{
		{"pass", StatusPass, "✓ test check\n"},
		{"fail", StatusFail, "✗ test check\n"},
		{"warning", StatusWarn, "⚠ test check\n"},
		{"info", StatusInfo, "ℹ test check\n"},
		{"unknown", CheckStatus("OTHER"), "• test check\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, out, _ := newTestLogger(false, false)
			l.CheckResult("test check", tt.status, "")
			if got := out.String(); got != tt.want {
				t.Errorf("CheckResult() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckResultWithDetails(t *testing.T) {
	l, out, _ := newTestLogger(false, false)
	l.CheckResult("test check", StatusPass, "some details")
	if got, want := out.String(), "✓ test check\n    some details\n"; got != want {
		t.Errorf("CheckResult() = %q, want %q", got, want)
	}
}

func TestFromEnvironmentReadsGates(t *testing.T) {
	env := map[string]string{"DEBUG": "1", "TERM": "xterm"}
	var out, errOut bytes.Buffer
	l := FromEnvironment(func(k string) string { return env[k] }, &out, &errOut)
	l.Debug("on")
	if got, want := errOut.String(), colors.Purple+"[DEBUG]"+colors.Reset+" on\n"; got != want {
		t.Errorf("FromEnvironment() logger wrote %q, want %q", got, want)
	}

	errOut.Reset()
	env["NO_COLOR"] = "1"
	delete(env, "DEBUG")
	l = FromEnvironment(func(k string) string { return env[k] }, &out, &errOut)
	l.Debug("off")
	l.Info("plain")
	if got, want := errOut.String(), "[INFO] plain\n"; got != want {
		t.Errorf("FromEnvironment() logger wrote %q, want %q", got, want)
	}
}
