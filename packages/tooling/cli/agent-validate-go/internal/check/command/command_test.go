package command_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/command"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, ".")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, command.Check{}, ctx)
}

func TestIdentity(t *testing.T) {
	c := command.Check{}
	if c.Name() != "command" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
	if _, err := c.Run(checktest.WithOptions(t, checktest.Context(t, nil, "."), map[string]any{"x": 1})); err == nil {
		t.Fatal("unknown option")
	}
}

func file(frontmatter, body string) map[string]string {
	return map[string]string{".claude/commands/do.md": "---\n" + frontmatter + "\n---\n" + body}
}

func TestValidCommands(t *testing.T) {
	good := []string{
		"description: Do it\nargument-hint: <target> [flag]\nallowed-tools: Bash(git status:*), Read\n",
		"description: Do it\nallowed-tools:\n  - Read\n  - Edit\nmodel: opus\ndisable-model-invocation: true",
		"description: Do it\nargument-hint: free text",
	}
	for i, fm := range good {
		got := run(t, file(fm, "Run $1 then $ARGUMENTS and $2.\n"))
		if len(got) != 0 {
			t.Errorf("good %d: %v", i, got)
		}
	}
}

func TestInvalidCommands(t *testing.T) {
	cases := map[string]struct {
		files map[string]string
		code  string
	}{
		"no description": {file("model: x", "body"), "command.description"},
		"unknown":        {file("description: d\nbogus: 1", "body"), "command.unknown-field"},
		"empty body":     {file("description: d", "\n"), "command.body"},
		"no frontmatter": {map[string]string{"commands/a.md": "text"}, "command.frontmatter"},
		"tools type":     {file("description: d\nallowed-tools: []", "x"), "command.allowed-tools"},
		"tools empty":    {file("description: d\nallowed-tools: 'Read,,Edit'", "x"), "command.allowed-tools"},
		"tools twice":    {file("description: d\nallowed-tools: Read, Read", "x"), "command.allowed-tools"},
		"tools paren":    {file("description: d\nallowed-tools: Bash(git", "x"), "command.allowed-tools"},
		"tools name":     {file("description: d\nallowed-tools: 9lives", "x"), "command.allowed-tools"},
		"hint type":      {file("description: d\nargument-hint: 3", "x"), "command.argument-hint"},
		"hint unused":    {file("description: d\nargument-hint: <a>", "never refers"), "command.argument-unused"},
		"hint excess":    {file("description: d\nargument-hint: <a>", "use $1 and $3"), "command.argument-excess"},
		"fenced":         {file("description: d\nargument-hint: <a>", "```\n$1\n```\n"), "command.argument-unused"},
		"long":           {file("description: d", strings.Repeat("line\n", 90)), "command.body-length"},
	}
	for name, tc := range cases {
		if got := run(t, tc.files); !checktest.Has(got, tc.code) {
			t.Errorf("%s: codes = %v", name, checktest.Codes(got))
		}
	}
	got := run(t, file("description: d", strings.Repeat("line\n", 11)), map[string]any{"maxBodyLines": 10})
	f, ok := checktest.Find(got, "command.body-length")
	if !ok || f.Line != 4+10 {
		t.Fatalf("findings = %v", got)
	}
}
