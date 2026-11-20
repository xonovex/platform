//go:build integration

package integration

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binary returns the built agent-audit. The ci-integration task builds it and
// names it in AGENT_AUDIT_BINARY.
func binary(t *testing.T) string {
	t.Helper()
	path := os.Getenv("AGENT_AUDIT_BINARY")
	if path == "" {
		t.Skip("AGENT_AUDIT_BINARY is not set; run the go-integration task")
	}
	return path
}

type result struct {
	stdout, stderr string
	code           int
}

func run(t *testing.T, dir string, args ...string) result {
	t.Helper()
	command := exec.CommandContext(context.Background(), binary(t), args...)
	command.Dir = dir
	// An empty user configuration keeps the agentic.json of the developer out
	// of the run.
	command.Env = append(os.Environ(), "XDG_CONFIG_HOME="+t.TempDir())
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

func write(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func catalog(reviewed, extra string) string {
	return `{"version":1,"sources":[{"id":"doc","kind":"original","title":"Doc","provenance":"written here","reviewedAt":"` + reviewed + `","feeds":["references/a.md"]` + extra + `}]}` + "\n"
}

func TestBinaryAuditExitCodes(t *testing.T) {
	root := t.TempDir()
	write(t, root, "skill/references/a.md", "# A\n")
	write(t, root, "skill/sources.json", catalog("2000-01-01", ""))

	stale := run(t, root, "sources", "skill")
	if stale.code != 1 || !strings.Contains(stale.stdout, "audit.review-overdue") {
		t.Fatalf("stale: %+v", stale)
	}

	marked := run(t, root, "sources", "skill", "--mark-reviewed")
	if marked.code != 0 || !strings.Contains(marked.stderr, "marked 1 source(s) as reviewed") {
		t.Fatalf("marked: %+v", marked)
	}
	if run(t, root, "sources", "skill").code != 0 {
		t.Error("a stamped catalog must audit clean")
	}

	sarif := run(t, root, "sources", "skill", "--format", "sarif")
	if sarif.code != 0 || !strings.Contains(sarif.stdout, `"version": "2.1.0"`) {
		t.Errorf("sarif: %+v", sarif)
	}

	// Without a git repository the comparison is a fault, never a pass.
	if fault := run(t, root, "sources", "skill", "--base", "main"); fault.code != 2 {
		t.Errorf("fault: %+v", fault)
	}
	if usage := run(t, root, "sources", "--no-such-flag"); usage.code != 2 {
		t.Errorf("usage: %+v", usage)
	}
}

func TestBinaryConvert(t *testing.T) {
	root := t.TempDir()
	write(t, root, "skill/SOURCES.md", "# Sources\n\n## Doc\n\n- **URL:** https://example.test/doc\n- **Last reviewed:** 2026-03-01\n- **Used for:** references/a.md\n")
	write(t, root, "skill/references/a.md", "# A\n")
	if got := run(t, root, "convert", "skill"); got.code != 0 {
		t.Fatalf("convert: %+v", got)
	}
	if _, err := os.Stat(filepath.Join(root, "skill/sources.json")); err != nil {
		t.Fatal(err)
	}
	if again := run(t, root, "convert", "skill"); again.code != 2 {
		t.Errorf("a second convert must refuse: %+v", again)
	}
	if audit := run(t, root, "sources", "skill", "--max-age", "99999", "--version-max-age", "99999"); audit.code != 0 {
		t.Errorf("audit of the conversion: %+v", audit)
	}
}

func TestBinaryCompletion(t *testing.T) {
	if got := run(t, t.TempDir(), "completion", "bash"); got.code != 0 || got.stdout == "" {
		t.Errorf("completion: %+v", got)
	}
}
