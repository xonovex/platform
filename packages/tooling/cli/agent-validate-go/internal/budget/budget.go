// Package budget collects the prose files of a skill catalog and measures their
// size in words, for the drift check that holds them to a word budget and for
// the command that records the current sizes as the budgets.
package budget

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

// Kinds of file that a budget covers.
const (
	KindSkill     = "skill"
	KindReference = "reference"
)

// File is one budgeted file of a skill.
type File struct {
	Path string
	Kind string
	Text string
}

// Words counts the prose words of the file. Fenced code does not count.
func (f File) Words() int { return markdown.ProseWords(f.Text) }

// Collect reads the SKILL.md and the reference files of every skill below the
// paths.
func Collect(t tree.Tree, paths []string) ([]File, error) {
	dirs, err := catalog.SkillDirs(t, paths)
	if err != nil {
		return nil, err
	}
	var files []File
	for _, dir := range dirs {
		collected, err := collectSkill(t, dir)
		if err != nil {
			return nil, err
		}
		files = append(files, collected...)
	}
	return files, nil
}

func collectSkill(t tree.Tree, dir string) ([]File, error) {
	var out []File
	skill := path.Join(dir, catalog.SkillFile)
	if text, found, err := t.Optional(skill); err != nil {
		return nil, err
	} else if found {
		out = append(out, File{Path: skill, Kind: KindSkill, Text: text})
	}
	entries, err := t.List(path.Join(dir, "references"))
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir || !strings.HasSuffix(entry.Name, ".md") {
			continue
		}
		p := path.Join(dir, "references", entry.Name)
		text, err := t.Read(p)
		if err != nil {
			return nil, err
		}
		out = append(out, File{Path: p, Kind: KindReference, Text: text})
	}
	return out, nil
}

// Seed records the current size of each file that has prose, so the ratchet
// starts from the catalog as it is.
func Seed(files []File) map[string]int {
	seeded := make(map[string]int, len(files))
	for _, f := range files {
		if words := f.Words(); words > 0 {
			seeded[f.Path] = words
		}
	}
	return seeded
}

// Marshal writes a manifest with its keys in code point order, so a reseed
// produces the same bytes on every machine.
func Marshal(manifest map[string]int) ([]byte, error) {
	if manifest == nil {
		manifest = map[string]int{}
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	return append(data, '\n'), err
}
