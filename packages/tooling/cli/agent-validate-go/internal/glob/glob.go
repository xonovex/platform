// Package glob matches slash paths against the glob syntax of the repository
// tools: ** spans directories, * and ? stop at a separator.
package glob

import (
	"regexp"
	"strings"
)

// Compile turns a glob into an expression that matches a whole path.
func Compile(pattern string) (*regexp.Regexp, error) {
	var b strings.Builder
	for i := 0; i < len(pattern); i++ {
		switch c := pattern[i]; {
		case strings.HasPrefix(pattern[i:], "**/"):
			b.WriteString("(?:.*/)?")
			i += 2
		case strings.HasPrefix(pattern[i:], "**"):
			b.WriteString(".*")
			i++
		case c == '*':
			b.WriteString("[^/]*")
		case c == '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	return regexp.Compile("^" + b.String() + "$")
}
