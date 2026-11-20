package processtest_test

import (
	"context"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/process"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/process/processtest"
)

func TestSequenceAnswersInOrderThenFails(t *testing.T) {
	fake := processtest.Sequence(
		processtest.Reply{Lines: []string{"a", "b"}, ExitCode: 3, Stderr: "e"},
		processtest.Reply{TimedOut: true},
	)
	var seen []string
	res, err := fake.Run(context.Background(), process.Request{Name: "x"}, func(l string) bool { seen = append(seen, l); return false })
	if err != nil || res.ExitCode != 3 || res.Stderr != "e" || len(seen) != 2 {
		t.Fatalf("%v %+v %v", err, res, seen)
	}
	res, err = fake.Run(context.Background(), process.Request{Name: "y"}, nil)
	if err != nil || !res.TimedOut {
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
	fake := processtest.Sequence(processtest.Reply{Lines: []string{"a", "b", "c"}})
	count := 0
	res, err := fake.Run(context.Background(), process.Request{}, func(string) bool { count++; return count == 2 })
	if err != nil || !res.Stopped || count != 2 {
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
