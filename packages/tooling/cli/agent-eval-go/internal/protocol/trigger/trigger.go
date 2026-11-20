// Package trigger is the protocol that measures whether a skill loads when it
// should and stays unloaded when it should not. Each query runs several times
// with only the target skill on offer; the share of runs that loaded the skill
// is compared with a threshold.
package trigger

import (
	"context"
	"fmt"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/store"
)

// Protocol is the trigger protocol.
type Protocol struct{}

// New returns the protocol.
func New() Protocol { return Protocol{} }

// Name is the protocol name.
func (Protocol) Name() model.Protocol { return model.Trigger }

// Plan makes one unit of each query of the skill that is in the split.
func (Protocol) Plan(in protocol.Input) (protocol.Plan, error) {
	skill, err := in.Env.Catalog.Skill(in.Target)
	if err != nil {
		return protocol.Plan{}, err
	}
	if skill.TriggerFile == "" {
		return protocol.Plan{}, fmt.Errorf("skill %s has no %s", skill.Name, model.TriggerFile)
	}
	doc, err := in.Env.Reader.Trigger(skill.TriggerFile)
	if err != nil {
		return protocol.Plan{}, err
	}
	s := in.Settings
	offer := protocol.Offers(skill)
	plan := protocol.Plan{Subject: skill.Name}
	for _, q := range doc.Queries {
		if s.Split != "" && q.Split != s.Split {
			continue
		}
		plan.Inputs = append(plan.Inputs, q.ID, q.Text, fmt.Sprint(q.ShouldTrigger))
		query := q
		plan.Units = append(plan.Units, protocol.Unit{
			ID: q.ID, ModelRuns: s.Runs,
			Run: func(ctx context.Context) (model.CaseResult, error) {
				return runQuery(ctx, in, skill.Name, offer, query)
			},
		})
	}
	if len(plan.Units) == 0 {
		return protocol.Plan{}, fmt.Errorf("skill %s has no queries in split %q", skill.Name, s.Split)
	}
	plan.Inputs = append(plan.Inputs, skill.Name, fmt.Sprint(s.Runs, s.Threshold, s.Harness, s.Model, s.Effort, s.Split),
		store.Digest(skill.Dir))
	plan.Decide = decide
	return plan, nil
}

func runQuery(ctx context.Context, in protocol.Input, skill string, offer []harness.Offer, q model.TriggerQuery) (model.CaseResult, error) {
	s := in.Settings
	result := model.CaseResult{ID: q.ID, Runs: s.Runs}
	var usage model.Usage
	hits := 0
	for i := 0; i < s.Runs; i++ {
		obs, err := in.Env.Exec.Execute(ctx, protocol.TriggerRequest(s, q.Text, skill, offer))
		if err != nil {
			return result, err
		}
		usage = usage.Add(obs.Usage)
		result.DurationMs += obs.DurationMs
		result.Selected = append(result.Selected, obs.Selected)
		if obs.Selected == harness.SelectedTarget {
			hits++
		}
	}
	rate := float64(hits) / float64(s.Runs)
	triggered := rate >= s.Threshold
	result.Score, result.SkillTriggered, result.Usage = &rate, &triggered, protocol.UsagePtr(usage)
	result.Outcome = model.Failed
	if triggered == q.ShouldTrigger {
		result.Outcome = model.Passed
	}
	return result, nil
}

func decide(cases []model.CaseResult) protocol.Verdict {
	lines := protocol.Failures(cases)
	passed := len(cases) - len(lines)
	return protocol.Verdict{
		Pass:  len(lines) == 0,
		Lines: append([]string{fmt.Sprintf("%d of %d queries behaved as expected", passed, len(cases))}, lines...),
	}
}
