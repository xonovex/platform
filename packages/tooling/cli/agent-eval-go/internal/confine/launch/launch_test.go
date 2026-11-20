package launch_test

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation/docker"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
	provnone "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision/none"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine/launch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

const (
	home      = "/home/u"
	workspace = "/tmp/agent-eval-1"
	nixRev    = "0123456789abcdef0123456789abcdef01234567"
	image     = "registry.example/eval@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

// harnessRun is the process the claude leaf plans for one prompt.
func harnessRun() confine.Request {
	return confine.Request{Harness: "claude", Process: process.Request{
		Dir: workspace, Name: "claude", Args: []string{"-p", "--", "hi"},
		Env: map[string]string{"HOME": workspace + "/home", "MODE": "hermetic"}, Timeout: time.Minute,
	}}
}

func host() *oshosttest.Fake {
	h := oshosttest.New()
	h.Home = home
	h.Env = map[string]string{"PATH": "/usr/bin", "USER": "u"}
	h.AddDir(home).AddDir(workspace)
	h.Outcomes[oshosttest.CommandLine("docker", "info")] = oshosttest.Outcome{}
	return h
}

func dockerPorts(h *oshosttest.Fake) launch.Ports {
	log := logging.New(&strings.Builder{}, &strings.Builder{}, false, false)
	return launch.Ports{
		Sandboxes: sandbox.NewRegistry[isoshared.Isolator]().
			RegisterIsolator(isolation.IsolationDocker, func() isoshared.Isolator { return docker.NewIsolator(h, log) }).
			RegisterProvisioner(provision.ProvisionNone, func() provshared.Provisioner { return provnone.New() }),
		Harnesses: agents.Embedded(),
		Home:      home,
	}
}

func TestTheHarnessRunsInANamedContainerThatBindsOnlyItsWorkspace(t *testing.T) {
	h := host()
	h.AddDir(home + "/.claude")
	confiner := launch.New(dockerPorts(h), model.Sandbox{Isolation: "docker", Image: image, CustomEnv: []string{"MODE=ci", "EXTRA=1"}})
	confined, err := confiner.Confine(harnessRun())
	if err != nil {
		t.Fatal(err)
	}
	run := confined.Process
	args := strings.Join(run.Args, " ")
	name := launch.InstanceName(workspace)
	if run.Name != "docker" || !strings.HasPrefix(args, "run --rm --name "+name+" ") || strings.Contains(args, "-it") {
		t.Errorf("args = %s, want a removed, named container without a terminal", args)
	}
	var mounts []string
	for i, arg := range run.Args {
		if arg == "-v" {
			mounts = append(mounts, run.Args[i+1])
		}
	}
	if !slices.Equal(mounts, []string{workspace + ":" + workspace}) {
		t.Errorf("mounts = %v, want only the workspace, read-write", mounts)
	}
	if !strings.HasSuffix(args, " "+image+" claude -p -- hi") {
		t.Errorf("args = %s, want the image and then the program of the leaf", args)
	}
	if run.Env["HOME"] != workspace+"/home" || run.Env["MODE"] != "hermetic" || run.Env["EXTRA"] != "1" {
		t.Errorf("env = %v, want the environment of the leaf over the environment of the sandbox", run.Env)
	}
	if run.Dir != workspace || run.Timeout != time.Minute {
		t.Errorf("dir %s, timeout %s, want those of the request", run.Dir, run.Timeout)
	}
	if confined.Stop == nil || confined.Stop.Name != "docker" || strings.Join(confined.Stop.Args, " ") != "rm --force "+name || confined.Stop.Timeout <= 0 {
		t.Errorf("stop = %+v, want the container removed within a bound", confined.Stop)
	}
	confined.Release()
}

func TestEveryWorkspaceGetsItsOwnName(t *testing.T) {
	first, second := launch.InstanceName("/tmp/agent-eval-1"), launch.InstanceName("/tmp/agent-eval-2")
	if first == second || !strings.HasPrefix(first, "agent-eval-") || first != launch.InstanceName("/tmp/agent-eval-1") {
		t.Errorf("%s %s: a name is stable for one workspace and differs between two", first, second)
	}
}

func TestADockerSandboxWithoutAnImageNamesTheImageSetting(t *testing.T) {
	const hint = "set eval.sandbox.image in the user or managed agentic.json, or pass --sandbox-image"
	confiner := launch.New(dockerPorts(host()), model.Sandbox{Isolation: "docker"})
	if err := confiner.Check(); !errors.Is(err, isoshared.ErrNoImage) || !strings.Contains(err.Error(), hint) {
		t.Errorf("Check() = %v, want ErrNoImage naming eval.sandbox.image and --sandbox-image", err)
	}
	if _, err := confiner.Confine(harnessRun()); !errors.Is(err, isoshared.ErrNoImage) || !strings.Contains(err.Error(), hint) {
		t.Errorf("Confine() = %v, want ErrNoImage naming eval.sandbox.image and --sandbox-image", err)
	}
}

func TestASandboxThatCanStartPassesTheCheck(t *testing.T) {
	for _, box := range []model.Sandbox{
		{Isolation: "docker", Image: image},
		{Isolation: "bwrap"},
		{RequirePinnedProvision: true},
		{Isolation: "chroot"},
	} {
		if err := launch.New(dockerPorts(host()), box).Check(); err != nil {
			t.Errorf("Check(%+v) = %v, want nil: Confine reports what remains", box, err)
		}
	}
}

func TestAGuaranteeTheSandboxCannotGiveStopsTheRun(t *testing.T) {
	_, err := launch.New(dockerPorts(host()), model.Sandbox{Isolation: "docker", Image: image, RequireKernelIsolation: true}).Confine(harnessRun())
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "kernel") {
		t.Fatalf("err = %v", err)
	}
}

