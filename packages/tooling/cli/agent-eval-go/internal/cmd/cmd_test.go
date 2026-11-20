package cmd_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/cmd"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/protocoltest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/registry"
)

var today = time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

const triggerDoc = `{"version":1,"skill":"alpha","queries":[
 {"id":"yes","prompt":"load it","shouldTrigger":true,"split":"validation"},
 {"id":"no","prompt":"leave it","shouldTrigger":false,"split":"train"}]}`

const project = `{"version":1,"pluginDirs":["plugins"],"defaultProfile":"ci","profiles":{"ci":{"harness":"claude","model":"m"}}}`

func baseFiles() map[string]string {
	return map[string]string{
		"/w/agentic.eval.json":                           project,
		"/w/plugins/core/.claude-plugin/plugin.json":     `{"name":"core","skills":"./skills/"}`,
		"/w/plugins/core/skills/alpha/SKILL.md":          "a",
		"/w/plugins/core/skills/alpha/eval-trigger.json": triggerDoc,
	}
}

type fixture struct {
	deps cmd.Dependencies
	out  *bytes.Buffer
	fs   fs.FileSystem
	exec *protocoltest.Exec
}

func selects(by map[string]string) *protocoltest.Exec {
	return &protocoltest.Exec{Respond: func(r harness.Request) (harness.Observation, error) {
		return harness.Observation{Selected: by[r.Prompt], Usage: model.Usage{CostUSD: 0.01, InputTokens: 5}}, nil
	}}
}

func newFixture(t *testing.T, files map[string]string, exec *protocoltest.Exec) *fixture {
	t.Helper()
	fsys := memfs.New(files)
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{out: &bytes.Buffer{}, fs: fsys, exec: exec}
	f.deps = cmd.Dependencies{
		FS: fsys, Validator: validator, Protocols: registry.Protocols(), Exec: exec,
		Predicates: registry.Predicates(exec), Env: map[string]string{}, Dir: "/w",
		Now: func() time.Time { return today }, Out: f.out, Err: &bytes.Buffer{},
	}
	return f
}

// run executes the command and returns the exit status as the binary maps it.
func (f *fixture) run(args ...string) (int, error) {
	f.out.Reset()
	root := cmd.NewRootCommand(f.deps)
	root.SetArgs(args)
	err := root.Execute()
	var exit cmd.ExitError
	switch {
	case err == nil:
		return 0, nil
	case errors.As(err, &exit):
		return exit.Code, nil
	}
	return 2, err
}

var good = map[string]string{"load it": "target", "leave it": "none"}

func TestTriggerPassesAndRecordsResultsAndCost(t *testing.T) {
	f := newFixture(t, baseFiles(), selects(good))
	code, err := f.run("trigger", "alpha", "--runs", "1")
	if err != nil || code != 0 || !strings.Contains(f.out.String(), "PASS") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
	if calls := f.exec.Calls(); len(calls) != 2 || calls[0].Model != "m" || calls[0].Harness != "claude" {
		t.Fatalf("the profile of the project file applies: %+v", calls)
	}
	id := ""
	entries, _ := f.fs.ReadDir("/w/.agent-eval/results")
	for _, e := range entries {
		if e.IsDir {
			id = e.Name
		}
	}
	if id == "" {
		t.Fatal("no run directory")
	}
	costCode, err := f.run("cost")
	if err != nil || costCode != 0 || !strings.Contains(f.out.String(), "trigger") || !strings.Contains(f.out.String(), "total") {
		t.Fatalf("%v\n%s", err, f.out)
	}
	if _, err := f.run("cost", "--protocol", "output"); err != nil || !strings.Contains(f.out.String(), "0 runs") {
		t.Fatalf("%v\n%s", err, f.out)
	}

	results := "/w/.agent-eval/results/" + id + "/results.json"
	if code, err := f.run("baseline", "record", results); err != nil || code != 0 || !strings.Contains(f.out.String(), "recorded 2 cases") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
	if code, err := f.run("baseline", "compare", results); err != nil || code != 0 || !strings.Contains(f.out.String(), "no regression") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
}

