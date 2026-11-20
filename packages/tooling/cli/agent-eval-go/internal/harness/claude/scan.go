package claude

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// toolResult says what the harness did with a Skill call. A name it cannot
// resolve launches nothing, so only a launch settles which skill answered.
type toolResult int

const (
	resultNone toolResult = iota
	resultLaunched
	resultRejected
)

// scanner reads the stream of one Claude run.
type scanner struct {
	req harness.Request

	text       string
	structured string
	usage      model.Usage
	durationMs int64
	events     []harness.Event
	failure    string
	raw        []string

	matched         bool
	targetAvailable bool
	pendingSkill    string
	hasPending      bool
	competing       string
	outputChars     int
	outputExceeded  bool
}

// Scan returns a scanner for the stream of a request.
func (Harness) Scan(req harness.Request) harness.Scanner { return &scanner{req: req} }

func (s *scanner) settled() bool {
	return s.req.Role == harness.RoleTrigger && (s.matched || s.competing != "" || s.outputExceeded)
}

// Feed reads one line of the stream.
func (s *scanner) Feed(raw string) bool {
	if s.settled() {
		return true
	}
	line := strings.TrimSpace(raw)
	if line == "" {
		return false
	}
	s.raw = append(s.raw, line)
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		return false
	}
	if s.req.Role == harness.RoleTrigger {
		s.scanTrigger(obj)
		return s.settled()
	}
	s.scanSession(obj)
	if s.req.Caps.MaxOutputChars > 0 {
		s.outputChars += deltaLength(obj)
		if s.outputChars > s.req.Caps.MaxOutputChars {
			s.outputExceeded = true
			return true
		}
	}
	return false
}

func (s *scanner) matches(name string) bool {
	short := s.req.Target
	return name == short || strings.HasSuffix(name, ":"+short)
}

func (s *scanner) scanTrigger(obj map[string]any) {
	if obj["type"] == "system" && obj["subtype"] == "init" {
		if skills, ok := obj["skills"].([]any); ok {
			for _, skill := range skills {
				if name, ok := skill.(string); ok && s.matches(name) {
					s.targetAvailable = true
				}
			}
		}
	}
	if obj["type"] == "result" {
		// A run that ends without a skill call still reports what it spent.
		s.usage = usageOf(obj)
		if d, ok := obj["duration_ms"].(float64); ok {
			s.durationMs = int64(d)
		}
	}
	invoked := invokedSkills(obj)
	for _, name := range invoked {
		if s.matches(name) {
			s.matched = true
			return
		}
	}
	if len(invoked) > 0 {
		s.pendingSkill, s.hasPending = invoked[0], true
		return
	}
	if s.hasPending {
		switch toolResultOf(obj) {
		case resultLaunched:
			s.competing = s.pendingSkill
		case resultRejected:
			s.pendingSkill, s.hasPending = "", false
		}
		return
	}
	s.outputChars += deltaLength(obj)
	if s.outputChars > 2000 {
		s.outputExceeded = true
	}
}

func (s *scanner) scanSession(obj map[string]any) {
	switch obj["type"] {
	case "system":
		if obj["subtype"] == "init" {
			if skills, ok := obj["skills"].([]any); ok {
				for _, skill := range skills {
					if name, ok := skill.(string); ok && s.req.Target != "" && s.matches(name) {
						s.targetAvailable = true
					}
				}
			}
		}
	case "assistant":
		s.scanAssistant(obj)
	case "result":
		if result, ok := obj["result"].(string); ok {
			s.text = result
		}
		if structured, ok := obj["structured_output"]; ok && structured != nil {
			if data, err := json.Marshal(structured); err == nil {
				s.structured = string(data)
			}
		}
		s.usage = usageOf(obj)
		if d, ok := obj["duration_ms"].(float64); ok {
			s.durationMs = int64(d)
		}
		if obj["is_error"] == true {
			s.failure = failureText(obj)
		} else if sub, _ := obj["subtype"].(string); strings.HasPrefix(sub, "error") {
			s.failure = failureText(obj)
		}
	}
}

func (s *scanner) scanAssistant(obj map[string]any) {
	message, _ := obj["message"].(map[string]any)
	content, _ := message["content"].([]any)
	for _, item := range content {
		block, _ := item.(map[string]any)
		switch block["type"] {
		case "text":
			if text, ok := block["text"].(string); ok {
				s.events = append(s.events, harness.Event{Kind: harness.EventMessage, Text: text})
			}
		case "tool_use":
			name, _ := block["name"].(string)
			input, _ := block["input"].(map[string]any)
			s.events = append(s.events, harness.Event{Kind: harness.EventToolCall, Name: name})
			switch name {
			case "Skill":
				if skill, ok := input["skill"].(string); ok && s.matches(skill) {
					s.matched = true
				}
			case "Read":
				s.events = append(s.events, harness.Event{Kind: harness.EventFileRead, Text: str(input["file_path"])})
			case "Write", "Edit", "MultiEdit", "NotebookEdit":
				s.events = append(s.events, harness.Event{Kind: harness.EventFileWrite, Text: str(input["file_path"])})
			case "Bash":
				s.events = append(s.events, harness.Event{Kind: harness.EventCommand, Text: str(input["command"])})
			}
		}
	}
}

