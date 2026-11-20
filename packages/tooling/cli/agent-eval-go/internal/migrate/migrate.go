// Package migrate converts the legacy skill eval files into the documents the
// contract defines: evals.json becomes eval-output.json and eval-queries.json
// becomes eval-trigger.json. The conversion loses nothing: an assertion becomes
// a model predicate, a files entry becomes a fixture that reads its content from
// the same file, and a tier and a rationale keep their place.
package migrate

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// Legacy file names.
const (
	LegacyOutputFile  = "evals.json"
	LegacyTriggerFile = "eval-queries.json"
)

type legacyOutput struct {
	SkillName string `json:"skill_name"`
	Tier      string `json:"tier"`
	Evals     []struct {
		ID         json.RawMessage `json:"id"`
		Prompt     string          `json:"prompt"`
		Assertions []string        `json:"assertions"`
		Files      []string        `json:"files"`
	} `json:"evals"`
}

type legacyQuery struct {
	Query         string `json:"query"`
	ShouldTrigger bool   `json:"should_trigger"`
	Split         string `json:"split"`
	Rationale     string `json:"rationale"`
}

// Output converts the content of an evals.json.
func Output(data []byte) (model.OutputEval, error) {
	var legacy legacyOutput
	if err := json.Unmarshal(data, &legacy); err != nil {
		return model.OutputEval{}, fmt.Errorf("read %s: %w", LegacyOutputFile, err)
	}
	doc := model.OutputEval{Version: model.Version, Skill: legacy.SkillName, Tier: model.Tier(legacy.Tier)}
	for i, e := range legacy.Evals {
		id := strings.Trim(string(e.ID), `"`)
		if id == "" || id == "null" {
			id = fmt.Sprint(i + 1)
		}
		c := model.OutputCase{
			ID: "case-" + slug(id), Prompt: model.Prompt{Prompt: e.Prompt}, Gate: model.Required,
		}
		if len(e.Assertions) == 0 {
			return model.OutputEval{}, fmt.Errorf("%s: case %s has no assertions", LegacyOutputFile, c.ID)
		}
		for _, a := range e.Assertions {
			c.Expect = append(c.Expect, model.Predicate{Kind: model.KindLLM, Criterion: a})
		}
		for _, f := range e.Files {
			c.Fixtures = append(c.Fixtures, model.Fixture{Path: fixturePath(f), ContentFile: f})
		}
		doc.Cases = append(doc.Cases, c)
	}
	return doc, nil
}

// fixturePath drops the evals/files/<case>/ prefix, so the agent sees the file
// where a user would keep it.
func fixturePath(file string) string {
	parts := strings.Split(file, "/")
	if len(parts) > 4 && parts[0] == "evals" && parts[1] == "files" {
		return strings.Join(parts[3:], "/")
	}
	return file
}

// Trigger converts the content of an eval-queries.json for the named skill.
func Trigger(data []byte, skill string) (model.TriggerEval, error) {
	var legacy []legacyQuery
	if err := json.Unmarshal(data, &legacy); err != nil {
		return model.TriggerEval{}, fmt.Errorf("read %s: %w", LegacyTriggerFile, err)
	}
	doc := model.TriggerEval{Version: model.Version, Skill: skill}
	for i, q := range legacy {
		doc.Queries = append(doc.Queries, model.TriggerQuery{
			ID: fmt.Sprintf("query-%04d", i+1), Prompt: model.Prompt{Prompt: q.Query},
			ShouldTrigger: q.ShouldTrigger, Split: model.Split(q.Split), Rationale: q.Rationale,
		})
	}
	return doc, nil
}

func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// Result lists what the conversion of one skill directory wrote and removed.
type Result struct {
	Skill   string
	Written []string
	Removed []string
}

// Converter converts skill directories through the file-system port and checks
// each new document against its schema before it writes it.
type Converter struct {
	FS     fs.FileSystem
	Reader document.Reader
	// RemoveLegacy deletes the legacy files after their replacements are written.
	RemoveLegacy bool
}

// Skill converts the legacy files of one skill directory. A directory without
// legacy files returns an empty result.
func (c Converter) Skill(dir string) (Result, error) {
	result := Result{Skill: path.Base(dir)}
	type job struct {
		legacy, target, schema string
		build                  func([]byte) (any, error)
	}
	jobs := []job{
		{path.Join(dir, LegacyOutputFile), path.Join(dir, model.OutputFile), schemas.EvalOutput,
			func(b []byte) (any, error) { return Output(b) }},
		{path.Join(dir, LegacyTriggerFile), path.Join(dir, model.TriggerFile), schemas.EvalTrigger,
			func(b []byte) (any, error) { return Trigger(b, path.Base(dir)) }},
	}
	for _, j := range jobs {
		exists, err := fs.Exists(c.FS, j.legacy)
		if err != nil {
			return result, err
		}
		if !exists {
			continue
		}
		raw, err := c.FS.ReadFile(j.legacy)
		if err != nil {
			return result, err
		}
		doc, err := j.build(raw)
		if err != nil {
			return result, err
		}
		out, err := model.Encode(doc)
		if err != nil {
			return result, err
		}
		violations, err := c.Reader.Validate(j.schema, out)
		if err != nil {
			return result, err
		}
		if len(violations) > 0 {
			return result, fmt.Errorf("converted %s is invalid: %s: %s", j.legacy, violations[0].Pointer, violations[0].Message)
		}
		if err := c.FS.WriteFile(j.target, out); err != nil {
			return result, err
		}
		result.Written = append(result.Written, j.target)
		if c.RemoveLegacy {
			if err := c.FS.Remove(j.legacy); err != nil {
				return result, err
			}
			result.Removed = append(result.Removed, j.legacy)
		}
	}
	return result, nil
}
