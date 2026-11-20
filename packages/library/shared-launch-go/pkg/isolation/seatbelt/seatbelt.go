// Package seatbelt is the isolation=seatbelt leaf: the agent runs on the host
// under a macOS Seatbelt profile that denies file writes outside the workspace,
// hides the home directory, and confines the network.
package seatbelt

import (
	"fmt"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/hostlaunch"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
)

// program is the macOS command that runs a process under a profile.
const program = "sandbox-exec"

// Isolator confines the agent with a Seatbelt profile. The agent keeps the host
// tools and the host kernel, so the isolator hides neither: it limits what the
// agent can write, what it can read in the home directory, and where it can
// connect. A profile has no way to name the interfaces of a network namespace, so
// a managed run reaches an allowlist proxy on the host loopback, and the profile
// allows that one port.
type Isolator struct {
	host hostlaunch.Host
	log  logging.Logger
}

// NewIsolator creates a seatbelt isolator over host, writing through log.
func NewIsolator(host hostlaunch.Host, log logging.Logger) *Isolator {
	return &Isolator{host: host, log: log}
}

// Available reports whether sandbox-exec is installed, which only macOS does.
func (i *Isolator) Available() (bool, error) {
	_, err := i.host.LookPath(program)
	return err == nil, nil
}

// HidesHost reports false: the agent runs the host tools.
func (i *Isolator) HidesHost(bool, string) bool { return false }

// PinnedProvision mirrors the selected provisioner's immutability because the
// agent runs from the host, with no image input.
func (i *Isolator) PinnedProvision(_ provision.ProvisionMethod, provisionerPinned bool, _ string) bool {
	return provisionerPinned
}

// KernelIsolated reports false: a profile restricts a process of the host kernel.
func (i *Isolator) KernelIsolated(string) bool { return false }

// ConfinesNetwork reports true: the profile denies network operations.
func (i *Isolator) ConfinesNetwork() bool { return true }

// EgressTransport declares loopback: the profile allows only the loopback port of
// the allowlist proxy.
func (i *Isolator) EgressTransport() network.EgressTransport { return network.TransportLoopback }

// Bounded reports true: the profile confines what the agent writes and reaches.
func (i *Isolator) Bounded() bool { return true }

// AcceptsHostBinds reports true: the profile grants access to a host path.
func (i *Isolator) AcceptsHostBinds() bool { return true }

// RequiresImage reports false: Seatbelt confines the host file system.
func (i *Isolator) RequiresImage() bool { return false }

// CheckAgent requires the agent executable on the PATH of the host.
func (i *Isolator) CheckAgent(cfg isoshared.RunConfig) error {
	if _, err := i.host.LookPath(cfg.Agent.Binary); err != nil {
		return fmt.Errorf("agent executable %q is not available on host PATH: %w", cfg.Agent.Binary, err)
	}
	return nil
}

// Run executes the agent under the profile.
func (i *Isolator) Run(cfg isoshared.RunConfig, c provision.Contribution) (int, error) {
	cmd, env, err := i.command(cfg, c)
	if err != nil {
		return 1, err
	}
	return hostlaunch.SpawnSandbox(i.host, i.log, cmd[0], cmd[1:], env, cfg.Process.Dir, "Seatbelt sandbox")
}

// Command returns the sandbox-exec command (for display and terminal wrappers).
func (i *Isolator) Command(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	cmd, _, err := i.command(cfg, c)
	return cmd, err
}

// StopCommand returns nil: sandbox-exec replaces itself with the agent, so the
// agent is the process of Command.
func (i *Isolator) StopCommand(isoshared.RunConfig) []string { return nil }

// TerminalCommand returns the command plus the environment used to launch it.
func (i *Isolator) TerminalCommand(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string, error) {
	return i.command(cfg, c)
}

func (i *Isolator) command(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string, error) {
	spec, err := i.spec(cfg, c)
	if err != nil {
		return nil, nil, err
	}
	profile, err := renderProfile(spec)
	if err != nil {
		return nil, nil, err
	}
	agent, env, err := hostlaunch.HostExecution(i.host, cfg, c)
	if err != nil {
		return nil, nil, err
	}
	return append([]string{program, "-p", profile}, agent...), env, nil
}

// spec resolves every path the run names to its canonical form.
func (i *Isolator) spec(cfg isoshared.RunConfig, c provision.Contribution) (profileSpec, error) {
	home, err := hostlaunch.ResolveDirectory(i.host, cfg.HomeDir, "home directory")
	if err != nil {
		return profileSpec{}, err
	}
	workDir, err := hostlaunch.ResolveDirectory(i.host, cfg.Process.Dir, "work directory")
	if err != nil {
		return profileSpec{}, err
	}
	spec := profileSpec{Home: home, Network: cfg.Network, Writable: []string{workDir}, Readable: []string{workDir}}
	if cfg.Network == network.NetworkManaged {
		if cfg.Egress.Transport != network.TransportLoopback {
			return profileSpec{}, fmt.Errorf("managed egress under seatbelt needs the loopback transport, not %q", cfg.Egress.Transport)
		}
		if err := cfg.Egress.Validate(); err != nil {
			return profileSpec{}, err
		}
		spec.ProxyAddress = cfg.Egress.Proxy
	}
	if cfg.RepoDir != "" {
		repo, err := hostlaunch.ResolveDirectory(i.host, cfg.RepoDir, "repository directory")
		if err != nil {
			return profileSpec{}, err
		}
		spec.Readable = append(spec.Readable, repo)
	}
	// The agent keeps its login and its sessions in its configuration, so it reads
	// and writes those places; the rest of the home directory stays hidden.
	for _, configPath := range cfg.Agent.UserConfigPaths {
		source, exists, err := hostlaunch.ResolveContainedOptionalPath(i.host, home, configPath, "user config path")
		if err != nil {
			return profileSpec{}, err
		}
		if exists {
			spec.Writable = append(spec.Writable, source)
			spec.Readable = append(spec.Readable, source)
		}
	}
	for _, path := range c.RoBindPaths {
		resolved, err := hostlaunch.ResolveExistingPath(i.host, path, "provisioned read-only bind")
		if err != nil {
			return profileSpec{}, err
		}
		spec.Readable = append(spec.Readable, resolved)
	}
	for _, path := range cfg.BindPaths {
		resolved, err := hostlaunch.ResolveExistingPath(i.host, filepath.Clean(path), "read-write bind")
		if err != nil {
			return profileSpec{}, err
		}
		spec.Writable = append(spec.Writable, resolved)
		spec.Readable = append(spec.Readable, resolved)
	}
	for _, path := range cfg.RoBindPaths {
		resolved, err := hostlaunch.ResolveExistingPath(i.host, path, "read-only bind")
		if err != nil {
			return profileSpec{}, err
		}
		spec.Readable = append(spec.Readable, resolved)
	}
	return spec, nil
}