// Finish combines the stream with how the program ended.
func (s *scanner) Finish(exit process.Result, elapsedMs int64) harness.Observation {
	obs := harness.Observation{
		Text: s.text, Structured: s.structured, Usage: s.usage, DurationMs: s.durationMs,
		Events: s.events, TargetAvailable: s.targetAvailable, TargetInvoked: s.matched,
		OutputLimit: s.outputExceeded, TimedOut: exit.TimedOut,
	}
	if obs.DurationMs == 0 {
		obs.DurationMs = elapsedMs
	}
	if s.req.Role == harness.RoleJudge && s.structured == "" && len(s.raw) > 0 {
		// A plain json answer is one object: read it from the whole output.
		var whole map[string]any
		if err := json.Unmarshal([]byte(strings.Join(s.raw, "")), &whole); err == nil {
			s.scanSession(whole)
			obs.Text, obs.Structured, obs.Usage = s.text, s.structured, s.usage
			if s.failure != "" {
				obs.Failure = s.failure
			}
		}
	}
	if s.req.Role == harness.RoleTrigger {
		return s.finishTrigger(obs, exit)
	}
	if !exit.TimedOut && !exit.Stopped && !s.outputExceeded && exit.ExitCode != 0 {
		detail := strings.TrimSpace(exit.Stderr)
		if detail == "" {
			detail = s.failure
		}
		obs.Failure = exitFailure(exit.ExitCode, detail)
	}
	return obs
}

func (s *scanner) finishTrigger(obs harness.Observation, exit process.Result) harness.Observation {
	switch {
	case s.matched:
		obs.Selected = harness.SelectedTarget
		obs.TargetInvoked = true
	case exit.TimedOut:
		obs.Failure = "timeout"
	case s.competing != "" || s.outputExceeded:
		if s.competing != "" {
			obs.Selected = harness.SelectedCompetitor + s.competing
		} else {
			obs.Selected = harness.SelectedOutputLimit
		}
		if !s.targetAvailable {
			obs.Failure = "target skill unavailable"
		}
	case exit.ExitCode != 0:
		obs.Failure = exitFailure(exit.ExitCode, strings.TrimSpace(exit.Stderr))
	case !s.targetAvailable:
		obs.Failure = "target skill unavailable"
	default:
		obs.Selected = harness.SelectedNone
	}
	return obs
}

func exitFailure(code int, detail string) string {
	if detail == "" {
		return fmt.Sprintf("claude exited %d", code)
	}
	return fmt.Sprintf("claude exited %d: %s", code, detail)
}

func failureText(obj map[string]any) string {
	if result, ok := obj["result"].(string); ok && strings.TrimSpace(result) != "" {
		return truncate(strings.TrimSpace(result), 500)
	}
	if sub, ok := obj["subtype"].(string); ok {
		return truncate(sub, 500)
	}
	return "error"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

// invokedSkills lists every skill the line invokes, whether the harness allowed
// the call or denied it.
func invokedSkills(obj map[string]any) []string {
	var names []string
	if message, ok := obj["message"].(map[string]any); ok {
		content, _ := message["content"].([]any)
		for _, item := range content {
			block, _ := item.(map[string]any)
			if block["type"] == "tool_use" && block["name"] == "Skill" {
				if input, ok := block["input"].(map[string]any); ok {
					if skill, ok := input["skill"].(string); ok {
						names = append(names, skill)
					}
				}
			}
		}
	}
	denials, _ := obj["permission_denials"].([]any)
	for _, d := range denials {
		denial, _ := d.(map[string]any)
		if denial["tool_name"] == "Skill" {
			if input, ok := denial["tool_input"].(map[string]any); ok {
				if skill, ok := input["skill"].(string); ok {
					names = append(names, skill)
				}
			}
		}
	}
	return names
}

func toolResultOf(obj map[string]any) toolResult {
	message, ok := obj["message"].(map[string]any)
	if !ok {
		return resultNone
	}
	content, _ := message["content"].([]any)
	for _, item := range content {
		block, _ := item.(map[string]any)
		if block["type"] != "tool_result" {
			continue
		}
		if block["is_error"] == true {
			return resultRejected
		}
		text, isText := block["content"].(string)
		if !isText {
			data, _ := json.Marshal(block["content"])
			text = string(data)
		}
		if strings.Contains(text, "tool_use_error") {
			return resultRejected
		}
		return resultLaunched
	}
	return resultNone
}

// deltaLength is the number of answer characters a stream event carries.
func deltaLength(obj map[string]any) int {
	if obj["type"] != "stream_event" {
		return 0
	}
	event, _ := obj["event"].(map[string]any)
	if event["type"] != "content_block_delta" {
		return 0
	}
	delta, _ := event["delta"].(map[string]any)
	if delta["type"] != "text_delta" {
		return 0
	}
	text, _ := delta["text"].(string)
	return len(text)
}

func usageOf(obj map[string]any) model.Usage {
	usage, _ := obj["usage"].(map[string]any)
	number := func(key string) int {
		v, _ := usage[key].(float64)
		return int(v)
	}
	cost, _ := obj["total_cost_usd"].(float64)
	return model.Usage{
		InputTokens:  number("input_tokens") + number("cache_creation_input_tokens") + number("cache_read_input_tokens"),
		OutputTokens: number("output_tokens"),
		CostUSD:      cost,
	}
}
