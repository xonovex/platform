package kiro

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

func TestPlaceMakesTheRootTheKiroHome(t *testing.T) {
	set := setOf(t, map[string]string{
		"skills/using-x/SKILL.md": "---\nname: using-x\ndescription: d\nmodel: opus\n---\n",
		"agents/reviewer.md":      "---\ndescription: Reviews.\n---\nBe strict.\n",
		"commands/fix.md":         "---\ndescription: Fix.\n---\nFix the bug.\n",
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
	want := "agents/reviewer.json,prompts/fix.md,settings/cli.json,skills/using-x/SKILL.md,steering/instructions.md"
	if strings.Join(paths, ",") != want {
		t.Errorf("paths = %s", strings.Join(paths, ","))
	}
	if strings.Join(placement.Environment, " ") != "KIRO_HOME=/tmp/o" || len(placement.Arguments) != 0 || !placement.RootWritable {
		t.Errorf("placement = %+v", placement)
	}
	if got := contentOf(placement, "prompts/fix.md"); got != "Fix the bug.\n" {
		t.Errorf("prompt = %q", got)
	}
	if got := contentOf(placement, "steering/instructions.md"); got != "---\ninclusion: always\n---\nRules.\n" {
		t.Errorf("steering = %q", got)
	}
	agent := contentOf(placement, "agents/reviewer.json")
	for _, part := range []string{`"name": "reviewer"`, `"prompt": "Be strict."`, `file:///tmp/o/skills/**/SKILL.md`, `file:///tmp/o/steering/instructions.md`} {
		if !strings.Contains(agent, part) {
			t.Errorf("agent lacks %s: %s", part, agent)
		}
	}
	if !strings.Contains(contentOf(placement, "settings/cli.json"), `"telemetry.enabled": false`) {
		t.Error("telemetry stays on")
	}
}

func TestPlaceAgentWithNoResources(t *testing.T) {
	placement, err := New().Place(shared.Input{Artifacts: setOf(t, map[string]string{"agents/a.md": "Body."}), Root: "/tmp/o"})
	if err != nil {
		t.Fatal(err)
	}
	agent := contentOf(placement, "agents/a.json")
	if !strings.Contains(agent, `"resources": []`) || strings.Contains(agent, "description") {
		t.Errorf("agent = %s", agent)
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
	if len(placement.Skipped) != 2 || contentOf(placement, "steering/instructions.md") != "" {
		t.Errorf("placement = %+v", placement)
	}
}

func TestPlaceOfNothingAndBadRoot(t *testing.T) {
	if placement, err := New().Place(shared.Input{Root: "/tmp/o"}); err != nil || len(placement.Files) != 0 || len(placement.Environment) != 0 {
		t.Errorf("placement = %+v, %v", placement, err)
	}
	if _, err := New().Place(shared.Input{Root: "x"}); err == nil {
		t.Error("a relative root passed")
	}
}

func TestFacts(t *testing.T) {
	if facts := New().Facts(); !facts.ReplacesHome || len(facts.NativeInstructions) != 1 {
		t.Errorf("facts = %+v", facts)
	}
}
