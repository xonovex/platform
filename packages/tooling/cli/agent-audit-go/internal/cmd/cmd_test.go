package cmd_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/cmd"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/process/processtest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/transport/plugins"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web/webtest"
)

var today = time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

const current = `{"version":1,"sources":[{"id":"doc","kind":"web","title":"Doc","urls":["https://e.test/doc"],"reviewedAt":"2026-11-01","feeds":["references/a.md"]}]}`

const stale = `{"version":1,"sources":[{"id":"doc","kind":"web","title":"Doc","urls":["https://e.test/doc"],"reviewedAt":"2025-01-01","feeds":["references/a.md"]}]}`

type harness struct {
	deps cmd.Dependencies
	out  *bytes.Buffer
	err  *bytes.Buffer
	fs   fs.FileSystem
	web  *webtest.Fake
	git  *processtest.Fake
}

func newHarness(t *testing.T, files map[string]string) *harness {
	t.Helper()
	fsys := memfs.New(files)
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{out: &bytes.Buffer{}, err: &bytes.Buffer{}, fs: fsys}
	h.web = webtest.New(map[string]webtest.Page{"https://e.test/doc": {{Response: web.Response{Status: 200, Body: []byte("body")}}}})
	h.git = processtest.New()
	h.deps = cmd.Dependencies{
		FS: fsys, Runner: h.git,
		Transports: plugins.DefaultRegistry(plugins.Ports{FS: fsys, Runner: h.git, Web: h.web}),
		Validator:  validator, LoaderFor: wiring.NewSourcesLoader,
		Now: func() time.Time { return today }, Out: h.out, Err: h.err,
	}
	return h
}

// run executes the command and returns the exit status: 0 for a clean run, the
// code of an ExitError, and 2 for any other error, as the binary does.
func (h *harness) run(t *testing.T, args ...string) (int, error) {
	t.Helper()
	root := cmd.NewRootCommand(h.deps)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	var exit cmd.ExitError
	switch {
	case err == nil:
		return 0, nil
	case errors.As(err, &exit):
		return exit.Code, nil
	default:
		return 2, err
	}
}

func files(catalog string) map[string]string {
	return map[string]string{"skill/sources.json": catalog, "skill/references/a.md": "# A\n"}
}

func TestSourcesCleanCatalog(t *testing.T) {
	h := newHarness(t, files(current))
	code, err := h.run(t, "sources", "skill")
	if code != 0 || err != nil || !strings.Contains(h.out.String(), "0 warn, 0 drift, 0 fail") {
		t.Fatalf("code = %d, err = %v, out:\n%s", code, err, h.out)
	}
	if h.web.Requests["https://e.test/doc"] != 0 {
		t.Error("the network must stay unused without --fetch")
	}
}

func TestSourcesOverdueFailsAndStrictWidens(t *testing.T) {
	h := newHarness(t, files(stale))
	code, _ := h.run(t, "sources", "skill")
	if code != 1 || !strings.Contains(h.out.String(), "audit.review-overdue") {
		t.Fatalf("code = %d, out:\n%s", code, h.out)
	}

	// A wide window turns the overdue finding off.
	h = newHarness(t, files(stale))
	if code, _ := h.run(t, "sources", "skill", "--max-age", "9999", "--version-max-age", "9999"); code != 0 {
		t.Errorf("a wide window: code = %d, out:\n%s", code, h.out)
	}
}

func TestSourcesStrictFailsOnWarnings(t *testing.T) {
	h := newHarness(t, map[string]string{"skill/sources.json": current})
	// The reference feed is missing: a warning only when strict.
	if code, _ := h.run(t, "sources", "skill", "--strict"); code != 1 {
		t.Errorf("strict: code = %d, out:\n%s", code, h.out)
	}
}

func TestSourcesFetchUsesTheNetwork(t *testing.T) {
	h := newHarness(t, files(current))
	if code, _ := h.run(t, "sources", "skill", "--fetch"); code != 0 {
		t.Fatalf("code = %d, out:\n%s", code, h.out)
	}
	if h.web.Requests["https://e.test/doc"] == 0 {
		t.Error("--fetch must read the document")
	}
}

