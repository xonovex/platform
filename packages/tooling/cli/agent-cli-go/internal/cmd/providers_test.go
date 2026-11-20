package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/trust"
)

const (
	userProviders    = "/home/user/.config/sandboxed-claude/providers"
	projectProviders = "/work/.agent-cli/providers"
	ledgerFile       = "/home/user/.local/share/agent-cli/approvals.json"
)

const projectRoute = `{"name":"proj","agent":"claude","models":{"main":{"id":"model-main"}},"environment":{"BASE_URL":"https://models.example.test/api","ROUTE_MODEL":"{model:main}"}}`

func projectRoot() *testRoot {
	r := newTestRoot()
	r.host.AddFile(projectProviders+"/proj.json", projectRoute, 0o644)
	return r
}

func TestProvidersListNamesTheDirectoriesWhenNothingIsDefined(t *testing.T) {
	r := newTestRoot()

	if err := r.execute("providers", "list"); err != nil {
		t.Fatalf("providers list error = %v", err)
	}

	for _, want := range []string{"No provider is defined", "managed: /etc/agent-cli/providers", "user: " + userProviders, "project: " + projectProviders} {
		if !strings.Contains(r.out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, r.out.String())
		}
	}
}

func TestProvidersListShowsTheLayerAndTheApprovalOfEachProvider(t *testing.T) {
	r := projectRoot()
	r.host.AddFile(userProviders+"/alpha.json", `{"name":"alpha","agent":"claude","portable":true}`, 0o644)
	r.host.AddFile("/etc/agent-cli/providers/gamma.json", `{"name":"gamma","agent":"codex","extends":"alpha"}`, 0o644)

	if err := r.execute("providers", "list"); err != nil {
		t.Fatalf("providers list error = %v", err)
	}

	want := "alpha claude user " + userProviders + "/alpha.json portable\n" +
		"proj claude project " + projectProviders + "/proj.json (not approved)\n" +
		"gamma codex managed /etc/agent-cli/providers/gamma.json extends alpha\n"
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

func TestProvidersApproveWithoutProjectContentSaysSo(t *testing.T) {
	r := newTestRoot()
	if err := r.execute("providers", "approve"); err != nil {
		t.Fatalf("providers approve error = %v", err)
	}
	if !strings.Contains(r.out.String(), "supplies no provider files and no model overrides") {
		t.Errorf("output = %q", r.out.String())
	}
}

func TestProvidersApprovePreviewsTheMaterialAndRecordsTheDigestOnYes(t *testing.T) {
	r := projectRoot()
	r.host.AddFile("/work/.agent-cli.yaml", "models: [main=project-id]\n", 0o644)
	r.host.Terminal = true
	r.host.Input = []string{"yes"}

	if err := r.execute("providers", "approve"); err != nil {
		t.Fatalf("providers approve error = %v", err)
	}

	out := r.out.String()
	for _, want := range []string{
		"File: " + projectProviders + "/proj.json", "Provider: proj (agent claude)", "Route hosts: models.example.test",
		"Model slots: main=model-main", "Model overrides: main=project-id", "Digest: sha256:", "Approve this digest? [y/N] ", "Approved sha256:",
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
	if err := r.execute("providers", "approve"); !errors.Is(err, trust.ErrDeclined) {
		t.Errorf("providers approve error = %v, want ErrDeclined", err)
	}

	r = projectRoot()
	err := r.execute("providers", "approve")
	if err == nil || !strings.Contains(err.Error(), "--approve-digest sha256:") {
		t.Fatalf("providers approve without a terminal error = %v, want the digest to approve", err)
	}
	if r.host.Exists(ledgerFile) {
		t.Error("a refused approval changed the ledger")
	}
	_, digest, _ := strings.Cut(err.Error(), "--approve-digest ")
	if err := r.execute("providers", "approve", "--approve-digest", digest); err != nil || !r.host.Exists(ledgerFile) {
		t.Errorf("providers approve --approve-digest error = %v", err)
	}
	if err := r.execute("providers", "approve", "--approve-digest", "sha256:"+strings.Repeat("0", 64)); err == nil || !strings.Contains(err.Error(), "differs") {
		t.Errorf("a different digest error = %v", err)
	}
}

func TestProvidersApproveStopsWhenTheLedgerCannotBeOpenedOrWritten(t *testing.T) {
	r := projectRoot()
	r.ports.Host = &homeless{Fake: r.host}
	r.host.Terminal = true
	r.host.Input = []string{"y"}
	if err := r.execute("providers", "approve"); err == nil || !strings.Contains(err.Error(), "no home") {
		t.Errorf("providers approve error = %v, want the home failure", err)
	}

	r = projectRoot()
	r.host.AddFile(ledgerFile, "{", 0o600)
	r.host.Terminal = true
	r.host.Input = []string{"y"}
	if err := r.execute("providers", "approve"); err == nil {
		t.Error("providers approve replaced a corrupt ledger")
	}
}

// homeless is a host whose home directory cannot be resolved.
type homeless struct{ *oshosttest.Fake }

func (homeless) UserHomeDir() (string, error) { return "", errors.New("no home") }

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
	if err == nil || !strings.Contains(err.Error(), "--approve-project-digest sha256:") {
		t.Fatalf("run error = %v, want the digest to approve", err)
	}
	if r.host.Exists(ledgerFile) {
		t.Error("a refused approval changed the ledger")
	}
	_, digest, _ := strings.Cut(err.Error(), "--approve-project-digest ")
	if err := r.execute("run", "--dry-run", "--provider", "proj", "--approve-project-digest", digest); err != nil {
		t.Errorf("run with the digest error = %v", err)
	}
}

func TestRunWithoutApprovalPortsReturnsTheApprovalError(t *testing.T) {
	r := projectRoot()
	root := NewRootCommand(r.deps, PluginPorts{}, RecordPorts{}, r.bridge)
	root.SetArgs([]string{"run", "--dry-run", "--provider", "proj"})
	root.SetOut(&r.out)
	root.SetErr(&r.out)

	var needed *launch.ApprovalRequiredError
	if err := root.Execute(); !errors.As(err, &needed) {
		t.Errorf("run error = %v, want ApprovalRequiredError", err)
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

func TestDefinedProvidersOffersNothingWhenTheProvidersCannotBeRead(t *testing.T) {
	r := newTestRoot()
	r.host.HomeErr = errors.New("no home")
	if got := definedProviders(r.deps, launch.Options{}, "claude"); got != nil {
		t.Errorf("definedProviders() = %v, want none", got)
	}
}
