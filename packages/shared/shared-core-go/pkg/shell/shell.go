package shell

import "strings"

// Quote wraps s in single quotes, escaping any embedded single quotes.
// Safe for POSIX sh arguments even if the value contains spaces or special chars.
func Quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// ContainsMetachars returns true if s contains POSIX shell metacharacters
// that could enable command injection.
func ContainsMetachars(s string) bool {
	const metachars = ";|&$`\\\"'<>(){}!#~\n\r"
	return strings.ContainsAny(s, metachars)
}
