// Package launch is the confinement that runs the harness inside the shared
// isolator of agent-cli. It plans each run with the shared-launch-go planner,
// which selects the isolator and the provisioner from a registry and refuses a
// sandbox that falls short of the guarantees the eval demands, and it turns the
// plan into one command line through the Launcher of the isolator. It names no
// isolator: the composition root registers them.
package launch

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	agentprovision "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	launchcore "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/launch"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// namePrefix starts the name of every sandbox instance, so a leftover one is
// recognizable as the work of agent-eval.
const namePrefix = "agent-eval-"

// stopTimeout bounds the command that ends a sandbox instance.
const stopTimeout = 30 * time.Second

// ErrUnavailable reports an isolation whose tool this machine lacks.
var ErrUnavailable = errors.New("isolation is not available")

// EgressOpener starts the allowlist enforcement of a managed network run and
// returns the route the isolator applies and the function that ends it.
type EgressOpener interface {
	Open(transport network.EgressTransport, allow egress.Allowlist) (isoshared.EgressRoute, func(), error)
}

// Ports are what the confinement runs through.
type Ports struct {
	// Sandboxes holds the isolators and provisioners this build offers.
	Sandboxes *sandbox.Registry[isoshared.Isolator]
	// Credentials maps a credential scheme to its store.
	Credentials *credshared.Registry
	// Egress starts the proxy of a managed network run.
	Egress EgressOpener
	// Harnesses holds the facts of every harness: its program and the package
	// a nix provision installs.
	Harnesses agents.Catalog
	// Home is the home directory of the user, which an isolator hides.
	Home string
}

// Confiner confines every harness process in the sandbox of the eval section.
// It provisions each harness once and resolves the credentials once, and it
// is safe for the concurrent runs of one eval.
type Confiner struct {
	ports   Ports
	sandbox model.Sandbox

	mu        sync.Mutex
	available map[string]bool
	prepared  map[string]contribution
	secrets   map[string]string
	resolved  bool
}

// contribution is what the provisioner supplied for one harness.
type contribution struct {
	value agentprovision.Contribution
	// handled reports that the provisioner supplies the harness program, so the
	// isolator need not.
	handled bool
}

var _ confine.Confiner = (*Confiner)(nil)

// New returns the confinement that runs every harness in sandbox.
func New(ports Ports, sandbox model.Sandbox) *Confiner {
	return &Confiner{ports: ports, sandbox: sandbox, available: map[string]bool{}, prepared: map[string]contribution{}}
}

// Confine plans the run of request in its workspace and returns the command
// that starts the harness inside the sandbox. The workspace is the only host
// directory the sandbox binds, read-write. The sandbox gets no terminal, and it
// gets a name derived from the workspace, so Stop can end it after a timeout.
func (c *Confiner) Confine(request confine.Request) (confine.Confined, error) {
	agent, err := c.facts(request.Harness)
	if err != nil {
		return confine.Confined{}, err
	}
	workspace := request.Process.Dir
	spec, choice, err := launchcore.PlanSpec(c.planRequest(request.Harness), agent, nil, workspace, workspace, c.ports.Home, c.ports.Sandboxes, true)
	if err != nil {
		return confine.Confined{}, err
	}
	isolator := choice.Isolator
	if err := c.checkAvailable(string(spec.IsolationName), isolator); err != nil {
		return confine.Confined{}, err
	}
	provided, err := c.provision(request.Harness, spec, choice)
	if err != nil {
		return confine.Confined{}, err
	}
	run, err := c.runConfig(spec, request)
	if err != nil {
		return confine.Confined{}, err
	}
	if !provided.handled {
		if err := isolator.CheckAgent(run); err != nil {
			return confine.Confined{}, err
		}
	}

	release := func() {}
	if run.Network == network.NetworkManaged {
		if c.ports.Egress == nil {
			return confine.Confined{}, errors.New("network managed: this build starts no egress proxy")
		}
		route, stop, err := c.ports.Egress.Open(run.Egress.Transport, spec.Allow)
		if err != nil {
			return confine.Confined{}, fmt.Errorf("network managed: %w", err)
		}
		run.Egress, release = route, stop
	}
	command, env, err := isolator.TerminalCommand(run, provided.value)
	if err != nil {
		release()
		return confine.Confined{}, fmt.Errorf("prepare the %s command: %w", spec.IsolationName, err)
	}
	if len(command) == 0 {
		release()
		return confine.Confined{}, fmt.Errorf("prepare the %s command: the isolator returned no command", spec.IsolationName)
	}
	return confine.Confined{
		Process: process.Request{
			Dir: workspace, Name: command[0], Args: command[1:],
			Env: envutil.ParseEnv(env), Timeout: request.Process.Timeout,
		},
		Stop:    stopRequest(isolator.StopCommand(run)),
		Release: release,
	}, nil
}

