package claude_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness/claude"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/process"
)

func has(args []string, flag string, value ...string) bool {
	for i, a := range args {
		if a != flag {
			continue
		}
		if len(value) == 0 {
			return true
		}
		if i+1 < len(args) && args[i+1] == value[0] {
			return true
		}
	}
	return false
}

func offer() []harness.Offer {
	return []harness.Offer{{Name: "plug", Dir: "/p", SkillDirs: []string{"/p/skills/one", "/p/skills/two"}}}
}

func TestNameAndBinary(t *testing.T) {
	h := claude.New()
	if h.Name() != "claude" || h.Binary() != "claude" {
		t.Fatal(h.Name(), h.Binary())
	}
}

func TestTriggerPlanIsHermeticAndStagesOnlyTheOffer(t *testing.T) {
	plan, err := claude.New().Plan(harness.Request{
		Role: harness.RoleTrigger, Model: "m", Prompt: "q", Target: "one", Offer: offer(),
		Caps: harness.Caps{BudgetUSD: 0.05}, Effort: model.Effort("low"),
	}, "/ws")
	if err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{"--no-session-persistence", "--no-chrome", "--strict-mcp-config", "--verbose"} {
		if !has(plan.Args, flag) {
			t.Errorf("missing %s in %v", flag, plan.Args)
		}
	}
	if !has(plan.Args, "--setting-sources", "") || !has(plan.Args, "--tools", "Skill") ||
		!has(plan.Args, "--max-turns", "2") || !has(plan.Args, "--max-budget-usd", "0.05") ||
		!has(plan.Args, "--model", "m") || !has(plan.Args, "--effort", "low") ||
		!has(plan.Args, "--plugin-dir", "/ws/plugin-0") {
		t.Fatalf("%v", plan.Args)
	}
	if plan.Prompt != "q" || plan.Dir != "/ws" {
		t.Fatalf("%+v", plan)
	}
	var copies int
	var manifest string
	for _, s := range plan.Stage {
		switch s.Kind {
		case harness.StageCopy:
			copies++
		case harness.StageWrite:
			manifest = s.Content
		}
	}
	var m struct {
		Name   string   `json:"name"`
		Skills []string `json:"skills"`
	}
	if err := json.Unmarshal([]byte(manifest), &m); err != nil || copies != 2 || m.Name != "plug" || len(m.Skills) != 2 || m.Skills[0] != "./skills/one" {
		t.Fatalf("copies=%d manifest=%s", copies, manifest)
	}
}

func TestGeneratePlanArms(t *testing.T) {
	with, err := claude.New().Plan(harness.Request{Role: harness.RoleGenerate, Prompt: "do", Target: "one", Offer: offer(), WithSkill: true}, "/ws")
	if err != nil {
		t.Fatal(err)
	}
	if with.Prompt != "/one do" || !has(with.Args, "--tools", "Skill,Read") || !has(with.Args, "--max-turns", "12") || len(with.Stage) == 0 {
		t.Fatalf("%+v", with)
	}
	without, err := claude.New().Plan(harness.Request{Role: harness.RoleGenerate, Prompt: "do", Caps: harness.Caps{MaxTurns: 5}}, "/ws")
	if err != nil {
		t.Fatal(err)
	}
	if without.Prompt != "do" || !has(without.Args, "--tools", "Read") || !has(without.Args, "--disallowedTools=Skill") ||
		!has(without.Args, "--max-turns", "5") || len(without.Stage) != 0 {
		t.Fatalf("%+v", without)
	}
}

