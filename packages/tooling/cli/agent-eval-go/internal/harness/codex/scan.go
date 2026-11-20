package codex

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/process"
)

// triggerOutputLimit is how many answer characters a trigger run may stream
// before the routing outcome is settled: a long plain answer was written
// without the skill.
const triggerOutputLimit = 2000

type scanner struct {
	req harness.Request

	text        string
	usage       model.Usage
	events      []harness.Event
	matched     bool
	outputChars int
	exceeded    bool
}

// Scan returns a scanner for the stream of a request.
func (Harness) Scan(req harness.Request) harness.Scanner { return &scanner{req: req} }

func (s *scanner) limit() int {
	if s.req.Role == harness.RoleTrigger {
		return triggerOutputLimit
	}
	return s.req.Caps.MaxOutputChars
}

func (s *scanner) settled() bool {
	return s.matched || s.exceeded
}

// Feed reads one JSONL line.
func (s *scanner) Feed(raw string) bool {
	if s.settled() {
		return true
	}
	var event map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &event); err != nil {
		return false
	}
	switch event["type"] {
	case "item.completed":
		s.item(event)
	case "turn.completed":
		s.usage = usageOf(event)
	}
	if limit := s.limit(); limit > 0 && s.outputChars > limit {
		s.exceeded = true
	}
	return s.settled()
}

func (s *scanner) item(event map[string]any) {
	item, _ := event["item"].(map[string]any)
	switch item["type"] {
	case "agent_message":
		text, _ := item["text"].(string)
		s.text = text
		s.outputChars += len(text)
		s.events = append(s.events, harness.Event{Kind: harness.EventMessage, Text: text})
		if s.req.Role == harness.RoleTrigger && strings.Contains(text, TriggerSignal) {
			s.matched = true
		}
	case "command_execution":
		command, _ := item["command"].(string)
		s.events = append(s.events,
			harness.Event{Kind: harness.EventToolCall, Name: "command_execution"},
			harness.Event{Kind: harness.EventCommand, Text: command})
	case "file_change":
		s.events = append(s.events, harness.Event{Kind: harness.EventToolCall, Name: "file_change"})
		changes, _ := item["changes"].([]any)
		for _, c := range changes {
			change, _ := c.(map[string]any)
			p, _ := change["path"].(string)
			s.events = append(s.events, harness.Event{Kind: harness.EventFileWrite, Text: p})
		}
	}
}

// Finish combines the stream with how the program ended.
func (s *scanner) Finish(exit process.Result, elapsedMs int64) harness.Observation {
	obs := harness.Observation{
		Text: s.text, Usage: s.usage, DurationMs: elapsedMs, Events: s.events,
		OutputLimit: s.exceeded, TimedOut: exit.TimedOut,
	}
	if s.req.Role == harness.RoleJudge {
		obs.Structured = s.text
	}
	if s.req.Role == harness.RoleTrigger {
		switch {
		case s.matched:
			obs.Selected, obs.TargetInvoked = harness.SelectedTarget, true
		case exit.TimedOut:
			obs.Failure = "timeout"
		case s.exceeded:
			obs.Selected = harness.SelectedOutputLimit
		case exit.ExitCode != 0:
			obs.Failure = exitFailure(exit)
		default:
			obs.Selected = harness.SelectedNone
		}
		return obs
	}
	if s.req.WithSkill && strings.HasPrefix(s.text, UseSignal) {
		obs.TargetInvoked = true
		obs.Text = strings.TrimLeft(strings.TrimPrefix(s.text, UseSignal), " \t\r\n")
	}
	if !exit.TimedOut && !exit.Stopped && !s.exceeded && exit.ExitCode != 0 {
		obs.Failure = exitFailure(exit)
	}
	return obs
}

func exitFailure(exit process.Result) string {
	detail := strings.TrimSpace(exit.Stderr)
	if detail == "" {
		return fmt.Sprintf("codex exited %d", exit.ExitCode)
	}
	return fmt.Sprintf("codex exited %d: %s", exit.ExitCode, detail)
}

func usageOf(event map[string]any) model.Usage {
	usage, _ := event["usage"].(map[string]any)
	number := func(key string) int {
		v, _ := usage[key].(float64)
		return int(v)
	}
	return model.Usage{InputTokens: number("input_tokens"), OutputTokens: number("output_tokens")}
}
