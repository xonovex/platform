package audit_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process/processtest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/audit"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/rules"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/transport/plugins"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web/webtest"
)

var now = time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

func options() audit.Options {
	return audit.Options{Now: now, Windows: rules.Windows{Default: 180, Versioned: 90}, CoverDir: "references"}
}

func deps(t *testing.T, fsys fs.FileSystem, runner *processtest.Fake, client web.Client) audit.Deps {
	t.Helper()
	loader, err := wiring.NewSourcesLoader(fsys)
	if err != nil {
		t.Fatal(err)
	}
	return audit.Deps{FS: fsys, Loader: loader, Transports: plugins.DefaultRegistry(plugins.Ports{FS: fsys, Runner: runner, Web: client})}
}

func codes(all []findings.Finding) []string {
	var out []string
	for _, f := range all {
		out = append(out, f.Code)
	}
	return out
}

func has(all []findings.Finding, code string) bool {
	for _, f := range all {
		if f.Code == code {
			return true
		}
	}
	return false
}

const cleanCatalog = `{"version":1,"sources":[
  {"id":"spec","kind":"original","title":"Spec","provenance":"written here","reviewedAt":"2026-11-01","feeds":["references/a.md"]}]}`

func TestCleanCatalogPasses(t *testing.T) {
	fsys := memfs.New(map[string]string{"s/sources.json": cleanCatalog, "s/references/a.md": "x"})
	report, err := audit.Catalog(context.Background(), deps(t, fsys, processtest.Script(), webtest.New(nil)), "s", options())
	if err != nil {
		t.Fatal(err)
	}
	if got := codes(report.Findings); len(got) != 2 || got[0] != "sources.ok" || got[1] != rules.CodeOK {
		t.Fatalf("findings = %v", report.Findings)
	}
	if report.Dir != "s" || len(report.Sources) != 1 || report.Sources[0].ID != "spec" || len(report.Catalog.Sources) != 1 {
		t.Errorf("report = %+v", report)
	}
}

func TestMissingCatalogStopsTheAudit(t *testing.T) {
	report, err := audit.Catalog(context.Background(), deps(t, memfs.New(map[string]string{"s/x": ""}), processtest.Script(), webtest.New(nil)), "s", options())
	if err != nil || len(report.Findings) != 1 || report.Findings[0].Code != sources.CodeMissing {
		t.Fatalf("report = %+v, err = %v", report, err)
	}
}

func TestEveryRuleReachesTheReport(t *testing.T) {
	catalog := `{"version":1,"sources":[
  {"id":"old","kind":"original","title":"Old","provenance":"here","reviewedAt":"2020-01-01","feeds":["references/missing.md"]},
  {"id":"unmapped","kind":"original","title":"Unmapped","provenance":"here","reviewedAt":"2026-11-01"},
  {"id":"unpinned","kind":"web","title":"Unpinned","urls":["https://e.test/doc"],"version":"1.0.0","reviewedAt":"2026-11-01","feeds":["references/a.md"]}]}`
	fsys := memfs.New(map[string]string{"s/sources.json": catalog, "s/references/a.md": "x", "s/references/b.md": "x"})
	report, err := audit.Catalog(context.Background(), deps(t, fsys, processtest.Script(), webtest.New(nil)), "s", options())
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{rules.CodeReviewOverdue, rules.CodeReferenceDangling, rules.CodeReferenceMissing, rules.CodePinMissing, rules.CodeReferenceUncovered} {
		if !has(report.Findings, code) {
			t.Errorf("missing %s in %v", code, codes(report.Findings))
		}
	}
	if has(report.Findings, rules.CodeOK) {
		t.Error("a catalog with problems must not report ok")
	}
}

func TestRemoteTransportsRunOnlyWhenFetching(t *testing.T) {
	catalog := `{"version":1,"sources":[{"id":"doc","kind":"web","title":"Doc","urls":["https://e.test/doc"],"reviewedAt":"2026-11-01","feeds":["references/a.md"],
    "claims":[{"id":"c","statement":"s","anchor":"gone"}]}]}`
	files := map[string]string{"s/sources.json": catalog, "s/references/a.md": "x"}
	client := webtest.New(map[string]webtest.Page{"https://e.test/doc": {{Response: web.Response{Status: 200, Body: []byte("# Other\n")}}}})

	offline, err := audit.Catalog(context.Background(), deps(t, memfs.New(files), processtest.Script(), client), "s", options())
	if err != nil || has(offline.Findings, rules.CodeAnchorMissing) || len(client.Requests) != 0 {
		t.Fatalf("offline: %v, err %v, requests %v", codes(offline.Findings), err, client.Requests)
	}
	if len(offline.Sources[0].Observations) != 0 {
		t.Errorf("observations = %+v", offline.Sources[0].Observations)
	}

	opts := options()
	opts.Fetch = true
	online, err := audit.Catalog(context.Background(), deps(t, memfs.New(files), processtest.Script(), client), "s", opts)
	if err != nil || !has(online.Findings, rules.CodeAnchorMissing) {
		t.Fatalf("online: %v, err %v", codes(online.Findings), err)
	}
	if got := online.Sources[0].Observations; len(got) != 1 || got[0].Transport != "http" {
		t.Errorf("observations = %+v", got)
	}
}

