package herdrcore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

const (
	script  = "/tmp/xonovex-agent-terminal-1.sh"
	runID   = "20261120T120000Z"
	started = `{"address":"t7","labels":{},"argv":["sh"],"cwd":"/work","state":"live","created_ms":1,"screen":"exact","cols":200,"rows":120}`
	exited3 = `{"terminal":"t7","ended":true,"reason":"exit","exit_code":3,"turn":0,"children":[]}`
	checkIn = `{"terminal":"t7","ended":false,"reason":"timeout","turn":0,"children":[]}`
	idle    = `{"terminal":"t7","ended":true,"reason":"status","status":"idle","turn":1,"children":[]}`
	live    = `{"address":"t7","state":"live","exit_code":null}`
)

// scriptedHost answers repeated command lines from a queue of outcomes, so a
// wait that is asked again answers anew; a line without a queue falls back to
// the outcomes of the Fake.
type scriptedHost struct {
	*oshosttest.Fake
	queue    map[string][]oshosttest.Outcome
	sleepErr error
}

var quiet = logging.New(io.Discard, io.Discard, false, false)

func newHost() *scriptedHost {
	fake := oshosttest.New()
	fake.Env["HC_TERMINAL_ID"] = "t1"
	return &scriptedHost{Fake: fake, queue: map[string][]oshosttest.Outcome{}}
}

func (h *scriptedHost) Output(ctx context.Context, name string, args []string, dir string) (string, string, int, error) {
	line := oshosttest.CommandLine(name, args...)
	if queued := h.queue[line]; len(queued) > 0 {
		h.Outcomes[line], h.queue[line] = queued[0], queued[1:]
	}
	return h.Fake.Output(ctx, name, args, dir)
}

func (h *scriptedHost) Sleep(ctx context.Context, d time.Duration) error {
	if err := h.Fake.Sleep(ctx, d); err != nil {
		return err
	}
	return h.sleepErr
}

func (h *scriptedHost) answer(line string, outcomes ...oshosttest.Outcome) {
	h.queue[line] = append(h.queue[line], outcomes...)
}

func (h *scriptedHost) calls() []string {
	var lines []string
	for _, call := range h.Calls {
		lines = append(lines, call.Line())
	}
	return lines
}

func firstName(h *scriptedHost) string {
	return "agent-cli-" + strconv.FormatInt(h.Clock.UnixNano(), 36)
}

func runLine(name string, extra ...string) string {
	args := append([]string{"hc", "run", "--bg", "--json", "--name", name, "--cwd", "/work", "--label", "agent-cli.run=" + runID}, extra...)
	return strings.Join(append(args, "--", "sh", script), " ")
}

func attached() *termshared.TerminalConfig { return &termshared.TerminalConfig{RunID: runID} }

func execute(h *scriptedHost, config *termshared.TerminalConfig) (termshared.Outcome, string, error) {
	var log bytes.Buffer
	outcome, err := NewExecutor(h, logging.New(&log, &log, true, false)).Execute(config, []string{"claude", "--model", "opus"}, []string{"TOKEN=s3cret"}, "/work")
	return outcome, log.String(), err
}

func TestExecutorIsAvailableWhenHcResolves(t *testing.T) {
	h := newHost()
	if NewExecutor(h, quiet).IsAvailable() {
		t.Error("IsAvailable() = true without hc")
	}
	h.AddExecutable("/usr/bin/hc")
	h.Env["PATH"] = "/usr/bin"
	if !NewExecutor(h, quiet).IsAvailable() {
		t.Error("IsAvailable() = false with hc on the path")
	}
}

func TestExecutorIsInsideOnlyInAHerdrCoreTerminal(t *testing.T) {
	h := newHost()
	executor := NewExecutor(h, quiet)
	if !executor.IsInside() {
		t.Error("IsInside() = false with HC_TERMINAL_ID set")
	}
	delete(h.Env, "HC_TERMINAL_ID")
	if executor.IsInside() {
		t.Error("IsInside() = true without HC_TERMINAL_ID")
	}
}

