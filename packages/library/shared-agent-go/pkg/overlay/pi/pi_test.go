package pi

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
		{Path: "agents/reviewer.md", Content: []byte("Be strict.\n")},
		{Path: "commands/fix.md", Content: []byte("---\nmodel: opus\n---\nFix the bug.\n")},
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

func TestPlaceNamesEveryEntryInAFlag(t *testing.T) {
	placement, err := New().Place(shared.Input{Artifacts: sample(t), Root: "/tmp/o"})
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, file := range placement.Files {
		paths = append(paths, file.Path)
	}
	if got, want := strings.Join(paths, ","), "AGENTS.md,prompts/fix.md,skills/using-x/SKILL.md"; got != want {
		t.Errorf("paths = %s", got)
	}
	want := "--skill /tmp/o/skills/using-x --prompt-template /tmp/o/prompts/fix.md --append-system-prompt /tmp/o/AGENTS.md"
	if strings.Join(placement.Arguments, " ") != want {
		t.Errorf("arguments = %v", placement.Arguments)
	}
	if len(placement.Environment) != 0 || placement.RootWritable {
		t.Errorf("placement = %+v", placement)
	}
	if len(placement.Skipped) != 1 || placement.Skipped[0].Reason != "pi has no agents" {
		t.Errorf("skipped = %+v", placement.Skipped)
	}
}

func TestPlaceLetsTheProjectInstructionWin(t *testing.T) {
	placement, err := New().Place(shared.Input{Artifacts: sample(t), Root: "/tmp/o", NativeInstructions: []string{"CLAUDE.md"}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(placement.Arguments, " "), "--append-system-prompt") || len(placement.Skipped) != 2 {
		t.Errorf("placement = %+v", placement)
	}
}

func TestPlaceFailsOnABadRoot(t *testing.T) {
	if _, err := New().Place(shared.Input{Root: "x"}); err == nil {
		t.Error("a relative root passed")
	}
}

func TestFacts(t *testing.T) {
	if facts := New().Facts(); facts.ReplacesHome || len(facts.NativeInstructions) != 2 {
		t.Errorf("facts = %+v", facts)
	}
}
