// Package judge grades a response against natural-language assertions with a
// model. It is a thin composition over the harness executor: it builds the
// grading prompt, runs a judge request and reads the strict JSON answer. An
// answer that cannot be read fails every assertion, so an unreliable judge
// never turns into a pass.
package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// UnparseableEvidence is the evidence of every assertion when the judge answer
// cannot be read.
const UnparseableEvidence = "unparseable judge output"

// Spec selects the harness and model that grade, and an optional rubric that
// replaces the default grading rules.
type Spec struct {
	Harness string
	Model   string
	Rubric  string
	// BudgetUSD caps the spend of one grading run.
	BudgetUSD float64
}

// Verdict is the grade of one assertion.
type Verdict struct {
	Text     string
	Passed   bool
	Evidence string
}

// Grader grades responses through a harness executor.
type Grader struct {
	Executor harness.Executor
}

const defaultRubric = "Grade each assertion as a binary pass or fail. " +
	"Pass an assertion only when the response itself shows it holds. " +
	"Cite the evidence in the response for each grade. Do not give partial credit."

// Prompt builds the grading prompt.
func Prompt(rubric, task, response string, assertions []string) string {
	if strings.TrimSpace(rubric) == "" {
		rubric = defaultRubric
	}
	var b strings.Builder
	b.WriteString(rubric)
	b.WriteString("\n\n## Task given to the assistant\n\n")
	b.WriteString(task)
	b.WriteString("\n\n## Response to grade\n\n")
	b.WriteString(response)
	b.WriteString("\n\n## Assertions\n\n")
	for i, assertion := range assertions {
		fmt.Fprintf(&b, "%d. %s\n", i+1, assertion)
	}
	b.WriteString("\nReturn only JSON: {\"assertion_results\":[{\"text\":\"<assertion>\",\"passed\":true|false,\"evidence\":\"<evidence>\"}]} " +
		"with one entry per assertion, in order.\n")
	return b.String()
}

// Grade grades the response against each assertion. It returns an error only
// for an infrastructure fault of the judge run.
func (g Grader) Grade(ctx context.Context, spec Spec, task, response string, assertions []string) ([]Verdict, model.Usage, error) {
	if len(assertions) == 0 {
		return nil, model.Usage{}, nil
	}
	obs, err := g.Executor.Execute(ctx, harness.Request{
		Role: harness.RoleJudge, Harness: spec.Harness, Model: spec.Model,
		Prompt: Prompt(spec.Rubric, task, response, assertions), Assertions: len(assertions),
		Caps: harness.Caps{BudgetUSD: spec.BudgetUSD},
	})
	if err != nil {
		return nil, model.Usage{}, err
	}
	return Parse(obs, assertions), obs.Usage, nil
}

type answer struct {
	Results []struct {
		Text     string `json:"text"`
		Passed   *bool  `json:"passed"`
		Evidence string `json:"evidence"`
	} `json:"assertion_results"`
}

// Parse reads a judge observation. It fails every assertion unless the answer
// holds exactly one verdict for each.
func Parse(obs harness.Observation, assertions []string) []Verdict {
	var parsed answer
	for _, candidate := range []string{obs.Structured, obs.Text} {
		if object, ok := ExtractJSON(candidate); ok && json.Unmarshal([]byte(object), &parsed) == nil && len(parsed.Results) > 0 {
			break
		}
		parsed = answer{}
	}
	verdicts := make([]Verdict, len(assertions))
	valid := len(parsed.Results) == len(assertions)
	for i := range parsed.Results {
		if parsed.Results[i].Passed == nil {
			valid = false
		}
	}
	for i, text := range assertions {
		if !valid {
			verdicts[i] = Verdict{Text: text, Evidence: UnparseableEvidence}
			continue
		}
		result := parsed.Results[i]
		verdicts[i] = Verdict{Text: text, Passed: *result.Passed, Evidence: result.Evidence}
	}
	return verdicts
}

// ExtractJSON finds the JSON object in text: the body of a fenced block, else
// the span from the first brace to the last.
func ExtractJSON(text string) (string, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", false
	}
	if open := strings.Index(text, "```"); open >= 0 {
		rest := text[open+3:]
		rest = strings.TrimPrefix(rest, "json")
		if end := strings.Index(rest, "```"); end >= 0 {
			if body := strings.TrimSpace(rest[:end]); strings.HasPrefix(body, "{") {
				return body, true
			}
		}
	}
	start, end := strings.Index(text, "{"), strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return "", false
	}
	return text[start : end+1], true
}
