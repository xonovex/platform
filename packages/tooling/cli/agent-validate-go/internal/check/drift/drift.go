// Package drift checks that a skill catalog does not grow past its word
// budgets and does not restate one invariant in many files.
package drift

import (
	"encoding/json"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/budget"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

// Name is the name of the check.
const Name = "drift"

// Options tune the check.
type Options struct {
	// BudgetFile is the manifest that maps a file to its word budget.
	BudgetFile string `json:"budgetFile"`
	// SkillCap is the word ceiling of a SKILL.md that has no budget entry.
	SkillCap int `json:"skillCap"`
	// ReferenceCap is the word ceiling of a reference file with no budget entry.
	ReferenceCap int `json:"referenceCap"`
	// Similarity is the bigram overlap at which two sentences state one rule.
	Similarity float64 `json:"similarity"`
	// MinFiles is the number of files that restate a rule before it is a finding.
	MinFiles int `json:"minFiles"`
}

func defaults() Options {
	return Options{BudgetFile: "budgets.json", SkillCap: 900, ReferenceCap: 650, Similarity: 0.75, MinFiles: 3}
}

const minSentenceWords = 6

var (
	normative = regexp.MustCompile(`(?i)\b(?:must|never|always|require[sd]?|shall|may not|cannot|do not|don't)\b`)
	markup    = regexp.MustCompile("[`*_\\[\\]()]")
	notWord   = regexp.MustCompile(`[^\da-z\s-]`)
	spaces    = regexp.MustCompile(`\s+`)
)

// Check is the drift check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string {
	return "file word budgets and invariants restated across skill files"
}

// DefaultPaths returns the catalog root.
func (Check) DefaultPaths() []string { return []string{"packages/runtime/plugin"} }

// Run measures every skill file below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	files, err := budget.Collect(ctx.Tree, ctx.Paths)
	if err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	budgets, ok := readBudgets(ctx, report, options.BudgetFile)
	if ok {
		checkBudgets(report, options, files, budgets)
	}
	checkDuplication(report, options, files)
	return report.Findings(), nil
}

// readBudgets reads the manifest. A missing manifest is an empty one; a broken
// manifest is a failure and stops the budget rule.
func readBudgets(ctx check.Context, report *check.Reporter, name string) (map[string]int, bool) {
	text, found, err := ctx.Tree.Optional(name)
	if err != nil {
		report.Fail("drift.budget-unreadable", name, 1, "budget manifest cannot be read: %v", err)
		return nil, false
	}
	if !found {
		return map[string]int{}, true
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		report.Fail("drift.budget-json", name, 1, "budget manifest is not valid JSON: %v", err)
		return nil, false
	}
	budgets := make(map[string]int, len(raw))
	for key, value := range raw {
		number, isNumber := value.(float64)
		if !isNumber || number < 1 || number != math.Trunc(number) {
			report.Fail("drift.budget-entry", name, 1, "budget manifest entries must map a path to a positive word budget")
			return nil, false
		}
		budgets[key] = int(number)
	}
	return budgets, true
}

// A file in the manifest may not exceed its recorded budget, and an unlisted
// file may not exceed the ceiling of its kind.
func checkBudgets(report *check.Reporter, options Options, files []budget.File, budgets map[string]int) {
	for _, f := range files {
		words := f.Words()
		if recorded, listed := budgets[f.Path]; listed {
			if words > recorded {
				report.Drift("drift.budget", f.Path, 1, "budget: %s is %d words, over its %d-word budget; bump the budget in the same change", f.Path, words, recorded)
			}
			continue
		}
		cap := options.ReferenceCap
		if f.Kind == budget.KindSkill {
			cap = options.SkillCap
		}
		if words > cap {
			report.Drift("drift.budget", f.Path, 1, "budget: %s is %d words, over the %d-word cap for a new %s file", f.Path, words, cap, f.Kind)
		}
	}
}

func normalizeWords(sentence string) []string {
	s := strings.ToLower(sentence)
	s = markup.ReplaceAllString(s, "")
	s = notWord.ReplaceAllString(s, " ")
	return strings.Fields(s)
}

// splitSentences breaks a text after sentence punctuation that precedes
// whitespace, and at each blank line.
func splitSentences(text string) []string {
	var out []string
	start := 0
	for i := 0; i < len(text); {
		if !isSpace(text[i]) {
			i++
			continue
		}
		j := i
		for j < len(text) && isSpace(text[j]) {
			j++
		}
		run := text[i:j]
		if (i > 0 && strings.IndexByte(".!?", text[i-1]) >= 0) || strings.Contains(run, "\n\n") {
			out = append(out, text[start:i])
			start = j
		}
		i = j
	}
	return append(out, text[start:])
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}

// normativeSentences returns the sentences that state a rule, which are the
// only ones worth de-duplicating.
func normativeSentences(text string) []string {
	var out []string
	for _, sentence := range splitSentences(markdown.MaskFences(text)) {
		sentence = strings.TrimSpace(spaces.ReplaceAllString(sentence, " "))
		if normative.MatchString(sentence) && len(normalizeWords(sentence)) >= minSentenceWords {
			out = append(out, sentence)
		}
	}
	return out
}

// bigrams tolerate the single-word substitutions that a restated invariant
// picks up; trigrams miss them while offering no better separation.
func bigrams(words []string) map[string]struct{} {
	set := map[string]struct{}{}
	size := 2
	if len(words) < size {
		size = len(words)
	}
	for i := 0; i+size <= len(words); i++ {
		set[strings.Join(words[i:i+size], " ")] = struct{}{}
	}
	return set
}

// jaccard is the share of bigrams two sentences hold in common. The ratio of
// the smaller set to the larger one bounds it from above, so two sets of very
// different size are not compared gram by gram.
func jaccard(a, b map[string]struct{}, floor float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	small, large := len(a), len(b)
	if small > large {
		small, large = large, small
	}
	if float64(small)/float64(large) < floor {
		return 0
	}
	shared := 0
	for gram := range a {
		if _, ok := b[gram]; ok {
			shared++
		}
	}
	return float64(shared) / float64(len(a)+len(b)-shared)
}

type cluster struct {
	paths map[string]bool
	grams map[string]struct{}
}

func checkDuplication(report *check.Reporter, options Options, files []budget.File) {
	var clusters []*cluster
	for _, f := range files {
		for _, sentence := range normativeSentences(f.Text) {
			grams := bigrams(normalizeWords(sentence))
			var found *cluster
			for _, c := range clusters {
				if jaccard(c.grams, grams, options.Similarity) >= options.Similarity {
					found = c
					break
				}
			}
			if found == nil {
				clusters = append(clusters, &cluster{paths: map[string]bool{f.Path: true}, grams: grams})
				continue
			}
			found.paths[f.Path] = true
		}
	}
	for _, c := range clusters {
		if len(c.paths) < options.MinFiles {
			continue
		}
		paths := make([]string, 0, len(c.paths))
		for p := range c.paths {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		report.Drift("drift.duplication", paths[0], 1,
			"duplication: an invariant is restated in %d files (%s); keep one normative source and cross-reference it", len(paths), strings.Join(paths, ", "))
	}
}
