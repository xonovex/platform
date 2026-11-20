package cmd

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	credshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
	pluginorigins "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/plugins"
	provshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/trust"
)

// recordingIsolator records the RunConfig the command line produced; it runs
// with exitCode and is pinned for every request.
type recordingIsolator struct {
	exitCode int
	unpinned bool
	run      isoshared.RunConfig
}

func (r *recordingIsolator) Available() (bool, error) { return true, nil }
func (r *recordingIsolator) Run(isoshared.RunConfig, provision.Contribution) (int, error) {
	return r.exitCode, nil
}

func (r *recordingIsolator) Command(cfg isoshared.RunConfig, _ provision.Contribution) ([]string, error) {
	r.run = cfg
	return []string{"sandbox", cfg.Agent.Binary}, nil
}

func (r *recordingIsolator) TerminalCommand(isoshared.RunConfig, provision.Contribution) ([]string, []string, error) {
	return nil, nil, nil
}
func (r *recordingIsolator) PinnedProvision(provision.ProvisionMethod, bool, string) bool {
	return !r.unpinned
}
func (r *recordingIsolator) HidesHost(bool, string) bool { return true }
func (r *recordingIsolator) KernelIsolated(string) bool  { return true }
func (r *recordingIsolator) EgressTransport() netshared.Transport {
	return netshared.TransportUnixSocket
}
func (r *recordingIsolator) ConfinesNetwork() bool                { return true }
func (r *recordingIsolator) CheckAgent(isoshared.RunConfig) error { return nil }

type recordingProvisioner struct{ input provshared.Input }

func (p *recordingProvisioner) Contribute(in provshared.Input) (provision.Contribution, error) {
	p.input = in
	return provision.Contribution{}, nil
}
func (p *recordingProvisioner) Pinned() bool { return false }
func (p *recordingProvisioner) CheckAgent(provision.Contribution, *types.AgentConfig) (bool, error) {
	return true, nil
}

// recordingBridge records the arguments of the egress-bridge command.
type recordingBridge struct {
	socket, listen string
	command        []string
	code           int
	err            error
}

func (b *recordingBridge) Run(socket, listen string, command []string) (int, error) {
	b.socket, b.listen, b.command = socket, listen, command
	return b.code, b.err
}

type testRoot struct {
	bridge      *recordingBridge
	isolator    *recordingIsolator
	provisioner *recordingProvisioner
	out         bytes.Buffer
	deps        launch.Dependencies
	ports       PluginPorts
	host        *systemtest.Fake
}

func newTestRoot() *testRoot {
	r := &testRoot{bridge: &recordingBridge{}, isolator: &recordingIsolator{}, provisioner: &recordingProvisioner{}}
	registry := sandbox.NewRegistry().
		RegisterProvisioner(provision.ProvisionNone, func() provshared.Provisioner { return r.provisioner }).
		RegisterProvisioner(provision.ProvisionCommand, func() provshared.Provisioner { return r.provisioner })
	for _, method := range []isolation.IsolationMethod{isolation.IsolationNone, isolation.IsolationBwrap, isolation.IsolationDocker} {
		registry.RegisterIsolator(method, func() isoshared.Isolator { return r.isolator })
	}
	r.host = systemtest.New()
	approvals, _ := trust.NewStore(r.host, trust.DefaultFile(r.host.Getenv, r.host.Home))
	r.ports = PluginPorts{Host: r.host, Origins: pluginorigins.DefaultRegistry(r.host), Out: &r.out}
	r.deps = launch.Dependencies{
		Host:      r.host,
		Log:       logging.New(io.Discard, io.Discard, false, false),
		Out:       &r.out,
		Sandboxes: registry,
		Approvals: approvals,
	}
	return r
}

func (r *testRoot) execute(args ...string) error {
	root := NewRootCommand(r.deps, r.ports, RecordPorts{}, r.bridge)
	root.SetArgs(args)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	return root.Execute()
}