func TestAnUnregisteredIsolationIsAnError(t *testing.T) {
	_, err := launch.New(dockerPorts(host()), model.Sandbox{Isolation: "bwrap"}).Confine(harnessRun())
	if !errors.Is(err, sandbox.ErrNoIsolator) {
		t.Fatalf("err = %v", err)
	}
}

func TestAnUnknownHarnessIsAnError(t *testing.T) {
	request := harnessRun()
	request.Harness = "vim"
	if _, err := launch.New(dockerPorts(host()), model.Sandbox{Isolation: "docker", Image: image}).Confine(request); err == nil {
		t.Fatal("a harness without facts cannot be confined")
	}
}

// fakeIsolator records what it confines and declares what a test asks of it.
type fakeIsolator struct {
	available   bool
	asked       int
	checkErr    error
	terminalErr error
	transport   network.EgressTransport
	runs        []isoshared.RunConfig
	// needsImage makes the double start the harness from a container image.
	needsImage bool
}

func (f *fakeIsolator) PinnedProvision(provision.ProvisionMethod, bool, string) bool { return true }
func (f *fakeIsolator) HidesHost(bool, string) bool                                  { return true }
func (f *fakeIsolator) KernelIsolated(string) bool                                   { return false }
func (f *fakeIsolator) ConfinesNetwork() bool                                        { return true }
func (f *fakeIsolator) EgressTransport() network.EgressTransport                     { return f.transport }
func (f *fakeIsolator) Bounded() bool                                                { return true }
func (f *fakeIsolator) AcceptsHostBinds() bool                                       { return true }
func (f *fakeIsolator) RequiresImage() bool                                          { return f.needsImage }
func (f *fakeIsolator) Available() (bool, error)                                     { f.asked++; return f.available, nil }
func (f *fakeIsolator) CheckAgent(isoshared.RunConfig) error                         { return f.checkErr }
func (f *fakeIsolator) Run(isoshared.RunConfig, provision.Contribution) (int, error) {
	return 0, nil
}
func (f *fakeIsolator) Command(cfg isoshared.RunConfig, _ provision.Contribution) ([]string, error) {
	return append([]string{"box"}, cfg.Process.Argv...), nil
}
func (f *fakeIsolator) StopCommand(isoshared.RunConfig) []string { return nil }
func (f *fakeIsolator) TerminalCommand(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string, error) {
	f.runs = append(f.runs, cfg)
	if f.terminalErr != nil {
		return nil, nil, f.terminalErr
	}
	command, _ := f.Command(cfg, c)
	return command, append(append([]string{}, cfg.Process.Env...), "PROXY="+cfg.Egress.Proxy), nil
}

