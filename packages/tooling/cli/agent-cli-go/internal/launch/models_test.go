package launch

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
)

const (
	projectProviders = "/work/repo/.agent-cli/providers"
	managedProviders = "/etc/agent-cli/providers"
)

// fakeApprovals approves the digests it holds.
type fakeApprovals struct {
	digests map[string]bool
	err     error
}

func (f fakeApprovals) Approved(kind, digest string) (bool, error) {
	return f.digests[kind+" "+digest] && f.err == nil, f.err
}

func approving(digests ...string) fakeApprovals {
	approved := map[string]bool{}
	for _, digest := range digests {
		approved[ApprovalKind+" "+digest] = true
	}
	return fakeApprovals{digests: approved}
}

func dryRunOptions() Options {
	options := bwrapOptions()
	options.DryRun = true
	return options
}

func TestModelOverridesApplyInLayerOrderAndNameTheirSource(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(userConfig, "provider: alpha\nmodels: [main=user-id]\n", 0o600)
	options := dryRunOptions()
	options.Config = userConfig
	options.Models = []string{"main=flag-id"}

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	log := f.log.String()
	for _, want := range []string{
		"models: main=user-id (user " + userConfig + ")",
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
	f.host.AddFile("/etc/agent-cli/managed.yaml", "models: [main=managed-id]\n", 0o644)
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
	f := newExecuteFixture()
	file := providerFile("proj", "claude", "Project route")
	f.host.AddFile(projectProviders+"/proj.json", file, 0o644)
	options := dryRunOptions()
	options.Provider = "proj"

	err := Execute(f.deps(), options)

	var needed *ApprovalRequiredError
	if !errors.As(err, &needed) {
		t.Fatalf("Execute() error = %v, want ApprovalRequiredError", err)
	}
	digest := needed.Material.Digest()
	if !strings.Contains(err.Error(), "--approve-digest "+digest) {
		t.Errorf("error = %q, want it to name the digest", err)
	}
	if f.sandbox.isolator.commandRun.Provider != nil {
		t.Error("the run reached the isolator without approval")
	}

	deps := f.deps()
	deps.Approvals = approving(digest)
	if err := Execute(deps, options); err != nil {
		t.Fatalf("Execute() with the approval error = %v", err)
	}

	f.host.AddFile(projectProviders+"/proj.json", file+"\n", 0o644)
	if err := Execute(deps, options); !errors.As(err, &needed) {
		t.Errorf("Execute() after the file changed error = %v, want ApprovalRequiredError", err)
	}
}

func TestProjectModelOverridesNeedApprovalEvenForAUserProvider(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile("/work/repo/.agent-cli.yaml", "provider: alpha\nmodels: [main=project-id]\n", 0o644)

	err := Execute(f.deps(), dryRunOptions())

	var needed *ApprovalRequiredError
	if !errors.As(err, &needed) || len(needed.Material.Models) != 1 {
		t.Fatalf("Execute() error = %v, want ApprovalRequiredError for the model override", err)
	}
	deps := f.deps()
	deps.Approvals = approving(needed.Material.Digest())
	if err := Execute(deps, dryRunOptions()); err != nil {
		t.Fatalf("Execute() with the approval error = %v", err)
	}
	if got := f.sandbox.isolator.commandRun.Provider.Slots[0]; got.ID != "project-id" || !strings.HasPrefix(got.Source, "project ") {
		t.Errorf("slot = %+v, want the project override", got)
	}
}

func TestAProjectFileThatRefinesAUserProviderNeedsApproval(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(projectProviders+"/child.json", `{"name":"child","agent":"claude","extends":"alpha","models":{"main":{"id":"child-id"}}}`, 0o644)
	options := dryRunOptions()
	options.Provider = "child"

	var needed *ApprovalRequiredError
	if err := Execute(f.deps(), options); !errors.As(err, &needed) {
		t.Fatalf("Execute() error = %v, want ApprovalRequiredError", err)
	}
}

func TestAProjectFileNeverChangesAManagedProvider(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(managedProviders+"/alpha.json", `{"name":"alpha","agent":"claude","models":{"main":{"id":"managed-id"}},"environment":{"ROUTE_MODEL":"{model:main}"}}`, 0o644)
	f.host.AddFile(projectProviders+"/alpha.json", `{"name":"alpha","agent":"claude","models":{"main":{"id":"project-id"}},"environment":{"ROUTE_MODEL":"{model:main}"}}`, 0o644)
	options := dryRunOptions()
	options.Provider = "alpha"

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v, want the managed provider to run without approval", err)
	}
	if got := f.sandbox.isolator.commandRun.Provider.Environment["ROUTE_MODEL"]; got != "managed-id" {
		t.Errorf("ROUTE_MODEL = %q, want the managed value", got)
	}
}

func TestApprovalsThatCannotBeReadStopTheRun(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(projectProviders+"/proj.json", providerFile("proj", "claude", "P"), 0o644)
	deps := f.deps()
	deps.Approvals = fakeApprovals{err: errors.New("corrupt")}
	options := dryRunOptions()
	options.Provider = "proj"

	if err := Execute(deps, options); err == nil || !strings.Contains(err.Error(), "read the approvals: corrupt") {
		t.Errorf("Execute() error = %v", err)
	}
	if _, err := ListProviders(deps, options); err == nil || !strings.Contains(err.Error(), "read the approvals") {
		t.Errorf("ListProviders() error = %v", err)
	}
}

func TestListProvidersShowsTheLayerAndTheApprovalOfProjectFiles(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(projectProviders+"/proj.json", providerFile("proj", "claude", "P"), 0o644)
	f.host.AddFile(managedProviders+"/zeta.json", `{"name":"zeta","agent":"codex","portable":true}`, 0o644)

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
	if len(listing.Dirs) != 3 || listing.Material.Empty() {
		t.Errorf("Dirs = %v, Material = %+v", listing.Dirs, listing.Material)
	}

	deps := f.deps()
	deps.Approvals = approving(listing.Material.Digest())
	approved, err := ListProviders(deps, Options{})
	if err != nil || approved.Rows[2].Approval != "approved" {
		t.Errorf("ListProviders() = %+v, %v, want the project file approved", approved.Rows, err)
	}
	material, err := ProjectMaterial(deps, Options{})
	if err != nil || material.Digest() != listing.Material.Digest() {
		t.Errorf("ProjectMaterial() = %v, %v", material.Digest(), err)
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

func TestNeedsApprovalIsFalseWithoutAProviderOrMaterial(t *testing.T) {
	material := providers.Material{Models: []string{"main=x"}}
	if needsApproval(Options{}, material) {
		t.Error("a run without a provider needs no approval")
	}
	if needsApproval(Options{Provider: "alpha", Agent: "claude"}, providers.Material{}) {
		t.Error("a run without project material needs no approval")
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
