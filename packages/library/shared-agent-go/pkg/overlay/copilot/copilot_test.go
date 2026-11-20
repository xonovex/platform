package copilot

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

func TestPlaceBuildsAPluginDirectoryAndAnInstructionDirectory(t *testing.T) {
	set := setOf(t, map[string]string{
		"skills/using-x/SKILL.md": "---\nname: using-x\ndescription: d\nmodel: opus\n---\n",
		"agents/reviewer.md":      "---\ndescription: r\nmodel: opus\n---\nb",
		"commands/fix.md":         "---\nmodel: opus\n---\nfix",
		"AGENTS.md":               "Rules.\n",
	})
	placement, err := New().Place(shared.Input{Artifacts: set, Root: "/tmp/o"})
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, file := range placement.Files {
		paths = append(paths, file.Path)
	}
	want := "instructions/AGENTS.md,plugin/agents/reviewer.md,plugin/commands/fix.md,plugin/plugin.json,plugin/skills/using-x/SKILL.md"
	if strings.Join(paths, ",") != want {
		t.Errorf("paths = %s", strings.Join(paths, ","))
	}
	if strings.Join(placement.Arguments, " ") != "--plugin-dir /tmp/o/plugin" || strings.Join(placement.Environment, " ") != "COPILOT_CUSTOM_INSTRUCTIONS_DIRS=/tmp/o/instructions" || placement.RootWritable {
		t.Errorf("placement = %+v", placement)
	}
	for _, file := range placement.Files {
		if strings.Contains(string(file.Content), "model") {
			t.Errorf("%s keeps the model field", file.Path)
		}
	}
}

func TestPlaceLetsASkillShadowACommandAndTheProjectWin(t *testing.T) {
	set := setOf(t, map[string]string{
		"skills/fix/SKILL.md": "---\nname: fix\ndescription: d\n---\n",
		"commands/fix.md":     "fix",
		"AGENTS.md":           "x",
	})
	placement, err := New().Place(shared.Input{Artifacts: set, Root: "/tmp/o", NativeInstructions: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(placement.Skipped) != 2 || len(placement.Environment) != 0 {
		t.Errorf("placement = %+v", placement)
	}
}

func TestPlaceOfOnlyAnInstructionHasNoPluginDirectory(t *testing.T) {
	placement, err := New().Place(shared.Input{Artifacts: setOf(t, map[string]string{"AGENTS.md": "x"}), Root: "/tmp/o"})
	if err != nil || len(placement.Arguments) != 0 || len(placement.Files) != 1 || len(placement.Environment) != 1 {
		t.Errorf("placement = %+v, %v", placement, err)
	}
}

func TestPlaceOfNothingAndBadRoot(t *testing.T) {
	if placement, err := New().Place(shared.Input{Root: "/tmp/o"}); err != nil || len(placement.Files) != 0 {
		t.Errorf("placement = %+v, %v", placement, err)
	}
	if _, err := New().Place(shared.Input{Root: "x"}); err == nil {
		t.Error("a relative root passed")
	}
}

func TestFacts(t *testing.T) {
	if facts := New().Facts(); facts.ReplacesHome || len(facts.NativeInstructions) != 4 {
		t.Errorf("facts = %+v", facts)
	}
}
