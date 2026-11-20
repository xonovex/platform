// Package bwrap is the isolation=bwrap leaf: bubblewrap-namespace confinement.
package bwrap

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/hostlaunch"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
)

// Isolator confines the agent with bubblewrap namespaces and applies a
// Contribution via read-only binds plus an explicit setenv allowlist.
//
// Deny-default (HostPassthrough off): a sandbox-local HOME (tmpfs, no host-$HOME
// bind) with only the curated config paths bound back read-only; only the
// workspace (rw) and RepoDir (ro) are bound; /dev is a minimal devtmpfs (not a
// full --dev-bind, which would expose /dev/sda, /dev/mem and input devices); the
// environment is reduced to an explicit allowlist with --unsetenv.
// HostPassthrough on restores the leaky behavior: host /usr,/lib,/bin,/etc are
// ro-bound and the host PATH is appended. It does not bind /nix/store, so a run
// fails before it starts when the agent or its interpreter resolves outside the
// binds.
//
// bwrap and default runc are attack-surface reduction, not a kernel trust
// boundary (see KernelIsolated); no-new-privs and cap-drop are bwrap defaults.
type Isolator struct {
	host hostlaunch.Host
	log  logging.Logger
}

// NewIsolator creates a bwrap isolator over host, writing through log.
func NewIsolator(host hostlaunch.Host, log logging.Logger) *Isolator {
	return &Isolator{host: host, log: log}
}

// Available reports whether bwrap is installed.
func (i *Isolator) Available() (bool, error) {
	_, err := i.host.LookPath("bwrap")
	return err == nil, nil
}

// HidesHost reports that host tools are off PATH and not bind-reachable in
// deny-default mode; HostPassthrough forfeits the guarantee.
func (i *Isolator) HidesHost(passthrough bool, _ string) bool { return !passthrough }

// PinnedProvision mirrors the selected provisioner's immutability because
// bubblewrap supplies no image-based tools of its own.
func (i *Isolator) PinnedProvision(_ provision.ProvisionMethod, provisionerPinned bool, _ string) bool {
	return provisionerPinned
}

// KernelIsolated reports false: bwrap is attack-surface reduction, not a kernel
// trust boundary.
func (i *Isolator) KernelIsolated(_ string) bool { return false }

// ConfinesNetwork reports true: bwrap unshares the network namespace.
func (i *Isolator) ConfinesNetwork() bool { return true }

// EgressTransport declares the unix socket: a managed run has no network
// namespace interface, and reaches the allowlist proxy through a bound socket.
func (i *Isolator) EgressTransport() network.EgressTransport { return network.TransportUnixSocket }

// Bounded reports true: the namespaces confine what the agent writes and reaches.
func (i *Isolator) Bounded() bool { return true }

// AcceptsHostBinds reports true: bwrap binds a host path into the sandbox.
func (i *Isolator) AcceptsHostBinds() bool { return true }

// RequiresImage reports false: bwrap confines the host file system.
func (i *Isolator) RequiresImage() bool { return false }

// CheckAgent requires host passthrough, because deny-default bwrap exposes no
// host tool, and then the agent executable on the PATH it passes through.
func (i *Isolator) CheckAgent(cfg isoshared.RunConfig) error {
	if !cfg.HostPassthrough {
		return fmt.Errorf("bwrap with provision=none cannot supply agent executable %q while host passthrough is disabled; use provision=nix or enable --isolation-bwrap-passthrough", cfg.Agent.Binary)
	}
	if _, err := i.host.LookPath(cfg.Agent.Binary); err != nil {
		return fmt.Errorf("agent executable %q is not available on passthrough PATH: %w", cfg.Agent.Binary, err)
	}
	return nil
}

// Run executes the agent in the bubblewrap sandbox.
func (i *Isolator) Run(cfg isoshared.RunConfig, c provision.Contribution) (int, error) {
	args, err := i.buildArgs(cfg, c)
	if err != nil {
		return 1, err
	}
	env, err := i.processEnv(cfg, c)
	if err != nil {
		return 1, err
	}
	return hostlaunch.SpawnSandbox(i.host, i.log, "bwrap", args, env, "", "Bubblewrap sandbox")
}

