package skill

import (
	"encoding/json"
	"path"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
)

// The query shapes an early generator produced for every skill say nothing
// about one skill, so a catalog query must not match them.
var genericQueries = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^Help me handle .+ correctly in this .+ task, including the edge cases that usually get missed\.$`),
	regexp.MustCompile("(?i)^I'm reviewing `[^`]+` in an? .+ project\\. The happy path works, but .+ is unclear\\."),
	regexp.MustCompile(`(?i)^A teammate says our .+ change may mishandle .+ before release\. Inspect the likely risk around `),
	regexp.MustCompile(`(?i)^CI started failing after we changed .+ in an? .+ project\. Use `),
	regexp.MustCompile("(?i)^A review comment on `work/"),
	regexp.MustCompile(`(?i)^quick pre-merge sanity check: after the `),
	regexp.MustCompile(`(?i)^The clean Linux CI job fails only for the minimal `),
}

var nearMissOwner = regexp.MustCompile(`^near miss owned by (\S+)$`)

func checkEvals(ctx check.Context, report *check.Reporter, options Options, doc document.Document, dir string) error {
	name, _ := doc.String("name")
	if err := checkOutputEvals(ctx, report, options, dir, name); err != nil {
		return err
	}
	return checkTriggerEvals(ctx, report, options, dir)
}

// readJSON reads a catalog file and reports a missing one or a malformed one.
func readJSON(ctx check.Context, report *check.Reporter, required bool, dir, name string) (any, string, bool, error) {
	file := path.Join(dir, name)
	text, ok, err := ctx.Tree.Optional(file)
	if err != nil {
		return nil, file, false, err
	}
	if !ok {
		if required {
			report.Fail("skill.evals-missing", file, 1, "catalog: missing %s", name)
		}
		return nil, file, false, nil
	}
	var value any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		report.Fail("skill.evals-json", file, 1, "catalog: %s is not valid JSON: %v", name, err)
		return nil, file, false, nil
	}
	return value, file, true, nil
}

// validateSchema reports each violation of a document that declares a version.
func validateSchema(ctx check.Context, report *check.Reporter, schema, file string) (bool, error) {
	text, err := ctx.Tree.Read(file)
	if err != nil {
		return false, err
	}
	violations, err := ctx.Validator.Validate(schema, []byte(text))
	if err != nil {
		return false, err
	}
	for _, v := range violations {
		report.Fail("skill.evals-schema", file, 1, "catalog: %s%s: %s", path.Base(file), v.Pointer, v.Message)
	}
	return len(violations) == 0, nil
}

func isVersioned(value any) bool {
	object, ok := value.(map[string]any)
	if !ok {
		return false
	}
	_, versioned := object["version"]
	return versioned
}

func nonEmptyString(value any) bool {
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) != ""
}

func checkOutputEvals(ctx check.Context, report *check.Reporter, options Options, dir, skillName string) error {
	raw, file, ok, err := readJSON(ctx, report, options.RequireEvals, dir, "evals.json")
	if err != nil || !ok {
		return err
	}
	if isVersioned(raw) {
		valid, err := validateSchema(ctx, report, schemas.EvalOutput, file)
		if err != nil || !valid {
			return err
		}
		object := raw.(map[string]any)
		if object["skill"] != skillName {
			report.Fail("skill.evals-name", file, 1, "catalog: evals.json skill '%v' does not match '%s'", object["skill"], skillName)
		}
		if cases, _ := object["cases"].([]any); len(cases) < 3 {
			report.Fail("skill.evals-count", file, 1, "catalog: evals.json must contain at least 3 output evals")
		}
		return nil
	}
	object, isObject := raw.(map[string]any)
	if !isObject {
		report.Fail("skill.evals-shape", file, 1, "catalog: evals.json must be an object")
		return nil
	}
	if object["skill_name"] != skillName {
		report.Fail("skill.evals-name", file, 1, "catalog: evals.json skill_name '%v' does not match '%s'", object["skill_name"], skillName)
	}
	switch object["tier"] {
	case "aggressive", "moderate", "conservative":
	default:
		report.Fail("skill.evals-tier", file, 1, "catalog: evals.json tier must be aggressive, moderate, or conservative")
	}
	evals, _ := object["evals"].([]any)
	if len(evals) < 3 {
		report.Fail("skill.evals-count", file, 1, "catalog: evals.json must contain at least 3 output evals")
		return nil
	}
	for i, item := range evals {
		position := i + 1
		entry, isEntry := item.(map[string]any)
		if !isEntry {
			report.Fail("skill.evals-entry", file, 1, "catalog: evals.json entry %d must be an object", position)
			continue
		}
		if !nonEmptyString(entry["prompt"]) {
			report.Fail("skill.evals-prompt", file, 1, "catalog: evals.json entry %d needs a non-empty prompt", position)
		}
		if !nonEmptyStrings(entry["assertions"]) {
			report.Fail("skill.evals-assertions", file, 1, "catalog: evals.json entry %d needs non-empty string assertions", position)
		}
	}
	return nil
}

func nonEmptyStrings(value any) bool {
	list, ok := value.([]any)
	if !ok || len(list) == 0 {
		return false
	}
	for _, item := range list {
		if !nonEmptyString(item) {
			return false
		}
	}
	return true
}

