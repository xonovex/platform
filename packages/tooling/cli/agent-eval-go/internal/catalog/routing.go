package catalog

import (
	"sort"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// Scenario is one routing question: a prompt that exactly one skill wants and
// at least one other skill does not, offered to the model next to every skill
// that carries the prompt.
type Scenario struct {
	Prompt string
	// Positive is the skill that should be loaded.
	Positive string
	// Candidates lists every skill that carries the prompt, sorted. It includes
	// the positive skill.
	Candidates []string
	// Split is the split of the positive skill's query.
	Split model.Split
}

// Scenarios finds the routing scenarios in the trigger documents, keyed by
// skill name. A prompt that two skills both mark positive, or that no skill
// marks positive, or that no other skill marks negative, is no scenario.
func Scenarios(docs map[string]model.TriggerEval) []Scenario {
	type carrier struct {
		skill string
		query model.TriggerQuery
	}
	byPrompt := map[string][]carrier{}
	for skill, doc := range docs {
		for _, q := range doc.Queries {
			byPrompt[q.Text] = append(byPrompt[q.Text], carrier{skill, q})
		}
	}
	var out []Scenario
	for prompt, carriers := range byPrompt {
		var positive *carrier
		positives, negatives := 0, 0
		for i, c := range carriers {
			if c.query.ShouldTrigger {
				positives++
				positive = &carriers[i]
			} else {
				negatives++
			}
		}
		if positives != 1 || negatives < 1 {
			continue
		}
		candidates := make([]string, len(carriers))
		for i, c := range carriers {
			candidates[i] = c.skill
		}
		sort.Strings(candidates)
		out = append(out, Scenario{Prompt: prompt, Positive: positive.skill, Candidates: candidates, Split: positive.query.Split})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Positive != out[j].Positive {
			return out[i].Positive < out[j].Positive
		}
		return out[i].Prompt < out[j].Prompt
	})
	return out
}
