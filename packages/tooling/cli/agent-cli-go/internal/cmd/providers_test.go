package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
)

// The places of the Fake host, whose home is /home/user and whose working
// directory is /work.
const (
	userProviders    = "/home/user/.config/agentic/providers.d"
	managedProviders = "/etc/agentic/providers.d"
	projectFile      = "/work/agentic.json"
	projectProviders = "/work/providers.d"
	ledgerFile       = "/home/user/.local/state/agentic/trust.json"
)

const projectRoute = `{"name":"proj","agent":"claude","models":{"main":{"id":"model-main"}},"environment":{"BASE_URL":"https://models.example.test/api","ROUTE_MODEL":"{model:main}"}}`

// projectDigest is the digest the trust gate approves for the project route.
var projectDigest = layer.Digest([]byte(projectRoute))

// projectRoot is a root whose project defines the provider proj in a shard
// beside an agentic.json that sets no gated section.
func projectRoot() *testRoot {
	r := newTestRoot()
	r.host.AddFile(projectFile, `{"version":1}`, 0o644)
	r.host.AddFile(projectProviders+"/proj.json", projectRoute, 0o644)
	return r
}

func TestProvidersListNamesTheDirectoriesWhenNothingIsDefined(t *testing.T) {
	r := newTestRoot()
	r.host.AddFile(projectFile, `{"version":1}`, 0o644)

	if err := r.execute("providers", "list"); err != nil {
		t.Fatalf("providers list error = %v", err)
	}

	for _, want := range []string{"No provider is defined", "managed: " + managedProviders, "user: " + userProviders, "project: " + projectProviders} {
		if !strings.Contains(r.out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, r.out.String())
		}
	}
}

func TestProvidersListShowsTheLayerAndTheApprovalOfEachProvider(t *testing.T) {
	r := projectRoot()
	r.host.AddFile(userProviders+"/alpha.json", `{"name":"alpha","agent":"claude","portable":true}`, 0o644)
	r.host.AddFile("/etc/agentic/agentic.json", `{"version":1,"providers":{"gamma":{"agent":"codex","extends":"alpha"}}}`, 0o644)

	if err := r.execute("providers", "list"); err != nil {
		t.Fatalf("providers list error = %v", err)
	}

	want := "alpha claude user " + userProviders + "/alpha.json portable\n" +
		"proj claude project " + projectProviders + "/proj.json (not approved)\n" +
		"gamma codex managed /etc/agentic/agentic.json extends alpha\n"
	if r.out.String() != want {
		t.Errorf("output = %q, want %q", r.out.String(), want)
	}
}

func TestProvidersListReportsAFailure(t *testing.T) {
	r := newTestRoot()
	r.host.AddFile(userProviders+"/broken.json", "{", 0o644)
	if err := r.execute("providers", "list"); err == nil || !strings.Contains(err.Error(), "broken.json") {
		t.Errorf("providers list error = %v", err)
	}
	if err := writeProviders(brokenWriter{}, launch.ProviderListing{}); err == nil {
		t.Error("writeProviders() ignored a write failure on the empty listing")
	}
	rows := launch.ProviderListing{Rows: make([]launch.ProviderRow, 1)}
	if err := writeProviders(brokenWriter{}, rows); err == nil {
		t.Error("writeProviders() ignored a write failure")
	}
}

func TestProvidersShowDescribesTheRouteAndTheApproval(t *testing.T) {
	r := projectRoot()
	r.host.AddFile(userProviders+"/proj.json", `{"name":"proj","agent":"codex"}`, 0o644)

	if err := r.execute("providers", "show", "proj"); err != nil {
		t.Fatalf("providers show error = %v", err)
	}

	for _, want := range []string{
		"Name: proj\nAgent: claude\nLayer: project\nFile: " + projectProviders + "/proj.json",
		"Route hosts: models.example.test", "Model slots: main=model-main", "Approval: not approved",
		"\n\nName: proj\nAgent: codex\nLayer: user", "Route hosts: (none)",
	} {
		if !strings.Contains(r.out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, r.out.String())
		}
	}
}

func TestProvidersShowNamesExtendsAndRefusesAnUnknownProvider(t *testing.T) {
	r := newTestRoot()
	r.host.AddFile(userProviders+"/child.json", `{"name":"child","agent":"claude","extends":"base"}`, 0o644)
	if err := r.execute("providers", "show", "child"); err != nil || !strings.Contains(r.out.String(), "Extends: base") {
		t.Errorf("providers show error = %v, output %q", err, r.out.String())
	}

	var usage UsageError
	if err := r.execute("providers", "show", "ghost"); !errors.As(err, &usage) || !strings.Contains(err.Error(), "provider ghost is not defined") {
		t.Errorf("providers show ghost error = %v, want a usage error", err)
	}
	r.host.AddFile(userProviders+"/broken.json", "{", 0o644)
	if err := r.execute("providers", "show", "child"); err == nil {
		t.Error("providers show read a broken file")
	}
}

