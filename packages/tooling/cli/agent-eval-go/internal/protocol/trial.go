package protocol

import (
	"context"
	"time"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

// TriggerRequest builds the request that asks which skill a prompt makes the
// model load. The caps are the trigger caps: a short run that ends at the first
// answer.
func TriggerRequest(s settings.Settings, prompt, target string, offer []harness.Offer) harness.Request {
	return harness.Request{
		Role: harness.RoleTrigger, Harness: s.Harness, Model: s.Model, Effort: s.Effort,
		Prompt: prompt, Target: target, Offer: offer,
		Caps: harness.Caps{
			Timeout:        time.Duration(RunSeconds(s.TimeoutSeconds, settings.TriggerTimeoutSeconds)) * time.Second,
			BudgetUSD:      settings.TriggerRunBudgetUSD,
			MaxOutputChars: 2000,
		},
	}
}

// Trial is one request run several times and checked against predicates.
type Trial struct {
	Request    harness.Request
	Runs       int
	Predicates []model.Predicate
	// Task is the prompt the judge sees.
	Task  string
	Judge judge.Spec
}

// TrialResult is what a trial showed.
type TrialResult struct {
	Observations []harness.Observation
	Predicates   []model.PredicateResult
	Usage        model.Usage
	DurationMs   int64
}

// RunTrial runs the request the given number of times, evaluates the predicates
// on each run and folds the verdicts. An infrastructure fault is an error.
func RunTrial(ctx context.Context, env Env, t Trial) (TrialResult, error) {
	var out TrialResult
	verdicts := make([][]predicate.Verdict, 0, t.Runs)
	for i := 0; i < t.Runs; i++ {
		obs, err := env.Exec.Execute(ctx, t.Request)
		if err != nil {
			return out, err
		}
		out.Observations = append(out.Observations, obs)
		out.Usage = out.Usage.Add(obs.Usage)
		out.DurationMs += obs.DurationMs
		got, used, err := env.Predicates.Evaluate(ctx, t.Predicates, predicate.Subject{Task: t.Task, Obs: obs, Judge: t.Judge})
		out.Usage = out.Usage.Add(used)
		if err != nil {
			return out, err
		}
		verdicts = append(verdicts, got)
	}
	out.Predicates = predicate.Aggregate(t.Predicates, verdicts)
	return out, nil
}

// HasKind reports whether any predicate has the kind.
func HasKind(predicates []model.Predicate, kind string) bool {
	for _, p := range predicates {
		if p.Kind == kind {
			return true
		}
	}
	return false
}

// OutcomeOf is passed when every predicate passed.
func OutcomeOf(results []model.PredicateResult) model.Outcome {
	if predicate.AllPassed(results) {
		return model.Passed
	}
	return model.Failed
}