func TestLocalTransportsRunOffline(t *testing.T) {
	catalog := `{"version":1,"sources":[{"id":"file","kind":"filesystem","title":"File","path":"docs/spec.md","reviewedAt":"2026-11-01","feeds":["references/a.md"],
    "claims":[{"id":"c","statement":"s","anchor":"gone"}]}]}`
	fsys := memfs.New(map[string]string{"s/sources.json": catalog, "s/references/a.md": "x", "s/docs/spec.md": "# Spec\n"})
	report, err := audit.Catalog(context.Background(), deps(t, fsys, processtest.Script(), webtest.New(nil)), "s", options())
	if err != nil || !has(report.Findings, rules.CodeAnchorMissing) {
		t.Fatalf("%v, err %v", codes(report.Findings), err)
	}
}

func TestATransportThatCannotRunIsAFinding(t *testing.T) {
	catalog := `{"version":1,"sources":[{"id":"up","kind":"git","title":"Up","checkout":"../up","commit":"abc1234","reviewedAt":"2026-11-01","feeds":["references/a.md"]}]}`
	fsys := memfs.New(map[string]string{"s/sources.json": catalog, "s/references/a.md": "x", "up/.git/HEAD": "ref"})
	runner := processtest.Script().On("git tag --list", processtest.Reply{Err: errors.New("git is not installed")})
	report, err := audit.Catalog(context.Background(), deps(t, fsys, runner, webtest.New(nil)), "s", options())
	if err != nil {
		t.Fatal(err)
	}
	var fault findings.Finding
	for _, f := range report.Findings {
		if f.Code == rules.CodeProviderError {
			fault = f
		}
	}
	if !strings.Contains(fault.Message, "the git transport failed: git is not installed") {
		t.Fatalf("findings = %v", report.Findings)
	}
}

func TestCoverDirCanBeTurnedOffOrBeAbsent(t *testing.T) {
	fsys := memfs.New(map[string]string{"s/sources.json": cleanCatalog, "s/references/a.md": "x", "s/references/extra.md": "x"})
	opts := options()
	opts.CoverDir = ""
	report, err := audit.Catalog(context.Background(), deps(t, fsys, processtest.Script(), webtest.New(nil)), "s", opts)
	if err != nil || has(report.Findings, rules.CodeReferenceUncovered) {
		t.Fatalf("%v, err %v", codes(report.Findings), err)
	}
	// The cleanCatalog feeds references/a.md, which this file system lacks.
	noDir := memfs.New(map[string]string{"s/sources.json": cleanCatalog})
	opts.CoverDir = "references"
	report, err = audit.Catalog(context.Background(), deps(t, noDir, processtest.Script(), webtest.New(nil)), "s", opts)
	if err != nil || has(report.Findings, rules.CodeReferenceUncovered) || !has(report.Findings, rules.CodeReferenceDangling) {
		t.Fatalf("%v, err %v", codes(report.Findings), err)
	}
}

type faulty struct {
	fs.FileSystem
	readDir error
	only    string
}

func (f faulty) ReadDir(p string) ([]fs.Entry, error) {
	if f.readDir != nil && (f.only == "" || p == f.only) {
		return nil, f.readDir
	}
	return f.FileSystem.ReadDir(p)
}

func TestFileSystemFaultsAreErrors(t *testing.T) {
	boom := errors.New("boom")
	base := memfs.New(map[string]string{"s/sources.json": cleanCatalog, "s/references/a.md": "x"})

	if _, err := audit.Catalog(context.Background(), deps(t, faulty{FileSystem: base, readDir: boom, only: "s/sources.d"}, processtest.Script(), webtest.New(nil)), "s", options()); !errors.Is(err, boom) {
		t.Errorf("loader fault: err = %v", err)
	}
	if _, err := audit.Catalog(context.Background(), deps(t, faulty{FileSystem: base, readDir: boom, only: "s/references"}, processtest.Script(), webtest.New(nil)), "s", options()); !errors.Is(err, boom) {
		t.Errorf("listing fault: err = %v", err)
	}
}

func TestMappingIgnoresAFailingExistenceCheck(t *testing.T) {
	// A feed whose existence cannot be decided is reported as dangling.
	boom := errors.New("boom")
	fsys := statFaulty{FileSystem: memfs.New(map[string]string{"s/sources.json": cleanCatalog}), err: boom}
	report, err := audit.Catalog(context.Background(), deps(t, fsys, processtest.Script(), webtest.New(nil)), "s", options())
	if err == nil && !has(report.Findings, rules.CodeReferenceDangling) {
		t.Fatalf("%v", codes(report.Findings))
	}
}

type statFaulty struct {
	fs.FileSystem
	err error
}

func (s statFaulty) Stat(p string) (fs.Info, error) {
	if strings.HasSuffix(p, "a.md") {
		return fs.Info{}, s.err
	}
	return s.FileSystem.Stat(p)
}
