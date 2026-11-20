package codex

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/shared"
)

func setOf(t *testing.T, files map[string]string) artifact.Set {
	t.Helper()
	var list []artifact.File
	for path, content := range files {
		list = append(list, artifact.File{Path: path, Content: []byte(content)})
	}
	tree, err := artifact.NewTree(list)
	if err != nil {
		t.Fatal(err)
	}
	set, err := artifact.FromTree("src", tree)
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func contentOf(placement shared.Placement, path string) string {
	for _, file := range placement.Files {
		if file.Path == path {
			return string(file.Content)
		}
	}
	return ""
}

func TestPlaceMakesTheRootTheCodexHome(t *testing.T) {
	set := setOf(t, map[string]string{
		"skills/using-x/SKILL.md": "---\nname: using-x\ndescription: Use when x.\nmodel: opus\n---\n# X\n",
		"agents/reviewer.md":      "---\ndescription: \"Reviews \\\"it\\\"\"\nmodel: opus\n---\nBe strict.\nLine two.\n",
		"commands/fix.md":         "---\ndescription: Fix it.\n---\nFix the bug.\n",
		"AGENTS.md":               "Rules \"here\"\\\n",
	})
	placement, err := New().Place(shared.Input{Artifacts: set, Root: "/tmp/o"})
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, file := range placement.Files {
		paths = append(paths, file.Path)
	}
	want := "agentic.config.toml,agents/reviewer.toml,skills/fix/SKILL.md,skills/fix/agents/openai.yaml,skills/using-x/SKILL.md"
	if strings.Join(paths, ",") != want {
		t.Errorf("paths = %s", strings.Join(paths, ","))
	}
	if strings.Join(placement.Environment, " ") != "CODEX_HOME=/tmp/o" || strings.Join(placement.Arguments, " ") != "-p agentic" || !placement.RootWritable {
		t.Errorf("placement = %+v", placement)
	}
	if got := contentOf(placement, "agents/reviewer.toml"); got != "name = \"reviewer\"\ndescription = \"Reviews \\\"it\\\"\"\ndeveloper_instructions = \"Be strict.\\nLine two.\"\n" {
		t.Errorf("agent = %q", got)
	}
	if got := contentOf(placement, "agentic.config.toml"); got != "developer_instructions = \"Rules \\\"here\\\"\\\\\\n\"\n" {
		t.Errorf("profile = %q", got)
	}
	if got := contentOf(placement, "skills/fix/SKILL.md"); got != "---\nname: fix\ndescription: \"Fix it.\"\n---\nFix the bug.\n" {
		t.Errorf("command skill = %q", got)
	}
	if got := contentOf(placement, "skills/fix/agents/openai.yaml"); !strings.Contains(got, "allow_implicit_invocation: false") {
		t.Errorf("policy = %q", got)
	}
	if strings.Contains(contentOf(placement, "skills/using-x/SKILL.md"), "model") {
		t.Error("the model field was kept")
	}
}

func TestPlaceSkipsWhatCodexCannotTake(t *testing.T) {
	set := setOf(t, map[string]string{
		"skills/fix/SKILL.md": "---\nname: fix\ndescription: d\n---\n",
		"commands/fix.md":     "---\ndescription: d\n---\nx",
		"commands/bare.md":    "no description",
		"agents/empty.md":     "---\ndescription: d\n---\n   \n",
		"AGENTS.md":           "x",
	})
	placement, err := New().Place(shared.Input{Artifacts: set, Root: "/tmp/o", NativeInstructions: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(placement.Files) != 1 || len(placement.Skipped) != 4 || len(placement.Arguments) != 0 {
		t.Errorf("placement = %+v", placement)
	}
}

func TestPlaceAgentWithoutDescription(t *testing.T) {
	set := setOf(t, map[string]string{"agents/a.md": "Body.\n"})
	placement, err := New().Place(shared.Input{Artifacts: set, Root: "/tmp/o"})
	if err != nil || contentOf(placement, "agents/a.toml") != "name = \"a\"\ndeveloper_instructions = \"Body.\"\n" {
		t.Errorf("placement = %+v, %v", placement, err)
	}
}

func TestPlaceOfNothingAndBadRoot(t *testing.T) {
	placement, err := New().Place(shared.Input{Root: "/tmp/o"})
	if err != nil || len(placement.Files) != 0 || len(placement.Environment) != 0 {
		t.Errorf("placement = %+v, %v", placement, err)
	}
	if _, err := New().Place(shared.Input{Root: "x"}); err == nil {
		t.Error("a relative root passed")
	}
}

func TestTomlStringEscapesControlCharacters(t *testing.T) {
	if got, want := tomlString("a\tb\rc\x01d\x7fe é"), `"a\tb\rc\u0001d\u007Fe é"`; got != want {
		t.Errorf("tomlString = %s, want %s", got, want)
	}
}

func TestFacts(t *testing.T) {
	if facts := New().Facts(); !facts.ReplacesHome || len(facts.NativeInstructions) != 1 {
		t.Errorf("facts = %+v", facts)
	}
}