// countingProvisioner counts the provisions it makes.
type countingProvisioner struct {
	made    int
	handles bool
	err     error
}

func (p *countingProvisioner) Contribute(provshared.Input) (provision.Contribution, error) {
	p.made++
	return provision.Contribution{PathEntries: []string{"/nix/store/x/bin"}}, p.err
}
func (p *countingProvisioner) Pinned() bool { return true }
func (p *countingProvisioner) CheckAgent(provision.Contribution, *types.AgentConfig) (bool, error) {
	return p.handles, nil
}

func fakePorts(iso *fakeIsolator, prov *countingProvisioner) launch.Ports {
	return launch.Ports{
		Sandboxes: sandbox.NewRegistry[isoshared.Isolator]().
			RegisterIsolator(isolation.IsolationBwrap, func() isoshared.Isolator { return iso }).
			RegisterProvisioner(provision.ProvisionNix, func() provshared.Provisioner { return prov }),
		Harnesses: agents.Embedded(),
		Home:      home,
	}
}

func TestEachHarnessIsProvisionedOnceAndTheToolIsAskedForOnce(t *testing.T) {
	iso, prov := &fakeIsolator{available: true, checkErr: errors.New("not reached")}, &countingProvisioner{handles: true}
	confiner := launch.New(fakePorts(iso, prov), model.Sandbox{Isolation: "bwrap", Provision: "nix", NixRev: nixRev})
	for range 3 {
		confined, err := confiner.Confine(harnessRun())
		if err != nil {
			t.Fatal(err)
		}
		if confined.Stop != nil {
			t.Errorf("stop = %+v, want none from an isolator that leaves nothing running", confined.Stop)
		}
	}
	if prov.made != 1 || iso.asked != 1 {
		t.Errorf("provisioned %d times and asked %d times, want once each", prov.made, iso.asked)
	}
	run := iso.runs[0]
	if run.RepoDir != "" || len(run.BindPaths) != 0 || run.Interactive || run.Agent.UserConfigPaths != nil {
		t.Errorf("run = %+v, want no repository, no extra bind, no terminal and no user configuration", run)
	}
}

func TestTheSandboxFailsClosed(t *testing.T) {
	for name, tc := range map[string]struct {
		iso  *fakeIsolator
		prov *countingProvisioner
		want string
	}{
		"unavailable":     {&fakeIsolator{}, &countingProvisioner{}, "isolation is not available"},
		"provision fails": {&fakeIsolator{available: true}, &countingProvisioner{err: errors.New("nix down")}, "provision nix: nix down"},
		"no harness":      {&fakeIsolator{available: true, checkErr: errors.New("no claude in the base")}, &countingProvisioner{}, "no claude in the base"},
		"no command":      {&fakeIsolator{available: true, terminalErr: errors.New("cannot")}, &countingProvisioner{handles: true}, "prepare the bwrap command: cannot"},
	} {
		_, err := launch.New(fakePorts(tc.iso, tc.prov), model.Sandbox{Isolation: "bwrap", Provision: "nix", NixRev: nixRev}).Confine(harnessRun())
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want %q", name, err, tc.want)
		}
	}
}

func TestAPinnedProvisionDemandAloneSelectsThePinnedPair(t *testing.T) {
	iso, prov := &fakeIsolator{available: true}, &countingProvisioner{handles: true}
	if _, err := launch.New(fakePorts(iso, prov), model.Sandbox{RequirePinnedProvision: true, NixRev: nixRev}).Confine(harnessRun()); err != nil || prov.made != 1 {
		t.Fatalf("%v, provisioned %d times", err, prov.made)
	}
}

// opener records the proxies it opens and closes.
type opener struct {
	opened, closed int
	allow          egress.Allowlist
	err            error
}

func (o *opener) Open(transport network.EgressTransport, allow egress.Allowlist) (isoshared.EgressRoute, func(), error) {
	if o.err != nil {
		return isoshared.EgressRoute{}, nil, o.err
	}
	o.opened++
	o.allow = allow
	return isoshared.EgressRoute{Transport: transport, Allow: allow, Proxy: "127.0.0.1:3128"}, func() { o.closed++ }, nil
}

