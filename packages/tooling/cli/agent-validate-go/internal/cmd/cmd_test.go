package cmd_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"

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
}

func (h *harness) run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	deps := cmd.Dependencies{
		Open: func(root string) (cmd.Environment, error) {
			h.root = root
			if h.openErr != nil {
				return cmd.Environment{}, h.openErr
			}
			return cmd.Environment{Tree: tree.Tree{FS: memfs.New(h.files), Root: "."}, Repository: checktest.Repo{}}, nil
		},
		Validator: validator,
		Checks:    h.checks,
		Out:       &out,
		Version:   "1.2.3",
	}
	root := cmd.NewRootCommand(deps)
	root.SetArgs(args)
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
		"format":        {&harness{}, []string{"--format", "xml"}, "unknown format"},
		"flag":          {&harness{}, []string{"--nope"}, "unknown flag"},
		"open":          {&harness{openErr: boom}, nil, "boom"},
		"unknown check": {&harness{checks: []check.Check{stub{name: "alpha"}}}, []string{"beta"}, "unknown check"},
		"check error":   {&harness{checks: []check.Check{stub{name: "alpha", err: boom}}}, nil, "alpha: boom"},
		"project file":  {&harness{files: map[string]string{"agentic.validate.json": `{"version": 3}`}}, nil, "invalid project file"},
		"project named": {&harness{}, []string{"--project", "missing.json"}, "does not exist"},
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
	files := map[string]string{"agentic.validate.json": `{"version": 1, "skip": ["vendor"], "checks": {
		"alpha": {"paths": ["from-project"], "level": "warn", "options": {"k": 1}},
		"beta": {"enabled": false}}}`}
	h := &harness{files: files, checks: []check.Check{
		stub{name: "alpha", level: findings.Fail, paths: &paths, opts: &opts},
		stub{name: "beta", level: findings.Fail},
	}}
	out, err := h.run(t)
	if err != nil {
		t.Fatalf("the cap lowers a failure to a warning, and beta is off: %v\n%s", err, out)
	}
	if strings.Join(paths, ",") != "from-project" || opts != `{"k": 1}` {
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
	h.files = map[string]string{"custom.json": `{"version": 1, "checks": {"gamma": {}}}`}
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
