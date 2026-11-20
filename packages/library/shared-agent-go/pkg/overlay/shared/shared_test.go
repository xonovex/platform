package shared

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
)

func TestCheckRoot(t *testing.T) {
	if err := CheckRoot("/tmp/overlay-1"); err != nil {
		t.Errorf("a good root failed: %v", err)
	}
	for _, root := range []string{"", "tmp/x", "/tmp/../x", "/tmp//x", "/", "/tmp/a*b", "/tmp/a?b", "/tmp/a[b", "/tmp/a{b", `/tmp/a\b`, "/tmp/a\nb"} {
		if err := CheckRoot(root); err == nil {
			t.Errorf("CheckRoot(%q) passed", root)
		}
	}
}

func TestFinishSortsAndFailsClosed(t *testing.T) {
	placement, err := Finish(Placement{
		Files:     []File{{Path: "b"}, {Path: "a/x"}},
		Skipped:   []Skip{{Kind: artifact.KindSkill, Name: "z"}, {Kind: artifact.KindAgent, Name: "y"}, {Kind: artifact.KindAgent, Name: "a"}},
		Arguments: []string{"--flag"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if placement.Files[0].Path != "a/x" || placement.Skipped[0].Name != "a" || placement.Skipped[2].Kind != artifact.KindSkill || len(placement.Arguments) != 1 {
		t.Errorf("placement = %+v", placement)
	}
	if _, err := Finish(Placement{Files: []File{{Path: "../x"}}}); err == nil {
		t.Error("a path that leaves the root passed")
	}
	if _, err := Finish(Placement{Files: []File{{Path: "a"}, {Path: "A"}}}); err == nil {
		t.Error("colliding paths passed")
	}
}

func TestFinishDropsTheRestOfAPlacementWithoutFiles(t *testing.T) {
	placement, err := Finish(Placement{Arguments: []string{"--x"}, Environment: []string{"A=b"}, RootWritable: true, Skipped: []Skip{{Name: "n"}}})
	if err != nil || len(placement.Arguments) != 0 || len(placement.Environment) != 0 || placement.RootWritable || len(placement.Skipped) != 1 {
		t.Errorf("placement = %+v, %v", placement, err)
	}
}

func TestSkillFilesDropFieldsAndKeepExtras(t *testing.T) {
	skill := artifact.Skill{
		Name:     "using-x",
		Document: "---\nname: using-x\nmodel: opus\n---\nbody\n",
		Files:    []artifact.File{{Path: "scripts/run.sh", Executable: true, Content: []byte("#!/bin/sh")}},
	}
	files := SkillFiles("skills", skill, "model")
	if len(files) != 2 || files[0].Path != "skills/using-x/SKILL.md" || strings.Contains(string(files[0].Content), "model") || files[0].Kind != artifact.KindSkill {
		t.Errorf("files = %+v", files)
	}
	if files[1].Path != "skills/using-x/scripts/run.sh" || !files[1].Executable {
		t.Errorf("extra = %+v", files[1])
	}
}

func TestDocumentFile(t *testing.T) {
	file := DocumentFile("a/b.md", artifact.KindAgent, "---\nmodel: x\ndescription: d\n---\nbody", "model")
	if file.Path != "a/b.md" || file.Kind != artifact.KindAgent || string(file.Content) != "---\ndescription: d\n---\nbody" {
		t.Errorf("file = %+v / %s", file, file.Content)
	}
}

func TestInstructionSkip(t *testing.T) {
	with := artifact.Set{Instructions: []artifact.Instruction{{Content: "x"}}}
	if _, ok := InstructionSkip(Input{Artifacts: with}); ok {
		t.Error("an instruction without a native file was skipped")
	}
	if _, ok := InstructionSkip(Input{NativeInstructions: []string{"AGENTS.md"}}); ok {
		t.Error("a native file without an instruction was reported")
	}
	skip, ok := InstructionSkip(Input{Artifacts: with, NativeInstructions: []string{"AGENTS.md", "CLAUDE.md"}})
	if !ok || skip.Name != "AGENTS.md" || skip.Kind != artifact.KindInstruction || !strings.Contains(skip.Reason, "wins") {
		t.Errorf("skip = %+v, %v", skip, ok)
	}
}

func TestSmallHelpers(t *testing.T) {
	if got := Join("/r", "a/b"); got != "/r/a/b" {
		t.Errorf("Join = %q", got)
	}
	if skip := CommandShadowed("fix"); skip.Kind != artifact.KindCommand || skip.Name != "fix" {
		t.Errorf("CommandShadowed = %+v", skip)
	}
	if skip := Unsupported(artifact.KindAgent, "r", "pi"); skip.Reason != "pi has no agents" {
		t.Errorf("Unsupported = %+v", skip)
	}
	names := SkillNames(artifact.Set{Skills: []artifact.Skill{{Name: "a"}, {Name: "b"}}})
	if len(names) != 2 || !names["a"] || names["c"] {
		t.Errorf("SkillNames = %v", names)
	}
}
