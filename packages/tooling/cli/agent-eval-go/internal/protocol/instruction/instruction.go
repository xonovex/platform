// Package instruction is the protocol that measures whether one rule of an
// instruction file changes what an agent does. A case runs the prompt with the
// rule present, removed (absent) or with a phrase of it removed (ablated), and
// checks what the agent did. The rule earns its place when the behavior it asks
// for appears with it and not without it.
package instruction

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/ablation"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

// Protocol is the instruction protocol.
type Protocol struct{}

// New returns the protocol.
func New() Protocol { return Protocol{} }

// Name is the protocol name.
func (Protocol) Name() model.Protocol { return model.Instruction }

// Plan makes a unit of each case and each condition the case declares. The
// target is the instruction eval document.
func (Protocol) Plan(in protocol.Input) (protocol.Plan, error) {
	doc, err := in.Env.Reader.Instruction(in.Target)
	if err != nil {
		return protocol.Plan{}, err
	}
	instructionFile := path.Join(path.Dir(in.Target), doc.Instruction)
	data, err := in.Env.FS.ReadFile(instructionFile)
	if err != nil {
		return protocol.Plan{}, fmt.Errorf("read instruction file %s: %w", instructionFile, err)
	}
	s := in.Settings
	plan := protocol.Plan{Subject: path.Base(instructionFile), Inputs: []string{string(data)}}
	for _, c := range doc.Cases {
		plan.Inputs = append(plan.Inputs, c.ID, c.Text, fmt.Sprint(c.Anchor, c.Expect, c.Ablate))
		for _, condition := range c.Conditions() {
			text, err := ablation.Apply(string(data), c.Anchor, c.Ablate, condition)
			if err != nil {
				return protocol.Plan{}, fmt.Errorf("case %s (%s): %w", c.ID, condition, err)
			}
			c, condition := c, condition
			plan.Units = append(plan.Units, protocol.Unit{
				ID: c.ID, Condition: condition, ModelRuns: s.Runs * calls(c, condition),
				Run: func(ctx context.Context) (model.CaseResult, error) {
					return runCase(ctx, in, c, condition, text)
				},
			})
		}
	}
	plan.Inputs = append(plan.Inputs, fmt.Sprint(s.Runs, s.MinimumLift, s.Harness, s.Model, s.JudgeModel, s.Effort))
	plan.Decide = func(cases []model.CaseResult) protocol.Verdict { return decide(cases, s.MinimumLift) }
	return plan, nil
}

// predicates lists what a unit evaluates: its own condition's expectations, then
// the present expectations when the condition is not present, to score the
// control.
func predicates(c model.InstructionCase, condition model.Condition) (own, control []model.Predicate) {
	own = c.Expect.For(condition)
	if condition != model.Present {
		control = c.Expect.Present
	}
	return own, control
}

func calls(c model.InstructionCase, condition model.Condition) int {
	own, control := predicates(c, condition)
	if protocol.HasKind(append(append([]model.Predicate{}, own...), control...), model.KindLLM) {
		return 2
	}
	return 1
}

func runCase(ctx context.Context, in protocol.Input, c model.InstructionCase, condition model.Condition, text string) (model.CaseResult, error) {
	s := in.Settings
	own, control := predicates(c, condition)
	all := append(append([]model.Predicate{}, own...), control...)
	trial, err := protocol.RunTrial(ctx, in.Env, protocol.Trial{
		Request: harness.Request{
			Role: harness.RoleAgent, Harness: s.Harness, Model: s.Model, Effort: s.Effort,
			Prompt: c.Text, Instructions: text,
			Caps: harness.Caps{
				Timeout:   time.Duration(protocol.RunSeconds(s.TimeoutSeconds, settings.AgentTimeoutSeconds)) * time.Second,
				BudgetUSD: settings.AgentRunBudgetUSD, MaxTurns: 12,
			},
		},
		Runs: s.Runs, Predicates: all, Task: c.Text,
		Judge: judge.Spec{Harness: s.Harness, Model: s.JudgeModel, BudgetUSD: settings.JudgeRunBudgetUSD},
	})
	if err != nil {
		return model.CaseResult{}, err
	}
	ownResults := trial.Predicates[:len(own)]
	score := predicate.Score(ownResults)
	result := model.CaseResult{
		ID: c.ID, Condition: condition, Outcome: protocol.OutcomeOf(ownResults), Score: &score, Runs: s.Runs,
		DurationMs: trial.DurationMs, Usage: protocol.UsagePtr(trial.Usage), Predicates: ownResults,
	}
	if len(control) > 0 {
		controlScore := predicate.Score(trial.Predicates[len(own):])
		result.ControlScore = &controlScore
	}
	return result, nil
}

func decide(cases []model.CaseResult, minimumLift float64) protocol.Verdict {
	present := map[string]float64{}
	for _, c := range cases {
		if c.Condition == model.Present && c.Score != nil {
			present[c.ID] = *c.Score
		}
	}
	sum, n := 0.0, 0
	for _, c := range cases {
		if p, ok := present[c.ID]; ok && c.Condition != model.Present && c.ControlScore != nil {
			sum += p - *c.ControlScore
			n++
		}
	}
	failures := protocol.Failures(cases)
	pass := len(failures) == 0
	lines := []string{}
	if n > 0 {
		lift := sum / float64(n)
		lines = append(lines, fmt.Sprintf("mean lift: %+.2f over %d comparisons (minimum %+.2f)", lift, n, minimumLift))
		pass = pass && lift >= minimumLift
	} else {
		lines = append(lines, "no case compares the rule present with the rule removed")
	}
	return protocol.Verdict{Pass: pass, Lines: append(lines, failures...)}
}
