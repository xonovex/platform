package opencode

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
		{Path: "agents/reviewer.md", Content: []byte("---\ndescription: Reviews.\nmodel: opus\ntools: Read\n---\nBe strict.\n")},
		{Path: "commands/fix.md", Content: []byte("---\ndescription: Fix it.\nmodel: opus\n---\nFix the bug.\n")},
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

func contentOf(placement shared.Placement, path string) string {
	for _, file := range placement.Files {
		if file.Path == path {
			return string(file.Content)
		}
	}
	return ""
}

func TestPlaceMakesTheRootTheConfigurationDirectory(t *testing.T) {
	placement, err := New().Place(shared.Input{Artifacts: sample(t), Root: "/tmp/o"})
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, file := range placement.Files {
		paths = append(paths, file.Path)
	}
	if got, want := strings.Join(paths, ","), ".gitignore,AGENTS.md,agents/reviewer.md,commands/fix.md,opencode.json,skills/using-x/SKILL.md"; got != want {
		t.Errorf("paths = %s", got)
	}
	if strings.Join(placement.Environment, " ") != "OPENCODE_CONFIG_DIR=/tmp/o" || len(placement.Arguments) != 0 || !placement.RootWritable {
		t.Errorf("placement = %+v", placement)
	}
	agent := contentOf(placement, "agents/reviewer.md")
	if strings.Contains(agent, "model") || strings.Contains(agent, "tools") || !strings.Contains(agent, "description: Reviews.") {
		t.Errorf("agent = %q", agent)
	}
	if strings.Contains(contentOf(placement, "commands/fix.md"), "model") || strings.Contains(contentOf(placement, "skills/using-x/SKILL.md"), "model") {
		t.Error("a model field was kept")
	}
	if configuration := contentOf(placement, "opencode.json"); !strings.Contains(configuration, `"/tmp/o/AGENTS.md"`) || !strings.Contains(configuration, "$schema") {
		t.Errorf("opencode.json = %s", configuration)
	}
}

func TestPlaceLetsTheProjectInstructionWin(t *testing.T) {
	placement, err := New().Place(shared.Input{Artifacts: sample(t), Root: "/tmp/o", NativeInstructions: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatal(err)
	}
	if contentOf(placement, "AGENTS.md") != "" || contentOf(placement, "opencode.json") != "" || len(placement.Skipped) != 1 {
		t.Errorf("placement = %+v", placement)
	}
}

func TestPlaceOfNothingPlacesNothing(t *testing.T) {
	placement, err := New().Place(shared.Input{Root: "/tmp/o"})
	if err != nil || len(placement.Files) != 0 || len(placement.Environment) != 0 || placement.RootWritable {
		t.Errorf("placement = %+v, %v", placement, err)
	}
	if _, err := New().Place(shared.Input{Root: ""}); err == nil {
		t.Error("an empty root passed")
	}
}

func TestFacts(t *testing.T) {
	if facts := New().Facts(); facts.ReplacesHome || len(facts.NativeInstructions) != 1 {
		t.Errorf("facts = %+v", facts)
	}
}
