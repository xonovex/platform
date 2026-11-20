package launch

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
)

const (
	projectProviders = "/work/repo/providers.d"
	managedProviders = "/etc/agentic/providers.d"
)

// withProject gives the Fake host a project agentic.json that sets no gated
// section, so the provider shards beside it are the only project content.
func withProject(f *executeFixture) *executeFixture {
	f.host.AddFile(projectFile, `{"version":1}`, 0o644)
	return f
}

func dryRunOptions() Options {
	options := bwrapOptions()
	options.DryRun = true
	return options
}

func TestModelOverridesApplyInLayerOrderAndNameTheirSource(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(userFile, runFile(`{"provider":"alpha","models":["main=user-id"]}`), 0o600)
	options := dryRunOptions()
	options.Models = []string{"main=flag-id"}

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	log := f.log.String()
	for _, want := range []string{
		"models: main=user-id (user " + userFile + ")",
		"models: main=flag-id (flags command line)",
		"Model main: flag-id (flags command line)",
	} {
		if !strings.Contains(log, want) {
			t.Errorf("dry run lacks %q:\n%s", want, log)
		}
	}
	if got := f.sandbox.isolator.commandRun.Provider.Environment["ROUTE_MODEL"]; got != "flag-id" {
		t.Errorf("ROUTE_MODEL = %q, want the command line override", got)
	}
}

func TestAManagedModelOverrideBeatsTheLayersBelowAndRefusesAContradiction(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(managedFile, runFile(`{"models":["main=managed-id"]}`), 0o644)
	options := dryRunOptions()
	options.Provider = "alpha"

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := f.sandbox.isolator.commandRun.Provider.Slots[0]; got.ID != "managed-id" || !strings.HasPrefix(got.Source, "managed ") {
		t.Errorf("slot = %+v, want the managed override", got)
	}

	options.Models = []string{"main=flag-id"}
	err := Execute(f.deps(), options)
	if err == nil || !strings.Contains(err.Error(), "the managed value cannot be changed") {
		t.Errorf("Execute() error = %v, want the managed contradiction", err)
	}

	options.Models = []string{"main=managed-id"}
	if err := Execute(f.deps(), options); err != nil {
		t.Errorf("Execute() with the managed value error = %v", err)
	}
}

func TestModelOverridesFailClosed(t *testing.T) {
	cases := map[string]struct {
		provider string
		models   []string
		phrase   string
	}{
		"no provider": {"", []string{"main=x"}, "needs a provider"},
		"no equals":   {"alpha", []string{"main"}, "must be slot=id"},
		"no slot":     {"alpha", []string{"=x"}, "must be slot=id"},
		"no id":       {"alpha", []string{"main="}, "must be slot=id"},
		"unknown":     {"alpha", []string{"ghost=x"}, `no model slot "ghost"`},
	}
	for name, test := range cases {
		f := newExecuteFixture()
		options := dryRunOptions()
		options.Provider = test.provider
		options.Models = test.models

		if err := Execute(f.deps(), options); err == nil || !strings.Contains(err.Error(), test.phrase) {
			t.Errorf("%s: Execute() error = %v, want %q", name, err, test.phrase)
		}
	}
}

func TestAProviderThatIsNotDefinedNamesTheDirectoryToAddItTo(t *testing.T) {
	f := newExecuteFixture()
	options := dryRunOptions()
	options.Provider = "ghost"

	err := Execute(f.deps(), options)

	if err == nil || !strings.Contains(err.Error(), "provider ghost is not defined; add ghost.json to "+userProviders) {
		t.Errorf("Execute() error = %v, want the directory to add the file to", err)
	}

	empty := newExecuteFixture()
	empty.host = oshosttest.New()
	empty.host.WorkingDir = "/work/repo"
	if err := Execute(empty.deps(), options); err == nil || !strings.Contains(err.Error(), "no provider is defined") {
		t.Errorf("Execute() without providers error = %v", err)
	}
}

