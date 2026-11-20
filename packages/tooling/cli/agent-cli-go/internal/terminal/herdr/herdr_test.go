package herdr

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

var testLog = logging.New(io.Discard, io.Discard, false, false)

const splitReply = `{"id":"1","result":{"type":"pane_info","pane":{"pane_id":"w1-3","tab_id":"t1"}}}`

func insideHost() *systemtest.Fake {
	host := systemtest.New()
	host.Env["HERDR_ENV"] = "1"
	host.Outcomes[systemtest.CommandLine("herdr", "pane", "split", "--current", "--direction", "right", "--cwd", "/work")] = systemtest.Outcome{Stdout: splitReply}
	return host
}

func calls(host *systemtest.Fake) []string {
	var lines []string
	for _, call := range host.Calls {
		lines = append(lines, call.Line())
	}
	return lines
}

func TestExecutorIsAvailableWhenHerdrResolves(t *testing.T) {
	host := systemtest.New()
	if NewExecutor(host, testLog).IsAvailable() {
		t.Error("IsAvailable() = true without herdr")
	}
	host.AddExecutable("/usr/bin/herdr")
	host.Env["PATH"] = "/usr/bin"
	if !NewExecutor(host, testLog).IsAvailable() {
		t.Error("IsAvailable() = false with herdr on the path")
	}
}

func TestExecutorIsInsideOnlyWhenHerdrSaysSo(t *testing.T) {
	host := systemtest.New()
	for value, want := range map[string]bool{"": false, "0": false, "1": true} {
		host.Env["HERDR_ENV"] = value
		if got := NewExecutor(host, testLog).IsInside(); got != want {
			t.Errorf("IsInside() with HERDR_ENV=%q = %v, want %v", value, got, want)
		}
	}
}

func TestExecuteRefusesOutsideHerdrWithoutTouchingAnySession(t *testing.T) {
	host := systemtest.New()

	code, err := NewExecutor(host, testLog).Execute(&termshared.TerminalConfig{}, []string{"claude"}, nil, "/work")

	if !errors.Is(err, ErrOutside) || code != 1 {
		t.Errorf("Execute() = (%d, %v), want the outside refusal", code, err)
	}
	if len(host.Calls) != 0 {
		t.Errorf("calls = %v, want none: never fall back to the focused session", calls(host))
	}
}

func TestExecuteSplitsTheCurrentPaneAndRunsTheLaunchScript(t *testing.T) {
	host := insideHost()
	host.Outcomes[systemtest.CommandLine("herdr", "pane", "run", "w1-3", "sh '/tmp/xonovex-agent-terminal-1.sh'")] = systemtest.Outcome{}

	code, err := NewExecutor(host, testLog).Execute(&termshared.TerminalConfig{}, []string{"claude", "--model", "opus"}, []string{"TOKEN=s3cret"}, "/work")

	if err != nil || code != 0 {
		t.Fatalf("Execute() = (%d, %v)", code, err)
	}
	got := calls(host)
	if len(got) != 2 || strings.Contains(strings.Join(got, "\n"), "s3cret") {
		t.Errorf("calls = %v, want a split and a run with no secret on the command line", got)
	}
	if strings.Contains(strings.Join(got, "\n"), "--no-focus") {
		t.Error("an attached run must keep the focus behaviour of herdr")
	}
	content, _ := host.File("/tmp/xonovex-agent-terminal-1.sh")
	if !strings.Contains(content, "export TOKEN='s3cret'") || !strings.Contains(content, "exec 'claude' '--model' 'opus'") {
		t.Errorf("launch script = %q", content)
	}
}