func TestJudgePlanHasStrictSchemaAndNoTools(t *testing.T) {
	plan, err := claude.New().Plan(harness.Request{Role: harness.RoleJudge, Prompt: "grade", Assertions: 3, Model: "j", Caps: harness.Caps{BudgetUSD: 0.1}}, "/ws")
	if err != nil {
		t.Fatal(err)
	}
	if !has(plan.Args, "--tools", "") || !has(plan.Args, "--max-turns", "1") || !has(plan.Args, "--output-format", "json") || !has(plan.Args, "--model", "j") {
		t.Fatalf("%v", plan.Args)
	}
	var schema string
	for i, a := range plan.Args {
		if a == "--json-schema" {
			schema = plan.Args[i+1]
		}
	}
	if !strings.Contains(schema, `"minItems":3`) || !strings.Contains(schema, `"maxItems":3`) {
		t.Fatal(schema)
	}
}

func TestAgentPlanStagesInstructionsAndDeniesBash(t *testing.T) {
	plan, err := claude.New().Plan(harness.Request{Role: harness.RoleAgent, Prompt: "work", Instructions: "# Rules"}, "/ws")
	if err != nil {
		t.Fatal(err)
	}
	if !has(plan.Args, "--allowedTools", "Read,Write,Edit,Glob,Grep") || !strings.Contains(strings.Join(plan.Args, " "), "Bash") {
		t.Fatalf("%v", plan.Args)
	}
	if len(plan.Stage) != 1 || plan.Stage[0].Dst != "/ws/CLAUDE.md" || plan.Stage[0].Content != "# Rules" {
		t.Fatalf("%+v", plan.Stage)
	}
	bare, _ := claude.New().Plan(harness.Request{Role: harness.RoleAgent, Prompt: "work"}, "/ws")
	if len(bare.Stage) != 0 {
		t.Fatal("no instructions, no stage")
	}
}

func TestUnknownRoleFails(t *testing.T) {
	if _, err := claude.New().Plan(harness.Request{Role: "x"}, "/ws"); err == nil {
		t.Fatal("want an error")
	}
}

func scan(req harness.Request, lines ...string) (harness.Scanner, bool) {
	s := claude.New().Scan(req)
	settled := false
	for _, l := range lines {
		if s.Feed(l) {
			settled = true
			break
		}
	}
	return s, settled
}

const (
	initLine    = `{"type":"system","subtype":"init","skills":["plug:one","other"]}`
	targetCall  = `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Skill","input":{"skill":"plug:one"}}]}}`
	otherCall   = `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Skill","input":{"skill":"other"}}]}}`
	launched    = `{"type":"user","message":{"content":[{"type":"tool_result","content":"Launching skill"}]}}`
	rejected    = `{"type":"user","message":{"content":[{"type":"tool_result","is_error":true,"content":"x"}]}}`
	rejectedTxt = `{"type":"user","message":{"content":[{"type":"tool_result","content":[{"type":"text","text":"<tool_use_error>no</tool_use_error>"}]}]}}`
)

func trigger() harness.Request {
	return harness.Request{Role: harness.RoleTrigger, Target: "one"}
}

func TestTriggerMatchesTheTarget(t *testing.T) {
	s, settled := scan(trigger(), "", "not json", initLine, targetCall)
	if !settled {
		t.Fatal("a match settles the run")
	}
	obs := s.Finish(process.Result{Stopped: true}, 5)
	if obs.Selected != harness.SelectedTarget || !obs.TargetInvoked || obs.Failure != "" {
		t.Fatalf("%+v", obs)
	}
	if s.Feed("anything") != true {
		t.Fatal("a settled scanner stays settled")
	}
}

func TestTriggerRunThatEndsWithoutASkillReportsItsUsage(t *testing.T) {
	result := `{"type":"result","result":"no","duration_ms":40,"total_cost_usd":0.02,"usage":{"input_tokens":7,"output_tokens":3}}`
	s, settled := scan(trigger(), initLine, result)
	if settled {
		t.Fatal("a result settles nothing")
	}
	obs := s.Finish(process.Result{}, 5)
	if obs.Selected != harness.SelectedNone || obs.Usage.CostUSD != 0.02 || obs.Usage.InputTokens != 7 || obs.DurationMs != 40 {
		t.Fatalf("%+v", obs)
	}
}

