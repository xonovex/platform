package colors

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

// Supported reports whether a terminal that sets TERM and NO_COLOR as given
// accepts color output.
func Supported(term, noColor string) bool {
	if term == "" || term == "dumb" {
		return false
	}
	return noColor == ""
}

// Colorize wraps text with the color code when enabled and returns it unchanged otherwise.
func Colorize(text, color string, enabled bool) string {
	if !enabled {
		return text
	}
	return color + text + Reset
}
