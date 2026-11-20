// Package output is the protocol that measures whether a skill improves the
// answer to a prompt. Each case runs with the skill and without it, a judge
// grades both answers against the case's expectations, and a tier gate compares
// the two pass rates over the required cases.
package output

import (
	"context"
	"fmt"
	"time"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

// Policy is the quality gate of a tier: the lowest pass rate with the skill and
// the smallest gain over the run without it.
type Policy struct {
	MinimumWithSkill float64
	MinimumLift      float64
}

// Policies maps each tier to its gate.
var Policies = map[model.Tier]Policy{
	model.Aggressive:   {MinimumWithSkill: 0.75, MinimumLift: 0.05},
	model.Moderate:     {MinimumWithSkill: 0.80, MinimumLift: 0.05},
	model.Conservative: {MinimumWithSkill: 0.90, MinimumLift: 0.10},
}

// Protocol is the output protocol.
type Protocol struct{}

// New returns the protocol.
func New() Protocol { return Protocol{} }

// Name is the protocol name.
func (Protocol) Name() model.Protocol { return model.Output }

// Plan makes a unit of each case and arm.
func (Protocol) Plan(in protocol.Input) (protocol.Plan, error) {
	skill, err := in.Env.Catalog.Skill(in.Target)
	if err != nil {
		return protocol.Plan{}, err
	}
	if skill.OutputFile == "" {
		return protocol.Plan{}, fmt.Errorf("skill %s has no %s", skill.Name, model.OutputFile)
	}
	doc, err := in.Env.Reader.Output(skill.OutputFile)
	if err != nil {
		return protocol.Plan{}, err
	}
	tier := doc.Tier
	if tier == "" {
		tier = model.Moderate
	}
	policy, ok := Policies[tier]
	if !ok {
		return protocol.Plan{}, fmt.Errorf("unknown tier %q", tier)
	}
	s := in.Settings
	spec := judgeSpec(doc, s)
	offer := protocol.Offers(skill)
	plan := protocol.Plan{Subject: skill.Name}
	gates := map[string]model.Gate{}
	for _, c := range doc.Cases {
		gates[c.ID] = c.Gate
		calls := 1
		if protocol.HasKind(c.Expect, model.KindLLM) {
			calls = 2
		}
		plan.Inputs = append(plan.Inputs, c.ID, c.Text, string(c.Gate), fmt.Sprint(c.Expect), fmt.Sprint(c.Fixtures))
		for _, arm := range []model.Condition{model.Present, model.Absent} {
			c, arm := c, arm
			plan.Units = append(plan.Units, protocol.Unit{
				ID: c.ID, Condition: arm, ModelRuns: s.Runs * calls,
				Run: func(ctx context.Context) (model.CaseResult, error) {
					return runCase(ctx, in, skill.Name, offer, spec, c, arm)
				},
			})
		}
	}
	plan.Inputs = append(plan.Inputs, string(tier), fmt.Sprint(s.Runs, s.Harness, s.Model, s.JudgeModel, s.Effort, spec))
	plan.Decide = func(cases []model.CaseResult) protocol.Verdict { return decide(cases, gates, policy, s.MinimumWinRate) }
	return plan, nil
}

func judgeSpec(doc model.OutputEval, s settings.Settings) judge.Spec {
	spec := judge.Spec{Harness: s.Harness, Model: s.JudgeModel, BudgetUSD: settings.JudgeRunBudgetUSD}
	if doc.Judge != nil {
		spec.Harness = doc.Judge.Harness
		if doc.Judge.Model != "" {
			spec.Model = doc.Judge.Model
		}
		spec.Rubric = doc.Judge.Rubric
	}
	return spec
}

func runCase(ctx context.Context, in protocol.Input, skill string, offer []harness.Offer, spec judge.Spec, c model.OutputCase, arm model.Condition) (model.CaseResult, error) {
	s := in.Settings
	files := make([]harness.File, len(c.Fixtures))
	prompt := c.Text
	for i, f := range c.Fixtures {
		files[i] = harness.File{Path: f.Path, Content: f.Text}
		if i == 0 {
			prompt += "\n\nRelevant input files (read them as needed):"
		}
		prompt += "\n- " + f.Path
	}
	trial, err := protocol.RunTrial(ctx, in.Env, protocol.Trial{
		Request: harness.Request{
			Role: harness.RoleGenerate, Harness: s.Harness, Model: s.Model, Effort: s.Effort,
			Prompt: prompt, Target: skill, Offer: offer, WithSkill: arm == model.Present, Fixtures: files,
			Caps: harness.Caps{
				Timeout:   time.Duration(protocol.RunSeconds(s.TimeoutSeconds, settings.GenerateTimeoutSeconds)) * time.Second,
				BudgetUSD: settings.GenerateRunBudgetUSD, MaxTurns: 12,
			},
		},
		Runs: s.Runs, Predicates: c.Expect, Task: c.Text, Judge: spec,
	})
	if err != nil {
		return model.CaseResult{}, err
	}
	score := predicate.Score(trial.Predicates)
	result := model.CaseResult{
		ID: c.ID, Condition: arm, Outcome: protocol.OutcomeOf(trial.Predicates), Score: &score, Runs: s.Runs,
		DurationMs: trial.DurationMs, Usage: protocol.UsagePtr(trial.Usage), Predicates: trial.Predicates,
	}
	if arm == model.Present {
		loaded := true
		for _, obs := range trial.Observations {
			loaded = loaded && obs.TargetInvoked
		}
		result.SkillTriggered = &loaded
		if !loaded {
			result.Outcome = model.Failed
			result.Error = "the skill was not loaded in every with-skill run"
		}
	}
	return result, nil
}

// rate is the share of predicates that passed over the required cases of an arm.
func rate(cases []model.CaseResult, gates map[string]model.Gate, arm model.Condition) float64 {
	passed, total := 0, 0
	for _, c := range cases {
		if c.Condition != arm || gates[c.ID] != model.Required {
			continue
		}
		for _, p := range c.Predicates {
			total++
			if p.Passed {
				passed++
			}
		}
	}
	if total == 0 {
		return 0
	}
	return float64(passed) / float64(total)
}

func decide(cases []model.CaseResult, gates map[string]model.Gate, policy Policy, minimumWinRate float64) protocol.Verdict {
	with, without := rate(cases, gates, model.Present), rate(cases, gates, model.Absent)
	lift := with - without
	pass := with >= policy.MinimumWithSkill && lift >= policy.MinimumLift
	lines := []string{
		fmt.Sprintf("with the skill: %.0f%% (minimum %.0f%%)", with*100, policy.MinimumWithSkill*100),
		fmt.Sprintf("without the skill: %.0f%%", without*100),
		fmt.Sprintf("lift: %+.0f points (minimum %+.0f)", lift*100, policy.MinimumLift*100),
	}
	byKey := map[string]model.CaseResult{}
	for _, c := range cases {
		byKey[c.Key()] = c
	}
	wins, required := 0, 0
	for _, c := range cases {
		if c.Condition != model.Present || gates[c.ID] != model.Required {
			continue
		}
		required++
		if c.SkillTriggered != nil && !*c.SkillTriggered {
			pass = false
			lines = append(lines, "skill not loaded: "+c.ID)
		}
		if control, ok := byKey[model.CaseKey(c.ID, model.Absent)]; ok && c.Score != nil && control.Score != nil && *c.Score > *control.Score {
			wins++
		}
	}
	if minimumWinRate > 0 && required > 0 {
		win := float64(wins) / float64(required)
		lines = append(lines, fmt.Sprintf("win rate: %.0f%% (minimum %.0f%%)", win*100, minimumWinRate*100))
		pass = pass && win >= minimumWinRate
	}
	for _, c := range cases {
		if c.Condition == model.Present && gates[c.ID] == model.Required && c.Outcome != model.Passed {
			lines = append(lines, "failed: "+c.ID)
		}
	}
	return protocol.Verdict{Pass: pass, Lines: lines}
}
