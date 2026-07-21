// Package bwrap is the isolation=bwrap leaf: bubblewrap-namespace confinement.
package bwrap

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	isoshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agentcmd"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/envutil"
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
type Isolator struct{}

// NewIsolator creates a bwrap isolator.
func NewIsolator() *Isolator { return &Isolator{} }

// Available reports whether bwrap is installed.
func (i *Isolator) Available() (bool, error) {
	_, err := exec.LookPath("bwrap")
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
	return isoshared.SpawnSandbox("bwrap", args, env, "Bubblewrap sandbox", cfg.Verbose)
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
	homeDir, err := isoshared.ResolveHomeDir(cfg.HomeDir)
	if err != nil {
		return nil, err
	}
	workDir, err := isoshared.ResolveDirectory(cfg.WorkDir, "work directory")
	if err != nil {
		return nil, err
	}
	repoDir := ""
	if cfg.RepoDir != "" {
		repoDir, err = isoshared.ResolveDirectory(cfg.RepoDir, "repository directory")
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
	inherited := envutil.ParseEnv(os.Environ())
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
	netArgs, err := networkArgs(cfg.Network)
	if err != nil {
		return nil, err
	}
	args = append(args, netArgs...)

	// Sandbox-local HOME: a tmpfs at the home path (no host-$HOME bind), with only
	// the curated config paths bound back read-only.
	args = append(args, "--tmpfs", homeDir)
	for _, configPath := range isolation.UserConfigPaths {
		src := filepath.Join(homeDir, configPath)
		exists, err := isoshared.OptionalPath(src, "user config path")
		if err != nil {
			return nil, err
		}
		if exists {
			args = append(args, "--ro-bind", src, src)
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
			if _, err := os.Stat(dir); err == nil {
				args = append(args, "--ro-bind", dir, dir)
			}
		}
	}

	// Contribution: read-only binds of the provisioned closure's requisites.
	for _, p := range c.RoBindPaths {
		resolved, err := isoshared.ResolveExistingPath(p, "provisioned read-only bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "--ro-bind", resolved, resolved)
	}

	// Caller-supplied extra binds.
	for _, path := range cfg.BindPaths {
		resolved, err := isoshared.ResolveExistingPath(path, "read-write bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "--bind", resolved, resolved)
	}
	for _, path := range cfg.RoBindPaths {
		resolved, err := isoshared.ResolveExistingPath(path, "read-only bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "--ro-bind", resolved, resolved)
	}

	args = append(args, "--chdir", workDir)

	// Agent command, wrapped with the provisioner's init commands.
	args = append(args, "--")
	agentCmd := agentcmd.BuildAgentCommand(cfg.Agent, cfg.Provider, cfg.AgentArgs, "")
	args = append(args, isoshared.WrapWithInitCommands(agentCmd, c.InitCommands)...)

	return args, nil
}

func (i *Isolator) processEnv(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	homeDir, err := isoshared.ResolveHomeDir(cfg.HomeDir)
	if err != nil {
		return nil, err
	}
	sandboxEnv, err := i.sandboxEnv(cfg, c, homeDir)
	if err != nil {
		return nil, err
	}
	return envutil.EnvMapToSlice(envutil.MergeEnvMaps(envutil.ParseEnv(os.Environ()), sandboxEnv)), nil
}

// sandboxEnv builds the explicit environment allowlist for inside the sandbox:
// HOME/PATH/TMPDIR/SHELL, the contribution's env, the provider tokens, and the
// caller's custom env. API keys reach the sandbox only through this allowlist.
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
		if hostPath := os.Getenv("PATH"); hostPath != "" {
			pathEntries = append(pathEntries, hostPath)
		}
	}
	env["PATH"] = strings.Join(pathEntries, ":")

	// Provider tokens + contribution env + custom env.
	providerEnv, err := agentcmd.BuildProviderEnv(cfg.Agent, cfg.Provider)
	if err != nil {
		return nil, err
	}
	for k, v := range providerEnv {
		env[k] = v
	}
	for k, v := range c.Env {
		env[k] = v
	}
	customEnv, err := envutil.ParseCustomEnv(cfg.CustomEnv)
	if err != nil {
		return nil, err
	}
	for k, v := range customEnv {
		env[k] = v
	}
	return env, nil
}