func TestExecuteRefusesOutsideWithoutTouchingTheRuntime(t *testing.T) {
	h := newHost()
	delete(h.Env, "HC_TERMINAL_ID")

	outcome, _, err := execute(h, attached())

	if !errors.Is(err, ErrOutside) || outcome != (termshared.Outcome{Code: 1, Ended: true}) || len(h.Calls) != 0 || h.Exists(script) {
		t.Errorf("Execute() = (%+v, %v) with calls %v, want the refusal and nothing started", outcome, err, h.calls())
	}
}

func TestExecuteWaitsForTheProgramAndReturnsItsExitCode(t *testing.T) {
	h := newHost()
	name := firstName(h)
	h.answer(runLine(name, "--no-notify"), oshosttest.Outcome{Stdout: started})
	h.answer("hc wait --json t7", oshosttest.Outcome{Stdout: checkIn}, oshosttest.Outcome{Stdout: exited3, ExitCode: 3})
	h.answer("hc rm --json t7", oshosttest.Outcome{Stdout: `{"removed":"t7"}`})

	outcome, log, err := execute(h, attached())

	if err != nil || outcome != (termshared.Outcome{Code: 3, Ended: true}) {
		t.Fatalf("Execute() = (%+v, %v), want the exit code of the ended program", outcome, err)
	}
	want := []string{runLine(name, "--no-notify"), "hc wait --json t7", "hc wait --json t7", "hc rm --json t7"}
	if !slices.Equal(h.calls(), want) {
		t.Errorf("calls = %v, want %v", h.calls(), want)
	}
	if strings.Contains(strings.Join(h.calls(), "\n"), "s3cret") || strings.Contains(log, "s3cret") {
		t.Error("a secret reached an hc command line or the log")
	}
	content, _ := h.File(script)
	if !strings.Contains(content, "export TOKEN='s3cret'") || !strings.Contains(content, "exec 'claude' '--model' 'opus'") {
		t.Errorf("launch script = %q", content)
	}
	if !strings.Contains(log, "herdr-core terminal t7") {
		t.Errorf("log = %q, want the terminal named", log)
	}
}

func TestExecuteReturnsAtTheStartWhenDetached(t *testing.T) {
	h := newHost()
	name := firstName(h)
	h.answer(runLine(name, "--rm", "--outlive-parent"), oshosttest.Outcome{Stdout: started})

	outcome, _, err := execute(h, &termshared.TerminalConfig{RunID: runID, Detach: true})

	if err != nil || outcome != (termshared.Outcome{}) || len(h.Calls) != 1 {
		t.Errorf("Execute() = (%+v, %v) with calls %v, want one start, no wait and a program that runs on", outcome, err, h.calls())
	}
}

func TestExecuteOmitsTheDirectoryAndTheLabelWhenThereAreNone(t *testing.T) {
	h := newHost()
	line := "hc run --bg --json --name " + firstName(h) + " --no-notify -- sh " + script
	h.answer(line, oshosttest.Outcome{Stdout: started})
	h.answer("hc wait --json t7", oshosttest.Outcome{Stdout: exited3})
	h.answer("hc rm --json t7", oshosttest.Outcome{})

	outcome, err := NewExecutor(h, quiet).Execute(&termshared.TerminalConfig{}, []string{"claude"}, nil, "")

	if err != nil || outcome.Code != 3 {
		t.Errorf("Execute() = (%+v, %v) with calls %v", outcome, err, h.calls())
	}
}

func TestExecuteWaitsOnAProgramThatStoppedWorkingWithAGrowingPause(t *testing.T) {
	h := newHost()
	h.answer(runLine(firstName(h), "--no-notify"), oshosttest.Outcome{Stdout: started})
	blocked := `{"terminal":"t7","ended":true,"reason":"status","status":"blocked","turn":1,"children":[]}`
	failed := `{"terminal":"t7","ended":true,"reason":"status","status":"error","turn":1,"children":[]}`
	h.answer("hc wait --json t7",
		oshosttest.Outcome{Stdout: idle}, oshosttest.Outcome{Stdout: idle}, oshosttest.Outcome{Stdout: idle},
		oshosttest.Outcome{Stdout: blocked}, oshosttest.Outcome{Stdout: failed}, oshosttest.Outcome{Stdout: failed},
		oshosttest.Outcome{Stdout: exited3})
	h.answer("hc get --json t7", oshosttest.Outcome{Stdout: live})
	h.answer("hc rm --json t7", oshosttest.Outcome{})

	outcome, _, err := execute(h, attached())

	if err != nil || outcome.Code != 3 {
		t.Fatalf("Execute() = (%+v, %v)", outcome, err)
	}
	want := []time.Duration{0, firstPause, 2 * firstPause, 0, 0, firstPause}
	if !slices.Equal(h.Sleeps, want) {
		t.Errorf("pauses = %v, want %v: blocked and error wait like idle", h.Sleeps, want)
	}
}

