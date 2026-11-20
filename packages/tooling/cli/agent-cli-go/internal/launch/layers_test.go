package launch

import (
	"errors"
	"strings"
	"testing"

	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/record"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
)

const userConfig = "/home/user/.config/sandboxed-claude/config.yaml"

func TestFlagInputHoldsOnlyWhatTheUserChose(t *testing.T) {
	options := baseOptions()
	options.Explicit = config.KeysOf(config.KeyIsolation)
	options.Isolation = "bwrap"
	options.Provider = "beta"
	options.BindPaths = []string{"/data"}
	options.Policy.RequireEgressRestricted = true

	input := flagInput(options)

	if input.Layer != config.LayerFlags {
		t.Errorf("Layer = %q, want flags", input.Layer)
	}
	if _, ok := input.Entries[config.KeyNetwork]; ok {
		t.Error("the default network is in the command line layer, so it would shadow a file")
	}
	if input.Entries[config.KeyIsolation].Text != "bwrap" || input.Entries[config.KeyProvider].Text != "beta" {
		t.Errorf("Entries = %+v, want the chosen isolation and the provider", input.Entries)
	}
	if got := input.Entries[config.KeyBindPaths].Items; len(got) != 1 || got[0] != "/data" {
		t.Errorf("bindPaths = %v", got)
	}
	if !input.Entries[config.KeyRequireEgressRestricted].On {
		t.Error("requireEgressRestricted is missing")
	}
}

func TestExecuteReadsTheUserLayerAndTheCommandLineWins(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile(userConfig, "isolation: bwrap\nnetwork: none\nprovider: beta\n", 0o600)
	options := baseOptions()
	options.Config = userConfig
	options.Network = "host"
	options.Explicit = config.KeysOf(config.KeyNetwork)
	options.DryRun = true

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	run := f.sandbox.isolator.commandRun
	if string(run.Network) != "host" {
		t.Errorf("network = %q, want the command line to win", run.Network)
	}
	log := f.log.String()
	for _, want := range []string{"Configuration:", "isolation: bwrap (user " + userConfig + ")", "network: host (flags"} {
		if !strings.Contains(log, want) {
			t.Errorf("dry run lacks %q:\n%s", want, log)
		}
	}
}

func TestExecuteAppliesTheManagedLayerAndRefusesAContradiction(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile("/etc/agent-cli/managed.yaml", "requireEgressRestricted: true\nnetwork: none\n", 0o644)
	options := bwrapOptions()
	options.Network = "host"
	options.Explicit = config.KeysOf(config.KeyNetwork)

	err := Execute(f.deps(), options)

	if err == nil || !strings.Contains(err.Error(), "the managed value cannot be changed") {
		t.Fatalf("Execute() error = %v, want the managed contradiction", err)
	}
	if f.sandbox.isolator.ran {
		t.Error("the run started under a contradicted managed value")
	}

	options.Network, options.Explicit = "none", config.KeysOf(config.KeyNetwork)
	if err := Execute(f.deps(), options); err != nil {
		t.Errorf("Execute() with the managed value error = %v", err)
	}
}

func TestExecuteReadsTheProjectLayerFromTheWorkDirectory(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile("/work/repo/.agent-cli.yaml", "provider: alpha\n", 0o644)
	options := bwrapOptions()
	options.DryRun = true

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(f.log.String(), "provider: alpha (project /work/repo/.agent-cli.yaml)") {
		t.Errorf("dry run lacks the project provenance:\n%s", f.log.String())
	}

	f = newExecuteFixture()
	f.host.AddFile("/work/repo/.agent-cli.yaml", "bindPaths: [/etc]\n", 0o644)
	if err := Execute(f.deps(), bwrapOptions()); err == nil || !strings.Contains(err.Error(), "project") {
		t.Errorf("Execute() error = %v, want the project layer refused for bindPaths", err)
	}
}

func TestExecuteReportsAConfigurationFailure(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.Config = "/missing.yaml"

	if err := Execute(f.deps(), options); err == nil || !strings.Contains(err.Error(), "failed to load config") {
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
