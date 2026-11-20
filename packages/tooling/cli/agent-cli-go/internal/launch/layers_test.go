package launch

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/record"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

func TestFlagInputHoldsOnlyWhatTheUserChose(t *testing.T) {
	options := baseOptions()
	options.Explicit = runconfig.KeysOf(runconfig.KeyIsolation)
	options.Isolation = "bwrap"
	options.Provider = "beta"
	options.BindPaths = []string{"/data"}
	options.Policy.RequireEgressRestricted = true

	got := map[string]string{}
	for _, flag := range flagInput(options) {
		if flag.Section != "run" {
			t.Errorf("flag %+v is outside the run section", flag)
		}
		got[flag.Pointer] = string(flag.Value)
	}

	want := map[string]string{"/isolation": `"bwrap"`, "/provider": `"beta"`, "/bindPaths": `["/data"]`, "/requireEgressRestricted": "true"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("flags = %v, want %v; the default network would shadow a file", got, want)
	}
}

func TestExecuteReadsTheUserLayerAndTheCommandLineWins(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(userFile, runFile(`{"isolation":"bwrap","network":"none","provider":"beta"}`), 0o600)
	options := baseOptions()
	options.Network = "host"
	options.Explicit = runconfig.KeysOf(runconfig.KeyNetwork)
	options.DryRun = true

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	run := f.sandbox.isolator.commandRun
	if string(run.Network) != "host" {
		t.Errorf("network = %q, want the command line to win", run.Network)
	}
	log := f.log.String()
	for _, want := range []string{"Configuration:", "isolation: bwrap (user " + userFile + ")", "network: host (flags command line)"} {
		if !strings.Contains(log, want) {
			t.Errorf("dry run lacks %q:\n%s", want, log)
		}
	}
}

func TestExecuteAppliesTheManagedLayerAndRefusesAContradiction(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(managedFile, runFile(`{"requireEgressRestricted":true,"network":"none"}`), 0o644)
	options := bwrapOptions()
	options.Network = "host"
	options.Explicit = runconfig.KeysOf(runconfig.KeyNetwork)

	err := Execute(f.deps(), options)

	if err == nil || !strings.Contains(err.Error(), "but the managed file "+managedFile+" sets it to") {
		t.Fatalf("Execute() error = %v, want the managed contradiction", err)
	}
	if f.sandbox.isolator.ran {
		t.Error("the run started under a contradicted managed value")
	}

	options.Network, options.Explicit = "none", runconfig.KeysOf(runconfig.KeyNetwork)
	if err := Execute(f.deps(), options); err != nil {
		t.Errorf("Execute() with the managed value error = %v", err)
	}
}

func TestExecuteReadsTheProjectLayerFromTheWorkDirectoryAfterTheApproval(t *testing.T) {
	f := newExecuteFixture()
	data := runFile(`{"provider":"alpha"}`)
	f.host.AddFile(projectFile, data, 0o644)
	options := bwrapOptions()
	options.DryRun = true

	if err := Execute(f.deps(), options); err == nil || !strings.Contains(err.Error(), "--approve-config-digest "+digestOf(data)) {
		t.Fatalf("Execute() error = %v, want the project file to wait for its approval", err)
	}
	options.ApproveDigest = digestOf(data)
	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(f.log.String(), "provider: alpha (project "+projectFile+")") {
		t.Errorf("dry run lacks the project provenance:\n%s", f.log.String())
	}

	f = newExecuteFixture()
	f.host.AddFile(projectFile, runFile(`{"bindPaths":["/etc"]}`), 0o644)
	if err := Execute(f.deps(), bwrapOptions()); err == nil || !strings.Contains(err.Error(), "the project file sets run.bindPaths") {
		t.Errorf("Execute() error = %v, want the project layer refused for bindPaths", err)
	}
}

func TestExecuteReportsAConfigurationFailure(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.Project = "/missing.json"

	if err := Execute(f.deps(), options); err == nil || !strings.Contains(err.Error(), "resolve configuration: project configuration /missing.json") {
		t.Errorf("Execute() error = %v, want a load failure", err)
	}
}

