package launch

import (
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	provshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/record"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

// fakeIsolator records the calls of the executor and answers with its fields.
// It confines the network, is kernel-isolated only under runsc, and is pinned
// only for a pinned provisioner or a digest image.
type fakeIsolator struct {
	unavailable  bool
	noNetwork    bool
	checkErr     error
	command      []string
	commandErr   error
	runCode      int
	runErr       error
	terminalCmd  []string
	terminalEnv  []string
	terminalErr  error
	checked      bool
	commandRun   isoshared.RunConfig
	ran          bool
	terminalRuns int
}

func (f *fakeIsolator) Available() (bool, error) { return !f.unavailable, nil }

func (f *fakeIsolator) Run(cfg isoshared.RunConfig, _ provision.Contribution) (int, error) {
	f.ran = true
	return f.runCode, f.runErr
}

func (f *fakeIsolator) Command(cfg isoshared.RunConfig, _ provision.Contribution) ([]string, error) {
	f.commandRun = cfg
	return f.command, f.commandErr
}

func (f *fakeIsolator) TerminalCommand(isoshared.RunConfig, provision.Contribution) ([]string, []string, error) {
	f.terminalRuns++
	return f.terminalCmd, f.terminalEnv, f.terminalErr
}

func (f *fakeIsolator) PinnedProvision(_ provision.ProvisionMethod, provisionerPinned bool, image string) bool {
	return provisionerPinned || strings.Contains(image, "@sha256:")
}

func (f *fakeIsolator) HidesHost(passthrough bool, _ string) bool { return !passthrough }
func (f *fakeIsolator) KernelIsolated(runtime string) bool        { return runtime == "runsc" }
func (f *fakeIsolator) EgressTransport() network.EgressTransport  { return network.TransportUnixSocket }
func (f *fakeIsolator) ConfinesNetwork() bool                     { return !f.noNetwork }

func (f *fakeIsolator) CheckAgent(isoshared.RunConfig) error {
	f.checked = true
	return f.checkErr
}

type fakeProvisioner struct {
	pinned        bool
	contribution  provision.Contribution
	contributeErr error
	handled       bool
	checkErr      error
	input         provshared.Input
}

func (f *fakeProvisioner) Contribute(in provshared.Input) (provision.Contribution, error) {
	f.input = in
	return f.contribution, f.contributeErr
}

func (f *fakeProvisioner) Pinned() bool { return f.pinned }

func (f *fakeProvisioner) CheckAgent(provision.Contribution, *types.AgentConfig) (bool, error) {
	return f.handled, f.checkErr
}

type fakeVCS struct {
	dir     string
	err     error
	configs []wsshared.Config
	repos   []string
}

func (f *fakeVCS) Setup(config wsshared.Config, repoDir string, _ bool) (string, error) {
	f.configs = append(f.configs, config)
	f.repos = append(f.repos, repoDir)
	return f.dir, f.err
}

func (f *fakeVCS) Available() bool { return true }

type fakeTerminal struct {
	missing  bool
	code     int
	err      error
	config   *termshared.TerminalConfig
	command  []string
	env      []string
	workDir  string
	executed bool
}

func (f *fakeTerminal) IsAvailable() bool { return !f.missing }
func (f *fakeTerminal) IsInside() bool    { return false }

func (f *fakeTerminal) Execute(config *termshared.TerminalConfig, command, env []string, workDir string) (int, error) {
	f.executed = true
	f.config, f.command, f.env, f.workDir = config, command, env, workDir
	return f.code, f.err
}

// testSandbox is a registry whose every method resolves to the same doubles.
type testSandbox struct {
	isolator    *fakeIsolator
	provisioner *fakeProvisioner
	registry    *sandbox.Registry
}

func newTestSandbox() testSandbox {
	iso := &fakeIsolator{command: []string{"agent", "run"}}
	nonConfining := &fakeIsolator{noNetwork: true, command: []string{"agent"}}
	prov := &fakeProvisioner{}
	pinned := &fakeProvisioner{pinned: true, handled: true}
	registry := sandbox.NewRegistry().
		RegisterIsolator(isolation.IsolationNone, func() isoshared.Isolator { return nonConfining }).
		RegisterIsolator(isolation.IsolationBwrap, func() isoshared.Isolator { return iso }).
		RegisterIsolator(isolation.IsolationDocker, func() isoshared.Isolator { return iso }).
		RegisterProvisioner(provision.ProvisionNone, func() provshared.Provisioner { return prov }).
		RegisterProvisioner(provision.ProvisionCommand, func() provshared.Provisioner { return prov }).
		RegisterProvisioner(provision.ProvisionNix, func() provshared.Provisioner { return pinned })
	return testSandbox{isolator: iso, provisioner: prov, registry: registry}
}

// fakeRecorder keeps the records it is given.
type fakeRecorder struct {
	records []record.Record
	err     error
}

func (f *fakeRecorder) Write(r record.Record) (string, error) {
	f.records = append(f.records, r)
	return "20261120T120000Z", f.err
}