// Command returns the full bwrap command (for display / terminal wrappers).
func (i *Isolator) Command(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	args, err := i.buildArgs(cfg, c)
	if err != nil {
		return nil, err
	}
	return append([]string{"bwrap"}, args...), nil
}

// StopCommand returns nil: bwrap runs with --die-with-parent, so the end of the
// bwrap process ends the sandbox.
func (i *Isolator) StopCommand(isoshared.RunConfig) []string { return nil }

// TerminalCommand returns the bwrap command plus the environment used to launch it.
func (i *Isolator) TerminalCommand(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string, error) {
	command, err := i.Command(cfg, c)
	if err != nil {
		return nil, nil, err
	}
	env, err := i.processEnv(cfg, c)
	return command, env, err
}

func (i *Isolator) buildArgs(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	homeDir, err := hostlaunch.ResolveDirectory(i.host, cfg.HomeDir, "home directory")
	if err != nil {
		return nil, err
	}
	workDir, err := hostlaunch.ResolveDirectory(i.host, cfg.Process.Dir, "work directory")
	if err != nil {
		return nil, err
	}
	repoDir := ""
	if cfg.RepoDir != "" {
		repoDir, err = hostlaunch.ResolveDirectory(i.host, cfg.RepoDir, "repository directory")
		if err != nil {
			return nil, err
		}
	}

	sandboxEnv, err := i.sandboxEnv(cfg, c, homeDir)
	if err != nil {
		return nil, err
	}

	args := []string{
		"--unshare-uts",
		"--unshare-ipc",
		"--unshare-pid",
		"--unshare-cgroup",
		"--die-with-parent",
		// Minimal devtmpfs — NOT --dev-bind /dev /dev, which exposes /dev/sda,
		// /dev/mem and input devices.
		"--dev", "/dev",
		"--proc", "/proc",
		"--tmpfs", "/tmp",
	}
	// bound collects the destination of every bind, which is all that the
	// sandbox reaches of the host file system.
	var bound []string
	bind := func(flag, src, dest string) {
		args = append(args, flag, src, dest)
		bound = append(bound, dest)
	}
	inherited := envutil.ParseEnv(i.host.Environ())
	unset := make([]string, 0, len(inherited))
	for key := range inherited {
		if _, allowed := sandboxEnv[key]; !allowed {
			unset = append(unset, key)
		}
	}
	sort.Strings(unset)
	for _, key := range unset {
		args = append(args, "--unsetenv", key)
	}

	// Network is applied explicitly through the network bridge.
	args = append(args, networkArgs(cfg.Network)...)

	// Sandbox-local HOME: a tmpfs at the home path (no host-$HOME bind), with only
	// the curated config paths bound back read-only.
	args = append(args, "--tmpfs", homeDir)
	for _, configPath := range cfg.Agent.UserConfigPaths {
		src, exists, err := hostlaunch.ResolveContainedOptionalPath(i.host, homeDir, configPath, "user config path")
		if err != nil {
			return nil, err
		}
		if exists {
			bind("--ro-bind", src, filepath.Join(homeDir, configPath))
		}
	}

	// Workspace (rw) + source repo (ro for worktrees).
	bind("--bind", workDir, workDir)
	if repoDir != "" && repoDir != workDir {
		bind("--ro-bind", repoDir, repoDir)
	}

	// HostPassthrough: restore host tool reachability (leaky mode).
	if cfg.HostPassthrough {
		for _, dir := range []string{"/usr", "/lib", "/lib64", "/bin", "/etc"} {
			if _, err := i.host.Stat(dir); err == nil {
				bind("--ro-bind", dir, dir)
			}
		}
	}

	// Contribution: read-only binds of the provisioned closure's requisites.
	for _, p := range c.RoBindPaths {
		resolved, err := hostlaunch.ResolveExistingPath(i.host, p, "provisioned read-only bind")
		if err != nil {
			return nil, err
		}
		bind("--ro-bind", resolved, resolved)
	}

	// Caller-supplied extra binds.
	for _, path := range cfg.BindPaths {
		resolved, err := hostlaunch.ResolveExistingPath(i.host, path, "read-write bind")
		if err != nil {
			return nil, err
		}
		bind("--bind", resolved, resolved)
	}
	for _, path := range cfg.RoBindPaths {
		resolved, err := hostlaunch.ResolveExistingPath(i.host, path, "read-only bind")
		if err != nil {
			return nil, err
		}
		bind("--ro-bind", resolved, resolved)
	}

	args = append(args, "--chdir", workDir)

	// Agent process, wrapped with the provisioner's init commands.
	command := hostlaunch.WrapWithInitCommands(cfg.Process.Argv, c.InitCommands)
	// Managed egress: bind the proxy socket directory and the bridge executable,
	// and run the command under the bridge.
	if cfg.Network == network.NetworkManaged {
		route, err := hostlaunch.ResolveBridge(i.host, cfg.Egress)
		if err != nil {
			return nil, err
		}
		socketDir, bridge := isoshared.BridgeMounts(route)
		bind("--bind", socketDir, socketDir)
		bind("--ro-bind", bridge, bridge)
		command = isoshared.WrapWithBridge(route, command)
	}
	if cfg.HostPassthrough {
		if err := i.checkPrograms(cfg, c, sandboxView{binds: bound, path: filepath.SplitList(sandboxEnv["PATH"]), workDir: workDir}); err != nil {
			return nil, err
		}
	}
	args = append(args, "--")
	args = append(args, command...)

	return args, nil
}

