package builder

import "strings"

// shellQuote wraps a string in single quotes, escaping any embedded single quotes.
// This is safe for POSIX sh arguments even if the value contains spaces or special chars.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