func TestNextPauseStopsGrowingAtTheLongestPause(t *testing.T) {
	if got := nextPause(longestPause, true); got != longestPause {
		t.Errorf("nextPause() = %v, want %v", got, longestPause)
	}
}

func TestExecuteTakesTheExitCodeOfATerminalThatFinishedBetweenWaits(t *testing.T) {
	h := newHost()
	h.answer(runLine(firstName(h), "--no-notify"), oshosttest.Outcome{Stdout: started})
	h.answer("hc wait --json t7", oshosttest.Outcome{Stdout: idle})
	h.answer("hc get --json t7", oshosttest.Outcome{Stdout: `{"address":"t7","state":"finished","exit_code":null,"signal":15}`})
	h.answer("hc rm --json t7", oshosttest.Outcome{})

	if outcome, _, err := execute(h, attached()); err != nil || outcome.Code != 143 {
		t.Errorf("Execute() = (%+v, %v), want 128 plus the signal", outcome, err)
	}
}

func TestExecuteEndsAProgramThatExitedWithoutACode(t *testing.T) {
	h := newHost()
	h.answer(runLine(firstName(h), "--no-notify"), oshosttest.Outcome{Stdout: started})
	h.answer("hc wait --json t7", oshosttest.Outcome{Stdout: `{"terminal":"t7","ended":true,"reason":"exit","turn":0,"children":[]}`})
	h.answer("hc kill --json t7", oshosttest.Outcome{})
	h.answer("hc wait --json --timeout 10 t7", oshosttest.Outcome{})
	h.answer("hc rm --json t7", oshosttest.Outcome{})

	outcome, _, err := execute(h, attached())

	if err == nil || outcome.Code != 1 || !strings.Contains(err.Error(), "ended without an exit code") {
		t.Errorf("Execute() = (%+v, %v), want the missing code reported", outcome, err)
	}
	if !slices.Contains(h.calls(), "hc rm --json t7") || h.Exists(script) {
		t.Errorf("calls = %v, want the terminal removed and the script gone", h.calls())
	}
}

func TestExecuteRetriesATakenNameOnce(t *testing.T) {
	h := newHost()
	name := firstName(h)
	taken := oshosttest.Outcome{Stderr: `{"error":{"kind":"name_taken","message":"name taken"}}`, ExitCode: 3}
	h.answer(runLine(name, "--rm", "--outlive-parent"), taken)
	h.answer(runLine(name+"-2", "--rm", "--outlive-parent"), oshosttest.Outcome{Stdout: started})

	if outcome, _, err := execute(h, &termshared.TerminalConfig{RunID: runID, Detach: true}); err != nil || outcome.Code != 0 {
		t.Fatalf("Execute() = (%+v, %v), want the second name to start", outcome, err)
	}

	h = newHost()
	h.answer(runLine(name, "--rm", "--outlive-parent"), taken)
	h.answer(runLine(name+"-2", "--rm", "--outlive-parent"), taken)
	outcome, _, err := execute(h, &termshared.TerminalConfig{RunID: runID, Detach: true})
	if err == nil || outcome.Code != 1 || !strings.Contains(err.Error(), "name taken") || len(h.Calls) != 2 || h.Exists(script) {
		t.Errorf("Execute() = (%+v, %v) with calls %v, want one retry then the failure", outcome, err, h.calls())
	}
}

