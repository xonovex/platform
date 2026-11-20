//go:build integration

// Package integration drives the built agent-eval binary against a real disk and
// a fake claude program, so the tier starts real processes and spends nothing.
package integration

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// binary returns the built agent-eval. The ci-integration task builds it and
// names it in AGENT_EVAL_BINARY.
func binary(t *testing.T) string {
	t.Helper()
	path := os.Getenv("AGENT_EVAL_BINARY")
	if path == "" {
		t.Skip("AGENT_EVAL_BINARY is not set; run the go-integration task")
	}
	return path
}

type result struct {
	stdout, stderr string
	code           int
}

// workspace is a temporary project with a fake claude program ahead on PATH.
type workspace struct {
	root string
	bin  string
	log  string
}

// fakeClaude answers like claude's stream-json output. A prompt that contains
// "load it" calls the Skill tool for alpha; "hang" calls it and then keeps the
// stream open, which only a stopping reader survives; "crash" exits 1.
const fakeClaude = `#!/bin/sh
for arg; do last=$arg; done
printf '%s\n' "$*" >> "$FAKE_CLAUDE_LOG"
case "$last" in
*crash*) echo "claude is down" >&2; exit 1 ;;
esac
echo '{"type":"system","subtype":"init","skills":["alpha"]}'
case "$last" in
*"load it"*|*hang*)
  echo '{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Skill","input":{"skill":"alpha"}}]}}'
  case "$last" in *hang*) sleep 60 ;; esac
  ;;
*)
  echo '{"type":"assistant","message":{"content":[{"type":"text","text":"ok"}]}}'
  echo '{"type":"result","result":"ok","total_cost_usd":0.01,"usage":{"input_tokens":5,"output_tokens":2}}'
  ;;
esac
`

const triggerDoc = `{"version":1,"skill":"alpha","queries":[
 {"id":"yes","prompt":"please load it","shouldTrigger":true,"split":"validation"},
 {"id":"no","prompt":"leave it alone","shouldTrigger":false,"split":"train"}]}`

