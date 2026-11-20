package cmd

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	provshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

// recordingIsolator records the RunConfig the command line produced; it runs
// with exitCode and is pinned for every request.
type recordingIsolator struct {
	exitCode int
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
	return true
}
func (r *recordingIsolator) HidesHost(bool, string) bool          { return true }
func (r *recordingIsolator) KernelIsolated(string) bool           { return true }
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

type testRoot struct {
	isolator    *recordingIsolator
	provisioner *recordingProvisioner
	out         bytes.Buffer
	deps        launch.Dependencies
}

func newTestRoot() *testRoot {
	r := &testRoot{isolator: &recordingIsolator{}, provisioner: &recordingProvisioner{}}
	registry := sandbox.NewRegistry().
		RegisterProvisioner(provision.ProvisionNone, func() provshared.Provisioner { return r.provisioner }).
		RegisterProvisioner(provision.ProvisionCommand, func() provshared.Provisioner { return r.provisioner })
	for _, method := range []isolation.IsolationMethod{isolation.IsolationNone, isolation.IsolationBwrap, isolation.IsolationDocker} {
		registry.RegisterIsolator(method, func() isoshared.Isolator { return r.isolator })
	}
	r.deps = launch.Dependencies{
		Host:      systemtest.New(),
		Log:       logging.New(io.Discard, io.Discard, false, false),
		Out:       &r.out,
		Sandboxes: registry,
	}
	return r
}

func (r *testRoot) execute(args ...string) error {
	root := NewRootCommand(r.deps)
	root.SetArgs(args)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	return root.Execute()
}

func TestNewRootCommandRegistersSubcommandsAndFlags(t *testing.T) {
	root := NewRootCommand(newTestRoot().deps)

	for _, name := range []string{"run", "completion"} {
		if command, _, err := root.Find([]string{name}); err != nil || command.Name() != name {
			t.Errorf("Find(%q) = (%v, %v), want the %s command", name, command, err, name)
		}
	}
	run, _, _ := root.Find([]string{"run"})
	for _, flag := range []string{"agent", "isolation", "provision", "network", "require-egress-restricted", "dry-run"} {
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
	root := NewRootCommand(r.deps)
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
	root := NewRootCommand(r.deps)
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
