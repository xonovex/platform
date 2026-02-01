package logging

import (
	"fmt"
	"os"
	"strings"

	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/colors"
)

// LogInfo logs an info message with blue color
func LogInfo(args ...any) {
	msg := formatArgs(args...)
	if colors.IsColorSupported() {
		fmt.Fprintln(os.Stderr, colors.WithBlue("[INFO]")+" "+msg)
	} else {
		fmt.Fprintln(os.Stderr, "[INFO] "+msg)
	}
}

// LogSuccess logs a success message with green color
func LogSuccess(args ...any) {
	msg := formatArgs(args...)
	if colors.IsColorSupported() {
		fmt.Fprintln(os.Stderr, colors.WithGreen("[SUCCESS]")+" "+msg)
	} else {
		fmt.Fprintln(os.Stderr, "[SUCCESS] "+msg)
	}
}

// LogWarning logs a warning message with yellow color
func LogWarning(args ...any) {
	msg := formatArgs(args...)
	if colors.IsColorSupported() {
		fmt.Fprintln(os.Stderr, colors.WithYellow("[WARNING]")+" "+msg)
	} else {
		fmt.Fprintln(os.Stderr, "[WARNING] "+msg)
	}
}

// LogError logs an error message with red color
func LogError(args ...any) {
	msg := formatArgs(args...)
	if colors.IsColorSupported() {
		fmt.Fprintln(os.Stderr, colors.WithRed("[ERROR]")+" "+msg)
	} else {
		fmt.Fprintln(os.Stderr, "[ERROR] "+msg)
	}
}

// LogDebug logs a debug message with purple color (only if DEBUG env var is set)
func LogDebug(args ...any) {
	if os.Getenv("DEBUG") == "" {
		return
	}
	msg := formatArgs(args...)
	if colors.IsColorSupported() {
		fmt.Fprintln(os.Stderr, colors.WithPurple("[DEBUG]")+" "+msg)
	} else {
		fmt.Fprintln(os.Stderr, "[DEBUG] "+msg)
	}
}

// PrintSection prints a formatted section header
func PrintSection(title string, content string) {
	if colors.IsColorSupported() {
		fmt.Printf("\n%s\n", colors.WithCyan(title))
		fmt.Printf("%s\n", colors.WithCyan(strings.Repeat("=", len(title))))
	} else {
		fmt.Printf("\n%s\n", title)
		fmt.Printf("%s\n", strings.Repeat("=", len(title)))
	}
	if content != "" {
		fmt.Println(content)
	}
}

// PrintSubsection prints a subsection header
func PrintSubsection(title string) {
	if colors.IsColorSupported() {
		fmt.Printf("\n%s\n", colors.WithBlue(title))
		if len(title) > 0 {
			fmt.Printf("%s\n", colors.WithBlue(strings.Repeat("-", len(title))))
		}
	} else {
		fmt.Printf("\n%s\n", title)
		if len(title) > 0 {
			fmt.Printf("%s\n", strings.Repeat("-", len(title)))
		}
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

// CheckResult prints a check result with appropriate color
func CheckResult(checkName string, status CheckStatus, details string) {
	symbol := "•"
	var color string

	switch status {
	case StatusPass, StatusOK, StatusSuccess:
		symbol = "✓"
		color = colors.Green
	case StatusFail, StatusError, StatusFailed:
		symbol = "✗"
		color = colors.Red
	case StatusWarn, StatusWarning:
		symbol = "⚠"
		color = colors.Yellow
	case StatusInfo:
		symbol = "ℹ"
		color = colors.Blue
	}

	if colors.IsColorSupported() {
		fmt.Println(colors.Colorize(symbol+" "+checkName, color))
	} else {
		fmt.Println(symbol + " " + checkName)
	}

	if details != "" {
		if colors.IsColorSupported() {
			fmt.Println(colors.WithPurple("    " + details))
		} else {
			fmt.Println("    " + details)
		}
	}
}

func formatArgs(args ...any) string {
	strs := make([]string, len(args))
	for i, arg := range args {
		strs[i] = fmt.Sprint(arg)
	}
	return strings.Join(strs, " ")
}
