package launch

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/legacy"
)

const legacyUserConfig = "/home/user/.config/sandboxed-claude/config"

func TestExecuteReadsALegacyFileAndWarnsOfIt(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(legacyUserConfig, "provider=alpha\n", 0o600)
	deps := f.deps()
	deps.Legacy = legacy.Adapter{}

	if err := Execute(deps, dryRunOptions()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(f.log.String(), legacyUserConfig+" is a legacy user configuration file") {
		t.Errorf("log = %q, want the legacy notice", f.log.String())
	}
}

func TestListProvidersMarksALegacyProjectByThePreviousLedger(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile("/work/repo/.agent-cli/providers/proj.json", providerFile("proj", "claude", "Project"), 0o644)
	deps := f.deps()
	deps.Legacy = legacy.Adapter{}

	listing, err := ListProviders(deps, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if rows := listing.Named("proj"); len(rows) != 1 || rows[0].Approval != "not approved" || len(listing.Pending) != 1 {
		t.Fatalf("listing = %+v, want the legacy project pending", listing)
	}

	f.host.AddFile("/home/user/.local/share/agent-cli/approvals.json", `{"version":1,"approved":[{"kind":"project-providers","subject":"/work/repo","digest":"`+listing.Pending[0].Subject.Digest+`"}]}`, 0o600)
	listing, err = ListProviders(deps, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if rows := listing.Named("proj"); len(rows) != 1 || rows[0].Approval != "approved" || len(listing.Pending) != 0 {
		t.Errorf("listing = %+v, want the legacy project approved by the previous ledger", listing)
	}
}