func TestExecuteReportsAStartFailureAndRemovesTheScript(t *testing.T) {
	cases := map[string]oshosttest.Outcome{
		"hc missing":   {Err: errors.New("executable file not found")},
		"plain error":  {Stderr: "no core", ExitCode: 2},
		"api error":    {Stdout: `{"error":{"message":"","data":{"kind":"blocked"}}}`, ExitCode: 3},
		"silent error": {ExitCode: 4},
	}
	for name, reply := range cases {
		h := newHost()
		h.answer(runLine(firstName(h), "--no-notify"), reply)

		outcome, _, err := execute(h, attached())

		if err == nil || outcome != (termshared.Outcome{Code: 1, Ended: true}) || !strings.HasPrefix(err.Error(), "hc run: ") || h.Exists(script) {
			t.Errorf("%s: Execute() = (%+v, %v), want the start failure and no script", name, outcome, err)
		}
	}
}

func TestParseFailureNamesTheKindAndTheMessage(t *testing.T) {
	cases := map[string]failure{
		`{"error":{"kind":"not_found","message":"no terminal"}}`: {kind: "not_found", text: "no terminal"},
		`{"error":{"data":{"kind":"blocked"}}}`:                  {kind: "blocked", text: "blocked"},
		`{"error":{}}`:                                           {text: `{"error":{}}`},
		"plain words":                                            {text: "plain words"},
	}
	for stderr, want := range cases {
		if got := parseFailure("", stderr, 2); got != want {
			t.Errorf("parseFailure(%q) = %+v, want %+v", stderr, got, want)
		}
	}
	if got := parseFailure("", "", 4); got.text != "exit code 4" {
		t.Errorf("parseFailure(silent) = %+v", got)
	}
}

func TestExecuteEndsAStartedTerminalWhoseAddressIsUnreadable(t *testing.T) {
	h := newHost()
	name := firstName(h)
	h.answer(runLine(name, "--no-notify"), oshosttest.Outcome{Stdout: "started"})
	h.answer("hc kill --json "+name, oshosttest.Outcome{})
	h.answer("hc wait --json --timeout 10 "+name, oshosttest.Outcome{})
	h.answer("hc rm --json "+name, oshosttest.Outcome{})

	outcome, _, err := execute(h, attached())

	if err == nil || outcome != (termshared.Outcome{Code: 1}) || !strings.Contains(err.Error(), "named no terminal") {
		t.Errorf("Execute() = (%+v, %v), want the unreadable reply", outcome, err)
	}
	if !slices.Contains(h.calls(), "hc kill --json "+name) || !slices.Contains(h.calls(), "hc rm --json "+name) || h.Exists(script) {
		t.Errorf("calls = %v, want the terminal ended by its name", h.calls())
	}
}

func TestExecuteEndsTheTerminalOnAnInterruptDuringTheStart(t *testing.T) {
	h := newHost()
	name := firstName(h)
	h.answer(runLine(name, "--no-notify"), oshosttest.Outcome{Stdout: started})
	h.InterruptOn = runLine(name, "--no-notify")

	outcome, log, err := execute(h, attached())

	if err == nil || outcome != (termshared.Outcome{Code: 1}) || h.Exists(script) {
		t.Errorf("Execute() = (%+v, %v), want the interrupted start reported and no end confirmed", outcome, err)
	}
	if !slices.Contains(h.calls(), "hc kill --json "+name) || !strings.Contains(log, "Could not remove herdr-core terminal "+name) {
		t.Errorf("calls = %v, log = %q, want the terminal ended by its name and the failed removal reported", h.calls(), log)
	}
}

func TestExecuteEndsTheProgramOnAnInterruptDuringTheWait(t *testing.T) {
	h := newHost()
	h.answer(runLine(firstName(h), "--no-notify"), oshosttest.Outcome{Stdout: started})
	h.answer("hc wait --json t7", oshosttest.Outcome{})
	h.answer("hc kill --json t7", oshosttest.Outcome{})
	h.answer("hc wait --json --timeout 10 t7", oshosttest.Outcome{Stdout: exited3})
	h.answer("hc rm --json t7", oshosttest.Outcome{})
	h.InterruptOn = "hc wait --json t7"

	outcome, log, err := execute(h, attached())

	if err != nil || outcome != (termshared.Outcome{Code: 130, Ended: true}) {
		t.Fatalf("Execute() = (%+v, %v), want the interrupted code and the confirmed end", outcome, err)
	}
	tail := h.calls()[2:]
	if !slices.Equal(tail, []string{"hc kill --json t7", "hc wait --json --timeout 10 t7", "hc rm --json t7"}) {
		t.Errorf("calls after the interrupt = %v, want kill, a bounded wait and rm", tail)
	}
	if !strings.Contains(log, "Interrupted") || h.Exists(script) {
		t.Errorf("log = %q, want the interrupt reported and the script gone", log)
	}
}

