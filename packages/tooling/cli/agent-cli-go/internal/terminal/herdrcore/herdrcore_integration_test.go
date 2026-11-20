//go:build integration

package herdrcore

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

// insideRuntime returns the real host when the test runs in a herdr-core
// terminal with hc installed, and skips the test otherwise.
func insideRuntime(t *testing.T) oshost.System {
	t.Helper()
	host := oshost.New(os.Stdin, os.Stderr, os.Stderr)
	executor := NewExecutor(host, logging.New(os.Stderr, os.Stderr, false, false))
	if !executor.IsAvailable() || !executor.IsInside() {
		t.Skip("the test does not run in a herdr-core terminal with hc installed")
	}
	return host
}

// TestTheRepliesOfTheInstalledHcParse runs only hc commands that read: it starts,
// changes and ends no terminal.
func TestTheRepliesOfTheInstalledHcParse(t *testing.T) {
	host := insideRuntime(t)
	ctx := context.Background()

	stdout, stderr, code, err := host.Output(ctx, command, []string{"get", "--json", host.Getenv(insideVariable)}, "")
	if err != nil || code != 0 {
		t.Fatalf("hc get = (%d, %v): %s", code, err, stderr)
	}
	if state, err := parseTerminal(stdout); err != nil || state.State != "live" {
		t.Errorf("parseTerminal(own terminal) = (%+v, %v), want a live terminal", state, err)
	}

	missing := namePrefix + "missing-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	stdout, stderr, code, err = host.Output(ctx, command, []string{"get", "--json", missing}, "")
	if err != nil || code == 0 {
		t.Fatalf("hc get of a missing terminal = (%d, %v), want a failure", code, err)
	}
	if failure := parseFailure(stdout, stderr, code); failure.kind != "not_found" || failure.text == "" {
		t.Errorf("parseFailure(missing terminal) = %+v, want not_found with a message", failure)
	}

	stdout, stderr, code, err = host.Output(ctx, command, []string{"ls", "--all", "--json", "--label", runLabel + "=" + missing}, "")
	if err != nil || code != 0 {
		t.Fatalf("hc ls = (%d, %v): %s", code, err, stderr)
	}
	if terminals := labelled(t, stdout); len(terminals) != 0 {
		t.Errorf("hc ls --label of an unused run id = %v, want none", terminals)
	}
}

// TestExecuteRunsAProgramAndLeavesNoTerminal starts a terminal in the runtime the
// test runs in, so it runs only when AGENT_CLI_HC_SMOKE is 1.
func TestExecuteRunsAProgramAndLeavesNoTerminal(t *testing.T) {
	if os.Getenv("AGENT_CLI_HC_SMOKE") != "1" {
		t.Skip("AGENT_CLI_HC_SMOKE=1 allows the test to start a herdr-core terminal")
	}
	host := insideRuntime(t)
	executor := NewExecutor(host, logging.New(os.Stderr, os.Stderr, true, false))
	runID := "smoke-" + strconv.FormatInt(time.Now().UnixNano(), 36)

	code, err := executor.Execute(&termshared.TerminalConfig{RunID: runID}, []string{"sh", "-c", "exit 7"}, []string{"AGENT_CLI_SMOKE=1"}, t.TempDir())
	if err != nil || code != 7 {
		t.Errorf("Execute() = (%d, %v), want the exit code 7 of the program", code, err)
	}

	stdout, stderr, status, err := host.Output(context.Background(), command, []string{"ls", "--all", "--json", "--label", runLabel + "=" + runID}, "")
	if err != nil || status != 0 {
		t.Fatalf("hc ls = (%d, %v): %s", status, err, stderr)
	}
	if terminals := labelled(t, stdout); len(terminals) != 0 {
		t.Errorf("terminals left behind: %v", terminals)
	}
}

// labelled returns the addresses of the terminals an hc ls reply lists.
func labelled(t *testing.T, stdout string) []string {
	t.Helper()
	var reply struct {
		Terminals []struct {
			Address string `json:"address"`
		} `json:"terminals"`
	}
	if err := json.Unmarshal([]byte(stdout), &reply); err != nil {
		t.Fatalf("hc ls reply %q: %v", stdout, err)
	}
	addresses := make([]string, 0, len(reply.Terminals))
	for _, terminal := range reply.Terminals {
		addresses = append(addresses, terminal.Address)
	}
	return addresses
}