func TestNewRootCommandRegistersSubcommandsAndFlags(t *testing.T) {
	root := NewRootCommand(newTestRoot().deps, PluginPorts{}, RecordPorts{}, &recordingBridge{})

	for _, name := range []string{"run", "completion"} {
		if command, _, err := root.Find([]string{name}); err != nil || command.Name() != name {
			t.Errorf("Find(%q) = (%v, %v), want the %s command", name, command, err, name)
		}
	}
	run, _, _ := root.Find([]string{"run"})
	for _, flag := range []string{"agent", "isolation", "provision", "network", "allow-host", "credential", "require-egress-restricted", "dry-run"} {
		if run.Flags().Lookup(flag) == nil {
			t.Errorf("run command has no --%s flag", flag)
		}
	}
	if root.PersistentFlags().Lookup("verbose") == nil {
		t.Error("root command has no --verbose flag")
	}
	if root.Version != version {
		t.Errorf("Version = %q, want %q", root.Version, version)
	}
}

func TestRunCommandBindsTheFlagsIntoTheRun(t *testing.T) {
	r := newTestRoot()

	err := r.execute("run", "--dry-run", "--agent", "opencode", "--isolation", "docker",
		"--provision", "command", "--init-command", "npm ci", "--network", "none",
		"--image", "agent@sha256:abc", "--isolation-docker-runtime", "runsc",
		"--work-dir", "/src", "--ro-bind", "/data", "--env", "A=1",
		"--require-kernel-isolation", "--", "review", "--fast")

	if err != nil {
		t.Fatalf("run error = %v", err)
	}
	run := r.isolator.run
	if run.Agent.Binary != "opencode" || run.WorkDir != "/src" || run.Image != "agent@sha256:abc" || run.Runtime != "runsc" {
		t.Errorf("RunConfig = %+v", run)
	}
	if run.Network != "none" || !slices.Equal(run.RoBindPaths, []string{"/data"}) || !slices.Equal(run.CustomEnv, []string{"A=1"}) {
		t.Errorf("network, binds, env = %q %v %v", run.Network, run.RoBindPaths, run.CustomEnv)
	}
	if !slices.Equal(run.AgentArgs, []string{"review", "--fast"}) {
		t.Errorf("AgentArgs = %v", run.AgentArgs)
	}
	if !slices.Equal(r.provisioner.input.InitCommands, []string{"npm ci"}) {
		t.Errorf("InitCommands = %v", r.provisioner.input.InitCommands)
	}
	if r.out.String() != "sandbox opencode\n" {
		t.Errorf("Out = %q, want the dry-run command", r.out.String())
	}
}

func TestRunCommandYieldsToThePinnedPairOnlyWithoutAnExplicitAxis(t *testing.T) {
	// The test registry has no nix provisioner, so the pinned default fails on
	// the provisioner lookup; an explicit axis keeps the user's choice.
	r := newTestRoot()
	if err := r.execute("run", "--dry-run", "--require-pinned-provision"); err == nil || !strings.Contains(err.Error(), `provisioning "nix"`) {
		t.Fatalf("run without an axis error = %v, want the pinned nix default", err)
	}
	if err := r.execute("run", "--dry-run", "--require-pinned-provision", "--isolation", "bwrap"); err != nil {
		t.Fatalf("run with an explicit isolation error = %v", err)
	}
	if err := r.execute("run", "--dry-run", "--require-pinned-provision", "--provision", "none"); err != nil {
		t.Fatalf("run with an explicit provision error = %v", err)
	}
}

func TestRunCommandSilencesTheAgentExitCode(t *testing.T) {
	r := newTestRoot()
	r.isolator.exitCode = 4
	root := NewRootCommand(r.deps, r.ports, RecordPorts{}, r.bridge)
	root.SetArgs([]string{"run"})
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)

	err := root.Execute()

	var exit launch.ExitError
	if !errors.As(err, &exit) || exit.Code != 4 {
		t.Fatalf("run error = %v, want ExitError{4}", err)
	}
	run, _, _ := root.Find([]string{"run"})
	if !run.SilenceErrors || !run.SilenceUsage || output.Len() != 0 {
		t.Errorf("SilenceErrors, SilenceUsage = %v %v, output = %q", run.SilenceErrors, run.SilenceUsage, output.String())
	}
}