// triggerEntry is a routing query reduced to what the rules read.
type triggerEntry struct {
	positive bool
	split    string
	owner    string
}

func checkTriggerEvals(ctx check.Context, report *check.Reporter, options Options, dir string) error {
	raw, file, ok, err := readJSON(ctx, report, options.RequireEvals, dir, "eval-queries.json")
	if err != nil || !ok {
		return err
	}
	var entries []triggerEntry
	switch {
	case isVersioned(raw):
		valid, err := validateSchema(ctx, report, schemas.EvalTrigger, file)
		if err != nil || !valid {
			return err
		}
		queries, _, err := catalog.ReadQueries(ctx.Tree, dir)
		if err != nil {
			report.Fail("skill.evals-queries", file, 1, "catalog: %v", err)
			return nil
		}
		seen := map[string]bool{}
		for _, q := range queries {
			entries = append(entries, queryEntry(report, file, q.Text, q.ShouldTrigger, q.Split, q.Rationale, seen))
		}
	default:
		list, isList := raw.([]any)
		if !isList {
			report.Fail("skill.evals-shape", file, 1, "catalog: eval-queries.json must be an array")
			return nil
		}
		seen := map[string]bool{}
		for i, item := range list {
			if entry, valid := legacyEntry(report, file, i, item, seen); valid {
				entries = append(entries, entry)
			}
		}
	}
	checkTriggerMinimums(report, file, entries)
	return nil
}

func legacyEntry(report *check.Reporter, file string, index int, item any, seen map[string]bool) (triggerEntry, bool) {
	position := index + 1
	object, isObject := item.(map[string]any)
	if !isObject {
		report.Fail("skill.evals-entry", file, 1, "catalog: eval-queries.json entry %d must be an object", position)
		return triggerEntry{}, false
	}
	query, _ := object["query"].(string)
	should, hasShould := object["should_trigger"].(bool)
	rationale, _ := object["rationale"].(string)
	split, _ := object["split"].(string)
	if !nonEmptyString(object["query"]) {
		report.Fail("skill.evals-query", file, 1, "catalog: eval-queries.json entry %d needs a non-empty query", position)
	} else {
		duplicateOrGeneric(report, file, query, seen)
	}
	if !hasShould {
		report.Fail("skill.evals-should-trigger", file, 1, "catalog: eval-queries.json entry %d needs boolean should_trigger", position)
		return triggerEntry{}, false
	}
	if split != "train" && split != "validation" {
		report.Fail("skill.evals-split", file, 1, "catalog: eval-queries.json entry %d needs split train or validation", position)
		split = ""
	}
	return triggerEntry{positive: should, split: split, owner: ownerOf(should, rationale)}, true
}

func queryEntry(report *check.Reporter, file string, query string, should bool, split, rationale string, seen map[string]bool) triggerEntry {
	duplicateOrGeneric(report, file, query, seen)
	return triggerEntry{positive: should, split: split, owner: ownerOf(should, rationale)}
}

func ownerOf(should bool, rationale string) string {
	if should {
		return ""
	}
	if m := nearMissOwner.FindStringSubmatch(strings.TrimSpace(rationale)); m != nil {
		return m[1]
	}
	return ""
}

func duplicateOrGeneric(report *check.Reporter, file, query string, seen map[string]bool) {
	key := strings.ToLower(strings.TrimSpace(query))
	if seen[key] {
		report.Fail("skill.evals-duplicate", file, 1, "catalog: duplicate trigger query '%s'", query)
	}
	seen[key] = true
	for _, pattern := range genericQueries {
		if pattern.MatchString(strings.TrimSpace(query)) {
			report.Fail("skill.evals-generic", file, 1, "catalog: generic trigger eval query must be replaced: '%s'", query)
			return
		}
	}
}

func checkTriggerMinimums(report *check.Reporter, file string, entries []triggerEntry) {
	var positive int
	for _, e := range entries {
		if e.positive {
			positive++
		}
	}
	negative := len(entries) - positive
	if positive < 8 || negative < 8 {
		report.Fail("skill.evals-minimum", file, 1, "catalog: trigger evals need at least 8 positive and 8 negative queries (found %d/%d)", positive, negative)
	}
	for _, polarity := range []bool{true, false} {
		for _, split := range []string{"train", "validation"} {
			var count int
			for _, e := range entries {
				if e.positive == polarity && e.split == split {
					count++
				}
			}
			if count < 2 {
				label := "negative"
				if polarity {
					label = "positive"
				}
				report.Fail("skill.evals-split-minimum", file, 1, "catalog: trigger evals need at least 2 %s %s queries", label, split)
			}
		}
	}
	// A generated near miss names the sibling that owns the query; drawing
	// them from one sibling would test one boundary only.
	owners := map[string]bool{}
	var generated int
	for _, e := range entries {
		if e.owner != "" {
			generated++
			owners[e.owner] = true
		}
	}
	required := generated
	if required > 3 {
		required = 3
	}
	if len(owners) < required {
		report.Fail("skill.evals-owners", file, 1, "catalog: generated negative routes need at least %d sibling owners (found %d)", required, len(owners))
	}
}
