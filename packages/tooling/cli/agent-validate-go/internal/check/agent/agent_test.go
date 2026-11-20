package agent_test

import (
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/agent"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func run(t *testing.T, files map[string]string) []string {
	t.Helper()
	ctx := checktest.Context(t, files, ".")
	return checktest.Codes(checktest.Run(t, agent.Check{}, ctx))
}

func has(codes []string, code string) bool {
	for _, c := range codes {
		if c == code {
			return true
		}
	}
	return false
}

func TestIdentity(t *testing.T) {
	c := agent.Check{}
	if c.Name() != "agent" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
	ctx := checktest.WithOptions(t, checktest.Context(t, nil, "."), map[string]any{"x": 1})
	if _, err := c.Run(ctx); err == nil {
		t.Fatal("unknown option")
	}
}

const good = "---\nname: code-reviewer\ndescription: Reviews code\ntools: Read, Grep\nmaxTurns: 5\ncolor: blue\n---\nYou review code.\n"

func TestValidAgent(t *testing.T) {
	files := map[string]string{
		".claude/agents/code-reviewer.md":          good,
		"skills/using-x/agents/not-a-component.md": "no frontmatter",
		"docs/other.md": "x",
	}
	if got := run(t, files); len(got) != 0 {
		t.Fatalf("codes = %v", got)
	}
}

func TestInvalidAgents(t *testing.T) {
	cases := map[string]struct {
		file string
		code string
	}{
		"no frontmatter": {"just text", "agent.frontmatter"},
		"bad yaml":       {"---\na: [\n---\nbody", "agent.frontmatter"},
		"not a mapping":  {"---\n- a\n---\nbody", "agent.frontmatter"},
		"unknown":        {"---\nname: a\ndescription: d\nbogus: 1\n---\nx", "agent.unknown-field"},
		"no name":        {"---\ndescription: d\n---\nx", "agent.name"},
		"no description": {"---\nname: a\n---\nx", "agent.description"},
		"bad name":       {"---\nname: Bad_Name\ndescription: d\n---\nx", "agent.name-format"},
		"name vs file":   {"---\nname: other\ndescription: d\n---\nx", "agent.name-file"},
		"no prompt":      {"---\nname: a\ndescription: d\n---\n\n", "agent.body"},
		"tools":          {"---\nname: a\ndescription: d\ntools: []\n---\nx", "agent.tools"},
		"permission":     {"---\nname: a\ndescription: d\npermissionMode: wild\n---\nx", "agent.permissionMode"},
		"turns":          {"---\nname: a\ndescription: d\nmaxTurns: 0\n---\nx", "agent.maxTurns"},
	}
	for name, tc := range cases {
		got := run(t, map[string]string{"agents/a.md": tc.file})
		if !has(got, tc.code) {
			t.Errorf("%s: codes = %v", name, got)
		}
	}
}