func TestRunCommandReportsOtherErrorsWithUsage(t *testing.T) {
	r := newTestRoot()
	root := NewRootCommand(r.deps, r.ports, RecordPorts{}, r.bridge)
	root.SetArgs([]string{"run", "--agent", "unknown"})
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)

	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "available agents") {
		t.Fatalf("run error = %v", err)
	}
	if !strings.Contains(output.String(), "Error:") || !strings.Contains(output.String(), "Usage:") {
		t.Errorf("output = %q, want the error and the usage", output.String())
	}
}

func TestRunCommandRunsEveryHarness(t *testing.T) {
	for _, agent := range []string{"claude", "codex", "copilot", "kiro", "opencode", "pi"} {
		t.Run(agent, func(t *testing.T) {
			r := newTestRoot()

			if err := r.execute("run", "--dry-run", "--agent", agent, "--", "review"); err != nil {
				t.Fatalf("run --agent %s error = %v", agent, err)
			}
			if got := r.isolator.run.Agent.Type; string(got) != agent {
				t.Errorf("agent = %q, want %q", got, agent)
			}
		})
	}
}

func TestRunCommandRejectsAnUnknownAgentAndAProviderForALoginAgent(t *testing.T) {
	r := newTestRoot()
	if err := r.execute("run", "--dry-run", "--agent", "nosuch"); err == nil || !strings.Contains(err.Error(), "kiro") {
		t.Fatalf("run --agent nosuch error = %v, want the available agents", err)
	}
	if err := r.execute("run", "--dry-run", "--agent", "kiro", "--provider", "alpha"); err == nil || !strings.Contains(err.Error(), "own login") {
		t.Fatalf("run --agent kiro --provider error = %v, want a native login rejection", err)
	}
}

func TestRunCommandHelpListsEveryHarness(t *testing.T) {
	root := NewRootCommand(newTestRoot().deps, PluginPorts{}, RecordPorts{}, &recordingBridge{})
	run, _, _ := root.Find([]string{"run"})

	usage := run.Flags().Lookup("agent").Usage

	for _, agent := range []string{"claude", "codex", "copilot", "kiro", "opencode", "pi"} {
		if !strings.Contains(usage, agent) {
			t.Errorf("--agent usage = %q, want %q", usage, agent)
		}
	}
}

// complete returns the candidates the shell completion offers for args.
func complete(t *testing.T, args ...string) []string {
	t.Helper()
	return completeWith(t, newTestRoot(), args...)
}

// completeWith returns the candidates the shell completion offers for args on the
// root r.
func completeWith(t *testing.T, r *testRoot, args ...string) []string {
	t.Helper()
	root := NewRootCommand(r.deps, PluginPorts{}, RecordPorts{}, &recordingBridge{})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs(append([]string{"__complete"}, args...))
	if err := root.Execute(); err != nil {
		t.Fatalf("__complete error = %v", err)
	}
	var candidates []string
	for _, line := range strings.Split(out.String(), "\n") {
		if line != "" && !strings.HasPrefix(line, ":") && !strings.HasPrefix(line, "Completion ended") {
			candidates = append(candidates, line)
		}
	}
	return candidates
}

func TestCompletionOffersTheRegisteredAgents(t *testing.T) {
	got := complete(t, "run", "--agent", "")

	want := []string{"claude", "codex", "copilot", "kiro", "opencode", "pi"}
	if !slices.Equal(got, want) {
		t.Fatalf("agent completion = %v, want %v", got, want)
	}
}

