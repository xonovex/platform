package artifact

import (
	"strings"
	"testing"
)

const skillDoc = "---\nname: using-x\ndescription: Use when x.\n---\n# X\n"

func sourceTree(t *testing.T, files ...File) Tree {
	t.Helper()
	tree, err := NewTree(files)
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

func TestFromTreeReadsEveryClass(t *testing.T) {
	tree := sourceTree(t,
		file("skills/using-x/SKILL.md", skillDoc),
		file("skills/using-x/references/a.md", "ref"),
		File{Path: "skills/using-x/scripts/run.sh", Content: []byte("#!/bin/sh"), Executable: true},
		file("skills/a-skill/SKILL.md", "---\nname: a-skill\ndescription: d\n---\n"),
		file("skills/README.md", "loose file"),
		file("agents/reviewer.md", "---\ndescription: r\n---\nbody"),
		file("agents/notes.txt", "ignored"),
		file("agents/nested/deep.md", "ignored"),
		file("commands/fix.md", "do it"),
		file("AGENTS.md", "rules\n"),
		file("other/file", "ignored"),
	)
	set, err := FromTree("src", tree)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Skills) != 2 || set.Skills[0].Name != "a-skill" || set.Skills[1].Name != "using-x" {
		t.Fatalf("skills = %+v", set.Skills)
	}
	skill := set.Skills[1]
	if skill.Source != "src" || skill.Document != skillDoc || len(skill.Files) != 2 || skill.Files[0].Path != "references/a.md" || !skill.Files[1].Executable {
		t.Errorf("skill = %+v", skill)
	}
	if len(set.Agents) != 1 || set.Agents[0].Name != "reviewer" || len(set.Commands) != 1 || set.Commands[0].Name != "fix" {
		t.Errorf("agents = %+v, commands = %+v", set.Agents, set.Commands)
	}
	if len(set.Instructions) != 1 || set.Instructions[0].Content != "rules\n" || set.Instructions[0].Source != "src" {
		t.Errorf("instructions = %+v", set.Instructions)
	}
}

func TestFromTreeOfNothingIsEmpty(t *testing.T) {
	set, err := FromTree("src", nil)
	if err != nil || !set.Empty() {
		t.Errorf("set = %+v, %v", set, err)
	}
}

func TestFromTreeFailsClosed(t *testing.T) {
	for name, files := range map[string][]File{
		"bad skill directory": {file("skills/Bad_Name/SKILL.md", skillDoc)},
		"no SKILL.md":         {file("skills/using-x/notes.md", "n")},
		"bad front matter":    {file("skills/using-x/SKILL.md", "---\nname: using-x\n")},
		"name mismatch":       {file("skills/using-x/SKILL.md", "---\nname: other\ndescription: d\n---\n")},
		"no name":             {file("skills/using-x/SKILL.md", "---\ndescription: d\n---\n")},
		"no description":      {file("skills/using-x/SKILL.md", "---\nname: using-x\n---\n")},
		"bad agent name":      {file("agents/Bad.md", "x")},
		"bad agent document":  {file("agents/ok.md", "---\nname: x\n")},
		"bad command name":    {file("commands/Bad.md", "x")},
		"bad command doc":     {file("commands/ok.md", "---\nname: x\n")},
	} {
		if _, err := FromTree("src", sourceTree(t, files...)); err == nil {
			t.Errorf("%s: FromTree passed", name)
		} else if !strings.Contains(err.Error(), "src") {
			t.Errorf("%s: the error %q names no source", name, err)
		}
	}
}

func TestMergeCombinesAndSorts(t *testing.T) {
	a, _ := FromTree("a", sourceTree(t, file("skills/b-skill/SKILL.md", "---\nname: b-skill\ndescription: d\n---\n"), file("AGENTS.md", "A")))
	b, _ := FromTree("b", sourceTree(t, file("skills/a-skill/SKILL.md", "---\nname: a-skill\ndescription: d\n---\n"), file("agents/x.md", "x"), file("agents/w.md", "w"), file("commands/y.md", "y"), file("commands/v.md", "v"), file("AGENTS.md", "B")))
	merged, err := Merge(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(merged.Skills) != 2 || merged.Skills[0].Name != "a-skill" || len(merged.Agents) != 2 || merged.Agents[0].Name != "w" || len(merged.Commands) != 2 || merged.Commands[0].Name != "v" {
		t.Errorf("merged = %+v", merged)
	}
	if merged.InstructionText() != "A\n\nB\n" {
		t.Errorf("instructions = %q", merged.InstructionText())
	}
}

func TestMergeRejectsAName(t *testing.T) {
	skill, _ := FromTree("a", sourceTree(t, file("skills/using-x/SKILL.md", skillDoc)))
	other, _ := FromTree("b", sourceTree(t, file("skills/using-x/SKILL.md", skillDoc)))
	agentA, _ := FromTree("a", sourceTree(t, file("agents/x.md", "x")))
	agentB, _ := FromTree("b", sourceTree(t, file("agents/x.md", "x")))
	commandA, _ := FromTree("a", sourceTree(t, file("commands/x.md", "x")))
	commandB, _ := FromTree("b", sourceTree(t, file("commands/x.md", "x")))
	for name, pair := range map[string][2]Set{"skill": {skill, other}, "agent": {agentA, agentB}, "command": {commandA, commandB}} {
		_, err := Merge(pair[0], pair[1])
		if err == nil || !strings.Contains(err.Error(), "a and by b") {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
