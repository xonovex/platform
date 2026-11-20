package codex_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness/codex"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/process"
)

func offer() []harness.Offer {
	return []harness.Offer{{Name: "p", Dir: "/p", SkillDirs: []string{"/p/skills/one", "/p/skills/two"}}}
}

func argsLine(a []string) string { return strings.Join(a, " ") }

func TestNameAndBinary(t *testing.T) {
	if codex.New().Name() != "codex" || codex.New().Binary() != "codex" {
		t.Fatal("name")
	}
}

func TestTriggerPlanStagesProjectSkillsAndSignalsOnlyTheTarget(t *testing.T) {
	plan, err := codex.New().Plan(harness.Request{
		Role: harness.RoleTrigger, Model: "m", Prompt: "q", Target: "one", Offer: offer(), Effort: model.Effort("high"),
	}, "/ws")
	if err != nil {
		t.Fatal(err)
	}
	line := argsLine(plan.Args)
	for _, want := range []string{"exec --json --ephemeral", "--sandbox read-only", "--ignore-user-config", "--ignore-rules", "--model m", `model_reasoning_effort="high"`} {
		if !strings.Contains(line, want) {
			t.Errorf("missing %q in %s", want, line)
		}
	}
	if plan.Env["HOME"] != "/ws/home" || plan.Env["CODEX_HOME"] != "/ws/.codex" || plan.Dir != "/ws" {
		t.Fatalf("%+v", plan)
	}
	if !strings.Contains(plan.Prompt, "NOT_APPLICABLE") || !strings.HasPrefix(plan.Prompt, "q") {
		t.Fatal(plan.Prompt)
	}
	var appended []string
	var copied []string
	for _, s := range plan.Stage {
		switch s.Kind {
		case harness.StageAppend:
			appended = append(appended, s.Dst)
			if !strings.Contains(s.Content, codex.TriggerSignal) {
				t.Fatal(s.Content)
			}
		case harness.StageCopy:
			copied = append(copied, s.Dst)
		}
	}
	if len(appended) != 1 || appended[0] != "/ws/.agents/skills/one/SKILL.md" || len(copied) != 2 {
		t.Fatalf("%v %v", appended, copied)
	}
}

func TestGeneratePlanSkillGoesInTheIsolatedHome(t *testing.T) {
	with, err := codex.New().Plan(harness.Request{Role: harness.RoleGenerate, Prompt: "do", Target: "one", Offer: offer(), WithSkill: true}, "/ws")
	if err != nil {
		t.Fatal(err)
	}
	if with.Prompt != "$one\n\ndo" {
		t.Fatal(with.Prompt)
	}
	var appended string
	for _, s := range with.Stage {
		if s.Kind == harness.StageAppend {
			appended = s.Dst
			if !strings.Contains(s.Content, codex.UseSignal) {
				t.Fatal(s.Content)
			}
		}
	}
	if appended != "/ws/home/.agents/skills/one/SKILL.md" {
		t.Fatal(appended)
	}
	without, _ := codex.New().Plan(harness.Request{Role: harness.RoleGenerate, Prompt: "do"}, "/ws")
	if without.Prompt != "do" || len(without.Stage) != 2 {
		t.Fatalf("%+v", without)
	}
}

func TestJudgeAndAgentPlans(t *testing.T) {
	judge, _ := codex.New().Plan(harness.Request{Role: harness.RoleJudge, Prompt: "g"}, "/ws")
	if !strings.Contains(argsLine(judge.Args), "read-only") {
		t.Fatal(judge.Args)
	}
	agent, _ := codex.New().Plan(harness.Request{Role: harness.RoleAgent, Prompt: "w", Instructions: "rules"}, "/ws")
	if !strings.Contains(argsLine(agent.Args), "workspace-write") {
		t.Fatal(agent.Args)
	}
	last := agent.Stage[len(agent.Stage)-1]
	if last.Dst != "/ws/AGENTS.md" || last.Content != "rules" {
		t.Fatalf("%+v", last)
	}
	bare, _ := codex.New().Plan(harness.Request{Role: harness.RoleAgent}, "/ws")
	if len(bare.Stage) != 2 {
		t.Fatal("no instructions, no file")
	}
	if _, err := codex.New().Plan(harness.Request{Role: "x"}, "/ws"); err == nil {
		t.Fatal("unknown role")
	}
}

func message(text string) string {
	return `{"type":"item.completed","item":{"type":"agent_message","text":"` + text + `"}}`
}

func run(req harness.Request, lines ...string) (harness.Scanner, bool) {
	s := codex.New().Scan(req)
	for _, l := range lines {
		if s.Feed(l) {
			return s, true
		}
	}
	return s, false
}

