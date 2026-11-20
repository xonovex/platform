// Package dockersandboxes is the isolation=docker-sandboxes leaf: the agent runs
// in a Docker Sandboxes sandbox that the leaf creates, executes in and removes
// through the sbx program.
package dockersandboxes

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/hostlaunch"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
)

// program is the Docker Sandboxes command line tool.
const program = "sbx"

// namePrefix starts the name of every sandbox this leaf creates, so a leftover
// sandbox is recognizable.
const namePrefix = "agent-cli-"

// maxNames bounds the search for a free sandbox name.
const maxNames = 1000

// ErrUnsupported reports a run that Docker Sandboxes cannot hold.
var ErrUnsupported = errors.New("docker-sandboxes cannot run this")

// Isolator runs the agent in a Docker Sandboxes sandbox.
//
// The sandbox is a separate machine with its own agent image, so the agent sees
// no host tools and no home directory. The leaf declares no network confinement
// and no egress transport: sbx limits egress with deny rules only, so no allowlist
// can be installed, and a run that asks for none or managed fails closed in the
// registry instead of running with the host network.
type Isolator struct {
	host hostlaunch.Host
	log  logging.Logger
}

// NewIsolator creates a Docker Sandboxes isolator over host, writing through log.
func NewIsolator(host hostlaunch.Host, log logging.Logger) *Isolator {
	return &Isolator{host: host, log: log}
}

// Available reports whether sbx is installed.
func (i *Isolator) Available() (bool, error) {
	_, err := i.host.LookPath(program)
	return err == nil, nil
}

// HidesHost reports false: the isolator does not pin the agent image of the sandbox,
// so it cannot show that the tools the agent sees are fixed.
func (i *Isolator) HidesHost(bool, string) bool { return false }

// PinnedProvision reports false: the agent image belongs to sbx, not to the run.
func (i *Isolator) PinnedProvision(provision.ProvisionMethod, bool, string) bool { return false }

// KernelIsolated reports false: the isolator does not verify the virtual machine
// boundary of the installed sbx version.
func (i *Isolator) KernelIsolated(string) bool { return false }

// ConfinesNetwork reports false: sbx has deny rules only.
func (i *Isolator) ConfinesNetwork() bool { return false }

// EgressTransport declares none: an allowlist cannot be installed.
func (i *Isolator) EgressTransport() network.EgressTransport { return network.TransportNone }

// Bounded reports true: the microVM confines what the agent writes.
func (i *Isolator) Bounded() bool { return true }

// AcceptsHostBinds reports true: the sandbox mounts a host path as a
// workspace.
func (i *Isolator) AcceptsHostBinds() bool { return true }

// CheckAgent has nothing to check: the agent binary lives in the sandbox image.
func (i *Isolator) CheckAgent(isoshared.RunConfig) error { return nil }

// plan holds the commands of one run.
type plan struct {
	name   string
	create []string
	exec   []string
	// values are the credential, provider and custom environment values that
	// reach the sandbox through the environment of the exec command, by name.
	values map[string]string
}

// Command returns the exec command of the run, for display. It names the sandbox
// the run would create first.
func (i *Isolator) Command(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	p, err := i.plan(cfg, c, namePrefix+"1")
	if err != nil {
		return nil, err
	}
	return append([]string{program}, p.exec...), nil
}

// StopCommand returns nil: only Run creates a sandbox, and Run removes it on
// every path that leaves it.
func (i *Isolator) StopCommand(isoshared.RunConfig) []string { return nil }

// TerminalCommand cannot place the sandbox lifecycle inside a terminal wrapper:
// the sandbox must exist before exec and be removed after it.
func (i *Isolator) TerminalCommand(isoshared.RunConfig, provision.Contribution) ([]string, []string, error) {
	return nil, nil, fmt.Errorf("%w: a terminal wrapper cannot create and remove the sandbox", ErrUnsupported)
}

