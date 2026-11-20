// Package orchestration is the protocol that measures a flow in which several
// agents each run a step after the steps they depend on. A step sees the answers
// and the files of the steps it waits on, and its expectations are checked on its
// own run. The flow budget bounds the model runs, the cost and the time of one
// execution.
package orchestration

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

// Protocol is the orchestration protocol.
type Protocol struct{}

// New returns the protocol.
func New() Protocol { return Protocol{} }

// Name is the protocol name.
func (Protocol) Name() model.Protocol { return model.Orchestration }

// Plan makes one unit for each execution of the flow. The target is the
// orchestration document.
func (Protocol) Plan(in protocol.Input) (protocol.Plan, error) {
	doc, err := in.Env.Reader.Orchestration(in.Target)
	if err != nil {
		return protocol.Plan{}, err
	}
	steps, err := document.Order(doc.Steps)
	if err != nil {
		return protocol.Plan{}, err
	}
	s := in.Settings
	plan := protocol.Plan{Subject: doc.ID}
	for _, step := range steps {
		plan.Inputs = append(plan.Inputs, step.ID, step.Text, step.Agent, fmt.Sprint(step.After, step.Expect))
	}
	for _, agent := range doc.Agents {
		plan.Inputs = append(plan.Inputs, fmt.Sprint(agent))
	}
	plan.Inputs = append(plan.Inputs, fmt.Sprint(doc.Budget, s.Runs, s.Harness, s.Model, s.JudgeModel, s.Effort))
	runs := 0
	for _, step := range steps {
		runs += calls(step)
	}
	for k := 1; k <= s.Runs; k++ {
		id := doc.ID
		if s.Runs > 1 {
			id = fmt.Sprintf("%s-run%d", doc.ID, k)
		}
		plan.Units = append(plan.Units, protocol.Unit{
			ID: id, ModelRuns: runs,
			Run: func(ctx context.Context) (model.CaseResult, error) { return execute(ctx, in, doc, steps, id) },
		})
	}
	plan.Decide = func(cases []model.CaseResult) protocol.Verdict {
		lines := protocol.Failures(cases)
		return protocol.Verdict{
			Pass:  len(lines) == 0,
			Lines: append([]string{fmt.Sprintf("%d of %d flow executions passed", len(cases)-len(lines), len(cases))}, lines...),
		}
	}
	return plan, nil
}

// calls is how many model runs a step may start: its own and a judge call when it
// has a model expectation.
func calls(step model.Step) int {
	if protocol.HasKind(step.Expect, model.KindLLM) {
		return 2
	}
	return 1
}

func execute(ctx context.Context, in protocol.Input, doc model.OrchestrationEval, steps []model.Step, id string) (model.CaseResult, error) {
	flowCtx := ctx
	limited := doc.Budget != nil && doc.Budget.TimeoutSeconds > 0
	if limited {
		var cancel context.CancelFunc
		flowCtx, cancel = context.WithTimeout(ctx, time.Duration(doc.Budget.TimeoutSeconds)*time.Second)
		defer cancel()
	}
	agents := map[string]model.Agent{}
	for _, a := range doc.Agents {
		agents[a.ID] = a
	}
	outputs := map[string]harness.Observation{}
	result := model.CaseResult{ID: id, Runs: 1}
	var usage model.Usage
	used := 0
	var failure string
	for _, step := range steps {
		if limit := budgetFailure(doc.Budget, used+calls(step), usage); limit != "" {
			failure = limit
			break
		}
		agent := agents[step.Agent]
		obs, verdicts, spent, err := runStep(flowCtx, in, agent, step, outputs)
		usage = usage.Add(spent)
		if err != nil {
			if ctx.Err() == nil && limited && errors.Is(err, context.DeadlineExceeded) {
				failure = "flow timed out"
				break
			}
			return model.CaseResult{}, err
		}
		used += calls(step)
		outputs[step.ID] = obs
		result.DurationMs += obs.DurationMs
		for i, p := range predicate.Aggregate(step.Expect, [][]predicate.Verdict{verdicts}) {
			p.Detail = strings.TrimSpace("step " + step.ID + ": " + p.Detail)
			if i < len(verdicts) && verdicts[i].Passed {
				p.Detail = ""
			}
			result.Predicates = append(result.Predicates, p)
		}
	}
	score := predicate.Score(result.Predicates)
	result.Score, result.Usage = &score, protocol.UsagePtr(usage)
	result.Outcome = protocol.OutcomeOf(result.Predicates)
	if failure != "" {
		result.Outcome, result.Error = model.Failed, failure
	}
	return result, nil
}

// budgetFailure reports the flow budget that the next step would break.
func budgetFailure(b *model.FlowBudget, runs int, usage model.Usage) string {
	if b == nil {
		return ""
	}
	if b.MaxModelRuns > 0 && runs > b.MaxModelRuns {
		return fmt.Sprintf("flow budget exceeded: %d model runs of %d", runs, b.MaxModelRuns)
	}
	if b.MaxCostUSD > 0 && usage.CostUSD >= b.MaxCostUSD {
		return fmt.Sprintf("flow budget exceeded: $%.4f of $%.2f", usage.CostUSD, b.MaxCostUSD)
	}
	return ""
}

func runStep(ctx context.Context, in protocol.Input, agent model.Agent, step model.Step, outputs map[string]harness.Observation) (harness.Observation, []predicate.Verdict, model.Usage, error) {
	s := in.Settings
	modelName := agent.Model
	if modelName == "" && agent.Harness == s.Harness {
		modelName = s.Model
	}
	prompt, files := compose(step, outputs)
	req := harness.Request{
		Role: harness.RoleAgent, Harness: agent.Harness, Model: modelName, Effort: s.Effort,
		Prompt: prompt, Fixtures: files,
		Caps: harness.Caps{
			Timeout:   time.Duration(protocol.RunSeconds(s.TimeoutSeconds, settings.AgentTimeoutSeconds)) * time.Second,
			BudgetUSD: settings.AgentRunBudgetUSD, MaxTurns: 12,
		},
	}
	var usage model.Usage
	obs, err := in.Env.Exec.Execute(ctx, req)
	if err != nil {
		return obs, nil, usage, err
	}
	usage = obs.Usage
	verdicts, used, err := in.Env.Predicates.Evaluate(ctx, step.Expect, predicate.Subject{
		Task: step.Text, Obs: obs,
		Judge: judge.Spec{Harness: s.Harness, Model: s.JudgeModel, BudgetUSD: settings.JudgeRunBudgetUSD},
	})
	return obs, verdicts, usage.Add(used), err
}

// compose builds the prompt of a step: its own text, then the answer of each step
// it waits on. The files those steps left become the fixtures of this step.
func compose(step model.Step, outputs map[string]harness.Observation) (string, []harness.File) {
	var b strings.Builder
	b.WriteString(step.Text)
	var files []harness.File
	seen := map[string]int{}
	for _, dep := range step.After {
		obs, ok := outputs[dep]
		if !ok {
			continue
		}
		fmt.Fprintf(&b, "\n\n## Result of step %s\n\n%s", dep, obs.Text)
		for path, content := range obs.Files {
			if i, dup := seen[path]; dup {
				files[i].Content = content
				continue
			}
			seen[path] = len(files)
			files = append(files, harness.File{Path: path, Content: content})
		}
	}
	sortFiles(files)
	return b.String(), files
}
