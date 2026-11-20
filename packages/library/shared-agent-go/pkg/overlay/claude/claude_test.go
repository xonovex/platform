package claude

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/shared"
)

func sample(t *testing.T) artifact.Set {
	t.Helper()
	tree, err := artifact.NewTree([]artifact.File{
		{Path: "skills/using-x/SKILL.md", Content: []byte("---\nname: using-x\ndescription: Use when x.\nmodel: opus\n---\n# X\n")},
		{Path: "skills/using-x/references/a.md", Content: []byte("ref")},
		{Path: "agents/reviewer.md", Content: []byte("---\ndescription: Reviews.\nmodel: opus\ntools: Read\n---\nBe strict.\n")},
		{Path: "commands/fix.md", Content: []byte("---\ndescription: Fix it.\n---\nFix the bug.\n")},
		{Path: "AGENTS.md", Content: []byte("Rules.\n")},
	})
	if err != nil {
		t.Fatal(err)
	}
	set, err := artifact.FromTree("src", tree)
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func paths(placement shared.Placement) string {
	var out []string
	for _, file := range placement.Files {
		out = append(out, file.Path)
	}
	return strings.Join(out, ",")
}

func TestPlaceWritesTheClaudeLayoutAndAddsTheRoot(t *testing.T) {
	placement, err := New().Place(shared.Input{Artifacts: sample(t), Root: "/tmp/o"})
	if err != nil {
		t.Fatal(err)
	}
	want := ".claude/agents/reviewer.md,.claude/commands/fix.md,.claude/skills/using-x/SKILL.md,.claude/skills/using-x/references/a.md,CLAUDE.md"
	if paths(placement) != want {
		t.Errorf("paths = %s", paths(placement))
	}
	if strings.Join(placement.Arguments, " ") != "--add-dir /tmp/o" || strings.Join(placement.Environment, " ") != "CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1" {
		t.Errorf("arguments = %v, environment = %v", placement.Arguments, placement.Environment)
	}
	if placement.RootWritable {
		t.Error("Claude does not write the added directory")
	}
	for _, file := range placement.Files {
		if file.Path == ".claude/skills/using-x/SKILL.md" && !strings.Contains(string(file.Content), "model: opus") {
			t.Error("Claude keeps the model field")
		}
	}
}

func TestPlaceLetsTheProjectInstructionWin(t *testing.T) {
	placement, err := New().Place(shared.Input{Artifacts: sample(t), Root: "/tmp/o", NativeInstructions: []string{"CLAUDE.md"}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(paths(placement), "CLAUDE.md") || len(placement.Environment) != 0 || len(placement.Skipped) != 1 {
		t.Errorf("placement = %+v", placement)
	}
}

func TestPlaceOfNothingPlacesNothing(t *testing.T) {
	placement, err := New().Place(shared.Input{Root: "/tmp/o"})
	if err != nil || len(placement.Files) != 0 || len(placement.Arguments) != 0 {
		t.Errorf("placement = %+v, %v", placement, err)
	}
	only := artifact.Set{Instructions: []artifact.Instruction{{Content: "x"}}}
	placement, err = New().Place(shared.Input{Artifacts: only, Root: "/tmp/o", NativeInstructions: []string{"CLAUDE.md"}})
	if err != nil || len(placement.Files) != 0 || len(placement.Arguments) != 0 || len(placement.Skipped) != 1 {
		t.Errorf("a skipped instruction left %+v, %v", placement, err)
	}
}

func TestPlaceFailsOnABadRootAndOnCollidingPaths(t *testing.T) {
	if _, err := New().Place(shared.Input{Artifacts: sample(t), Root: "relative"}); err == nil {
		t.Error("a relative root passed")
	}
	clash := artifact.Set{Skills: []artifact.Skill{{Name: "a", Document: "x"}, {Name: "a", Document: "y"}}}
	if _, err := New().Place(shared.Input{Artifacts: clash, Root: "/tmp/o"}); err == nil {
		t.Error("two skills of one name passed")
	}
}

func TestFacts(t *testing.T) {
	facts := New().Facts()
	if facts.ReplacesHome || len(facts.NativeInstructions) != 2 {
		t.Errorf("facts = %+v", facts)
	}
}
