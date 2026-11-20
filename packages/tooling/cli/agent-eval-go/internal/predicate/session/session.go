// Package session is the predicate leaf that checks what the agent did during a
// run: the tools it called, the files it read or wrote, the commands it ran and
// the messages it wrote.
package session

import (
	"context"
	"fmt"
	"regexp"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
)

// Evaluator checks session predicates against the events of a run.
type Evaluator struct{}

// New returns the leaf.
func New() Evaluator { return Evaluator{} }

// Kind is the predicate kind the leaf checks.
func (Evaluator) Kind() string { return model.KindSession }

// Evaluate counts the events of the predicate's kind whose text matches. With
// neither bound the predicate needs at least one match; with only a maximum the
// minimum is zero.
func (Evaluator) Evaluate(_ context.Context, predicates []model.Predicate, subject predicate.Subject) ([]predicate.Verdict, model.Usage, error) {
	verdicts := make([]predicate.Verdict, len(predicates))
	for i, p := range predicates {
		re, err := regexp.Compile(p.Match)
		if err != nil {
			verdicts[i] = predicate.Verdict{Detail: "invalid pattern: " + err.Error()}
			continue
		}
		count := 0
		for _, event := range subject.Obs.Events {
			if event.Kind == p.Event && re.MatchString(text(event)) {
				count++
			}
		}
		verdicts[i] = bounds(p, count)
	}
	return verdicts, model.Usage{}, nil
}

// text is what a pattern matches: the tool name of a tool call, else the text.
func text(event harness.Event) string {
	if event.Kind == harness.EventToolCall {
		return event.Name
	}
	return event.Text
}

func bounds(p model.Predicate, count int) predicate.Verdict {
	minimum := 1
	switch {
	case p.Min != nil:
		minimum = *p.Min
	case p.Max != nil:
		minimum = 0
	}
	if count < minimum {
		return predicate.Verdict{Detail: fmt.Sprintf("%d %s events match %q, at least %d needed", count, p.Event, p.Match, minimum)}
	}
	if p.Max != nil && count > *p.Max {
		return predicate.Verdict{Detail: fmt.Sprintf("%d %s events match %q, at most %d allowed", count, p.Event, p.Match, *p.Max)}
	}
	return predicate.Verdict{Passed: true}
}