func TestTriggerOutcomes(t *testing.T) {
	req := harness.Request{Role: harness.RoleTrigger}
	s, settled := run(req, "garbage", message(codex.TriggerSignal))
	if !settled {
		t.Fatal("the signal settles the run")
	}
	if obs := s.Finish(process.Result{Stopped: true}, 3); obs.Selected != "target" || !obs.TargetInvoked {
		t.Fatalf("%+v", obs)
	}
	if !s.Feed("x") {
		t.Fatal("stays settled")
	}
	s, _ = run(req, message("NOT_APPLICABLE"))
	if obs := s.Finish(process.Result{}, 0); obs.Selected != "none" {
		t.Fatalf("%+v", obs)
	}
	s, _ = run(req)
	if obs := s.Finish(process.Result{TimedOut: true}, 0); obs.Failure != "timeout" {
		t.Fatalf("%+v", obs)
	}
	s, _ = run(req)
	if obs := s.Finish(process.Result{ExitCode: 2, Stderr: "bad"}, 0); obs.Failure != "codex exited 2: bad" {
		t.Fatalf("%+v", obs)
	}
	s, _ = run(req)
	if obs := s.Finish(process.Result{ExitCode: 2}, 0); obs.Failure != "codex exited 2" {
		t.Fatalf("%+v", obs)
	}
	s, settled = run(req, message(strings.Repeat("x", 2100)))
	if !settled {
		t.Fatal("a long answer settles the run")
	}
	if obs := s.Finish(process.Result{Stopped: true}, 0); obs.Selected != "output-limit" || obs.Failure != "" || !obs.OutputLimit {
		t.Fatalf("%+v", obs)
	}
}

func TestGenerateStripsTheUseSignalAndReadsUsage(t *testing.T) {
	req := harness.Request{Role: harness.RoleGenerate, WithSkill: true}
	s, _ := run(req,
		`{"type":"item.completed","item":{"type":"command_execution","command":"ls"}}`,
		`{"type":"item.completed","item":{"type":"file_change","changes":[{"path":"a.md"},{"path":"b.md"}]}}`,
		`{"type":"item.completed","item":{"type":"reasoning"}}`,
		message(codex.UseSignal+"\\n\\nthe answer"),
		`{"type":"turn.completed","usage":{"input_tokens":10,"output_tokens":4}}`,
	)
	obs := s.Finish(process.Result{}, 77)
	if !obs.TargetInvoked || obs.Text != "the answer" || obs.DurationMs != 77 {
		t.Fatalf("%+v", obs)
	}
	if obs.Usage != (model.Usage{InputTokens: 10, OutputTokens: 4}) {
		t.Fatalf("%+v", obs.Usage)
	}
	kinds := map[string]int{}
	for _, e := range obs.Events {
		kinds[e.Kind]++
	}
	if kinds[harness.EventCommand] != 1 || kinds[harness.EventFileWrite] != 2 || kinds[harness.EventMessage] != 1 || kinds[harness.EventToolCall] != 2 {
		t.Fatalf("%v", kinds)
	}
	plain, _ := run(harness.Request{Role: harness.RoleGenerate, WithSkill: true}, message("no marker"))
	if obs := plain.Finish(process.Result{}, 0); obs.TargetInvoked || obs.Text != "no marker" {
		t.Fatalf("%+v", obs)
	}
}

func TestGenerateFailureCapAndJudge(t *testing.T) {
	req := harness.Request{Role: harness.RoleGenerate, Caps: harness.Caps{MaxOutputChars: 5}}
	s, settled := run(req, message("0123456789"))
	if !settled {
		t.Fatal("the cap ends the run")
	}
	if obs := s.Finish(process.Result{Stopped: true}, 0); !obs.OutputLimit || obs.Failure != "" {
		t.Fatalf("%+v", obs)
	}
	s, _ = run(harness.Request{Role: harness.RoleGenerate})
	if obs := s.Finish(process.Result{ExitCode: 1}, 0); obs.Failure != "codex exited 1" {
		t.Fatalf("%+v", obs)
	}
	s, _ = run(harness.Request{Role: harness.RoleGenerate})
	if obs := s.Finish(process.Result{TimedOut: true}, 0); !obs.TimedOut || obs.Failure != "" {
		t.Fatalf("%+v", obs)
	}
	judge, _ := run(harness.Request{Role: harness.RoleJudge}, message("verdict"))
	if obs := judge.Finish(process.Result{}, 0); obs.Structured != "verdict" {
		t.Fatalf("%+v", obs)
	}
}