func TestAManagedRunHasItsOwnProxyUntilItIsReleased(t *testing.T) {
	iso, prov := &fakeIsolator{available: true, transport: network.TransportLoopback}, &countingProvisioner{handles: true}
	ports := fakePorts(iso, prov)
	proxy := &opener{}
	ports.Egress = proxy
	box := model.Sandbox{Isolation: "bwrap", Provision: "nix", NixRev: nixRev, Network: "managed", AllowHosts: []string{"api.example.com"}}
	confined, err := launch.New(ports, box).Confine(harnessRun())
	if err != nil {
		t.Fatal(err)
	}
	if proxy.opened != 1 || proxy.closed != 0 || confined.Process.Env["PROXY"] != "127.0.0.1:3128" {
		t.Fatalf("opened %d, closed %d, env %v: the run reaches the proxy it opened", proxy.opened, proxy.closed, confined.Process.Env)
	}
	confined.Release()
	if proxy.closed != 1 {
		t.Errorf("closed %d times, want once on release", proxy.closed)
	}

	iso.terminalErr = errors.New("cannot")
	if _, err := launch.New(ports, box).Confine(harnessRun()); err == nil || proxy.closed != 2 {
		t.Errorf("%v, closed %d: a run that cannot start closes its proxy", err, proxy.closed)
	}
	proxy.err = errors.New("no socket")
	if _, err := launch.New(ports, box).Confine(harnessRun()); err == nil || !strings.Contains(err.Error(), "network managed: no socket") {
		t.Errorf("err = %v", err)
	}
	ports.Egress = nil
	if _, err := launch.New(ports, box).Confine(harnessRun()); err == nil || !strings.Contains(err.Error(), "starts no egress proxy") {
		t.Errorf("err = %v", err)
	}
}

// vault is a credential store that counts its reads.
type vault struct{ reads int }

func (v *vault) Available() bool { return true }
func (v *vault) Lookup(locator string) (credshared.Secret, error) {
	v.reads++
	if locator == "missing" {
		return credshared.Secret{}, credshared.ErrNotFound
	}
	return credshared.NewSecret("s3cret-" + locator), nil
}

func TestCredentialsAreReadOnceAndReachTheHarnessAsEnvironment(t *testing.T) {
	iso, prov := &fakeIsolator{available: true}, &countingProvisioner{handles: true}
	ports := fakePorts(iso, prov)
	store := &vault{}
	ports.Credentials = credshared.NewRegistry().Register("vault", func() credshared.Store { return store })
	confiner := launch.New(ports, model.Sandbox{Isolation: "bwrap", Provision: "nix", NixRev: nixRev, Credentials: []string{"API_KEY=vault:api"}})
	for range 2 {
		if _, err := confiner.Confine(harnessRun()); err != nil {
			t.Fatal(err)
		}
	}
	if store.reads != 1 || iso.runs[1].Secrets["API_KEY"] != "s3cret-api" {
		t.Errorf("reads %d, secrets %v", store.reads, iso.runs[1].Secrets)
	}
	for _, args := range iso.runs[1].Process.Argv {
		if strings.Contains(args, "s3cret") {
			t.Errorf("argv %v holds the secret", iso.runs[1].Process.Argv)
		}
	}

	missing := launch.New(ports, model.Sandbox{Isolation: "bwrap", Provision: "nix", NixRev: nixRev, Credentials: []string{"API_KEY=vault:missing"}})
	if _, err := missing.Confine(harnessRun()); !errors.Is(err, credshared.ErrNotFound) {
		t.Errorf("err = %v", err)
	}
	ports.Credentials = nil
	if _, err := launch.New(ports, model.Sandbox{Isolation: "bwrap", Provision: "nix", NixRev: nixRev, Credentials: []string{"API_KEY=vault:api"}}).Confine(harnessRun()); err == nil || !strings.Contains(err.Error(), "reads no credential store") {
		t.Errorf("err = %v", err)
	}
}

func TestABadEnvironmentEntryIsAnError(t *testing.T) {
	iso, prov := &fakeIsolator{available: true}, &countingProvisioner{handles: true}
	if _, err := launch.New(fakePorts(iso, prov), model.Sandbox{Isolation: "bwrap", Provision: "nix", NixRev: nixRev, CustomEnv: []string{"=x"}}).Confine(harnessRun()); err == nil {
		t.Fatal("an environment entry without a name is refused")
	}
}
