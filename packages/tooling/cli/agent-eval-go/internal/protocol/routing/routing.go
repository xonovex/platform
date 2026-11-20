// Package routing is the protocol that measures whether the model picks the
// right skill when several skills are on offer. A scenario is a prompt that one
// skill wants and at least one other skill does not, offered next to every skill
// that carries it. The run passes when the wanted skill is the one loaded.
package routing

import (
	"context"
	"fmt"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/store"
)

// Protocol is the routing protocol.
type Protocol struct{}

// New returns the protocol.
func New() Protocol { return Protocol{} }

// Name is the protocol name.
func (Protocol) Name() model.Protocol { return model.Routing }

// Plan finds the scenarios among the trigger documents of the target plugin, or
// of every plugin when the target is empty.
func (Protocol) Plan(in protocol.Input) (protocol.Plan, error) {
	docs, skills, err := load(in)
	if err != nil {
		return protocol.Plan{}, err
	}
	subject := in.Target
	if subject == "" {
		subject = "all"
	}
	plan := protocol.Plan{Subject: subject}
	for _, scenario := range catalog.Scenarios(docs) {
		if in.Settings.Split != "" && scenario.Split != in.Settings.Split {
			continue
		}
		id := "route-" + store.Digest(scenario.Prompt)[len(store.DigestPrefix):len(store.DigestPrefix)+8]
		candidates := make([]catalog.Skill, len(scenario.Candidates))
		for i, key := range scenario.Candidates {
			candidates[i] = skills[key]
		}
		sc := scenario
		wanted := skills[sc.Positive]
		plan.Inputs = append(plan.Inputs, id, sc.Positive, fmt.Sprint(sc.Candidates))
		plan.Units = append(plan.Units, protocol.Unit{
			ID: id, ModelRuns: in.Settings.Runs,
			Run: func(ctx context.Context) (model.CaseResult, error) {
				return run(ctx, in, id, sc, wanted.Name, candidates)
			},
		})
	}
	if len(plan.Units) == 0 {
		return protocol.Plan{}, fmt.Errorf("no routing scenario in %s: a prompt needs one skill that wants it and another that does not", subject)
	}
	plan.Inputs = append(plan.Inputs, fmt.Sprint(in.Settings.Runs, in.Settings.Threshold, in.Settings.Harness, in.Settings.Model, in.Settings.Effort, in.Settings.Split))
	plan.Decide = decide
	return plan, nil
}

// load reads the trigger documents of the skills in scope, keyed plugin:skill.
func load(in protocol.Input) (map[string]model.TriggerEval, map[string]catalog.Skill, error) {
	docs := map[string]model.TriggerEval{}
	skills := map[string]catalog.Skill{}
	matched := false
	for _, plugin := range in.Env.Catalog.Plugins {
		if in.Target != "" && plugin.Name != in.Target {
			continue
		}
		matched = true
		for _, skill := range plugin.Skills {
			if skill.TriggerFile == "" {
				continue
			}
			doc, err := in.Env.Reader.Trigger(skill.TriggerFile)
			if err != nil {
				return nil, nil, err
			}
			key := plugin.Name + ":" + skill.Name
			docs[key], skills[key] = doc, skill
		}
	}
	if !matched {
		return nil, nil, fmt.Errorf("plugin %q not found", in.Target)
	}
	return docs, skills, nil
}

func run(ctx context.Context, in protocol.Input, id string, sc catalog.Scenario, wanted string, candidates []catalog.Skill) (model.CaseResult, error) {
	s := in.Settings
	offer := protocol.Offers(candidates...)
	result := model.CaseResult{ID: id, Runs: s.Runs}
	var usage model.Usage
	hits := 0
	for i := 0; i < s.Runs; i++ {
		obs, err := in.Env.Exec.Execute(ctx, protocol.TriggerRequest(s, sc.Prompt, wanted, offer))
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
	chosen := rate >= s.Threshold
	result.Score, result.SkillTriggered, result.Usage = &rate, &chosen, protocol.UsagePtr(usage)
	result.Outcome = model.Failed
	if chosen {
		result.Outcome = model.Passed
	}
	return result, nil
}

func decide(cases []model.CaseResult) protocol.Verdict {
	lines := protocol.Failures(cases)
	return protocol.Verdict{
		Pass:  len(lines) == 0,
		Lines: append([]string{fmt.Sprintf("%d of %d scenarios routed to the right skill", len(cases)-len(lines), len(cases))}, lines...),
	}
}
