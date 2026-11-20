package execute_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process/processtest"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/scratch/scratchtest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/execute"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
)

type leaf struct{ stages []harness.Stage }

func (leaf) Name() string   { return "leaf" }
func (leaf) Binary() string { return "leaf-bin" }
func (l leaf) Plan(req harness.Request, ws string) (harness.Plan, error) {
	if req.Prompt == "bad plan" {
		return harness.Plan{}, errors.New("no plan")
	}
	return harness.Plan{Args: []string{"-x"}, Prompt: req.Prompt, Dir: ws, Env: map[string]string{"K": "V"}, Stage: l.stages}, nil
}
func (leaf) Scan(harness.Request) harness.Scanner { return &scan{} }

type scan struct{ text string }

func (s *scan) Feed(line string) bool { s.text += line; return line == "stop" }
func (s *scan) Finish(exit process.Result, ms int64) harness.Observation {
	obs := harness.Observation{Text: s.text, DurationMs: ms}
	if exit.ExitCode != 0 {
		obs.Failure = "exit"
	}
	return obs
}

type fixture struct {
	exec    execute.Executor
	fs      *memfs.FileSystem
	scratch *scratchtest.Fake
	proc    *processtest.Fake
}

func build(l leaf, proc *processtest.Fake, files map[string]string) fixture {
	fsys := memfs.New(files)
	reg := harness.NewRegistry()
	reg.Register(l)
	sc := scratchtest.New()
	clock := time.Unix(0, 0)
	return fixture{
		exec: execute.Executor{
			Harnesses: reg, Runner: proc, Scratch: sc, FS: fsys,
			Now:   func() time.Time { clock = clock.Add(5 * time.Millisecond); return clock },
			Pause: func(int) {},
		},
		fs: fsys, scratch: sc, proc: proc,
	}
}

func TestExecuteStagesRunsAndCleans(t *testing.T) {
	stages := []harness.Stage{
		{Kind: harness.StageDir, Dst: "/scratch/agent-eval-1/home"},
		{Kind: harness.StageCopy, Src: "/skills/one", Dst: "/scratch/agent-eval-1/skills/one"},
		{Kind: harness.StageCopy, Src: "/skills/one/SKILL.md", Dst: "/scratch/agent-eval-1/lone.md"},
		{Kind: harness.StageAppend, Dst: "/scratch/agent-eval-1/skills/one/SKILL.md", Content: "+tail"},
		{Kind: harness.StageWrite, Dst: "/scratch/agent-eval-1/w.txt", Content: "w"},
	}
	f := build(leaf{stages}, processtest.Sequence(processtest.Reply{Stdout: processtest.Lines("a", "b")}),
		map[string]string{"/skills/one/SKILL.md": "skill", "/skills/one/ref/r.md": "ref"})
	obs, err := f.exec.Execute(context.Background(), harness.Request{
		Harness: "leaf", Prompt: "go", Fixtures: []harness.File{{Path: "in/a.txt", Content: "fx"}},
		Caps: harness.Caps{Timeout: time.Second},
	})
	if err != nil || obs.Text != "ab" || obs.DurationMs != 5 {
		t.Fatalf("%v %+v", err, obs)
	}
	got, _ := f.fs.ReadFile("/scratch/agent-eval-1/skills/one/SKILL.md")
	ref, _ := f.fs.ReadFile("/scratch/agent-eval-1/skills/one/ref/r.md")
	fx, _ := f.fs.ReadFile("/scratch/agent-eval-1/in/a.txt")
	lone, _ := f.fs.ReadFile("/scratch/agent-eval-1/lone.md")
	if string(got) != "skill+tail" || string(ref) != "ref" || string(fx) != "fx" || string(lone) != "skill" {
		t.Fatalf("%q %q %q %q", got, ref, fx, lone)
	}
	call := f.proc.Calls()[0]
	if call.Name != "leaf-bin" || strings.Join(call.Args, " ") != "-x -- go" || call.Env["K"] != "V" || call.Timeout != time.Second || call.Dir != "/scratch/agent-eval-1" {
		t.Fatalf("%+v", call)
	}
	if len(f.scratch.Removed()) != 1 {
		t.Fatal("the workspace must be removed")
	}
}

func TestExecuteRetriesThenSucceeds(t *testing.T) {
	f := build(leaf{}, processtest.Sequence(
		processtest.Reply{ExitCode: 1}, processtest.Reply{Err: errors.New("spawn")}, processtest.Reply{Stdout: processtest.Lines("ok")},
	), nil)
	obs, err := f.exec.Execute(context.Background(), harness.Request{Harness: "leaf"})
	if err != nil || obs.Text != "ok" || len(f.proc.Calls()) != 3 || len(f.scratch.Removed()) != 3 {
		t.Fatalf("%v %+v", err, obs)
	}
}