func TestExecuteKeepsTheFocusOfTheUserWhenDetached(t *testing.T) {
	host := systemtest.New()
	host.Env["HERDR_ENV"] = "1"
	host.Outcomes[systemtest.CommandLine("herdr", "pane", "split", "--current", "--direction", "right", "--cwd", "/work", "--no-focus")] = systemtest.Outcome{Stdout: splitReply}
	host.Outcomes[systemtest.CommandLine("herdr", "pane", "run", "w1-3", "sh '/tmp/xonovex-agent-terminal-1.sh'")] = systemtest.Outcome{}

	if code, err := NewExecutor(host, testLog).Execute(&termshared.TerminalConfig{Detach: true}, []string{"claude"}, nil, "/work"); err != nil || code != 0 {
		t.Fatalf("Execute() = (%d, %v)", code, err)
	}
}

func TestExecuteOmitsTheDirectoryWhenThereIsNone(t *testing.T) {
	host := systemtest.New()
	host.Env["HERDR_ENV"] = "1"
	host.Outcomes[systemtest.CommandLine("herdr", "pane", "split", "--current", "--direction", "right")] = systemtest.Outcome{Stdout: splitReply}
	host.Outcomes[systemtest.CommandLine("herdr", "pane", "run", "w1-3", "sh '/tmp/xonovex-agent-terminal-1.sh'")] = systemtest.Outcome{}

	if _, err := NewExecutor(host, testLog).Execute(&termshared.TerminalConfig{}, []string{"claude"}, nil, ""); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestExecuteRemovesTheScriptWhenTheSplitFails(t *testing.T) {
	host := insideHost()
	host.Outcomes[systemtest.CommandLine("herdr", "pane", "split", "--current", "--direction", "right", "--cwd", "/work")] = systemtest.Outcome{Err: errors.New("no server")}

	code, err := NewExecutor(host, testLog).Execute(&termshared.TerminalConfig{}, []string{"claude"}, nil, "/work")

	if err == nil || code != 1 || !strings.Contains(err.Error(), "herdr pane split") {
		t.Errorf("Execute() = (%d, %v), want the split failure", code, err)
	}
	if host.Exists("/tmp/xonovex-agent-terminal-1.sh") {
		t.Error("the launch script remains after a failed split")
	}
}

func TestExecuteRejectsAReplyWithoutAPane(t *testing.T) {
	for name, reply := range map[string]string{"not json": "ok", "no pane": `{"result":{}}`} {
		host := systemtest.New()
		host.Env["HERDR_ENV"] = "1"
		host.Outcomes[systemtest.CommandLine("herdr", "pane", "split", "--current", "--direction", "right", "--cwd", "/work")] = systemtest.Outcome{Stdout: reply}

		if _, err := NewExecutor(host, testLog).Execute(&termshared.TerminalConfig{}, []string{"claude"}, nil, "/work"); err == nil || !strings.Contains(err.Error(), "named no pane") {
			t.Errorf("%s: Execute() error = %v, want a missing pane", name, err)
		}
	}
}

func TestExecuteClosesThePaneThatCannotRunTheCommand(t *testing.T) {
	host := insideHost()
	host.Outcomes[systemtest.CommandLine("herdr", "pane", "run", "w1-3", "sh '/tmp/xonovex-agent-terminal-1.sh'")] = systemtest.Outcome{Err: errors.New("pane gone")}

	code, err := NewExecutor(host, testLog).Execute(&termshared.TerminalConfig{}, []string{"claude"}, nil, "/work")

	if err == nil || code != 1 {
		t.Fatalf("Execute() = (%d, %v), want the run failure", code, err)
	}
	if !slices.Contains(calls(host), "herdr pane close w1-3") {
		t.Errorf("calls = %v, want the new pane closed", calls(host))
	}
	if host.Exists("/tmp/xonovex-agent-terminal-1.sh") {
		t.Error("the launch script remains after a failed run")
	}
}

func TestExecuteReportsALaunchScriptFailure(t *testing.T) {
	host := insideHost()

	if code, err := NewExecutor(host, testLog).Execute(&termshared.TerminalConfig{}, nil, nil, "/work"); err == nil || code != 1 {
		t.Errorf("Execute() = (%d, %v), want the missing command refused", code, err)
	}
	if len(calls(host)) != 0 {
		t.Errorf("calls = %v, want no herdr call without a script", calls(host))
	}
}