func TestTriggerCompetitorMustLaunch(t *testing.T) {
	s, settled := scan(trigger(), initLine, otherCall, launched)
	if !settled {
		t.Fatal("a launched competitor settles the run")
	}
	if obs := s.Finish(process.Result{Stopped: true}, 0); obs.Selected != "competitor:other" || obs.Failure != "" {
		t.Fatalf("%+v", obs)
	}
	s, settled = scan(trigger(), initLine, otherCall, rejected, rejectedTxt)
	if settled {
		t.Fatal("a rejected call settles nothing")
	}
	if obs := s.Finish(process.Result{}, 0); obs.Selected != harness.SelectedNone {
		t.Fatalf("%+v", obs)
	}
	s, _ = scan(trigger(), initLine, otherCall, `{"type":"user","message":{"content":[{"type":"text"}]}}`, `{"type":"assistant"}`)
	if obs := s.Finish(process.Result{}, 0); obs.Selected != harness.SelectedNone {
		t.Fatalf("%+v", obs)
	}
}

func TestTriggerDenialCountsAsAnInvocation(t *testing.T) {
	denied := `{"type":"result","permission_denials":[{"tool_name":"Skill","tool_input":{"skill":"plug:one"}},{"tool_name":"Read"}]}`
	s, settled := scan(trigger(), initLine, denied)
	if !settled || s.Finish(process.Result{Stopped: true}, 0).Selected != harness.SelectedTarget {
		t.Fatal("a denied call still shows the model chose the skill")
	}
}

func TestTriggerOutputLimit(t *testing.T) {
	delta := `{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"` + strings.Repeat("x", 2100) + `"}}}`
	s, settled := scan(trigger(), initLine, delta)
	if !settled {
		t.Fatal("a long plain answer settles the run")
	}
	obs := s.Finish(process.Result{Stopped: true}, 0)
	if obs.Selected != harness.SelectedOutputLimit || !obs.OutputLimit || obs.Failure != "" {
		t.Fatalf("%+v", obs)
	}
	s, _ = scan(trigger(), delta)
	if obs := s.Finish(process.Result{Stopped: true}, 0); obs.Failure != "target skill unavailable" {
		t.Fatalf("%+v", obs)
	}
}

func TestTriggerFailures(t *testing.T) {
	s, _ := scan(trigger(), initLine)
	if obs := s.Finish(process.Result{TimedOut: true}, 0); obs.Failure != "timeout" {
		t.Fatalf("%+v", obs)
	}
	s, _ = scan(trigger(), initLine)
	if obs := s.Finish(process.Result{ExitCode: 3, Stderr: " boom "}, 0); obs.Failure != "claude exited 3: boom" {
		t.Fatalf("%+v", obs)
	}
	s, _ = scan(trigger(), initLine)
	if obs := s.Finish(process.Result{ExitCode: 3}, 0); obs.Failure != "claude exited 3" {
		t.Fatalf("%+v", obs)
	}
	s, _ = scan(trigger())
	if obs := s.Finish(process.Result{}, 0); obs.Failure != "target skill unavailable" {
		t.Fatalf("%+v", obs)
	}
}

