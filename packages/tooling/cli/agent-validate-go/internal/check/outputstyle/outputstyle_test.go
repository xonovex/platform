package outputstyle_test

import (
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/outputstyle"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func TestOutputStyle(t *testing.T) {
	c := outputstyle.Check{}
	if c.Name() != "output-style" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
	if _, err := c.Run(checktest.WithOptions(t, checktest.Context(t, nil, "."), map[string]any{"x": 1})); err == nil {
		t.Fatal("unknown option")
	}
	good := "---\nname: terse\ndescription: Short answers\nkeep-coding-instructions: true\n---\nBe brief.\n"
	files := map[string]string{"plugin/output-styles/terse.md": good}
	if got := checktest.Run(t, c, checktest.Context(t, files, ".")); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
	bad := "---\nname: terse\nkeep-coding-instructions: yes please\nforce-for-plugin: 3\nextra: 1\n---\n"
	got := checktest.Codes(checktest.Run(t, c, checktest.Context(t, map[string]string{"output-styles/x.md": bad}, ".")))
	want := map[string]bool{
		"output-style.description": true, "output-style.keep-coding-instructions": true,
		"output-style.force-for-plugin": true, "output-style.unknown-field": true, "output-style.body": true,
	}
	for _, code := range got {
		delete(want, code)
	}
	if len(want) != 0 {
		t.Fatalf("missing %v in %v", want, got)
	}
}
