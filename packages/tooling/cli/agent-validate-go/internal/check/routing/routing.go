// Package routing checks the routing declarations of a skill catalog: each
// skill owns a validation scenario, no query has two owners, and a rationale
// cites only operations that a skill documents.
package routing

import (
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
)

// Name is the name of the check.
const Name = "routing"

var operationMention = regexp.MustCompile("`([a-z0-9]+(?:-[a-z0-9]+)*)`\\s+operation")

// Check is the routing check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string {
	return "each skill owns a validation routing scenario, no query has two owners, and cited operations exist"
}

// DefaultPaths returns the catalog root.
func (Check) DefaultPaths() []string { return []string{"packages/runtime/plugin"} }

type candidate struct {
	name    string
	file    string
	queries []catalog.Query
	refs    []string
}

// Run reads every skill that declares queries.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	if err := check.DecodeOptions(ctx.Options, &struct{}{}); err != nil {
		return nil, err
	}
	dirs, err := catalog.SkillDirs(ctx.Tree, ctx.Paths)
	if err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	var candidates []candidate
	for _, dir := range dirs {
		c, ok, err := read(ctx, report, dir)
		if err != nil {
			return nil, err
		}
		if ok {
			candidates = append(candidates, c)
		}
	}
	checkOwners(report, candidates)
	checkConflicts(report, candidates)
	checkOperations(report, candidates)
	return report.Findings(), nil
}

func read(ctx check.Context, report *check.Reporter, dir string) (candidate, bool, error) {
	queries, exists, err := catalog.ReadQueries(ctx.Tree, dir)
	queriesFile := path.Join(dir, catalog.TriggerFile)
	if !exists {
		return candidate{}, false, nil
	}
	if err != nil {
		report.Fail("routing.queries-invalid", queriesFile, 1, "%v", err)
		return candidate{}, false, nil
	}
	skillFile := path.Join(dir, catalog.SkillFile)
	text, err := ctx.Tree.Read(skillFile)
	if err != nil {
		return candidate{}, false, err
	}
	doc, err := document.Parse(text)
	name, _ := doc.String("name")
	if err != nil || name == "" {
		report.Fail("routing.name-missing", skillFile, 1, "routing: skill name missing from %s", skillFile)
		return candidate{}, false, nil
	}
	entries, err := ctx.Tree.List(path.Join(dir, "references"))
	if err != nil {
		return candidate{}, false, err
	}
	var refs []string
	for _, entry := range entries {
		if !entry.IsDir && strings.HasSuffix(entry.Name, ".md") {
			refs = append(refs, strings.TrimSuffix(entry.Name, ".md"))
		}
	}
	return candidate{name: name, file: queriesFile, queries: queries, refs: refs}, true, nil
}

// A query has an owner when exactly one skill marks it should_trigger and at
// least one other skill carries it as a negative. Deleting a skill can silently
// strip another skill of its only pairing, which this finds.
func checkOwners(report *check.Reporter, candidates []candidate) {
	type tally struct {
		positives []candidate
		negatives int
		split     string
	}
	byQuery := map[string]*tally{}
	for _, c := range candidates {
		for _, q := range c.queries {
			t := byQuery[q.Text]
			if t == nil {
				t = &tally{}
				byQuery[q.Text] = t
			}
			if q.ShouldTrigger {
				t.positives = append(t.positives, c)
				t.split = q.Split
			} else {
				t.negatives++
			}
		}
	}
	owners := map[string]bool{}
	for _, t := range byQuery {
		if len(t.positives) != 1 || t.negatives == 0 {
			continue
		}
		if t.split == "validation" {
			owners[t.positives[0].name] = true
		}
	}
	for _, c := range candidates {
		if !owners[c.name] {
			report.Fail("routing.no-owner", c.file, 1,
				"routing: %s owns no validation-split routing scenario; another skill must carry one of its validation positives as a negative", c.name)
		}
	}
}

// normalize folds the differences that make a copied query look distinct: case,
// run-together whitespace and closing punctuation.
func normalize(query string) string {
	joined := strings.Join(strings.Fields(strings.ToLower(query)), " ")
	return strings.TrimRight(joined, ".?!")
}

func checkConflicts(report *check.Reporter, candidates []candidate) {
	type claim struct {
		query  string
		owners map[string]bool
		file   string
	}
	claims := map[string]*claim{}
	for _, c := range candidates {
		for _, q := range c.queries {
			if !q.ShouldTrigger {
				continue
			}
			key := normalize(q.Text)
			entry := claims[key]
			if entry == nil {
				entry = &claim{query: q.Text, owners: map[string]bool{}, file: c.file}
				claims[key] = entry
			}
			entry.owners[c.name] = true
		}
	}
	var conflicts []*claim
	for _, entry := range claims {
		if len(entry.owners) > 1 {
			conflicts = append(conflicts, entry)
		}
	}
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].query < conflicts[j].query })
	for _, entry := range conflicts {
		names := make([]string, 0, len(entry.owners))
		for name := range entry.owners {
			names = append(names, name)
		}
		sort.Strings(names)
		report.Fail("routing.conflict", entry.file, 1,
			"routing: %s both mark \"%s\" should_trigger; exactly one skill owns a query and the others carry it as a negative", strings.Join(names, " and "), entry.query)
	}
}

// Only a token in backticks cites an operation by name, so a plain English
// compound is not an identifier.
func checkOperations(report *check.Reporter, candidates []candidate) {
	known := map[string]bool{}
	for _, c := range candidates {
		for _, ref := range c.refs {
			known[ref] = true
		}
	}
	for _, c := range candidates {
		for _, q := range c.queries {
			for _, m := range operationMention.FindAllStringSubmatch(q.Rationale, -1) {
				if !known[m[1]] {
					report.Fail("routing.operation-unresolved", c.file, 1,
						"routing: %s cites the '%s' operation, which no skill owns a reference file for; name the operation that exists or drop it", c.name, m[1])
				}
			}
		}
	}
}