func TestExecuteEndsTheProgramWhenTheWaitFails(t *testing.T) {
	cases := map[string]func(h *scriptedHost){
		"wait error": func(h *scriptedHost) {
			h.answer("hc wait --json t7", oshosttest.Outcome{Err: errors.New("broken pipe")})
		},
		"wait reply": func(h *scriptedHost) {
			h.answer("hc wait --json t7", oshosttest.Outcome{Stderr: `{"error":{"kind":"not_found","message":"no terminal t7"}}`, ExitCode: 3})
		},
		"get error": func(h *scriptedHost) {
			h.answer("hc wait --json t7", oshosttest.Outcome{Stdout: idle})
			h.answer("hc get --json t7", oshosttest.Outcome{Err: errors.New("broken pipe")})
		},
		"get failure": func(h *scriptedHost) {
			h.answer("hc wait --json t7", oshosttest.Outcome{Stdout: idle})
			h.answer("hc get --json t7", oshosttest.Outcome{Stderr: "gone", ExitCode: 3})
		},
		"get reply": func(h *scriptedHost) {
			h.answer("hc wait --json t7", oshosttest.Outcome{Stdout: idle})
			h.answer("hc get --json t7", oshosttest.Outcome{Stdout: "{}"})
		},
		"pause": func(h *scriptedHost) {
			h.answer("hc wait --json t7", oshosttest.Outcome{Stdout: idle})
			h.answer("hc get --json t7", oshosttest.Outcome{Stdout: live})
			h.sleepErr = errors.New("clock gone")
		},
	}
	for name, arrange := range cases {
		h := newHost()
		h.answer(runLine(firstName(h), "--no-notify"), oshosttest.Outcome{Stdout: started})
		arrange(h)
		h.answer("hc kill --json t7", oshosttest.Outcome{})
		h.answer("hc wait --json --timeout 10 t7", oshosttest.Outcome{})
		h.answer("hc rm --json t7", oshosttest.Outcome{})

		outcome, _, err := execute(h, attached())

		if err == nil || outcome != (termshared.Outcome{Code: 1}) || !slices.Contains(h.calls(), "hc kill --json t7") || h.Exists(script) {
			t.Errorf("%s: Execute() = (%+v, %v) with calls %v, want the failure, the program ended and no end confirmed", name, outcome, err, h.calls())
		}
	}
}

func TestExecuteWarnsWhenTheFinishedTerminalCannotBeRemoved(t *testing.T) {
	h := newHost()
	h.answer(runLine(firstName(h), "--no-notify"), oshosttest.Outcome{Stdout: started})
	h.answer("hc wait --json t7", oshosttest.Outcome{Stdout: exited3})
	h.answer("hc rm --json t7", oshosttest.Outcome{Stderr: `{"error":{"kind":"busy","message":"still live"}}`, ExitCode: 3})

	outcome, log, err := execute(h, attached())

	if err != nil || outcome.Code != 3 || !strings.Contains(log, "Could not remove herdr-core terminal t7 (still live); remove it with hc kill t7 and hc rm t7") {
		t.Errorf("Execute() = (%+v, %v), log = %q, want the code and a warning", outcome, err, log)
	}
}

func TestExecuteReportsALaunchScriptFailure(t *testing.T) {
	h := newHost()

	outcome, err := NewExecutor(h, quiet).Execute(attached(), nil, nil, "/work")

	if err == nil || outcome != (termshared.Outcome{Code: 1, Ended: true}) || len(h.Calls) != 0 {
		t.Errorf("Execute() = (%+v, %v) with calls %v, want the missing command refused", outcome, err, h.calls())
	}
}