func TestSourcesFormats(t *testing.T) {
	for _, args := range [][]string{{"--json"}, {"--format", "json"}} {
		h := newHarness(t, files(current))
		if code, _ := h.run(t, append([]string{"sources", "skill"}, args...)...); code != 0 || !strings.Contains(h.out.String(), `"catalogs"`) {
			t.Errorf("%v: code = %d, out:\n%s", args, code, h.out)
		}
	}
	h := newHarness(t, files(stale))
	if code, _ := h.run(t, "sources", "skill", "--format", "sarif"); code != 1 || !strings.Contains(h.out.String(), `"version": "2.1.0"`) {
		t.Errorf("sarif: code = %d, out:\n%s", code, h.out)
	}
	h = newHarness(t, files(current))
	if code, err := h.run(t, "sources", "skill", "--format", "xml"); code != 2 || err == nil {
		t.Errorf("an unknown format: code = %d, err = %v", code, err)
	}
}

func TestSourcesWithoutACatalogWarns(t *testing.T) {
	h := newHarness(t, map[string]string{"empty/readme.md": "x"})
	code, _ := h.run(t, "sources", "empty")
	if code != 0 || !strings.Contains(h.out.String(), "audit.no-sources") {
		t.Fatalf("code = %d, out:\n%s", code, h.out)
	}
	if code, _ := h.run(t, "sources", "empty", "--strict"); code != 1 {
		t.Errorf("strict: code = %d", code)
	}
}

func TestSourcesDefaultsToTheCurrentDirectory(t *testing.T) {
	h := newHarness(t, map[string]string{"sources.json": current, "references/a.md": "x"})
	if code, err := h.run(t, "sources"); code != 0 || err != nil {
		t.Fatalf("code = %d, err = %v, out:\n%s", code, err, h.out)
	}
}

func TestSourcesMarkReviewed(t *testing.T) {
	h := newHarness(t, files(stale))
	code, _ := h.run(t, "sources", "skill", "--mark-reviewed", "--match", "doc")
	if code != 0 || !strings.Contains(h.err.String(), "marked 1 source(s) as reviewed on 2026-11-20") {
		t.Fatalf("code = %d, err:\n%s\nout:\n%s", code, h.err, h.out)
	}
	data, _ := h.fs.ReadFile("skill/sources.json")
	if !strings.Contains(string(data), `"reviewedAt": "2026-11-20"`) {
		t.Errorf("the catalog was not stamped:\n%s", data)
	}

	h = newHarness(t, files(current))
	h.run(t, "sources", "skill", "--mark-reviewed", "--match", "nothing")
	if h.err.Len() != 0 {
		t.Errorf("nothing matched, yet: %s", h.err)
	}
}

func TestSourcesCompareWithBase(t *testing.T) {
	changed := strings.Replace(current, `"reviewedAt":"2026-11-01"`, `"reviewedAt":"2026-11-01","version":"2.0.0"`, 1)
	h := newHarness(t, files(changed))
	h.git.On("git rev-parse --verify --quiet main^{commit}", processtest.Reply{}).
		On("git cat-file -t main:./sources.json", processtest.Reply{Stdout: "blob\n"}).
		On("git cat-file -s main:./sources.json", processtest.Reply{Stdout: "1\n"}).
		On("git show main:./sources.json", processtest.Reply{Stdout: current}).
		On("git cat-file -t main:./sources.d", processtest.Reply{ExitCode: 128}).
		On("git diff --name-only --relative main -- .", processtest.Reply{})
	code, _ := h.run(t, "sources", "skill", "--base", "main", "--max-age", "9999", "--version-max-age", "9999")
	if code != 1 || !strings.Contains(h.out.String(), "audit.metadata-changed") {
		t.Fatalf("code = %d, out:\n%s", code, h.out)
	}
}