func TestBaselineCompareFindsARegression(t *testing.T) {
	f := newFixture(t, baseFiles(), selects(good))
	_, _ = f.run("trigger", "alpha", "--runs", "1")
	entries, _ := f.fs.ReadDir("/w/.agent-eval/results")
	var results string
	for _, e := range entries {
		if e.IsDir {
			results = "/w/.agent-eval/results/" + e.Name + "/results.json"
		}
	}
	_, _ = f.run("baseline", "record", results)
	f.exec.Respond = func(harness.Request) (harness.Observation, error) {
		return harness.Observation{Selected: "none"}, nil
	}
	// A different threshold is a different run id, so this run is fresh.
	if code, _ := f.run("trigger", "alpha", "--runs", "1", "--threshold", "0.9"); code != 1 {
		t.Fatalf("the failing run exits 1, got %d\n%s", code, f.out)
	}
	recorded := results
	entries, _ = f.fs.ReadDir("/w/.agent-eval/results")
	for _, e := range entries {
		if path := "/w/.agent-eval/results/" + e.Name + "/results.json"; e.IsDir && path != recorded {
			results = path
		}
	}
	code, err := f.run("baseline", "compare", results)
	if err != nil || code != 1 || !strings.Contains(f.out.String(), "regressed yes") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
}