func TestCompletionOffersTheProvidersOfTheSelectedAgent(t *testing.T) {
	r := newTestRoot()
	r.host.AddFile(userProviders+"/gamma.json", `{"name":"gamma","agent":"codex"}`, 0o644)
	r.host.AddFile(userProviders+"/alpha.json", `{"name":"alpha","agent":"claude"}`, 0o644)
	if got := completeWith(t, r, "run", "--agent", "codex", "--provider", ""); !slices.Equal(got, []string{"gamma"}) {
		t.Fatalf("codex provider completion = %v, want [gamma]", got)
	}
	if got := complete(t, "run", "--agent", "codex", "--provider", ""); len(got) != 0 {
		t.Fatalf("provider completion without a provider file = %v, want none", got)
	}
	if got := complete(t, "run", "--agent", "kiro", "--provider", ""); len(got) != 0 {
		t.Fatalf("kiro provider completion = %v, want none", got)
	}
}

func TestRunCommandBindsTheNetworkAndCredentialFlags(t *testing.T) {
	r := newTestRoot()
	r.deps.Egress = &recordingOpener{}
	err := r.execute("run", "--dry-run", "--isolation", "bwrap", "--network", "managed",
		"--allow-host", "api.example.com", "--allow-host", "*.cdn.example.com:8443",
		"--credential", "API_KEY=env:SOURCE_KEY")
	if err != nil {
		t.Fatalf("run error = %v", err)
	}
	run := r.isolator.run
	if run.Network != "managed" || !run.Egress.Allow.Allows("x.cdn.example.com", 8443) || !run.Egress.Allow.Allows("api.example.com", 443) {
		t.Errorf("network = %q, allowlist = %v", run.Network, run.Egress.Allow.Rules())
	}
	if run.Secrets["API_KEY"] != "[redacted]" {
		t.Errorf("a dry run must carry only the redacted placeholder, got %q", run.Secrets["API_KEY"])
	}
}

type recordingOpener struct{}

func (recordingOpener) Open(transport netshared.Transport, allow egress.Allowlist) (isoshared.EgressRoute, func(), error) {
	return isoshared.EgressRoute{Transport: transport, Allow: allow}, func() {}, nil
}

func TestCompletionOffersTheRegisteredIsolationsAndNetworkModes(t *testing.T) {
	if got := complete(t, "run", "--isolation", ""); !slices.Equal(got, []string{"bwrap", "docker", "none"}) {
		t.Errorf("isolation completion = %v", got)
	}
	if got := complete(t, "run", "--network", ""); !slices.Equal(got, []string{"host", "none", "managed"}) {
		t.Errorf("network completion = %v", got)
	}
}

func TestCompletionOffersCredentialSchemesAfterTheVariableName(t *testing.T) {
	r := newTestRoot()
	r.deps.Credentials = credshared.NewRegistry().
		Register("env", func() credshared.Store { return nil }).
		Register("pass", func() credshared.Store { return nil })
	root := NewRootCommand(r.deps, r.ports, RecordPorts{}, r.bridge)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"__complete", "run", "--credential", "API_KEY="})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"API_KEY=env:", "API_KEY=pass:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("completion lacks %q:\n%s", want, out.String())
		}
	}
	if got := credentialCandidates(r.deps, "API"); got != nil {
		t.Errorf("before the equals sign the user names the variable: %v", got)
	}
	if got := credentialCandidates(launch.Dependencies{}, "API_KEY="); got != nil {
		t.Errorf("a build without stores offers nothing: %v", got)
	}
	if got := isolationNames(launch.Dependencies{}); got != nil {
		t.Errorf("a build without isolations offers nothing: %v", got)
	}
}

func TestCompletionOffersTheRegisteredTerminals(t *testing.T) {
	r := newTestRoot()
	r.deps.Terminals = map[termshared.TerminalType]termshared.TerminalExecutor{termshared.TerminalTmux: nil, termshared.TerminalHerdr: nil}
	root := NewRootCommand(r.deps, r.ports, RecordPorts{}, r.bridge)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"__complete", "run", "--terminal", ""})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "herdr\ntmux\n") {
		t.Errorf("terminal completion = %q, want herdr and tmux", out.String())
	}
}
