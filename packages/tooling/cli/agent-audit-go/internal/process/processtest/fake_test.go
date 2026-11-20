package processtest_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/process"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/process/processtest"
)

func TestFakeAnswersAScriptedCommandAndRecordsIt(t *testing.T) {
	fake := processtest.New().On("git status", processtest.Reply{Stdout: "clean", ExitCode: 3})
	got, err := fake.Run(context.Background(), process.Command{Dir: "d", Name: "git", Args: []string{"status"}})
	if err != nil || string(got.Stdout) != "clean" || got.ExitCode != 3 || got.Succeeded() {
		t.Fatalf("Run = %+v, %v", got, err)
	}
	if len(fake.Calls) != 1 || fake.Calls[0].Dir != "d" {
		t.Errorf("Calls = %+v", fake.Calls)
	}
}

func TestFakeFailsAnUnscriptedCommand(t *testing.T) {
	if _, err := processtest.New().Run(context.Background(), process.Command{Name: "ls"}); err == nil {
		t.Error("an unscripted command must be an error")
	}
}

func TestFakeReturnsAScriptedFault(t *testing.T) {
	boom := errors.New("no such program")
	fake := processtest.New().On("tool", processtest.Reply{Err: boom})
	if _, err := fake.Run(context.Background(), process.Command{Name: "tool"}); !errors.Is(err, boom) {
		t.Errorf("err = %v", err)
	}
}

func TestKey(t *testing.T) {
	if got := processtest.Key("git", "tag", "-l"); got != "git tag -l" {
		t.Errorf("Key = %q", got)
	}
	if got := processtest.Key("git"); got != "git" {
		t.Errorf("Key = %q", got)
	}
}
