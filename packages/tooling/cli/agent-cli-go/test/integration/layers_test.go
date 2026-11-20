//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/record"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/schematest"
)

// machine is one user: a home directory that holds the configuration and the run
// records, a working directory, and a directory of fake agents first on the path.
// The environment holds nothing of the test process but the rest of the path, so
// no herdr or tmux variable reaches the CLI.
type machine struct {
	home, work, bin string
}

func newMachine(t *testing.T, agentCode string) machine {
	t.Helper()
	return machine{home: t.TempDir(), work: t.TempDir(), bin: fakeAgent(t, agentCode)}
}

func (m machine) stateDir() string { return filepath.Join(m.home, "state") }

func (m machine) command(args ...string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
	cmd := exec.Command(os.Getenv("AGENT_CLI_BINARY"), args...)
	cmd.Dir = m.work
	cmd.Env = []string{"HOME=" + m.home, "XDG_CONFIG_HOME=" + m.home, "XDG_STATE_HOME=" + m.stateDir(), "PATH=" + m.bin + string(os.PathListSeparator) + os.Getenv("PATH")}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	return cmd, &stdout, &stderr
}

// fakeAgent writes a claude script that exits with code, and returns its directory.
func fakeAgent(t *testing.T, code string) string {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\nexit " + code + "\n"
	if err := os.WriteFile(filepath.Join(dir, "claude"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func exitCode(err error) int {
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode()
	}
	return 0
}

func TestARunLeavesARecordThatSatisfiesTheSchema(t *testing.T) {
	m := newMachine(t, "3")
	run, _, _ := m.command("run")

	if code := exitCode(run.Run()); code != 3 {
		t.Fatalf("run exit code = %d, want the 3 of the agent", code)
	}

	show, stdout, stderr := m.command("runs", "show", "latest", "--json")
	if err := show.Run(); err != nil {
		t.Fatalf("runs show: %v\n%s", err, stderr)
	}
	if err := schematest.Validate(record.Schema(), stdout.Bytes()); err != nil {
		t.Errorf("the record violates its schema: %v\n%s", err, stdout)
	}
	got, err := record.Unmarshal(stdout.Bytes())
	if err != nil || got.Status != record.StatusExited || got.ExitCode == nil || *got.ExitCode != 3 || got.Agent.Name != "claude" {
		t.Errorf("record = %+v, %v", got, err)
	}
	info, err := os.Stat(filepath.Join(m.stateDir(), "agent-cli", "runs", got.ID+".json"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("record file = %v, %v; want a private file", info, err)
	}
	list, listed, _ := m.command("runs", "list")
	if err := list.Run(); err != nil || !strings.Contains(listed.String(), got.ID) {
		t.Errorf("runs list = %q, %v; want the record", listed, err)
	}
}

func TestPreflightExitCodes(t *testing.T) {
	m := newMachine(t, "0")
	cases := map[string]struct {
		args []string
		want int
	}{
		"ready":              {[]string{"preflight"}, 0},
		"required guarantee": {[]string{"preflight", "--require-kernel-isolation"}, 1},
		"bad flag":           {[]string{"preflight", "--no-such-flag"}, 2},
		"bad argument":       {[]string{"preflight", "extra"}, 2},
	}
	for name, tc := range cases {
		cmd, stdout, stderr := m.command(tc.args...)
		if got := exitCode(cmd.Run()); got != tc.want {
			t.Errorf("%s: exit code = %d, want %d\n%s%s", name, got, tc.want, stdout, stderr)
		}
	}
}

func TestPreflightJSONReportsTheGuaranteeThatFails(t *testing.T) {
	cmd, stdout, _ := newMachine(t, "0").command("preflight", "--require-kernel-isolation", "--json")

	if got := exitCode(cmd.Run()); got != 1 {
		t.Fatalf("exit code = %d, want 1", got)
	}
	var report struct {
		Ready      bool `json:"ready"`
		Guarantees []struct {
			Guarantee string `json:"guarantee"`
			Provided  bool   `json:"provided"`
			Reason    string `json:"reason"`
		} `json:"guarantees"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil || report.Ready {
		t.Fatalf("report = %+v, %v", report, err)
	}
	for _, guarantee := range report.Guarantees {
		if guarantee.Guarantee == "kernel-isolation" && (guarantee.Provided || guarantee.Reason == "") {
			t.Errorf("kernel guarantee = %+v, want unmet with a reason", guarantee)
		}
	}
}

func TestTheProjectLayerShowsInTheDryRun(t *testing.T) {
	m := newMachine(t, "0")
	if err := os.WriteFile(filepath.Join(m.work, ".agent-cli.yaml"), []byte("provider: gemini\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd, stdout, stderr := m.command("run", "--dry-run")
	// The gemini preset reads its token from the environment.
	cmd.Env = append(cmd.Env, "CLI_PROXY_API_KEY=not-a-secret")

	if err := cmd.Run(); err != nil {
		t.Fatalf("dry run: %v\n%s", err, stderr)
	}
	if output := stdout.String() + stderr.String(); !strings.Contains(output, "provider: gemini (project ") {
		t.Errorf("dry run lacks the provenance of the project layer:\n%s", output)
	}
}

func TestSchemaCommandsPrintValidJSON(t *testing.T) {
	m := newMachine(t, "0")
	for _, name := range []string{"config", "run-record"} {
		cmd, stdout, stderr := m.command("schema", name)
		if err := cmd.Run(); err != nil {
			t.Fatalf("schema %s: %v\n%s", name, err, stderr)
		}
		if !json.Valid(stdout.Bytes()) {
			t.Errorf("schema %s is not JSON", name)
		}
	}
}

func TestHerdrRefusesToRunOutsideAHerdrPane(t *testing.T) {
	if _, err := exec.LookPath("herdr"); err != nil {
		t.Skip("herdr is not installed")
	}
	cmd, _, stderr := newMachine(t, "0").command("run", "--terminal", "herdr")

	if err := cmd.Run(); err == nil {
		t.Fatal("run --terminal herdr succeeded outside a herdr pane")
	}
	if !strings.Contains(stderr.String(), "does not run inside a herdr pane") {
		t.Errorf("stderr = %q, want the outside refusal", stderr.String())
	}
}