func newWorkspace(t *testing.T) *workspace {
	t.Helper()
	w := &workspace{root: t.TempDir()}
	w.bin = filepath.Join(w.root, "bin")
	w.log = filepath.Join(w.root, "claude.log")
	w.write(t, "bin/claude", fakeClaude)
	if err := os.Chmod(filepath.Join(w.bin, "claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	w.write(t, "work/agentic.eval.json", `{"version":1,"pluginDirs":["plugins"],"profiles":{"ci":{"harness":"claude","model":"fake"}},"defaultProfile":"ci"}`)
	w.write(t, "work/plugins/core/.claude-plugin/plugin.json", `{"name":"core","skills":"./skills/"}`)
	w.write(t, "work/plugins/core/skills/alpha/SKILL.md", "---\nname: alpha\ndescription: x\n---\n")
	w.write(t, "work/plugins/core/skills/alpha/eval-trigger.json", triggerDoc)
	return w
}

func (w *workspace) write(t *testing.T, name, content string) {
	t.Helper()
	path := filepath.Join(w.root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (w *workspace) run(t *testing.T, args ...string) result {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary(t), args...)
	command.Dir = filepath.Join(w.root, "work")
	command.Env = append(os.Environ(), "PATH="+w.bin+string(os.PathListSeparator)+os.Getenv("PATH"), "FAKE_CLAUDE_LOG="+w.log)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return result{stdout.String(), stderr.String(), code}
}

func (w *workspace) calls(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(w.log)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func TestTriggerPassesThroughARealProcessAndKeepsItsResults(t *testing.T) {
	w := newWorkspace(t)
	got := w.run(t, "trigger", "alpha", "--runs", "1")
	if got.code != 0 || !strings.Contains(got.stdout, "PASS") {
		t.Fatalf("%d\n%s\n%s", got.code, got.stdout, got.stderr)
	}
	calls := w.calls(t)
	if len(calls) != 2 {
		t.Fatalf("two queries, one run each: %v", calls)
	}
	for _, flag := range []string{"--setting-sources", "--strict-mcp-config", "--no-session-persistence", "--model fake"} {
		if !strings.Contains(calls[0], flag) {
			t.Errorf("the harness stays hermetic, %q is missing: %s", flag, calls[0])
		}
	}
	runs, err := filepath.Glob(filepath.Join(w.root, "work/.agent-eval/results/*/results.json"))
	if err != nil || len(runs) != 1 {
		t.Fatalf("one checkpoint: %v %v", runs, err)
	}
	if cost := w.run(t, "cost"); cost.code != 0 || !strings.Contains(cost.stdout, "trigger") {
		t.Fatalf("%d\n%s", cost.code, cost.stdout)
	}
	if rec := w.run(t, "baseline", "record", runs[0]); rec.code != 0 {
		t.Fatalf("%d\n%s%s", rec.code, rec.stdout, rec.stderr)
	}
	if cmp := w.run(t, "baseline", "compare", runs[0]); cmp.code != 0 || !strings.Contains(cmp.stdout, "no regression") {
		t.Fatalf("%d\n%s", cmp.code, cmp.stdout)
	}
}

func TestDryRunStartsNoProcess(t *testing.T) {
	w := newWorkspace(t)
	got := w.run(t, "trigger", "alpha", "--dry-run")
	if got.code != 0 || len(w.calls(t)) != 0 {
		t.Fatalf("%d\n%s\n%v", got.code, got.stdout, w.calls(t))
	}
}

func TestABudgetStopExitsThreeAndResumeFinishes(t *testing.T) {
	w := newWorkspace(t)
	if got := w.run(t, "trigger", "alpha", "--runs", "1", "--max-model-runs", "1"); got.code != 3 {
		t.Fatalf("%d\n%s\n%s", got.code, got.stdout, got.stderr)
	}
	if got := w.run(t, "trigger", "alpha", "--runs", "1", "--resume"); got.code != 0 {
		t.Fatalf("%d\n%s\n%s", got.code, got.stdout, got.stderr)
	}
	if n := len(w.calls(t)); n != 2 {
		t.Fatalf("resume runs only the unit that was left: %d calls", n)
	}
}

func TestAFailingHarnessInvalidatesTheRun(t *testing.T) {
	w := newWorkspace(t)
	w.write(t, "work/plugins/core/skills/alpha/eval-trigger.json",
		`{"version":1,"skill":"alpha","queries":[{"id":"x","prompt":"crash now","shouldTrigger":true,"split":"validation"}]}`)
	got := w.run(t, "trigger", "alpha", "--runs", "1")
	if got.code != 2 || !strings.Contains(got.stdout, "INVALID") {
		t.Fatalf("%d\n%s\n%s", got.code, got.stdout, got.stderr)
	}
}

func TestTheTriggerRunStopsAHarnessThatKeepsTheStreamOpen(t *testing.T) {
	w := newWorkspace(t)
	w.write(t, "work/plugins/core/skills/alpha/eval-trigger.json",
		`{"version":1,"skill":"alpha","queries":[{"id":"x","prompt":"hang here","shouldTrigger":true,"split":"validation"}]}`)
	start := time.Now()
	got := w.run(t, "trigger", "alpha", "--runs", "1")
	if got.code != 0 {
		t.Fatalf("%d\n%s\n%s", got.code, got.stdout, got.stderr)
	}
	if elapsed := time.Since(start); elapsed > 30*time.Second {
		t.Fatalf("the process was not stopped once the skill loaded: %s", elapsed)
	}
}

func TestMigrateFindsNothingLeftToConvertInThePlugins(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	plugins := filepath.Join(filepath.Dir(file), "../../../../../runtime/plugin")
	if _, err := os.Stat(plugins); err != nil {
		t.Skip("the plugin tree is not part of this checkout")
	}
	w := newWorkspace(t)
	copyTree(t, plugins, filepath.Join(w.root, "work/catalog"))
	got := w.run(t, "migrate", "catalog", "--remove-legacy")
	if got.code != 0 {
		t.Fatalf("%d\n%s\n%s", got.code, tail(got.stdout), got.stderr)
	}
	var legacy, current int
	_ = filepath.WalkDir(filepath.Join(w.root, "work/catalog"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		switch d.Name() {
		case "evals.json", "eval-queries.json":
			legacy++
		case "eval-trigger.json", "eval-output.json":
			current++
		}
		return nil
	})
	if legacy != 0 || current == 0 {
		t.Fatalf("legacy eval files = %d, current eval files = %d\n%s", legacy, current, tail(got.stdout))
	}
}

func tail(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) > 8 {
		lines = lines[len(lines)-8:]
	}
	return strings.Join(lines, "\n")
}

// copyTree copies the regular files below src, skipping dependencies.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "node_modules" {
			return filepath.SkipDir
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, _ := filepath.Rel(src, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}