func TestProvidersShowReportsAFailureToWrite(t *testing.T) {
	r := projectRoot()
	r.host.AddFile(userProviders+"/proj.json", `{"name":"proj","agent":"codex"}`, 0o644)
	r.deps.Out = brokenWriter{}
	if err := r.execute("providers", "show", "proj"); err == nil {
		t.Error("providers show ignored a write failure")
	}
	r.deps.Out = &failAfter{n: 1}
	if err := r.execute("providers", "show", "proj"); err == nil {
		t.Error("providers show ignored a write failure between the providers")
	}
}

func TestProvidersApproveWithoutProjectContentSaysSo(t *testing.T) {
	r := newTestRoot()
	if err := r.execute("providers", "approve"); err != nil {
		t.Fatalf("providers approve error = %v", err)
	}
	if !strings.Contains(r.out.String(), "supplies no file that needs approval") {
		t.Errorf("output = %q", r.out.String())
	}
}

func TestProvidersApproveShowsEachFileAndRecordsTheDigestOnYes(t *testing.T) {
	r := projectRoot()
	project := `{"version":1,"run":{"models":["main=project-id"]}}`
	r.host.AddFile(projectFile, project, 0o644)
	r.host.Terminal = true
	r.host.Input = []string{"yes", "y"}

	if err := r.execute("providers", "approve"); err != nil {
		t.Fatalf("providers approve error = %v", err)
	}

	out := r.out.String()
	for _, want := range []string{
		projectFile + " sets the run configuration", project, projectProviders + "/proj.json sets the provider configuration", projectRoute,
		"Digest: " + projectDigest, "Approve this digest? [y/N] ", "Approved " + projectDigest + " (" + projectProviders + "/proj.json)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if !r.host.Exists(ledgerFile) {
		t.Fatal("the approval was not recorded")
	}
	if err := r.execute("providers", "list"); err != nil || !strings.Contains(r.out.String(), "(approved)") {
		t.Errorf("providers list after the approval = %v, %q", err, r.out.String())
	}
}

func TestProvidersApproveDeclinesAndFailsClosedWithoutATerminal(t *testing.T) {
	r := projectRoot()
	r.host.Terminal = true
	r.host.Input = []string{"n"}
	if err := r.execute("providers", "approve"); !errors.Is(err, trust.ErrNotApproved) || !strings.Contains(err.Error(), "declined") {
		t.Errorf("providers approve error = %v, want the refusal", err)
	}

	r = projectRoot()
	err := r.execute("providers", "approve")
	if err == nil || !strings.Contains(err.Error(), "--approve-config-digest "+projectDigest) {
		t.Fatalf("providers approve without a terminal error = %v, want the digest to approve", err)
	}
	if r.host.Exists(ledgerFile) {
		t.Error("a refused approval changed the ledger")
	}
	if err := r.execute("providers", "approve", "--approve-config-digest", projectDigest); err != nil || !r.host.Exists(ledgerFile) {
		t.Errorf("providers approve --approve-config-digest error = %v", err)
	}
	if err := r.execute("providers", "approve", "--approve-config-digest", "sha256:"+strings.Repeat("0", 64)); err == nil || !strings.Contains(err.Error(), "digest of no project file") {
		t.Errorf("a different digest error = %v", err)
	}
}

func TestProvidersApproveStopsWhenTheLedgerCannotBeReadOrTheOutputWritten(t *testing.T) {
	r := projectRoot()
	r.host.AddFile(ledgerFile, "{", 0o600)
	r.host.Terminal = true
	r.host.Input = []string{"y"}
	if err := r.execute("providers", "approve"); err == nil {
		t.Error("providers approve replaced a corrupt ledger")
	}

	r = projectRoot()
	r.deps.Out = brokenWriter{}
	if err := r.execute("providers", "approve", "--approve-config-digest", projectDigest); err == nil {
		t.Error("providers approve ignored a write failure")
	}
	r = newTestRoot()
	r.deps.Out = brokenWriter{}
	if err := r.execute("providers", "approve"); err == nil {
		t.Error("providers approve ignored a write failure without project content")
	}
	r = newTestRoot()
	r.host.AddFile(userProviders+"/broken.json", "{", 0o644)
	if err := r.execute("providers", "approve"); err == nil {
		t.Error("providers approve read a broken file")
	}
}