func TestExecuteVoidsTheBatchAfterThreeFailures(t *testing.T) {
	f := build(leaf{}, processtest.New(func(process.Request) processtest.Reply { return processtest.Reply{ExitCode: 1} }), nil)
	_, err := f.exec.Execute(context.Background(), harness.Request{Harness: "leaf"})
	var infra *execute.InfrastructureError
	if !errors.As(err, &infra) || infra.Attempts != 3 || !strings.Contains(err.Error(), "failed after 3 attempts: exit") {
		t.Fatalf("%v", err)
	}
}

func TestExecuteErrors(t *testing.T) {
	f := build(leaf{}, processtest.Sequence(), nil)
	if _, err := f.exec.Execute(context.Background(), harness.Request{Harness: "nope"}); err == nil {
		t.Fatal("unknown harness")
	}
	if _, err := f.exec.Execute(context.Background(), harness.Request{Harness: "leaf", Prompt: "bad plan"}); err == nil {
		t.Fatal("plan failure")
	}
	f.scratch.MakeErr = errors.New("disk")
	if _, err := f.exec.Execute(context.Background(), harness.Request{Harness: "leaf"}); err == nil {
		t.Fatal("scratch failure")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := build(leaf{}, processtest.Sequence(), nil)
	if _, err := g.exec.Execute(ctx, harness.Request{Harness: "leaf"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("%v", err)
	}
	h := build(leaf{}, processtest.New(func(process.Request) processtest.Reply { return processtest.Reply{Err: context.Canceled} }), nil)
	if _, err := h.exec.Execute(context.Background(), harness.Request{Harness: "leaf"}); err == nil {
		t.Fatal("a program that never starts voids the batch")
	}
}

func TestStageFailures(t *testing.T) {
	for name, stage := range map[string]harness.Stage{
		"append missing": {Kind: harness.StageAppend, Dst: "/nowhere"},
		"copy missing":   {Kind: harness.StageCopy, Src: "/nowhere", Dst: "/x"},
		"unknown kind":   {Kind: "weird", Dst: "/x"},
	} {
		f := build(leaf{[]harness.Stage{stage}}, processtest.Sequence(), nil)
		if _, err := f.exec.Execute(context.Background(), harness.Request{Harness: "leaf"}); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

type agentLeaf struct{ leaf }

func (agentLeaf) Plan(req harness.Request, ws string) (harness.Plan, error) {
	return harness.Plan{Prompt: "p", Dir: ws, Stage: []harness.Stage{
		{Kind: harness.StageWrite, Dst: ws + "/CLAUDE.md", Content: "rules"},
		{Kind: harness.StageDir, Dst: ws + "/home"},
	}}, nil
}

func TestAgentRunSnapshotsTheRunAndSkipsTheStagedSetup(t *testing.T) {
	var f fixture
	proc := processtest.New(func(req process.Request) processtest.Reply {
		_ = f.fs.WriteFile(req.Dir+"/out/a.txt", []byte("A"))
		_ = f.fs.WriteFile(req.Dir+"/big.txt", []byte(strings.Repeat("x", 1<<20+1)))
		_ = f.fs.WriteFile(req.Dir+"/home/h.txt", []byte("H"))
		return processtest.Reply{}
	})
	f = build(leaf{}, proc, nil)
	reg := harness.NewRegistry()
	reg.Register(agentLeaf{})
	f.exec.Harnesses = reg
	obs, err := f.exec.Execute(context.Background(), harness.Request{Harness: "leaf", Role: harness.RoleAgent, Fixtures: []harness.File{{Path: "f.txt", Content: "F"}}})
	if err != nil || obs.Files["out/a.txt"] != "A" || obs.Files["f.txt"] != "F" {
		t.Fatalf("%v %+v", err, obs.Files)
	}
	for _, left := range []string{"big.txt", "CLAUDE.md", "home/h.txt"} {
		if _, ok := obs.Files[left]; ok {
			t.Errorf("%s must be left out of the snapshot", left)
		}
	}
	g := build(leaf{}, processtest.Sequence(processtest.Reply{}), nil)
	plain, _ := g.exec.Execute(context.Background(), harness.Request{Harness: "leaf", Role: harness.RoleJudge})
	if plain.Files != nil {
		t.Fatal("a judge run has no snapshot")
	}
}