// checkPrograms requires the programs that the sandbox starts by name to resolve
// inside its binds: the agent, and the shell that runs the init commands of the
// provisioner.
func (i *Isolator) checkPrograms(cfg isoshared.RunConfig, c provision.Contribution, view sandboxView) error {
	var programs []string
	if len(cfg.Process.Argv) > 0 {
		programs = append(programs, cfg.Process.Argv[0])
	}
	if len(c.InitCommands) > 0 {
		programs = append(programs, hostlaunch.WrapWithInitCommands(nil, c.InitCommands)[0])
	}
	for _, program := range programs {
		if err := i.checkReachable(view, program); err != nil {
			return err
		}
	}
	return nil
}

func (i *Isolator) processEnv(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	homeDir, err := hostlaunch.ResolveDirectory(i.host, cfg.HomeDir, "home directory")
	if err != nil {
		return nil, err
	}
	sandboxEnv, err := i.sandboxEnv(cfg, c, homeDir)
	if err != nil {
		return nil, err
	}
	return envutil.EnvMapToSlice(envutil.MergeEnvMaps(envutil.ParseEnv(i.host.Environ()), sandboxEnv)), nil
}

// sandboxEnv builds the explicit environment allowlist for inside the sandbox:
// HOME/PATH/TMPDIR/SHELL, the provider tokens, the contribution's env, and the
// environment of the process. API keys reach the sandbox only through this allowlist.
func (i *Isolator) sandboxEnv(cfg isoshared.RunConfig, c provision.Contribution, homeDir string) (map[string]string, error) {
	env := map[string]string{
		"HOME":   homeDir,
		"TMPDIR": "/tmp",
		"SHELL":  "/bin/bash",
	}

	// PATH: the contribution's tool dirs first, then standard dirs (resolvable
	// only under HostPassthrough), then the host PATH when passing through.
	pathEntries := append([]string{}, c.PathEntries...)
	pathEntries = append(pathEntries, "/usr/local/bin", "/usr/bin", "/bin")
	if cfg.HostPassthrough {
		if hostPath := i.host.Getenv("PATH"); hostPath != "" {
			pathEntries = append(pathEntries, hostPath)
		}
	}
	env["PATH"] = strings.Join(pathEntries, ":")

	// Provider tokens + contribution env + process env.
	for k, v := range cfg.ProviderEnv {
		env[k] = v
	}
	for k, v := range c.Env {
		env[k] = v
	}
	customEnv, err := hostlaunch.RunEnv(cfg)
	if err != nil {
		return nil, err
	}
	for k, v := range customEnv {
		env[k] = v
	}
	return env, nil
}
