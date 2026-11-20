package cmd_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/configtest"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/cmd"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

// stub is a check that reports what it is told to.
type stub struct {
	name  string
	level findings.Level
	err   error
	paths *[]string
	opts  *string
}

func (s stub) Name() string           { return s.name }
func (s stub) Summary() string        { return "the " + s.name + " stub" }
func (s stub) DefaultPaths() []string { return []string{"default"} }

func (s stub) Run(ctx check.Context) ([]findings.Finding, error) {
	if s.paths != nil {
		*s.paths = ctx.Paths
	}
	if s.opts != nil {
		*s.opts = string(ctx.Options)
	}
	if s.err != nil {
		return nil, s.err
	}
	if s.level == findings.Pass {
		return nil, nil
	}
	return []findings.Finding{{Check: s.name, Level: s.level, Code: s.name + ".code", Path: "a.md", Line: 2, Message: "stubbed"}}, nil
}

type harness struct {
	files   map[string]string
	checks  []check.Check
	openErr error
	root    string
	disk    *memfs.FileSystem
	// brokenDir fails the listing of the directories that end with it.
	brokenDir string
	// brokenOut makes the output writer fail.
	brokenOut bool
	// diskRoot, when set, is the absolute root of the tree, and the files of
	// the harness sit below it on the disk.
	diskRoot string
}

func (h *harness) run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	var sink io.Writer = &out
	if h.brokenOut {
		sink = failingWriter{}
	}
	// The working directory /w is the root of a repository, and the files of the
	// harness sit in it, so the project search finds /w/agentic.json.
	config := map[string]string{"/w/.git/HEAD": ""}
	for name, content := range h.files {
		config["/w/"+name] = content
	}
	deps := cmd.Dependencies{
		ConfigFiles: configtest.NewFiles(config),
		Places:      configtest.Places{User: "/home/u/.config/agentic", Wd: "/w"},
		Open: func(root string) (cmd.Environment, error) {
			h.root = root
			if h.openErr != nil {
				return cmd.Environment{}, h.openErr
			}
			treeRoot, onDisk := ".", h.files
			if h.diskRoot != "" {
				treeRoot, onDisk = h.diskRoot, map[string]string{}
				for name, content := range h.files {
					onDisk[h.diskRoot+"/"+name] = content
				}
			}
			h.disk = memfs.New(onDisk)
			var disk fs.FileSystem = h.disk
			if h.brokenDir != "" {
				disk = brokenDir{FileSystem: h.disk, suffix: h.brokenDir}
			}
			return cmd.Environment{Tree: tree.Tree{FS: disk, Root: treeRoot}, Repository: checktest.Repo{}}, nil
		},
		Validator: validator,
		Checks:    h.checks,
		Out:       sink,
		Version:   "1.2.3",
	}
	root := cmd.NewRootCommand(deps)
	root.SetArgs(args)
	root.SetOut(sink)
	root.SetErr(&out)
	err = root.Execute()
	return out.String(), err
}

func code(err error) int { return cmd.ExitCode(err) }

func TestCleanRun(t *testing.T) {
	h := &harness{checks: []check.Check{stub{name: "alpha", level: findings.Pass}}}
	out, err := h.run(t)
	if err != nil || out != "alpha: 0 fail, 0 drift, 0 warn\n" {
		t.Fatalf("out = %q err = %v", out, err)
	}
	if h.root != "." {
		t.Fatalf("root = %q", h.root)
	}
}

func TestExitCodes(t *testing.T) {
	cases := map[string]struct {
		level  findings.Level
		args   []string
		expect int
	}{
		"fail":        {findings.Fail, nil, 1},
		"drift":       {findings.Drift, nil, 1},
		"warn":        {findings.Warn, nil, 0},
		"warn strict": {findings.Warn, []string{"--strict"}, 1},
	}
	for name, tc := range cases {
		h := &harness{checks: []check.Check{stub{name: "alpha", level: tc.level}}}
		out, err := h.run(t, tc.args...)
		if code(err) != tc.expect {
			t.Errorf("%s: code = %d (%v)", name, code(err), err)
		}
		if !strings.Contains(out, "[alpha] a.md:2:") {
			t.Errorf("%s: out = %q", name, out)
		}
	}
}

func TestUsageAndInfrastructureFaults(t *testing.T) {
	boom := errors.New("boom")
	cases := map[string]struct {
		h    *harness
		args []string
		want string
	}{
		"format":        {&harness{}, []string{"--format", "xml"}, "unknown output format"},
		"flag":          {&harness{}, []string{"--nope"}, "unknown flag"},
		"open":          {&harness{openErr: boom}, nil, "boom"},
		"unknown check": {&harness{checks: []check.Check{stub{name: "alpha"}}}, []string{"beta"}, "unknown check"},
		"check error":   {&harness{checks: []check.Check{stub{name: "alpha", err: boom}}}, nil, "alpha: boom"},
		"project file":  {&harness{files: map[string]string{"agentic.json": `{"version": 1, "validate": {"skip": 3}}`}}, nil, "invalid validate section"},
		"envelope":      {&harness{files: map[string]string{"agentic.json": `{"version": 3}`}}, nil, "version is 3"},
		"project named": {&harness{}, []string{"--project", "missing.json"}, "project configuration /w/missing.json"},
		"config show":   {&harness{files: map[string]string{"agentic.json": `{"version": 3}`}}, []string{"config", "show"}, "version is 3"},
		"list args":     {&harness{}, []string{"list", "x"}, "unknown command"},
	}
	for name, tc := range cases {
		_, err := tc.h.run(t, tc.args...)
		if code(err) != 2 || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: code = %d err = %v", name, code(err), err)
		}
	}
}