func TestAProjectProviderFileTakesEffectOnlyAfterTheDigestIsApproved(t *testing.T) {
	f := withProject(newExecuteFixture())
	file := providerFile("proj", "claude", "Project route")
	f.host.AddFile(projectProviders+"/proj.json", file, 0o644)
	options := dryRunOptions()
	options.Provider = "proj"

	err := Execute(f.deps(), options)

	if err == nil || !strings.Contains(err.Error(), "--approve-config-digest "+digestOf(file)) {
		t.Fatalf("Execute() error = %v, want the digest to approve", err)
	}
	if f.sandbox.isolator.commandRun.Provider != nil {
		t.Error("the run reached the isolator without approval")
	}

	options.ApproveDigest = digestOf(file)
	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() with the approval error = %v", err)
	}
	options.ApproveDigest = ""
	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() after the approval error = %v, want the ledger to hold it", err)
	}

	f.host.AddFile(projectProviders+"/proj.json", file+"\n", 0o644)
	if err := Execute(f.deps(), options); err == nil || !strings.Contains(err.Error(), "not approved") {
		t.Errorf("Execute() after the file changed error = %v, want a new approval", err)
	}
}

func TestProjectModelOverridesNeedApprovalEvenForAUserProvider(t *testing.T) {
	f := newExecuteFixture()
	data := runFile(`{"provider":"alpha","models":["main=project-id"]}`)
	f.host.AddFile(projectFile, data, 0o644)

	if err := Execute(f.deps(), dryRunOptions()); err == nil || !strings.Contains(err.Error(), "sets the run configuration") {
		t.Fatalf("Execute() error = %v, want the project file to wait for its approval", err)
	}
	options := dryRunOptions()
	options.ApproveDigest = digestOf(data)
	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() with the approval error = %v", err)
	}
	if got := f.sandbox.isolator.commandRun.Provider.Slots[0]; got.ID != "project-id" || !strings.HasPrefix(got.Source, "project ") {
		t.Errorf("slot = %+v, want the project override", got)
	}
}

func TestTheTrustGateAsksAtATerminal(t *testing.T) {
	f := withProject(newExecuteFixture())
	file := `{"name":"child","agent":"claude","extends":"alpha","models":{"main":{"id":"child-id"}}}`
	f.host.AddFile(projectProviders+"/child.json", file, 0o644)
	f.host.Terminal = true
	f.host.Input = []string{"no", "yes"}
	options := dryRunOptions()
	options.Provider = "child"

	if err := Execute(f.deps(), options); err == nil || !strings.Contains(err.Error(), "was declined") {
		t.Fatalf("Execute() error = %v, want the refusal", err)
	}
	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() after yes error = %v", err)
	}
	for _, want := range []string{file, "Digest: " + digestOf(file), "Approve this digest? [y/N]"} {
		if !strings.Contains(f.out.String(), want) {
			t.Errorf("the gate output lacks %q:\n%s", want, f.out.String())
		}
	}
}

func TestAFlagThatNamesNoProjectFileStopsTheRun(t *testing.T) {
	f := withProject(newExecuteFixture())
	f.host.AddFile(projectProviders+"/proj.json", providerFile("proj", "claude", "P"), 0o644)
	options := dryRunOptions()
	options.ApproveDigest = digestOf("other")

	if err := Execute(f.deps(), options); err == nil || !strings.Contains(err.Error(), "is the digest of no project file") {
		t.Errorf("Execute() error = %v, want the stale digest refused", err)
	}
}

func TestAProjectFileNeverChangesAManagedProvider(t *testing.T) {
	f := withProject(newExecuteFixture())
	f.host.AddFile(managedFile, `{"version":1}`, 0o644)
	f.host.AddFile(managedProviders+"/alpha.json", `{"name":"alpha","agent":"claude","models":{"main":{"id":"managed-id"}},"environment":{"ROUTE_MODEL":"{model:main}"}}`, 0o644)
	project := `{"name":"alpha","agent":"claude","models":{"main":{"id":"project-id"}},"environment":{"ROUTE_MODEL":"{model:main}"}}`
	f.host.AddFile(projectProviders+"/alpha.json", project, 0o644)
	options := dryRunOptions()
	options.Provider = "alpha"
	options.ApproveDigest = digestOf(project)

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := f.sandbox.isolator.commandRun.Provider.Environment["ROUTE_MODEL"]; got != "managed-id" {
		t.Errorf("ROUTE_MODEL = %q, want the managed value", got)
	}
}

