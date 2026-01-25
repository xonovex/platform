package colors

import (
	"os"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[0;31m"
	Green  = "\033[0;32m"
	Yellow = "\033[1;33m"
	Blue   = "\033[0;34m"
	Purple = "\033[0;35m"
	Cyan   = "\033[0;36m"
	Gray   = "\033[90m"
	Bold   = "\033[1m"
	NC     = Reset // No Color alias
)

// IsColorSupported checks if terminal supports color output
func IsColorSupported() bool {
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return true
}

// Colorize wraps text with color codes if supported
func Colorize(text, color string) string {
	if !IsColorSupported() {
		return text
	}
	return color + text + Reset
}

// WithRed returns red colored text
func WithRed(text string) string {
	return Colorize(text, Red)
}

// WithGreen returns green colored text
func WithGreen(text string) string {
	return Colorize(text, Green)
}

// WithYellow returns yellow colored text
func WithYellow(text string) string {
	return Colorize(text, Yellow)
}

// WithBlue returns blue colored text
func WithBlue(text string) string {
	return Colorize(text, Blue)
}

// WithPurple returns purple colored text
func WithPurple(text string) string {
	return Colorize(text, Purple)
}

// WithCyan returns cyan colored text
func WithCyan(text string) string {
	return Colorize(text, Cyan)
}

// WithGray returns gray colored text
func WithGray(text string) string {
	return Colorize(text, Gray)
}

// WithBold returns bold text
func WithBold(text string) string {
	return Colorize(text, Bold)
}