func TestGenerateReadsEventsTextAndUsage(t *testing.T) {
	req := harness.Request{Role: harness.RoleGenerate, Target: "one", WithSkill: true}
	s, _ := scan(req,
		initLine,
		targetCall,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"hello"},{"type":"tool_use","name":"Read","input":{"file_path":"a.md"}},{"type":"tool_use","name":"Write","input":{"file_path":"b.md"}},{"type":"tool_use","name":"Bash","input":{"command":"ls"}}]}}`,
		`{"type":"result","result":"final","usage":{"input_tokens":10,"cache_read_input_tokens":5,"cache_creation_input_tokens":1,"output_tokens":7},"total_cost_usd":0.02,"duration_ms":1234}`,
	)
	obs := s.Finish(process.Result{}, 9)
	if obs.Text != "final" || obs.DurationMs != 1234 || !obs.TargetInvoked || !obs.TargetAvailable {
		t.Fatalf("%+v", obs)
	}
	if obs.Usage != (model.Usage{InputTokens: 16, OutputTokens: 7, CostUSD: 0.02}) {
		t.Fatalf("%+v", obs.Usage)
	}
	kinds := map[string]string{}
	for _, e := range obs.Events {
		kinds[e.Kind] = e.Text
	}
	if kinds[harness.EventMessage] != "hello" || kinds[harness.EventFileRead] != "a.md" ||
		kinds[harness.EventFileWrite] != "b.md" || kinds[harness.EventCommand] != "ls" {
		t.Fatalf("%v", kinds)
	}
}

func TestGenerateFailuresAndOutputCap(t *testing.T) {
	req := harness.Request{Role: harness.RoleGenerate}
	s, _ := scan(req, `{"type":"result","is_error":true,"result":"max turns reached","subtype":"error_max_turns"}`)
	obs := s.Finish(process.Result{ExitCode: 1}, 40)
	if obs.Failure != "claude exited 1: max turns reached" || obs.DurationMs != 40 {
		t.Fatalf("%+v", obs)
	}
	s, _ = scan(req, `{"type":"result","subtype":"error_during_execution"}`)
	if obs := s.Finish(process.Result{ExitCode: 1}, 0); obs.Failure != "claude exited 1: error_during_execution" {
		t.Fatalf("%+v", obs)
	}
	s, _ = scan(req, `{"type":"result","is_error":true}`)
	if obs := s.Finish(process.Result{ExitCode: 1}, 0); obs.Failure != "claude exited 1: error" {
		t.Fatalf("%+v", obs)
	}
	capped := harness.Request{Role: harness.RoleGenerate, Caps: harness.Caps{MaxOutputChars: 10}}
	delta := `{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"0123456789ab"}}}`
	s, settled := scan(capped, delta)
	if !settled {
		t.Fatal("the output cap ends the run")
	}
	if obs := s.Finish(process.Result{Stopped: true}, 0); !obs.OutputLimit || obs.Failure != "" {
		t.Fatalf("%+v", obs)
	}
	s, _ = scan(req)
	if obs := s.Finish(process.Result{TimedOut: true}, 0); !obs.TimedOut || obs.Failure != "" {
		t.Fatalf("%+v", obs)
	}
}

func TestJudgeStructuredOutput(t *testing.T) {
	req := harness.Request{Role: harness.RoleJudge}
	s, _ := scan(req, `{"type":"result","result":"","structured_output":{"assertion_results":[{"passed":true}]},"usage":{"input_tokens":1,"output_tokens":2},"total_cost_usd":0.01}`)
	obs := s.Finish(process.Result{}, 0)
	if !strings.Contains(obs.Structured, "assertion_results") || obs.Usage.OutputTokens != 2 {
		t.Fatalf("%+v", obs)
	}
	// A pretty-printed answer spans several lines and is read as a whole.
	s, _ = scan(req, `{`, `"type":"result",`, `"result":"{\"assertion_results\":[]}"`, `}`)
	if obs := s.Finish(process.Result{}, 0); obs.Text != `{"assertion_results":[]}` {
		t.Fatalf("%+v", obs)
	}
	s, _ = scan(req, `{`, `"type":"result","is_error":true,"result":"no"`, `}`)
	if obs := s.Finish(process.Result{}, 0); obs.Failure != "no" {
		t.Fatalf("%+v", obs)
	}
}

func TestAgentSessionWithoutTarget(t *testing.T) {
	s, _ := scan(harness.Request{Role: harness.RoleAgent}, initLine, `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Edit","input":{"file_path":"c"}}]}}`)
	obs := s.Finish(process.Result{}, 0)
	if obs.TargetAvailable || len(obs.Events) != 2 {
		t.Fatalf("%+v", obs)
	}
}