// facts returns the facts of the harness without the configuration paths of
// the user: an eval run is hermetic, so no isolator binds the settings of the
// user into it.
func (c *Confiner) facts(harness string) (*types.AgentConfig, error) {
	agent, err := c.ports.Harnesses.Get(types.AgentType(harness))
	if err != nil {
		return nil, err
	}
	facts := *agent
	facts.UserConfigPaths = nil
	return &facts, nil
}

// planRequest is the request of the planner. The isolation or the provision
// counts as chosen when the sandbox names it, so a pinned-provision demand alone
// selects the pinned default pair.
func (c *Confiner) planRequest(harness string) launchcore.Request {
	s := c.sandbox
	return launchcore.Request{
		Agent:                     harness,
		Harnesses:                 c.ports.Harnesses,
		Isolation:                 s.Isolation,
		Provision:                 s.Provision,
		Network:                   s.Network,
		ExplicitAxes:              s.Isolation != "" || s.Provision != "",
		IsolationBwrapPassthrough: s.HostPassthrough,
		IsolationDockerRuntime:    s.Runtime,
		Image:                     s.Image,
		InitCommands:              s.InitCommands,
		NixRev:                    s.NixRev,
		NixPackages:               s.NixPackages,
		CustomEnv:                 s.CustomEnv,
		AllowHosts:                s.AllowHosts,
		Credentials:               s.Credentials,
		Policy: policy.SandboxPolicy{
			RequirePinnedProvision:      s.RequirePinnedProvision,
			RequireHostToolsUnreachable: s.RequireHostToolsUnreachable,
			RequireEgressRestricted:     s.RequireEgressRestricted,
			RequireKernelIsolation:      s.RequireKernelIsolation,
		},
	}
}

// runConfig confines the program of the harness leaf in place of the agent
// process the planner computed, with the environment of the sandbox under the
// environment of the leaf, which keeps the run hermetic.
func (c *Confiner) runConfig(spec launchcore.Spec, request confine.Request) (isoshared.RunConfig, error) {
	custom, err := envutil.ParseCustomEnv(spec.Run.Process.Env)
	if err != nil {
		return isoshared.RunConfig{}, err
	}
	secrets, err := c.resolveSecrets(spec.Credentials)
	if err != nil {
		return isoshared.RunConfig{}, err
	}
	run := spec.Run
	run.Process = isoshared.Process{
		Argv: append([]string{request.Process.Name}, request.Process.Args...),
		Env:  envutil.EnvMapToSlice(envutil.MergeEnvMaps(custom, request.Process.Env)),
		Dir:  request.Process.Dir,
	}
	run.RepoDir = ""
	run.BindPaths = nil
	run.Interactive = false
	run.Name = InstanceName(request.Process.Dir)
	run.Secrets = secrets
	return run, nil
}

// InstanceName names the sandbox instance of the run in workspace. Every run
// has its own workspace, so concurrent runs never share a name.
func InstanceName(workspace string) string {
	sum := sha256.Sum256([]byte(workspace))
	return namePrefix + hex.EncodeToString(sum[:])[:12]
}

// checkAvailable asks once per isolation whether its tool is installed.
func (c *Confiner) checkAvailable(name string, isolator isoshared.Isolator) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.available[name] {
		return nil
	}
	ok, err := isolator.Available()
	if err != nil || !ok {
		return fmt.Errorf("%w: %s; available isolations: %v", ErrUnavailable, name, sandbox.AvailableIsolations(c.ports.Sandboxes))
	}
	c.available[name] = true
	return nil
}

// provision runs the provisioner of the harness once and keeps what it
// supplied: a nix provision builds a closure, which every run reuses.
func (c *Confiner) provision(harness string, spec launchcore.Spec, choice sandbox.Choice[isoshared.Isolator]) (contribution, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if prepared, ok := c.prepared[harness]; ok {
		return prepared, nil
	}
	value, err := choice.Provisioner.Contribute(spec.Provision)
	if err != nil {
		return contribution{}, fmt.Errorf("provision %s: %w", spec.ProvisionName, err)
	}
	handled, err := choice.Provisioner.CheckAgent(value, spec.Agent)
	if err != nil {
		return contribution{}, err
	}
	prepared := contribution{value: value, handled: handled}
	c.prepared[harness] = prepared
	return prepared, nil
}

// resolveSecrets reads every bound credential once. A value reaches the
// harness only as an environment variable.
func (c *Confiner) resolveSecrets(bindings []credshared.Binding) (map[string]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.resolved || len(bindings) == 0 {
		return c.secrets, nil
	}
	if c.ports.Credentials == nil {
		return nil, errors.New("credentials: this build reads no credential store")
	}
	secrets, err := credshared.Resolve(c.ports.Credentials, bindings)
	if err != nil {
		return nil, err
	}
	c.secrets, c.resolved = credshared.Reveal(secrets), true
	return c.secrets, nil
}

func stopRequest(command []string) *process.Request {
	if len(command) == 0 {
		return nil
	}
	return &process.Request{Name: command[0], Args: command[1:], Timeout: stopTimeout}
}
