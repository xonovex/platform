package launch

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// ledgerFile is the approval ledger in the state directory of the Fake host.
const ledgerFile = "/home/user/.local/state/agentic/trust.json"

// pendingProject gives the fixture a project provider file that waits for its
// approval, and returns its digest.
func pendingProject(f *executeFixture) string {
	withProject(f)
	file := providerFile("proj", "claude", "Project route")
	f.host.AddFile(projectProviders+"/proj.json", file, 0o644)
	return digestOf(file)
}

func TestApproveProjectFilesApprovesTheNamedDigest(t *testing.T) {
	f := newExecuteFixture()
	if approved, err := ApproveProjectFiles(f.deps(), Options{}); err != nil || approved != nil {
		t.Fatalf("ApproveProjectFiles() without project content = %v, %v", approved, err)
	}
	digest := pendingProject(f)

	approved, err := ApproveProjectFiles(f.deps(), Options{ApproveDigest: digest})

	if err != nil || len(approved) != 1 || approved[0].Subject.Digest != digest {
		t.Fatalf("ApproveProjectFiles() = %+v, %v", approved, err)
	}
	listing, err := ListProviders(f.deps(), Options{})
	if err != nil || listing.Named("proj")[0].Approval != "approved" {
		t.Errorf("after the approval: %+v, %v", listing.Rows, err)
	}
}

func TestApproveProjectFilesFailsClosed(t *testing.T) {
	f := newExecuteFixture()
	pendingProject(f)
	if _, err := ApproveProjectFiles(f.deps(), Options{}); err == nil || !strings.Contains(err.Error(), "--approve-config-digest") {
		t.Errorf("ApproveProjectFiles() without a terminal error = %v", err)
	}
	g := newExecuteFixture()
	g.host.AddFile(userProviders+"/broken.json", "{", 0o644)
	if _, err := ApproveProjectFiles(g.deps(), Options{}); err == nil {
		t.Error("ApproveProjectFiles() read a broken file")
	}
}

func TestRevokeProjectFilesRemovesTheirApprovals(t *testing.T) {
	f := newExecuteFixture()
	digest := pendingProject(f)
	if revoked, err := RevokeProjectFiles(f.deps(), Options{}); err != nil || len(revoked) != 0 {
		t.Fatalf("RevokeProjectFiles() before an approval = %v, %v", revoked, err)
	}
	if _, err := ApproveProjectFiles(f.deps(), Options{ApproveDigest: digest}); err != nil {
		t.Fatal(err)
	}

	revoked, err := RevokeProjectFiles(f.deps(), Options{})

	if err != nil || len(revoked) != 1 || revoked[0].Subject.Digest != digest {
		t.Errorf("RevokeProjectFiles() = %+v, %v", revoked, err)
	}
	g := newExecuteFixture()
	g.host.HomeErr = errors.New("no home")
	if _, err := RevokeProjectFiles(g.deps(), Options{}); err == nil {
		t.Error("RevokeProjectFiles() without a home directory error = nil")
	}
}

func TestRevokeDigestFailsWhenTheLedgerHoldsNoApproval(t *testing.T) {
	f := newExecuteFixture()
	digest := pendingProject(f)
	if _, err := ApproveProjectFiles(f.deps(), Options{ApproveDigest: digest}); err != nil {
		t.Fatal(err)
	}
	if err := RevokeDigest(f.deps(), digest); err != nil {
		t.Fatalf("RevokeDigest() error = %v", err)
	}
	if err := RevokeDigest(f.deps(), digest); err == nil || !strings.Contains(err.Error(), "holds no approval of "+digest) {
		t.Errorf("a second RevokeDigest() error = %v", err)
	}
	f.host.AddFile(ledgerFile, "{", 0o600)
	if err := RevokeDigest(f.deps(), digest); err == nil {
		t.Error("RevokeDigest() read a corrupt ledger")
	}
}

func TestProviderRowsListTheirSlotsAndRouteHosts(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(userProviders+"/routed.json", `{"name":"routed","agent":"claude","environment":{"ANTHROPIC_BASE_URL":"https://api.example.com/v1"},"models":{"zeta":{"id":"z-1"},"main":{"id":"m-1"}}}`, 0o644)
	listing, err := ListProviders(f.deps(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	row := listing.Named("routed")[0]
	if got := row.ModelSlots(); !slices.Equal(got, []string{"main=m-1", "zeta=z-1"}) {
		t.Errorf("ModelSlots() = %v, want them sorted by slot", got)
	}
	if got := row.RouteHosts(); !slices.Equal(got, []string{"api.example.com"}) {
		t.Errorf("RouteHosts() = %v", got)
	}
}

func TestShowConfigReturnsTheSectionsAgentCLIReads(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(userFile, runFile(`{"network":"none"}`), 0o644)

	view, err := ShowConfig(f.deps(), Options{})

	if err != nil {
		t.Fatalf("ShowConfig() error = %v", err)
	}
	if len(view.Sections) != 2 || !strings.Contains(string(view.Sections["run"]), `"network":"none"`) {
		t.Errorf("Sections = %s", view.Sections)
	}
	if len(view.Merged.Values()) == 0 {
		t.Error("the merged view holds no value")
	}
	g := newExecuteFixture()
	g.host.HomeErr = errors.New("no home")
	if _, err := ShowConfig(g.deps(), Options{}); err == nil {
		t.Error("ShowConfig() without a home directory error = nil")
	}
}
