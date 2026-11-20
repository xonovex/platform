// Package catalog finds the skills of a repository and reads the routing
// queries that each skill declares.
package catalog

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/discovery"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

// SkillFile is the name of the file that makes a directory a skill.
const SkillFile = "SKILL.md"

// SkillDirs returns the skill directories below the paths, sorted. A path that
// is a skill directory, or its SKILL.md, names that skill. A skill directory is
// a leaf: the directories below it belong to the skill.
func SkillDirs(t tree.Tree, paths []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	skip := map[string]bool{}
	for _, name := range append(discovery.DefaultSkip(), t.Skip...) {
		skip[name] = true
	}
	var visit func(dir string) error
	visit = func(dir string) error {
		if t.IsFile(path.Join(dir, SkillFile)) {
			if !seen[dir] {
				seen[dir] = true
				out = append(out, dir)
			}
			return nil
		}
		entries, err := t.List(dir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir && !skip[entry.Name] && !strings.HasPrefix(entry.Name, ".") {
				if err := visit(path.Join(dir, entry.Name)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, p := range paths {
		p = path.Clean(p)
		if path.Base(p) == SkillFile && t.IsFile(p) {
			p = path.Dir(p)
		}
		if !t.IsDir(p) {
			continue
		}
		if err := visit(p); err != nil {
			return nil, err
		}
	}
	sortStrings(out)
	return out, nil
}

// TriggerFile and OutputFile name the eval documents of a skill.
const (
	TriggerFile = "eval-trigger.json"
	OutputFile  = "eval-output.json"
)

// Query is one routing query of a skill.
type Query struct {
	Text          string
	ShouldTrigger bool
	Split         string
	Rationale     string
}

// ReadQueries reads the trigger eval document of a skill and reports whether
// the file exists. A file that holds no valid query list is an error.
func ReadQueries(t tree.Tree, skillDir string) ([]Query, bool, error) {
	file := path.Join(skillDir, TriggerFile)
	text, ok, err := t.Optional(file)
	if err != nil || !ok {
		return nil, false, err
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return nil, true, fmt.Errorf("invalid JSON in %s: %w", file, err)
	}
	list, _ := raw["queries"].([]any)
	if len(list) == 0 {
		return nil, true, fmt.Errorf("invalid queries in %s: must contain at least one query", file)
	}
	queries := make([]Query, 0, len(list))
	for i, item := range list {
		entry, isEntry := item.(map[string]any)
		should, hasShould := entry["shouldTrigger"].(bool)
		prompt, _ := entry["prompt"].(string)
		if promptFile, ok := entry["promptFile"].(string); ok && prompt == "" {
			read, err := t.Read(path.Join(skillDir, promptFile))
			if err != nil {
				return nil, true, fmt.Errorf("invalid queries in %s: entry %d: %w", file, i+1, err)
			}
			prompt = strings.TrimSpace(read)
		}
		if !isEntry || prompt == "" || !hasShould {
			return nil, true, fmt.Errorf("invalid queries in %s: entry %d needs a prompt and shouldTrigger", file, i+1)
		}
		split, _ := entry["split"].(string)
		rationale, _ := entry["rationale"].(string)
		queries = append(queries, Query{Text: prompt, ShouldTrigger: should, Split: split, Rationale: rationale})
	}
	return queries, true, nil
}
