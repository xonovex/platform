package processtest_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process/processtest"
)

func TestSequenceAnswersInOrderThenFails(t *testing.T) {
	fake := processtest.Sequence(
		processtest.Reply{Stdout: processtest.Lines("a", "b"), ExitCode: 3, Stderr: "e"},
		processtest.Reply{TimedOut: true},
	)
	var seen []string
	res, err := fake.Run(context.Background(), process.Request{Name: "x"}, process.Lines(func(l string) bool {
		seen = append(seen, l)
		return false
	}))
	if err != nil || res.ExitCode != 3 || res.Stderr != "e" || len(seen) != 2 || seen[1] != "b" {
		t.Fatalf("%v %+v %v", err, res, seen)
	}
	res, err = fake.Run(context.Background(), process.Request{Name: "y"}, nil)
	if err != nil || !res.TimedOut || res.Succeeded() {
		t.Fatalf("%v %+v", err, res)
	}
	if _, err = fake.Run(context.Background(), process.Request{}, nil); err == nil {
		t.Fatal("an unscripted run must fail")
	}
	if calls := fake.Calls(); len(calls) != 3 || calls[0].Name != "x" {
		t.Fatalf("%+v", calls)
	}
}

func TestStopEndsTheRunEarly(t *testing.T) {
	fake := processtest.Sequence(processtest.Reply{Stdout: processtest.Lines("a", "b", "c"), Stderr: "e"})
	count := 0
	res, err := fake.Run(context.Background(), process.Request{}, process.Lines(func(string) bool { count++; return count == 2 }))
	if err != nil || !res.Stopped || res.Stderr != "e" || count != 2 {
		t.Fatalf("%v %+v %d", err, res, count)
	}
}

func TestReplyErrorMeansTheProgramDidNotStart(t *testing.T) {
	want := context.Canceled
	fake := processtest.New(func(process.Request) processtest.Reply { return processtest.Reply{Err: want} })
	if _, err := fake.Run(context.Background(), process.Request{}, nil); err != want {
		t.Fatalf("%v", err)
	}
}

func TestScriptAnswersACommandLineAndRecordsIt(t *testing.T) {
	fake := processtest.Script().On("git status", processtest.Reply{Stdout: "clean\nno end", ExitCode: 3})
	got, err := process.Capture(context.Background(), fake, process.Request{Dir: "d", Name: "git", Args: []string{"status"}})
	if err != nil || string(got.Stdout) != "clean\nno end" || got.ExitCode != 3 || got.Succeeded() {
		t.Fatalf("Capture = %+v, %v", got, err)
	}
	if calls := fake.Calls(); len(calls) != 1 || calls[0].Dir != "d" {
		t.Errorf("Calls = %+v", calls)
	}
}

func TestScriptFailsAnUnscriptedCommand(t *testing.T) {
	if _, err := processtest.Script().Run(context.Background(), process.Request{Name: "ls"}, nil); err == nil {
		t.Error("an unscripted command must be an error")
	}
}

func TestScriptComesBeforeTheResponder(t *testing.T) {
	fake := processtest.New(func(process.Request) processtest.Reply { return processtest.Reply{ExitCode: 9} }).
		On("tool", processtest.Reply{ExitCode: 1})
	scripted, _ := fake.Run(context.Background(), process.Request{Name: "tool"}, nil)
	other, _ := fake.Run(context.Background(), process.Request{Name: "other"}, nil)
	if scripted.ExitCode != 1 || other.ExitCode != 9 {
		t.Fatalf("scripted = %+v, other = %+v", scripted, other)
	}
}

func TestScriptReturnsAScriptedFault(t *testing.T) {
	boom := errors.New("no such program")
	fake := processtest.Script().On("tool", processtest.Reply{Err: boom})
	if _, err := fake.Run(context.Background(), process.Request{Name: "tool"}, nil); !errors.Is(err, boom) {
		t.Errorf("err = %v", err)
	}
}

func TestKeyAndLines(t *testing.T) {
	if got := processtest.Key("git", "tag", "-l"); got != "git tag -l" {
		t.Errorf("Key = %q", got)
	}
	if got := processtest.Key("git"); got != "git" {
		t.Errorf("Key = %q", got)
	}
	if got := processtest.Lines("a", "b"); got != "a\nb\n" {
		t.Errorf("Lines = %q", got)
	}
}