func TestSourcesFaultsExitWithTwo(t *testing.T) {
	// An unknown base ref is a fault, never a pass.
	h := newHarness(t, files(current))
	h.git.On("git rev-parse --verify --quiet nope^{commit}", processtest.Reply{ExitCode: 1})
	code, _ := h.run(t, "sources", "skill", "--base", "nope")
	if code != 2 || !strings.Contains(h.out.String(), "unknown git ref") {
		t.Errorf("base: code = %d, out:\n%s", code, h.out)
	}

	// The loader cannot be built.
	h = newHarness(t, files(current))
	h.deps.LoaderFor = func(fs.FileSystem) (sources.Loader, error) { return sources.Loader{}, errors.New("no schemas") }
	if code, _ := h.run(t, "sources", "skill"); code != 2 {
		t.Errorf("loader: code = %d", code)
	}

	// The root cannot be listed.
	h = newHarness(t, files(current))
	if code, _ := h.run(t, "sources", "missing-dir"); code != 2 {
		t.Errorf("walk: code = %d, out:\n%s", code, h.out)
	}

	// A file system that fails stops the audit and the stamp.
	boom := errors.New("disk fault")
	h = newHarness(t, files(stale))
	h.deps.FS = faulty{FileSystem: h.fs, read: boom}
	h.deps.Transports = plugins.DefaultRegistry(plugins.Ports{FS: h.deps.FS, Runner: h.git, Web: h.web})
	if code, _ := h.run(t, "sources", "skill"); code != 2 || !strings.Contains(h.out.String(), "disk fault") {
		t.Errorf("audit read: code = %d, out:\n%s", code, h.out)
	}
	if code, _ := h.run(t, "sources", "skill", "--mark-reviewed"); code != 2 {
		t.Errorf("stamp load: code = %d", code)
	}
	h = newHarness(t, files(stale))
	h.deps.FS = faulty{FileSystem: h.fs, write: boom}
	if code, _ := h.run(t, "sources", "skill", "--mark-reviewed"); code != 2 {
		t.Errorf("stamp write: code = %d", code)
	}
}

// faulty fails reads or writes of the file system it wraps.
type faulty struct {
	fs.FileSystem
	read, write error
}

func (f faulty) ReadFile(path string) ([]byte, error) {
	if f.read != nil {
		return nil, f.read
	}
	return f.FileSystem.ReadFile(path)
}

func (f faulty) WriteFile(path string, data []byte) error {
	if f.write != nil {
		return f.write
	}
	return f.FileSystem.WriteFile(path, data)
}

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestSourcesReportsAWriteFault(t *testing.T) {
	h := newHarness(t, files(current))
	h.deps.Out = brokenWriter{}
	if code, err := h.run(t, "sources", "skill"); code != 2 || err == nil {
		t.Errorf("code = %d, err = %v", code, err)
	}
}

func TestConvert(t *testing.T) {
	const legacy = "# Sources\n\n## Doc\n\n- **URL:** https://e.test/doc\n- **Last reviewed:** 2026-03-01\n- **Used for:** facts in references/a.md\n"
	h := newHarness(t, map[string]string{"skill/SOURCES.md": legacy, "skill/references/a.md": "# A\n"})
	if code, err := h.run(t, "convert", "skill"); code != 0 || err != nil {
		t.Fatalf("code = %d, err = %v, err out:\n%s", code, err, h.err)
	}
	data, err := h.fs.ReadFile("skill/sources.json")
	if err != nil || !strings.Contains(string(data), `"id": "doc"`) || !strings.Contains(h.err.String(), "wrote skill/sources.json (1 source(s))") {
		t.Fatalf("sources.json = %s, %v, note: %s", data, err, h.err)
	}
	if kept, _ := h.fs.ReadFile("skill/SOURCES.md"); string(kept) != legacy {
		t.Error("SOURCES.md must stay as it was")
	}

	// The converted catalog audits clean.
	if code, _ := h.run(t, "sources", "skill", "--cover-dir", "", "--max-age", "9999", "--version-max-age", "9999"); code != 0 {
		t.Errorf("audit of the conversion: code = %d, out:\n%s", code, h.out)
	}

	// A second run refuses to overwrite, and --force replaces.
	if code, err := h.run(t, "convert", "skill/SOURCES.md"); code != 2 || err == nil {
		t.Errorf("overwrite: code = %d, err = %v", code, err)
	}
	if code, _ := h.run(t, "convert", "skill/SOURCES.md", "--force"); code != 0 {
		t.Errorf("force: code = %d", code)
	}
}

