package logging

import (
	"fmt"
	"io"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/colors"
)

// Logger writes leveled messages to err and section or check output to out.
// It owns no global state: the writers, the debug gate and the color switch
// are inputs.
type Logger struct {
	out   io.Writer
	err   io.Writer
	debug bool
	color bool
}

// New returns a Logger over the given sinks. Debug messages are dropped unless
// debug is true, and ANSI colors are written only when color is true.
func New(out, err io.Writer, debug, color bool) Logger {
	return Logger{out: out, err: err, debug: debug, color: color}
}

// FromEnvironment returns a Logger whose debug gate and color switch come from
// the DEBUG, TERM and NO_COLOR variables that getenv resolves.
func FromEnvironment(getenv func(string) string, out, err io.Writer) Logger {
	return New(out, err, getenv("DEBUG") != "", colors.Supported(getenv("TERM"), getenv("NO_COLOR")))
}

// Info logs an info message with blue color.
func (l Logger) Info(args ...any) {
	l.logLevel("[INFO]", colors.Blue, args)
}

// Success logs a success message with green color.
func (l Logger) Success(args ...any) {
	l.logLevel("[SUCCESS]", colors.Green, args)
}

// Warning logs a warning message with yellow color.
func (l Logger) Warning(args ...any) {
	l.logLevel("[WARNING]", colors.Yellow, args)
}

// Error logs an error message with red color.
func (l Logger) Error(args ...any) {
	l.logLevel("[ERROR]", colors.Red, args)
}

// Debug logs a debug message with purple color when the debug gate is open.
func (l Logger) Debug(args ...any) {
	if !l.debug {
		return
	}
	l.logLevel("[DEBUG]", colors.Purple, args)
}

// Section prints a formatted section header followed by optional content.
func (l Logger) Section(title string, content string) {
	l.println(l.out, "")
	l.println(l.out, l.paint(title, colors.Cyan))
	l.println(l.out, l.paint(strings.Repeat("=", len(title)), colors.Cyan))
	if content != "" {
		l.println(l.out, content)
	}
}

// Subsection prints a subsection header.
func (l Logger) Subsection(title string) {
	l.println(l.out, "")
	l.println(l.out, l.paint(title, colors.Blue))
	if len(title) > 0 {
		l.println(l.out, l.paint(strings.Repeat("-", len(title)), colors.Blue))
	}
}

// CheckStatus represents the status of a check result
type CheckStatus string

const (
	StatusPass    CheckStatus = "PASS"
	StatusOK      CheckStatus = "OK"
	StatusSuccess CheckStatus = "SUCCESS"
	StatusFail    CheckStatus = "FAIL"
	StatusError   CheckStatus = "ERROR"
	StatusFailed  CheckStatus = "FAILED"
	StatusWarn    CheckStatus = "WARN"
	StatusWarning CheckStatus = "WARNING"
	StatusInfo    CheckStatus = "INFO"
)

// CheckResult prints a check result with the color of its status.
func (l Logger) CheckResult(checkName string, status CheckStatus, details string) {
	symbol, color := checkStyle(status)
	l.println(l.out, l.paint(symbol+" "+checkName, color))
	if details != "" {
		l.println(l.out, l.paint("    "+details, colors.Purple))
	}
}

func checkStyle(status CheckStatus) (symbol, color string) {
	switch status {
	case StatusPass, StatusOK, StatusSuccess:
		return "✓", colors.Green
	case StatusFail, StatusError, StatusFailed:
		return "✗", colors.Red
	case StatusWarn, StatusWarning:
		return "⚠", colors.Yellow
	case StatusInfo:
		return "ℹ", colors.Blue
	default:
		return "•", ""
	}
}

func (l Logger) logLevel(tag, color string, args []any) {
	l.println(l.err, l.paint(tag, color)+" "+formatArgs(args...))
}

func (l Logger) paint(text, color string) string {
	return colors.Colorize(text, color, l.color)
}

func (l Logger) println(w io.Writer, line string) {
	_, _ = fmt.Fprintln(w, line)
}

func formatArgs(args ...any) string {
	strs := make([]string, len(args))
	for i, arg := range args {
		strs[i] = fmt.Sprint(arg)
	}
	return strings.Join(strs, " ")
}