func TestSARIF(t *testing.T) {
	h := &harness{checks: []check.Check{stub{name: "alpha", level: findings.Fail}}}
	out, err := h.run(t, "--format", "sarif")
	if code(err) != 1 {
		t.Fatalf("code = %d", code(err))
	}
	var log struct {
		Runs []struct {
			Results []struct{ RuleID string } `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal([]byte(out), &log); err != nil || len(log.Runs) != 1 || len(log.Runs[0].Results) != 1 {
		t.Fatalf("out = %s", out)
	}
	h = &harness{openErr: errors.New("boom")}
	out, err = h.run(t, "--format", "sarif")
	if code(err) != 2 || !strings.Contains(out, "validate.failed") {
		t.Fatalf("a failed run leaves a failed SARIF log: %q %v", out, err)
	}
}

func TestProjectFile(t *testing.T) {
	var paths []string
	var opts string
	files := map[string]string{"agentic.json": `{"version": 1, "validate": {"skip": ["vendor"], "checks": {
		"alpha": {"paths": ["from-project"], "level": "warn", "options": {"k": 1}},
		"beta": {"enabled": false}}}}`}
	h := &harness{files: files, checks: []check.Check{
		stub{name: "alpha", level: findings.Fail, paths: &paths, opts: &opts},
		stub{name: "beta", level: findings.Fail},
	}}
	out, err := h.run(t)
	if err != nil {
		t.Fatalf("the cap lowers a failure to a warning, and beta is off: %v\n%s", err, out)
	}
	if strings.Join(paths, ",") != "from-project" || opts != `{"k":1}` {
		t.Fatalf("paths = %v opts = %s", paths, opts)
	}
	if !strings.Contains(out, "warn alpha.code") || strings.Contains(out, "beta") {
		t.Fatalf("out = %q", out)
	}
	_, err = h.run(t, "--path", "cli-path", "beta")
	if code(err) != 1 {
		t.Fatalf("a named check runs although disabled: %v", err)
	}
	if _, err := h.run(t, "--path", "cli-path"); err != nil || strings.Join(paths, ",") != "cli-path" {
		t.Fatalf("paths = %v err = %v", paths, err)
	}
	h.diskRoot = "/repo"
	if _, err := h.run(t, "--path", "/repo/cli-path/", "--path", "/elsewhere"); err != nil || strings.Join(paths, ",") != "cli-path,/elsewhere" {
		t.Fatalf("an absolute path below the root is relative to it: paths = %v err = %v", paths, err)
	}
	h.diskRoot = ""
	h.files = map[string]string{"custom.json": `{"version": 1, "validate": {"checks": {"gamma": {}}}}`}
	if _, err := h.run(t, "--project", "custom.json"); code(err) != 2 || !strings.Contains(err.Error(), "unknown check") {
		t.Fatalf("an unknown check in the project file is a usage error: %v", err)
	}
}

func TestRootFlag(t *testing.T) {
	h := &harness{checks: []check.Check{stub{name: "alpha"}}}
	if _, err := h.run(t, "--root", "/repo"); err != nil || h.root != "/repo" {
		t.Fatalf("root = %q err = %v", h.root, err)
	}
}

func TestListAndVersion(t *testing.T) {
	h := &harness{checks: []check.Check{stub{name: "zeta"}, stub{name: "alpha"}}}
	out, err := h.run(t, "list")
	if err != nil || out != "alpha\tthe alpha stub\nzeta\tthe zeta stub\n" {
		t.Fatalf("out = %q err = %v", out, err)
	}
	out, err = h.run(t, "--version")
	if err != nil || !strings.Contains(out, "1.2.3") {
		t.Fatalf("out = %q err = %v", out, err)
	}
}

func TestExitErrorMessage(t *testing.T) {
	if (cmd.ExitError{Code: 1}).Error() != "exit 1" {
		t.Fatal("a bare exit names its code")
	}
	cause := errors.New("cause")
	e := cmd.ExitError{Code: 2, Err: cause}
	if e.Error() != "cause" || !errors.Is(e, cause) {
		t.Fatal("an exit error wraps its cause")
	}
}

func TestSeedBudgetsRecordsTheCurrentWordCounts(t *testing.T) {
	h := &harness{files: map[string]string{
		"packages/runtime/plugin/p/skills/using-a/SKILL.md":        "one two three\n",
		"packages/runtime/plugin/p/skills/using-a/references/b.md": "four five\n",
	}}
	out, err := h.run(t, "seed-budgets")
	if err != nil || out != "recorded 2 budgets in budgets.json\n" {
		t.Fatalf("out = %q err = %v", out, err)
	}
	data, err := h.disk.ReadFile("budgets.json")
	if err != nil || !strings.Contains(string(data), `"packages/runtime/plugin/p/skills/using-a/SKILL.md": 3`) {
		t.Fatalf("manifest = %q, %v", data, err)
	}
	if _, err := h.run(t, "seed-budgets", "--path", "elsewhere", "--file", "out/b.json"); err != nil {
		t.Fatal(err)
	}
	if data, err := h.disk.ReadFile("out/b.json"); err != nil || string(data) != "{}\n" {
		t.Fatalf("empty manifest = %q, %v", data, err)
	}
}

func TestSeedBudgetsHonoursAbsolutePaths(t *testing.T) {
	h := &harness{diskRoot: "/repo", files: map[string]string{
		"packages/runtime/plugin/p/skills/using-a/SKILL.md": "one two three\n",
		"catalog/q/skills/using-b/SKILL.md":                 "four five\n",
	}}
	if _, err := h.run(t, "seed-budgets", "--path", "/repo/catalog", "--file", "/var/budgets.json"); err != nil {
		t.Fatal(err)
	}
	data, err := h.disk.ReadFile("/var/budgets.json")
	if err != nil || string(data) != "{\n  \"catalog/q/skills/using-b/SKILL.md\": 2\n}\n" {
		t.Fatalf("manifest at the absolute file = %q, %v", data, err)
	}
	if _, err := h.disk.ReadFile("/repo/var/budgets.json"); err == nil {
		t.Fatal("the absolute file landed below the root")
	}
	if _, err := h.run(t, "seed-budgets", "--path", "catalog", "--file", "out/b.json"); err != nil {
		t.Fatal(err)
	}
	if relative, err := h.disk.ReadFile("/repo/out/b.json"); err != nil || string(relative) != string(data) {
		t.Fatalf("a relative path resolves against the root: %q, %v", relative, err)
	}
	if _, err := h.run(t, "seed-budgets", "--path", "/repo/packages/runtime/plugin/p/skills/using-a/SKILL.md", "--file", "../outside.json"); err != nil {
		t.Fatal(err)
	}
	if outside, err := h.disk.ReadFile("/outside.json"); err != nil || !strings.Contains(string(outside), `"packages/runtime/plugin/p/skills/using-a/SKILL.md": 3`) {
		t.Fatalf("a path outside the root stays allowed: %q, %v", outside, err)
	}
}

func TestSeedBudgetsFaults(t *testing.T) {
	boom := errors.New("boom")
	if _, err := (&harness{openErr: boom}).run(t, "seed-budgets"); code(err) != 2 || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("open: %v", err)
	}
	if _, err := (&harness{}).run(t, "seed-budgets", "extra"); code(err) != 2 {
		t.Fatalf("args: %v", err)
	}
	if _, err := (&harness{}).run(t, "seed-budgets", "--file", ""); code(err) != 2 {
		t.Fatalf("write: %v", err)
	}
}

type brokenDir struct {
	fs.FileSystem
	suffix string
}

func (b brokenDir) ReadDir(path string) ([]fs.Entry, error) {
	if strings.HasSuffix(path, b.suffix) {
		return nil, errors.New("disk fault")
	}
	return b.FileSystem.ReadDir(path)
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestSeedBudgetsReportsAReadFaultAndAClosedOutput(t *testing.T) {
	files := map[string]string{"packages/runtime/plugin/p/skills/a/SKILL.md": "one\n"}
	if _, err := (&harness{files: files, brokenDir: "packages/runtime/plugin"}).run(t, "seed-budgets"); code(err) != 2 || !strings.Contains(err.Error(), "disk fault") {
		t.Fatalf("read: %v", err)
	}
	if _, err := (&harness{files: files, brokenOut: true}).run(t, "seed-budgets"); code(err) != 2 || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("output: %v", err)
	}
}

func TestConfigShow(t *testing.T) {
	h := &harness{files: map[string]string{"agentic.json": `{"version": 1, "validate": {"skip": ["vendor"], "checks": {"alpha": {"level": "warn"}}}}`}}
	out, err := h.run(t, "config", "show")
	want := "{\n  \"checks\": {\n    \"alpha\": {\n      \"level\": \"warn\"\n    }\n  },\n  \"skip\": [\n    \"vendor\"\n  ]\n}\n"
	if err != nil || out != want {
		t.Fatalf("out = %q err = %v", out, err)
	}
	out, err = h.run(t, "config", "show", "--sources")
	want = "validate.skip = [\"vendor\"]  (project /w/agentic.json)\nvalidate.checks.alpha = {\"level\":\"warn\"}  (project /w/agentic.json)\n"
	if err != nil || out != want {
		t.Fatalf("out = %q err = %v", out, err)
	}
}