func TestExecuteWritesARunRecord(t *testing.T) {
	f := newExecuteFixture()
	f.sandbox.isolator.runCode = 3
	options := bwrapOptions()
	options.Credentials = []string{"TOKEN=test:vault/token"}
	deps := f.deps()
	deps.Credentials = credshared.NewRegistry().Register("test", func() credshared.Store { return fakeStore{value: "s3cret"} })

	err := Execute(deps, options)

	var exit ExitError
	if !errors.As(err, &exit) || exit.Code != 3 {
		t.Fatalf("Execute() error = %v, want exit code 3", err)
	}
	if len(f.records.records) != 1 {
		t.Fatalf("records = %d, want one", len(f.records.records))
	}
	r := f.records.records[0]
	if r.Status != record.StatusExited || r.ExitCode == nil || *r.ExitCode != 3 {
		t.Errorf("status, code = %v, %v; want exited with 3", r.Status, r.ExitCode)
	}
	if r.Axes.Isolation != "bwrap" || r.Axes.Workspace.Mode != record.ModeInPlace || r.Agent.Name != "claude" {
		t.Errorf("axes, agent = %+v, %+v", r.Axes, r.Agent)
	}
	if len(r.Credentials) != 1 || r.Credentials[0] != "TOKEN" {
		t.Errorf("credentials = %v, want the name only", r.Credentials)
	}
	if len(r.Guarantees) == 0 {
		t.Error("the record holds no guarantee")
	}
	if data, err := record.Marshal(r); err != nil || strings.Contains(string(data), "s3cret") {
		t.Errorf("Marshal() = %v, %v; the record must not hold the secret", err, string(data))
	}
}

func TestExecuteRecordsARunThatFailedToStart(t *testing.T) {
	f := newExecuteFixture()
	f.sandbox.isolator.runErr = errors.New("sandbox crashed")

	if err := Execute(f.deps(), bwrapOptions()); err == nil {
		t.Fatal("Execute() error = nil, want the failure")
	}

	r := f.records.records[0]
	if r.Status != record.StatusFailed || r.ExitCode != nil || !strings.Contains(r.Error, "sandbox crashed") {
		t.Errorf("record = %+v, want failed-to-start with the error and no code", r)
	}
}

func TestExecuteRecordsACheckoutAndATerminalRun(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.WorktreeBranch = "feature"
	options.Terminal = "tmux"

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	r := f.records.records[0]
	if r.Axes.Workspace.Mode != record.ModeCheckout || r.Axes.Workspace.Branch != "feature" || r.Axes.Terminal != "tmux" {
		t.Errorf("axes = %+v, want a checkout run in tmux", r.Axes)
	}
}

func TestExecuteLeavesNoRecordForADryRunOrAPlanFailure(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.DryRun = true
	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	options = bwrapOptions()
	options.Agent = "no-such-agent"
	if err := Execute(f.deps(), options); err == nil {
		t.Fatal("Execute() error = nil, want an unknown agent")
	}
	if len(f.records.records) != 0 {
		t.Errorf("records = %d, want none", len(f.records.records))
	}
}

func TestExecuteWarnsWhenTheRecordCannotBeWritten(t *testing.T) {
	f := newExecuteFixture()
	f.records.err = errors.New("disk full")

	if err := Execute(f.deps(), bwrapOptions()); err != nil {
		t.Fatalf("Execute() error = %v, want the run to succeed", err)
	}
	if !strings.Contains(f.log.String(), "Could not write the run record: disk full") {
		t.Errorf("log = %q, want the warning", f.log.String())
	}

	deps := f.deps()
	deps.Records = nil
	if err := Execute(deps, bwrapOptions()); err != nil {
		t.Errorf("Execute() without a recorder error = %v", err)
	}
}

func TestExecuteRecordsTheOverlaySourcesWithTheirDigests(t *testing.T) {
	f := newExecuteFixture()
	digest := "sha256:" + strings.Repeat("b", 64)
	deps := f.deps()
	deps.Overlays = &fakeOverlays{result: OverlayResult{Sources: []OverlaySource{
		{Kind: SourcePlugin, Name: "hello", Digest: digest},
		{Kind: SourceDirectory, Name: "/work/repo/extras", Digest: digest},
	}}}

	if err := Execute(deps, overlayOptions()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	overlay := f.records.records[0].Overlay
	if len(overlay.Plugins) != 1 || overlay.Plugins[0].Name != "hello" || len(overlay.Directories) != 1 || overlay.Directories[0].Digest != digest {
		t.Errorf("overlay = %+v, want one plugin and one directory with digests", overlay)
	}
}

func TestExecuteStartsTheBinaryOfAHarnessOverride(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(userDir+"/harnesses.d/claude.json", `{"binary":"claude-next"}`, 0o644)

	if err := Execute(f.deps(), dryRunOptions()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := f.sandbox.isolator.commandRun.Process.Argv; len(got) == 0 || got[0] != "claude-next" {
		t.Errorf("agent argv = %q, want the binary of the override", got)
	}

	f.host.AddFile(userDir+"/harnesses.d/claude.json", `{"binary":7}`, 0o644)
	if err := Execute(f.deps(), dryRunOptions()); err == nil || !strings.Contains(err.Error(), "harness override") {
		t.Errorf("Execute() error = %v, want the broken override named", err)
	}
}