func TestDryRunRunsNoModel(t *testing.T) {
	f := newFixture(t, baseFiles(), selects(good))
	code, err := f.run("trigger", "alpha", "--dry-run")
	if err != nil || code != 0 || !strings.Contains(f.out.String(), "dry run: 2 units planned") || len(f.exec.Calls()) != 0 {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
}

func TestBudgetStopExitsThreeAndResumeFinishes(t *testing.T) {
	f := newFixture(t, baseFiles(), selects(good))
	code, err := f.run("trigger", "alpha", "--runs", "1", "--max-model-runs", "1")
	if err != nil || code != 3 || !strings.Contains(f.out.String(), "run again with --resume") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
	code, err = f.run("trigger", "alpha", "--runs", "1", "--resume", "--format", "json")
	if err != nil || code != 0 || !strings.Contains(f.out.String(), `"pass": true`) || !strings.Contains(f.out.String(), `"resumed": 1`) {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
}

func TestAFailingQueryExitsOne(t *testing.T) {
	f := newFixture(t, baseFiles(), selects(map[string]string{"load it": "none", "leave it": "none"}))
	if code, err := f.run("trigger", "alpha", "--runs", "1"); err != nil || code != 1 || !strings.Contains(f.out.String(), "FAIL") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
}

func TestInfrastructureFaultExitsTwo(t *testing.T) {
	f := newFixture(t, baseFiles(), &protocoltest.Exec{Respond: func(harness.Request) (harness.Observation, error) {
		return harness.Observation{}, errors.New("claude is down")
	}})
	if code, err := f.run("trigger", "alpha", "--runs", "1"); err != nil || code != 2 || !strings.Contains(f.out.String(), "INVALID") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
}

func TestUsageErrors(t *testing.T) {
	f := newFixture(t, baseFiles(), selects(good))
	for name, args := range map[string][]string{
		"unknown skill":   {"trigger", "zeta"},
		"bad format":      {"trigger", "alpha", "--format", "xml"},
		"unknown profile": {"trigger", "alpha", "--profile", "nope"},
		"too many runs":   {"trigger", "alpha", "--runs", "9"},
		"no argument":     {"trigger"},
		"missing project": {"trigger", "alpha", "--project", "none.json"},
		"bad plugin dir":  {"trigger", "alpha", "--plugin-dir", "/missing"},
		"baseline none":   {"baseline", "compare", "/w/missing.json"},
		"cost bad file":   {"cost", "--file", "/w/bad.json"},
	} {
		f.fs.WriteFile("/w/bad.json", []byte("{}"))
		if code, err := f.run(args...); code != 2 || err == nil {
			t.Errorf("%s: %d %v", name, code, err)
		}
	}
}

func TestWithoutAProjectFileTheWorkingDirectoryIsTheCatalog(t *testing.T) {
	files := baseFiles()
	delete(files, "/w/agentic.eval.json")
	f := newFixture(t, files, selects(good))
	if code, err := f.run("trigger", "alpha", "--runs", "1", "--plugin-dir", "plugins"); err != nil || code != 0 {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
	if _, err := f.run("trigger", "alpha", "--runs", "1"); err == nil {
		t.Fatal("the working directory holds no plugin, so the skill is not found")
	}
}

func TestInstructionAndOrchestrationTakePaths(t *testing.T) {
	files := baseFiles()
	files["/w/e/eval.json"] = `{"version":1,"instruction":"AGENTS.md","cases":[{"id":"c","prompt":"p","anchor":{"heading":"Tests"},
	 "expect":{"present":[{"kind":"session","event":"command","match":"test"}],"absent":[{"kind":"session","event":"command","match":"test","max":0}]}}]}`
	files["/w/e/AGENTS.md"] = "# R\n\n## Tests\n\nRun tests.\n"
	files["/w/flow.json"] = `{"version":1,"id":"f","agents":[{"id":"a","harness":"claude"}],"steps":[{"id":"s","prompt":"go","agent":"a"}]}`
	exec := &protocoltest.Exec{Respond: func(r harness.Request) (harness.Observation, error) {
		obs := harness.Observation{}
		if strings.Contains(r.Instructions, "Run tests.") {
			obs.Events = []harness.Event{{Kind: harness.EventCommand, Text: "go test"}}
		}
		return obs, nil
	}}
	f := newFixture(t, files, exec)
	if code, err := f.run("instruction", "e/eval.json", "--min-lift-unused"); err == nil && code == 0 {
		t.Fatal("an unknown flag is an error")
	}
	if code, err := f.run("instruction", "e/eval.json"); err != nil || code != 0 || !strings.Contains(f.out.String(), "mean lift") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
	if code, err := f.run("orchestration", "flow.json"); err != nil || code != 0 || !strings.Contains(f.out.String(), "1 of 1 flow executions passed") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
}

func TestRoutingTakesAnOptionalPlugin(t *testing.T) {
	files := baseFiles()
	files["/w/plugins/core/skills/beta/SKILL.md"] = "b"
	files["/w/plugins/core/skills/beta/eval-trigger.json"] = `{"version":1,"skill":"beta","queries":[{"id":"x","prompt":"load it","shouldTrigger":false,"split":"validation"}]}`
	f := newFixture(t, files, selects(map[string]string{"load it": "target"}))
	if code, err := f.run("routing", "--runs", "1"); err != nil || code != 0 || !strings.Contains(f.out.String(), "1 of 1 scenarios") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
	if _, err := f.run("routing", "nope"); err == nil {
		t.Fatal("unknown plugin")
	}
}

func TestMigrateConvertsSkillsBelowAPath(t *testing.T) {
	files := map[string]string{
		"/w/plugins/core/skills/alpha/SKILL.md":          "a",
		"/w/plugins/core/skills/alpha/eval-queries.json": `[{"query":"q","should_trigger":true,"split":"train","rationale":"r"}]`,
		"/w/plugins/core/skills/beta/SKILL.md":           "b",
		"/w/plugins/core/skills/beta/evals.json":         `{"skill_name":"beta","evals":[{"id":1,"prompt":"p","assertions":["a"]}]}`,
	}
	f := newFixture(t, files, selects(nil))
	if code, err := f.run("migrate", "plugins", "--remove-legacy"); err != nil || code != 0 ||
		!strings.Contains(f.out.String(), "2 documents written") || !strings.Contains(f.out.String(), "removed /w/plugins/core/skills/beta/evals.json") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
	if ok, _ := fs.Exists(f.fs, "/w/plugins/core/skills/alpha/eval-trigger.json"); !ok {
		t.Fatal("trigger document written")
	}
	if _, err := f.run("migrate", "/w/missing"); err == nil {
		t.Fatal("missing path")
	}
	f.fs.WriteFile("/w/plugins/core/skills/gamma/SKILL.md", []byte("g"))
	f.fs.WriteFile("/w/plugins/core/skills/gamma/evals.json", []byte("["))
	if _, err := f.run("migrate", "plugins"); err == nil {
		t.Fatal("a malformed legacy file is an error")
	}
}

func TestCompletionAndVersion(t *testing.T) {
	f := newFixture(t, nil, selects(nil))
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		if code, err := f.run("completion", shell); err != nil || code != 0 || f.out.Len() == 0 {
			t.Errorf("%s: %d %v", shell, code, err)
		}
	}
	if _, err := f.run("completion", "tcsh"); err == nil {
		t.Fatal("unknown shell")
	}
	if _, err := f.run("--version"); err != nil || !strings.Contains(f.out.String(), "dev") {
		t.Fatalf("%v %s", err, f.out)
	}
}

func TestBaselineRecordNeedsACompleteRun(t *testing.T) {
	f := newFixture(t, baseFiles(), selects(good))
	_, _ = f.run("trigger", "alpha", "--runs", "1", "--max-model-runs", "1")
	entries, _ := f.fs.ReadDir("/w/.agent-eval/results")
	var results string
	for _, e := range entries {
		if e.IsDir {
			results = "/w/.agent-eval/results/" + e.Name + "/results.json"
		}
	}
	if _, err := f.run("baseline", "record", results); err == nil || !strings.Contains(err.Error(), "only a complete run") {
		t.Fatalf("%v", err)
	}
	if _, err := f.run("baseline", "compare", results); err == nil || !strings.Contains(err.Error(), "only a complete run") {
		t.Fatalf("%v", err)
	}
}

func TestExitErrorNamesItsStatus(t *testing.T) {
	if got := (cmd.ExitError{Code: 3}).Error(); got != "exit status 3" {
		t.Fatal(got)
	}
}

func TestBaselineCompareNeedsARecordedBaseline(t *testing.T) {
	f := newFixture(t, baseFiles(), selects(good))
	_, _ = f.run("trigger", "alpha", "--runs", "1")
	entries, _ := f.fs.ReadDir("/w/.agent-eval/results")
	for _, e := range entries {
		if e.IsDir {
			_, err := f.run("baseline", "compare", "/w/.agent-eval/results/"+e.Name+"/results.json")
			if err == nil || !strings.Contains(err.Error(), "baseline record first") {
				t.Fatalf("%v", err)
			}
		}
	}
}

func TestMigrateDoesNotSearchBelowTheWalkDepth(t *testing.T) {
	deep := "/w/a/b/c/d/e/f/g/skills/x/SKILL.md"
	f := newFixture(t, map[string]string{deep: "x", "/w/a/b/c/d/e/f/g/skills/x/evals.json": `{"skill_name":"x","evals":[{"id":1,"prompt":"p","assertions":["a"]}]}`}, selects(nil))
	if code, err := f.run("migrate", "a"); err != nil || code != 0 || !strings.Contains(f.out.String(), "0 documents written") {
		t.Fatalf("%d %v\n%s", code, err, f.out)
	}
}

func TestAnUnreadableProjectFileIsAUsageError(t *testing.T) {
	files := baseFiles()
	files["/w/agentic.eval.json"] = `{"version":1,"profiles":3}`
	f := newFixture(t, files, selects(good))
	if code, err := f.run("trigger", "alpha"); code != 2 || err == nil {
		t.Fatalf("%d %v", code, err)
	}
}
