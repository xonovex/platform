// Package llm is the predicate leaf that asks a model to grade the answer of a
// run against natural-language criteria. It grades every criterion of a case in
// one judge call.
package llm

import (
	"context"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
)

// Evaluator grades model predicates through a judge.
type Evaluator struct {
	Grader judge.Grader
}

// New returns the leaf.
func New(grader judge.Grader) Evaluator { return Evaluator{Grader: grader} }

// Kind is the predicate kind the leaf checks.
func (Evaluator) Kind() string { return model.KindLLM }

// Evaluate grades the final answer of the run against each criterion.
func (e Evaluator) Evaluate(ctx context.Context, predicates []model.Predicate, subject predicate.Subject) ([]predicate.Verdict, model.Usage, error) {
	criteria := make([]string, len(predicates))
	for i, p := range predicates {
		criteria[i] = p.Criterion
	}
	graded, usage, err := e.Grader.Grade(ctx, subject.Judge, subject.Task, subject.Obs.Text, criteria)
	if err != nil {
		return nil, usage, err
	}
	verdicts := make([]predicate.Verdict, len(graded))
	for i, g := range graded {
		verdicts[i] = predicate.Verdict{Passed: g.Passed, Detail: g.Evidence}
	}
	return verdicts, usage, nil
}
