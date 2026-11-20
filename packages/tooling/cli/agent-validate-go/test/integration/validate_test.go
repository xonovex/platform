//go:build integration

// Package integration runs the built agent-validate binary over this repository
// and over temporary repositories. It is the only tier that starts the process,
// reads the real disk and asks git.
package integration

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func binary(t *testing.T) string {
	t.Helper()
	path := os.Getenv("AGENT_VALIDATE_BINARY")
	if path == "" {
		t.Fatal("AGENT_VALIDATE_BINARY names the built binary; run the go-integration task")
	}
	return path
}

// repoRoot is the root of this repository: the module sits four levels below it.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "..", ".."))
}

type result struct {
	stdout, stderr string
	code           int
}

func run(t *testing.T, dir string, args ...string) result {
	t.Helper()
	command := exec.Command(binary(t), args...)
	command.Dir = dir
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	res := result{stdout: stdout.String(), stderr: stderr.String()}
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		res.code = exit.ExitCode()
	case err != nil:
		t.Fatalf("run %v: %v", args, err)
	}
	return res
}

func write(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestRepositoryIsClean(t *testing.T) {
	res := run(t, repoRoot(t))
	if res.code != 0 {
		t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", res.code, res.stdout, res.stderr)
	}
	for _, name := range []string{"skill", "sources", "drift", "links", "routing", "plugin", "marketplace", "repository", "workspace", "text"} {
		if !strings.Contains(res.stdout, name+": 0 fail, 0 drift, 0 warn") {
			t.Errorf("no clean summary for %s in\n%s", name, res.stdout)
		}
	}
}

func TestRepositoryReportsSARIF(t *testing.T) {
	res := run(t, repoRoot(t), "--format", "sarif")
	if res.code != 0 {
		t.Fatalf("exit %d\n%s", res.code, res.stderr)
	}
	var log struct {
		Version string `json:"version"`
		Runs    []struct {
			Tool struct {
				Driver struct{ Name string } `json:"driver"`
			} `json:"tool"`
			Results []json.RawMessage `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &log); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, res.stdout)
	}
	if len(log.Runs) != 1 || log.Runs[0].Tool.Driver.Name != "agent-validate" || len(log.Runs[0].Results) != 0 {
		t.Fatalf("log = %+v", log)
	}
}

func TestSelectsChecksAndPaths(t *testing.T) {
	root := t.TempDir()
	write(t, root, "docs/a.md", "A sentence with a dash \u2014 here.\n")
	write(t, root, "docs/b.md", "Clean line.\n")
	res := run(t, root, "text", "--path", "docs/b.md")
	if res.code != 0 || !strings.Contains(res.stdout, "text: 0 fail") {
		t.Fatalf("exit %d\n%s%s", res.code, res.stdout, res.stderr)
	}
	res = run(t, root, "text")
	if res.code != 1 || !strings.Contains(res.stdout, "[text] docs/a.md:1: fail text.em-dash") {
		t.Fatalf("exit %d\n%s%s", res.code, res.stdout, res.stderr)
	}
	if res = run(t, filepath.Join(root, "docs"), "--root", root, "text"); res.code != 1 {
		t.Fatalf("--root names the repository whatever the working directory: exit %d", res.code)
	}
}

func TestProjectFileCapsAndSkips(t *testing.T) {
	root := t.TempDir()
	write(t, root, "docs/a.md", "A dash \u2014 here.\n")
	write(t, root, "vendor/b.md", "A dash \u2014 there.\n")
	write(t, root, "agentic.validate.json", `{"version": 1, "skip": ["vendor"], "checks": {"text": {"level": "warn"}}}`)
	res := run(t, root, "text")
	if res.code != 0 || !strings.Contains(res.stdout, "docs/a.md") || strings.Contains(res.stdout, "vendor") {
		t.Fatalf("exit %d\n%s%s", res.code, res.stdout, res.stderr)
	}
	if res = run(t, root, "text", "--strict"); res.code != 1 {
		t.Fatalf("--strict fails on a warning: exit %d", res.code)
	}
}

func TestGitStateReachesTheWorkspaceCheck(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init", "--quiet")
	write(t, root, ".moon/workspace.yml", "hasher:\n  walkStrategy: glob\n  ignorePatterns:\n    - '**/node_modules/**/*'\n")
	write(t, root, ".gitignore", "node_modules/\nbuild/\n")
	write(t, root, "node_modules/x/index.js", "x")
	write(t, root, "build/out.txt", "x")
	write(t, root, "packages/a/package.json", `{"name": "a", "private": true, "bin": {"a": "bin/a.js"}}`)
	write(t, root, "packages/a/moon.yml", "tags: []\n")
	res := run(t, root, "workspace", "--path", "packages")
	if res.code != 1 {
		t.Fatalf("exit %d\n%s%s", res.code, res.stdout, res.stderr)
	}
	for _, want := range []string{"workspace.hasher-uncovered", "build/", "workspace.bin-target"} {
		if !strings.Contains(res.stdout, want) {
			t.Errorf("missing %q in\n%s", want, res.stdout)
		}
	}
	if strings.Contains(res.stdout, "node_modules/ is ignored") {
		t.Errorf("a covered directory is not reported:\n%s", res.stdout)
	}
	write(t, root, "packages/a/bin/a.js", "#!/usr/bin/env node\n")
	git(t, root, "add", "packages/a/bin/a.js")
	res = run(t, root, "workspace", "--path", "packages")
	if strings.Contains(res.stdout, "workspace.bin-target") {
		t.Errorf("a committed target satisfies the rule:\n%s", res.stdout)
	}
}

func TestOutsideGitTheGitRulesFail(t *testing.T) {
	root := t.TempDir()
	write(t, root, "packages/a/package.json", `{"name": "a", "bin": {"a": "bin/a.js"}}`)
	write(t, root, "packages/a/moon.yml", "tags: []\n")
	res := run(t, root, "workspace", "--path", "packages")
	if res.code != 1 || !strings.Contains(res.stdout, "workspace.bin-git") {
		t.Fatalf("a missing git state is a finding, not a pass: exit %d\n%s%s", res.code, res.stdout, res.stderr)
	}
}

func TestUsageErrorsExitTwo(t *testing.T) {
	root := t.TempDir()
	cases := map[string][]string{
		"unknown format":  {"--format", "xml"},
		"unknown check":   {"nope"},
		"unknown flag":    {"--nope"},
		"missing project": {"--project", "missing.json"},
	}
	for name, args := range cases {
		res := run(t, root, args...)
		if res.code != 2 || res.stderr == "" {
			t.Errorf("%s: exit %d stderr %q", name, res.code, res.stderr)
		}
	}
	write(t, root, "agentic.validate.json", `{"version": 9}`)
	if res := run(t, root); res.code != 2 || !strings.Contains(res.stderr, "invalid project file") {
		t.Errorf("a broken project file: exit %d stderr %q", res.code, res.stderr)
	}
	res := run(t, root, "--format", "sarif", "nope")
	if res.code != 2 || !strings.Contains(res.stdout, "validate.failed") {
		t.Errorf("a failed run leaves a failed SARIF log: exit %d\n%s", res.code, res.stdout)
	}
}

func TestListNamesEveryCheck(t *testing.T) {
	res := run(t, t.TempDir(), "list")
	if res.code != 0 {
		t.Fatalf("exit %d: %s", res.code, res.stderr)
	}
	for _, name := range []string{"skill", "text", "workspace", "kit", "output-style"} {
		if !strings.Contains(res.stdout, name+"\t") {
			t.Errorf("list lacks %s:\n%s", name, res.stdout)
		}
	}
}