// Run creates the sandbox, executes the agent in it, and removes the sandbox on
// every path that leaves Run after the create call.
func (i *Isolator) Run(cfg isoshared.RunConfig, c provision.Contribution) (int, error) {
	name, err := i.freeName()
	if err != nil {
		return 1, err
	}
	p, err := i.plan(cfg, c, name)
	if err != nil {
		return 1, err
	}
	if _, err := i.host.Capture(program, p.create, ""); err != nil {
		// A create that failed half way can still have left the sandbox.
		_ = i.release(name)
		return 1, fmt.Errorf("create sandbox: %w", err)
	}
	env := append(i.host.Environ(), envutil.EnvMapToSlice(p.values)...)
	code, runErr := hostlaunch.SpawnSandbox(i.host, i.log, program, p.exec, env, "", "Docker Sandboxes")
	if err := i.release(name); err != nil {
		i.log.Warning(fmt.Sprintf("Could not remove sandbox %s: %v", name, err))
	}
	return code, runErr
}

// release stops and removes the sandbox.
func (i *Isolator) release(name string) error {
	_, stopErr := i.host.Capture(program, []string{"stop", name}, "")
	_, removeErr := i.host.Capture(program, []string{"rm", "--force", name}, "")
	if removeErr != nil {
		return errors.Join(removeErr, stopErr)
	}
	return nil
}

// freeName returns the first sandbox name that sbx does not list.
func (i *Isolator) freeName() (string, error) {
	listing, err := i.host.Capture(program, []string{"ls", "-q"}, "")
	if err != nil {
		return "", fmt.Errorf("list sandboxes: %w", err)
	}
	taken := map[string]bool{}
	for _, line := range strings.Fields(listing) {
		taken[line] = true
	}
	for n := 1; n <= maxNames; n++ {
		if name := namePrefix + strconv.Itoa(n); !taken[name] {
			return name, nil
		}
	}
	return "", fmt.Errorf("no free sandbox name under %s", namePrefix)
}

// mountable rejects a path that sbx would read as an option or as a mode suffix.
func mountable(path string) error {
	if strings.ContainsAny(path, ":\n\r") || strings.HasPrefix(path, "-") {
		return fmt.Errorf("%w: sbx cannot take the path %q", ErrUnsupported, path)
	}
	return nil
}

func (i *Isolator) plan(cfg isoshared.RunConfig, c provision.Contribution, name string) (plan, error) {
	switch {
	case cfg.Network != network.NetworkHost:
		return plan{}, fmt.Errorf("%w: network mode %q", ErrUnsupported, cfg.Network)
	case len(c.InitCommands) > 0:
		return plan{}, fmt.Errorf("%w: the sandbox starts the agent itself, so initialization commands cannot run before it", ErrUnsupported)
	case len(c.PathEntries) > 0:
		return plan{}, fmt.Errorf("%w: a provisioned tool directory cannot reach the sandbox PATH", ErrUnsupported)
	}
	workDir, err := hostlaunch.ResolveDirectory(i.host, cfg.Process.Dir, "work directory")
	if err != nil {
		return plan{}, err
	}
	runEnv, err := hostlaunch.RunEnv(cfg)
	if err != nil {
		return plan{}, err
	}
	values := envutil.MergeEnvMaps(cfg.ProviderEnv, runEnv, c.Env)

	mounts := []string{workDir}
	for _, path := range cfg.BindPaths {
		resolved, err := hostlaunch.ResolveExistingPath(i.host, path, "read-write bind")
		if err != nil {
			return plan{}, err
		}
		mounts = append(mounts, resolved)
	}
	var readOnly []string
	for _, path := range append(append([]string(nil), c.RoBindPaths...), cfg.RoBindPaths...) {
		resolved, err := hostlaunch.ResolveExistingPath(i.host, path, "read-only bind")
		if err != nil {
			return plan{}, err
		}
		readOnly = append(readOnly, resolved+":ro")
	}
	for _, mount := range append(append([]string(nil), mounts...), readOnly...) {
		if err := mountable(strings.TrimSuffix(mount, ":ro")); err != nil {
			return plan{}, err
		}
	}

	create := append([]string{"create", "--no-progress", "--name", name, string(cfg.Agent.Type)}, mounts...)
	create = append(create, readOnly...)

	exec := []string{"exec"}
	if cfg.Interactive {
		exec = append(exec, "-i", "-t")
	}
	exec = append(exec, "-w", workDir)
	names := make([]string, 0, len(values))
	for variable := range values {
		names = append(names, variable)
	}
	sort.Strings(names)
	for _, variable := range names {
		exec = append(exec, "-e", variable)
	}
	exec = append(exec, name)
	exec = append(exec, cfg.Process.Argv...)
	return plan{name: name, create: create, exec: exec, values: values}, nil
}
