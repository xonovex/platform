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
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
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
// ro-bound and the host PATH is appended.
//
// bwrap and default runc are attack-surface reduction, not a kernel trust
// boundary (see KernelIsolated); no-new-privs and cap-drop are bwrap defaults.
type Isolator struct {
	host isoshared.Host
	log  logging.Logger
}

// NewIsolator creates a bwrap isolator over host, writing through log.
func NewIsolator(host isoshared.Host, log logging.Logger) *Isolator {
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
	return isoshared.SpawnSandbox(i.host, i.log, "bwrap", args, env, "", "Bubblewrap sandbox")
}

// Command returns the full bwrap command (for display / terminal wrappers).
func (i *Isolator) Command(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	args, err := i.buildArgs(cfg, c)
	if err != nil {
		return nil, err
	}
	return append([]string{"bwrap"}, args...), nil
}

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
	homeDir, err := isoshared.ResolveDirectory(i.host, cfg.HomeDir, "home directory")
	if err != nil {
		return nil, err
	}
	workDir, err := isoshared.ResolveDirectory(i.host, cfg.Process.Dir, "work directory")
	if err != nil {
		return nil, err
	}
	repoDir := ""
	if cfg.RepoDir != "" {
		repoDir, err = isoshared.ResolveDirectory(i.host, cfg.RepoDir, "repository directory")
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
		src, exists, err := isoshared.ResolveContainedOptionalPath(i.host, homeDir, configPath, "user config path")
		if err != nil {
			return nil, err
		}
		if exists {
			args = append(args, "--ro-bind", src, filepath.Join(homeDir, configPath))
		}
	}

	// Workspace (rw) + source repo (ro for worktrees).
	args = append(args, "--bind", workDir, workDir)
	if repoDir != "" && repoDir != workDir {
		args = append(args, "--ro-bind", repoDir, repoDir)
	}

	// HostPassthrough: restore host tool reachability (leaky mode).
	if cfg.HostPassthrough {
		for _, dir := range []string{"/usr", "/lib", "/lib64", "/bin", "/etc"} {
			if _, err := i.host.Stat(dir); err == nil {
				args = append(args, "--ro-bind", dir, dir)
			}
		}
	}

	// Contribution: read-only binds of the provisioned closure's requisites.
	for _, p := range c.RoBindPaths {
		resolved, err := isoshared.ResolveExistingPath(i.host, p, "provisioned read-only bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "--ro-bind", resolved, resolved)
	}

	// Caller-supplied extra binds.
	for _, path := range cfg.BindPaths {
		resolved, err := isoshared.ResolveExistingPath(i.host, path, "read-write bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "--bind", resolved, resolved)
	}
	for _, path := range cfg.RoBindPaths {
		resolved, err := isoshared.ResolveExistingPath(i.host, path, "read-only bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "--ro-bind", resolved, resolved)
	}

	args = append(args, "--chdir", workDir)

	// Agent process, wrapped with the provisioner's init commands.
	command := isoshared.WrapWithInitCommands(cfg.Process.Argv, c.InitCommands)
	// Managed egress: bind the proxy socket directory and the bridge executable,
	// and run the command under the bridge.
	if cfg.Network == network.NetworkManaged {
		route, err := isoshared.ResolveBridge(i.host, cfg.Egress)
		if err != nil {
			return nil, err
		}
		socketDir, bridge := isoshared.BridgeMounts(route)
		args = append(args, "--bind", socketDir, socketDir, "--ro-bind", bridge, bridge)
		command = isoshared.WrapWithBridge(route, command)
	}
	args = append(args, "--")
	args = append(args, command...)

	return args, nil
}

func (i *Isolator) processEnv(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	homeDir, err := isoshared.ResolveDirectory(i.host, cfg.HomeDir, "home directory")
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
	customEnv, err := isoshared.RunEnv(cfg)
	if err != nil {
		return nil, err
	}
	for k, v := range customEnv {
		env[k] = v
	}
	return env, nil
}
