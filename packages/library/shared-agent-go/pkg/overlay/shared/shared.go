// Package shared is the overlay axis core: the Dialect port every harness leaf
// implements, the Input a dialect reads and the Placement it returns. A dialect
// turns the neutral artifacts into the files, variables and arguments one
// harness needs. It takes no effect to do so, so the launcher writes the files
// and a golden test can check the result. The core names no harness.
package shared

import (
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
)

// Facts are what a dialect declares about its harness. The launcher compares
// them with the facts of the agent and names no harness.
type Facts struct {
	// NativeInstructions are the project files, relative to the work directory,
	// that the harness reads as its own instructions. One that exists wins over
	// the instruction of the overlay.
	NativeInstructions []string
	// ReplacesHome is true when the overlay root becomes the configuration home of
	// the harness, which hides the configuration and the login that the home of the
	// user holds.
	ReplacesHome bool
}

// Dialect is one harness's reading of the neutral artifacts.
type Dialect interface {
	Facts() Facts
	// Place derives the overlay of one run. It is a pure function of the input.
	Place(Input) (Placement, error)
}

// Input is what a dialect reads.
type Input struct {
	Artifacts artifact.Set
	// Root is the absolute directory the launcher allocated for the overlay. The
	// dialect may name it in a variable or an argument, and the launcher writes
	// the files below it.
	Root string
	// NativeInstructions are the files of Facts.NativeInstructions that exist in
	// the work directory.
	NativeInstructions []string
}

// File is one file of the overlay.
type File struct {
	// Path is relative to the overlay root, in slash form.
	Path       string
	Kind       artifact.Kind
	Content    []byte
	Executable bool
}

// Skip is content the dialect did not place, and why.
type Skip struct {
	Kind   artifact.Kind
	Name   string
	Reason string
}

// Placement is the whole result of a dialect. Environment holds KEY=VALUE
// entries and Arguments are appended after every other argument of the harness,
// because a flag that takes a list would otherwise read what follows it as part
// of the list. A placement with no file has neither. RootWritable says the harness
// writes files of its own below the root.
type Placement struct {
	Files        []File
	Environment  []string
	Arguments    []string
	Skipped      []Skip
	RootWritable bool
}

// CheckRoot fails on a root a harness would not take: a relative or unclean
// path, the file system root, or a path with a control or glob character.
func CheckRoot(root string) error {
	switch {
	case root == "":
		return errors.New("the overlay root is empty")
	case !path.IsAbs(root):
		return fmt.Errorf("the overlay root %q is not an absolute path", root)
	case path.Clean(root) != root:
		return fmt.Errorf("the overlay root %q is not a clean path", root)
	case root == "/":
		return errors.New("the overlay root is the root of the file system")
	case strings.ContainsFunc(root, func(character rune) bool {
		return character < ' ' || character == 0x7f || strings.ContainsRune(`*?[]{}\`, character)
	}):
		return fmt.Errorf("the overlay root %q holds a glob, backslash or control character", root)
	}
	return nil
}

// Finish sorts the files and the skips of a placement, and fails closed on a
// file path that leaves the root or that collides with another. A placement with
// no file keeps only its skips.
func Finish(placement Placement) (Placement, error) {
	slices.SortFunc(placement.Skipped, func(a, b Skip) int {
		return strings.Compare(string(a.Kind)+"\x00"+a.Name, string(b.Kind)+"\x00"+b.Name)
	})
	if len(placement.Files) == 0 {
		return Placement{Skipped: placement.Skipped}, nil
	}
	files := make([]artifact.File, len(placement.Files))
	for index, file := range placement.Files {
		files[index] = artifact.File{Path: file.Path}
	}
	if _, err := artifact.NewTree(files); err != nil {
		return Placement{}, err
	}
	slices.SortFunc(placement.Files, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	return placement, nil
}

// SkillFiles places a skill below directory: its SKILL.md without the dropped
// front matter keys, and the files beside it as they are.
func SkillFiles(directory string, skill artifact.Skill, drop ...string) []File {
	base := directory + "/" + skill.Name + "/"
	files := []File{{
		Path:    base + artifact.SkillFile,
		Kind:    artifact.KindSkill,
		Content: []byte(artifact.DropFields(skill.Document, drop...)),
	}}
	for _, extra := range skill.Files {
		files = append(files, File{Path: base + extra.Path, Kind: artifact.KindSkill, Content: extra.Content, Executable: extra.Executable})
	}
	return files
}

// DocumentFile places a Markdown document of an agent or a command at filePath,
// without the dropped front matter keys.
func DocumentFile(filePath string, kind artifact.Kind, text string, drop ...string) File {
	return File{Path: filePath, Kind: kind, Content: []byte(artifact.DropFields(text, drop...))}
}

// InstructionSkip reports the instruction the overlay leaves out because the
// project has a file the harness reads as its own instructions. It reports
// nothing when the set has no instruction or the project has no such file.
func InstructionSkip(input Input) (Skip, bool) {
	if len(input.Artifacts.Instructions) == 0 || len(input.NativeInstructions) == 0 {
		return Skip{}, false
	}
	return Skip{Kind: artifact.KindInstruction, Name: input.NativeInstructions[0], Reason: "the project has " + input.NativeInstructions[0] + ", which wins"}, true
}

// CommandShadowed reports the skip of a command that shares its name with a
// skill, for a harness that offers both as one entry and lets the skill win.
func CommandShadowed(name string) Skip {
	return Skip{Kind: artifact.KindCommand, Name: name, Reason: "the harness offers a skill and a command of one name as one entry, and the skill wins"}
}

// Unsupported reports the skip of an artifact the harness has no place for.
func Unsupported(kind artifact.Kind, name, harness string) Skip {
	return Skip{Kind: kind, Name: name, Reason: harness + " has no " + string(kind) + "s"}
}

// Join returns root joined with a relative slash path.
func Join(root, relative string) string { return root + "/" + relative }

// SkillNames returns the names of the skills of a set as a lookup.
func SkillNames(set artifact.Set) map[string]bool {
	names := make(map[string]bool, len(set.Skills))
	for _, skill := range set.Skills {
		names[skill.Name] = true
	}
	return names
}