func TestALedgerThatCannotBeReadStopsTheRun(t *testing.T) {
	f := withProject(newExecuteFixture())
	f.host.AddFile(projectProviders+"/proj.json", providerFile("proj", "claude", "P"), 0o644)
	f.host.AddFile("/home/user/.local/state/agentic/trust.json", "{", 0o600)
	options := dryRunOptions()
	options.Provider = "proj"

	if err := Execute(f.deps(), options); err == nil || !strings.Contains(err.Error(), "approval ledger") {
		t.Errorf("Execute() error = %v", err)
	}
	if _, err := ListProviders(f.deps(), options); err == nil || !strings.Contains(err.Error(), "approval ledger") {
		t.Errorf("ListProviders() error = %v", err)
	}
}

func TestListProvidersShowsTheLayerAndTheApprovalOfProjectFiles(t *testing.T) {
	f := withProject(newExecuteFixture())
	project := providerFile("proj", "claude", "P")
	f.host.AddFile(projectProviders+"/proj.json", project, 0o644)
	f.host.AddFile(managedFile, `{"version":1,"providers":{"zeta":{"agent":"codex","portable":true}}}`, 0o644)

	listing, err := ListProviders(f.deps(), Options{})

	if err != nil {
		t.Fatalf("ListProviders() error = %v", err)
	}
	var lines []string
	for _, row := range listing.Rows {
		lines = append(lines, row.Name+" "+row.Layer+" "+row.Approval)
	}
	want := []string{"alpha user ", "beta user ", "proj project not approved", "zeta managed "}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Errorf("rows = %q, want %q", lines, want)
	}
	if len(listing.Dirs) != 3 || len(listing.Pending) != 1 || listing.Rows[3].Origin != managedFile {
		t.Errorf("Dirs = %v, Pending = %+v, rows = %+v", listing.Dirs, listing.Pending, listing.Rows)
	}
	if rows := listing.Named("proj"); len(rows) != 1 || string(rows[0].Document.Raw) != project {
		t.Errorf("Named(proj) = %+v, want the project document", rows)
	}

	options := Options{ApproveDigest: digestOf(project)}
	if _, err := LoadConfig(f.deps(), options); err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	approved, err := ListProviders(f.deps(), Options{})
	if err != nil || approved.Rows[2].Approval != "approved" {
		t.Errorf("ListProviders() = %+v, %v, want the project file approved", approved.Rows, err)
	}
}

func TestListProvidersFailsOnABrokenFileAndAnUnreadableHome(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(userProviders+"/broken.json", "{", 0o644)
	if _, err := ListProviders(f.deps(), Options{}); err == nil || !strings.Contains(err.Error(), "broken.json") {
		t.Errorf("ListProviders() error = %v, want the broken file named", err)
	}

	g := newExecuteFixture()
	g.host.HomeErr = errors.New("no home")
	if _, err := ListProviders(g.deps(), Options{}); err == nil || !strings.Contains(err.Error(), "no home") {
		t.Errorf("ListProviders() error = %v, want the home failure", err)
	}
}

func TestPreflightAndDryRunShowTheModelSlotsWithTheirEffort(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(userProviders+"/effortful.json", `{"name":"effortful","agent":"claude","models":{"main":{"id":"model-main","effort":"high"}}}`, 0o644)
	options := dryRunOptions()
	options.Provider = "effortful"

	report := Preflight(f.deps(), options)

	var shown string
	for _, axis := range report.Axes {
		if axis.Axis == "model main" {
			shown = axis.Choice + " from " + axis.Source
		}
	}
	if want := "model-main (effort high) from " + userProviders + "/effortful.json"; shown != want {
		t.Errorf("model axis = %q, want %q (problem %q)", shown, want, report.Problem)
	}
	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(f.log.String(), "Model main: model-main (effort high) ("+userProviders+"/effortful.json)") {
		t.Errorf("dry run lacks the slot:\n%s", f.log.String())
	}
}
