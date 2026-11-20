// Package predicate is the port for checks on what a run showed. A leaf checks
// one kind of predicate (the files a run left, the session it ran, or a model's
// judgement) and the registry dispatches a case's predicates to the leaves, so
// a protocol evaluates any mix of kinds without naming one.
package predicate

import (
	"context"
	"fmt"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// Subject is what one run showed, and the judge that grades it.
type Subject struct {
	// Task is the prompt the run answered.
	Task string
	// Obs is the observation of the run.
	Obs harness.Observation
	// Judge selects the model that grades a model predicate.
	Judge judge.Spec
}

// Verdict is the result of one predicate on one run.
type Verdict struct {
	Passed bool
	Detail string
}

// Evaluator checks the predicates of one kind. It returns one verdict for each
// predicate, in order, so a model leaf can grade a whole batch in one call.
type Evaluator interface {
	Kind() string
	Evaluate(ctx context.Context, predicates []model.Predicate, subject Subject) ([]Verdict, model.Usage, error)
}

// Registry holds the evaluators by kind.
type Registry struct {
	byKind map[string]Evaluator
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{byKind: map[string]Evaluator{}} }

// Register adds an evaluator. A second evaluator of one kind panics, because
// two leaves claiming one kind is a programming error.
func (r *Registry) Register(e Evaluator) {
	if _, dup := r.byKind[e.Kind()]; dup {
		panic(fmt.Sprintf("predicate kind %q registered twice", e.Kind()))
	}
	r.byKind[e.Kind()] = e
}

// Evaluate checks every predicate against one run and returns the verdicts in
// the order of the predicates.
func (r *Registry) Evaluate(ctx context.Context, predicates []model.Predicate, subject Subject) ([]Verdict, model.Usage, error) {
	verdicts := make([]Verdict, len(predicates))
	var usage model.Usage
	groups := map[string][]int{}
	var order []string
	for i, p := range predicates {
		if _, ok := r.byKind[p.Kind]; !ok {
			return nil, usage, fmt.Errorf("no evaluator for predicate kind %q", p.Kind)
		}
		if _, seen := groups[p.Kind]; !seen {
			order = append(order, p.Kind)
		}
		groups[p.Kind] = append(groups[p.Kind], i)
	}
	for _, kind := range order {
		indexes := groups[kind]
		batch := make([]model.Predicate, len(indexes))
		for j, i := range indexes {
			batch[j] = predicates[i]
		}
		got, used, err := r.byKind[kind].Evaluate(ctx, batch, subject)
		usage = usage.Add(used)
		if err != nil {
			return nil, usage, err
		}
		if len(got) != len(batch) {
			return nil, usage, fmt.Errorf("%s evaluator returned %d verdicts for %d predicates", kind, len(got), len(batch))
		}
		for j, i := range indexes {
			verdicts[i] = got[j]
		}
	}
	return verdicts, usage, nil
}

// Aggregate folds the verdicts of several runs into one result for each
// predicate. A predicate passes when the share of runs that passed it reaches
// its threshold; a predicate without a threshold needs every run.
func Aggregate(predicates []model.Predicate, runs [][]Verdict) []model.PredicateResult {
	results := make([]model.PredicateResult, len(predicates))
	for i, p := range predicates {
		passed, detail := 0, ""
		for _, run := range runs {
			if run[i].Passed {
				passed++
			} else if detail == "" {
				detail = run[i].Detail
			}
		}
		threshold := 1.0
		if p.Threshold != nil {
			threshold = *p.Threshold
		}
		rate := 0.0
		if len(runs) > 0 {
			rate = float64(passed) / float64(len(runs))
		}
		ok := len(runs) > 0 && rate >= threshold
		if ok {
			detail = ""
		} else if detail == "" {
			detail = "no run passed"
		}
		if len(runs) > 1 {
			detail = fmt.Sprintf("%d/%d runs passed", passed, len(runs)) + suffix(detail)
		}
		results[i] = model.PredicateResult{Kind: p.Kind, Passed: ok, Detail: detail}
	}
	return results
}

func suffix(detail string) string {
	if detail == "" {
		return ""
	}
	return ": " + detail
}

// Score is the share of results that passed.
func Score(results []model.PredicateResult) float64 {
	if len(results) == 0 {
		return 1
	}
	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}
	return float64(passed) / float64(len(results))
}

// AllPassed reports whether every result passed.
func AllPassed(results []model.PredicateResult) bool { return Score(results) == 1 }
