package glob_test

import (
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/glob"
)

func TestCompile(t *testing.T) {
	cases := []struct {
		pattern, path string
		want          bool
	}{
		{"a/*.json", "a/x.json", true},
		{"a/*.json", "a/b/x.json", false},
		{"a/**/x.json", "a/x.json", true},
		{"a/**/x.json", "a/b/c/x.json", true},
		{"**/x.json", "x.json", true},
		{"a/**", "a/b/c", true},
		{"a?c", "abc", true},
		{"a?c", "a/c", false},
		{"a.b", "axb", false},
		{"a+(b)", "a+(b)", true},
	}
	for _, tc := range cases {
		re, err := glob.Compile(tc.pattern)
		if err != nil {
			t.Fatal(err)
		}
		if got := re.MatchString(tc.path); got != tc.want {
			t.Errorf("%s ~ %s = %v", tc.pattern, tc.path, got)
		}
	}
}
