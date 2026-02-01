package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	_ = w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return ""
	}
	return buf.String()
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return ""
	}
	return buf.String()
}

func TestLogInfo(t *testing.T) {
	output := captureStderr(func() {
		LogInfo("test message")
	})

	if !strings.Contains(output, "[INFO]") || !strings.Contains(output, "test message") {
		t.Errorf("LogInfo() output = %q, want to contain [INFO] and test message", output)
	}
}

func TestLogSuccess(t *testing.T) {
	output := captureStderr(func() {
		LogSuccess("success message")
	})

	if !strings.Contains(output, "[SUCCESS]") || !strings.Contains(output, "success message") {
		t.Errorf("LogSuccess() output = %q, want to contain [SUCCESS] and success message", output)
	}
}

func TestLogWarning(t *testing.T) {
	output := captureStderr(func() {
		LogWarning("warning message")
	})

	if !strings.Contains(output, "[WARNING]") || !strings.Contains(output, "warning message") {
		t.Errorf("LogWarning() output = %q, want to contain [WARNING] and warning message", output)
	}
}

func TestLogError(t *testing.T) {
	output := captureStderr(func() {
		LogError("error message")
	})

	if !strings.Contains(output, "[ERROR]") || !strings.Contains(output, "error message") {
		t.Errorf("LogError() output = %q, want to contain [ERROR] and error message", output)
	}
}

func TestLogDebug(t *testing.T) {
	origDebug := os.Getenv("DEBUG")
	defer func() {
		if origDebug == "" {
			_ = os.Unsetenv("DEBUG")
		} else {
			_ = os.Setenv("DEBUG", origDebug)
		}
	}()

	_ = os.Unsetenv("DEBUG")
	output := captureStderr(func() {
		LogDebug("should not appear")
	})
	if output != "" {
		t.Errorf("LogDebug() without DEBUG env should not output, got %q", output)
	}

	_ = os.Setenv("DEBUG", "1")
	output = captureStderr(func() {
		LogDebug("debug message")
	})
	if !strings.Contains(output, "[DEBUG]") || !strings.Contains(output, "debug message") {
		t.Errorf("LogDebug() with DEBUG env output = %q, want to contain [DEBUG] and debug message", output)
	}
}

func TestLogMultipleArgs(t *testing.T) {
	output := captureStderr(func() {
		LogInfo("multiple", "args", 123)
	})

	if !strings.Contains(output, "multiple args 123") {
		t.Errorf("LogInfo() with multiple args output = %q, want to contain 'multiple args 123'", output)
	}
}

func TestPrintSection(t *testing.T) {
	output := captureStdout(func() {
		PrintSection("Test Section", "content here")
	})

	if !strings.Contains(output, "Test Section") {
		t.Errorf("PrintSection() output = %q, want to contain 'Test Section'", output)
	}
	if !strings.Contains(output, "content here") {
		t.Errorf("PrintSection() output = %q, want to contain 'content here'", output)
	}
}

func TestPrintSubsection(t *testing.T) {
	output := captureStdout(func() {
		PrintSubsection("Test Subsection")
	})

	if !strings.Contains(output, "Test Subsection") {
		t.Errorf("PrintSubsection() output = %q, want to contain 'Test Subsection'", output)
	}
}

func TestCheckResult(t *testing.T) {
	tests := []struct {
		name     string
		status   CheckStatus
		contains string
	}{
		{"pass", StatusPass, "✓"},
		{"fail", StatusFail, "✗"},
		{"warning", StatusWarn, "⚠"},
		{"info", StatusInfo, "ℹ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				CheckResult("test check", tt.status, "")
			})

			if !strings.Contains(output, tt.contains) {
				t.Errorf("CheckResult() with %s output = %q, want to contain %q", tt.name, output, tt.contains)
			}
		})
	}
}

func TestCheckResultWithDetails(t *testing.T) {
	output := captureStdout(func() {
		CheckResult("test check", StatusPass, "some details")
	})

	if !strings.Contains(output, "some details") {
		t.Errorf("CheckResult() with details output = %q, want to contain 'some details'", output)
	}
}