func TestProvidersRevokeRemovesTheApprovalOfTheProjectFiles(t *testing.T) {
	r := projectRoot()
	if err := r.execute("providers", "revoke"); err != nil || !strings.Contains(r.out.String(), "holds no approval of a project file") {
		t.Fatalf("providers revoke before an approval = %v, %q", err, r.out.String())
	}
	if err := r.execute("providers", "approve", "--approve-config-digest", projectDigest); err != nil {
		t.Fatal(err)
	}
	r.out.Reset()

	if err := r.execute("providers", "revoke"); err != nil || !strings.Contains(r.out.String(), "Revoked "+projectDigest+" ("+projectProviders+"/proj.json)") {
		t.Fatalf("providers revoke = %v, %q", err, r.out.String())
	}
	if err := r.execute("run", "--dry-run", "--provider", "proj"); err == nil || !strings.Contains(err.Error(), "not approved") {
		t.Errorf("run after the revocation error = %v, want the gate to ask again", err)
	}
}

func TestProvidersRevokeRemovesOneDigest(t *testing.T) {
	r := projectRoot()
	if err := r.execute("providers", "approve", "--approve-config-digest", projectDigest); err != nil {
		t.Fatal(err)
	}
	r.out.Reset()

	if err := r.execute("providers", "revoke", projectDigest); err != nil || r.out.String() != "Revoked "+projectDigest+"\n" {
		t.Errorf("providers revoke digest = %v, %q", err, r.out.String())
	}
	if err := r.execute("providers", "revoke", projectDigest); err == nil || !strings.Contains(err.Error(), "holds no approval of") {
		t.Errorf("a second revocation error = %v", err)
	}
}

func TestProvidersRevokeReportsAFailure(t *testing.T) {
	r := projectRoot()
	r.host.AddFile(ledgerFile, "{", 0o600)
	if err := r.execute("providers", "revoke", projectDigest); err == nil {
		t.Error("providers revoke digest read a corrupt ledger")
	}
	if err := r.execute("providers", "revoke"); err == nil {
		t.Error("providers revoke read a corrupt ledger")
	}

	r = newTestRoot()
	r.host.AddFile(userProviders+"/broken.json", "{", 0o644)
	if err := r.execute("providers", "revoke"); err == nil {
		t.Error("providers revoke read a broken file")
	}

	r = projectRoot()
	if err := r.execute("providers", "approve", "--approve-config-digest", projectDigest); err != nil {
		t.Fatal(err)
	}
	r.deps.Out = brokenWriter{}
	if err := r.execute("providers", "revoke"); err == nil {
		t.Error("providers revoke ignored a write failure")
	}
	r = newTestRoot()
	r.deps.Out = brokenWriter{}
	if err := r.execute("providers", "revoke"); err == nil {
		t.Error("providers revoke ignored a write failure without an approval")
	}
}

func TestRunAsksForTheProjectProviderAndRunsOnYes(t *testing.T) {
	r := projectRoot()
	r.host.Terminal = true
	r.host.Input = []string{"y"}

	if err := r.execute("run", "--dry-run", "--provider", "proj"); err != nil {
		t.Fatalf("run error = %v", err)
	}
	if !strings.Contains(r.out.String(), "Approve this digest? [y/N] ") || !r.host.Exists(ledgerFile) {
		t.Errorf("output = %q, ledger = %v", r.out.String(), r.host.Exists(ledgerFile))
	}

	r.out.Reset()
	if err := r.execute("run", "--dry-run", "--provider", "proj"); err != nil || strings.Contains(r.out.String(), "Approve") {
		t.Errorf("the second run error = %v, output = %q, want no question", err, r.out.String())
	}
}

func TestRunFailsClosedWithoutATerminalAndTakesTheDigestFromItsFlag(t *testing.T) {
	r := projectRoot()
	err := r.execute("run", "--dry-run", "--provider", "proj")
	if err == nil || !strings.Contains(err.Error(), "--approve-config-digest "+projectDigest) {
		t.Fatalf("run error = %v, want the digest to approve", err)
	}
	if r.host.Exists(ledgerFile) {
		t.Error("a refused approval changed the ledger")
	}
	if err := r.execute("run", "--dry-run", "--provider", "proj", "--approve-config-digest", projectDigest); err != nil {
		t.Errorf("run with the digest error = %v", err)
	}
}

func TestRunStopsWhenTheModelOverrideNamesNoSlot(t *testing.T) {
	r := newTestRoot()
	r.host.AddFile(userProviders+"/alpha.json", `{"name":"alpha","agent":"claude","models":{"main":{"id":"model-main"}}}`, 0o644)
	if err := r.execute("run", "--dry-run", "--provider", "alpha", "--model", "ghost=x"); err == nil || !strings.Contains(err.Error(), `no model slot "ghost"`) {
		t.Errorf("run error = %v", err)
	}
	if err := r.execute("run", "--dry-run", "--provider", "alpha", "--model", "main=other"); err != nil {
		t.Errorf("run with a model override error = %v", err)
	}
}

// failAfter accepts n writes, then fails every write.
type failAfter struct{ n int }

func (f *failAfter) Write(p []byte) (int, error) {
	if f.n == 0 {
		return 0, errors.New("closed")
	}
	f.n--
	return len(p), nil
}