func TestConvertStdout(t *testing.T) {
	h := newHarness(t, map[string]string{"skill/SOURCES.md": "## Doc\n\n- **URL:** https://e.test\n- **Last reviewed:** 2026-03-01\n"})
	if code, _ := h.run(t, "convert", "skill", "--stdout"); code != 0 || !strings.Contains(h.out.String(), `"version": 1`) {
		t.Fatalf("code = %d, out:\n%s", code, h.out)
	}
	if exists, _ := fs.Exists(h.fs, "skill/sources.json"); exists {
		t.Error("--stdout must not write a file")
	}
}

func TestConvertRefusesALossyFile(t *testing.T) {
	h := newHarness(t, map[string]string{"skill/SOURCES.md": "## Doc\n\n- **URL:** https://e.test\n"})
	code, _ := h.run(t, "convert", "skill")
	if code != 1 || !strings.Contains(h.err.String(), "convert.reviewed-missing") {
		t.Fatalf("code = %d, err:\n%s", code, h.err)
	}
	if exists, _ := fs.Exists(h.fs, "skill/sources.json"); exists {
		t.Error("a file with findings must not be written")
	}
}

func TestConvertRefusesAnInvalidCatalog(t *testing.T) {
	// A date the schema rejects.
	h := newHarness(t, map[string]string{"skill/SOURCES.md": "## Doc\n\n- **URL:** https://e.test\n- **Last reviewed:** soon\n"})
	code, _ := h.run(t, "convert", "skill")
	if code != 1 || !strings.Contains(h.err.String(), "convert.invalid") {
		t.Fatalf("code = %d, err:\n%s", code, h.err)
	}
}

func TestConvertFaults(t *testing.T) {
	h := newHarness(t, map[string]string{"skill/readme.md": "x"})
	if code, _ := h.run(t, "convert", "nowhere"); code != 2 {
		t.Errorf("missing target: code = %d", code)
	}
	if code, _ := h.run(t, "convert", "skill"); code != 2 {
		t.Errorf("missing SOURCES.md: code = %d", code)
	}
	h = newHarness(t, map[string]string{"skill/SOURCES.md": "## Doc\n\n- **URL:** https://e.test\n- **Last reviewed:** 2026-03-01\n"})
	h.deps.Out = brokenWriter{}
	if code, _ := h.run(t, "convert", "skill", "--stdout"); code != 2 {
		t.Errorf("stdout fault: code = %d", code)
	}
	h = newHarness(t, map[string]string{"skill/SOURCES.md": "## Doc\n\n- **URL:** https://e.test\n"})
	h.deps.Err = brokenWriter{}
	if code, _ := h.run(t, "convert", "skill"); code != 2 {
		t.Errorf("findings write fault: code = %d", code)
	}
}

func TestCompletionAndVersion(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		h := newHarness(t, nil)
		if code, err := h.run(t, "completion", shell); code != 0 || err != nil || h.out.Len() == 0 {
			t.Errorf("%s: code = %d, err = %v", shell, code, err)
		}
	}
	h := newHarness(t, nil)
	if code, err := h.run(t, "completion", "tcsh"); code != 2 || err == nil {
		t.Errorf("an unknown shell: code = %d, err = %v", code, err)
	}
	if code, _ := h.run(t, "--version"); code != 0 || !strings.Contains(h.out.String(), "agent-audit version") {
		t.Errorf("version: code = %d, out = %s", code, h.out)
	}
}

func TestExitErrorMessage(t *testing.T) {
	if got := (cmd.ExitError{Code: 1}).Error(); got != "exit status 1" {
		t.Errorf("Error() = %q", got)
	}
}
