package bwrap

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandboxutil"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agentcmd"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/sandbox"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/envutil"
)

// Isolator confines the agent with bubblewrap namespaces and applies a
// Contribution via read-only binds plus an explicit setenv allowlist.
//
// Deny-default (HostPassthrough off): a sandbox-local HOME (tmpfs, no host-$HOME
// bind) with only the curated config paths bound back read-only; only the
// workspace (rw) and RepoDir (ro) are bound; /dev is a minimal devtmpfs (not a
// full --dev-bind, which would expose /dev/sda, /dev/mem and input devices); the
// environment is cleared (--clearenv) and rebuilt from an explicit allowlist.
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

// KernelIsolated reports false: bwrap is attack-surface reduction, not a kernel
// trust boundary.
func (i *Isolator) KernelIsolated(_ string) bool { return false }

// Run executes the agent in the bubblewrap sandbox. The sandbox environment is
// built entirely from --setenv (the parent env is cleared by --clearenv), so the
// parent process only needs its own environment to launch bwrap.
func (i *Isolator) Run(cfg *types.SandboxConfig, c types.Contribution) (int, error) {
	args := i.buildArgs(cfg, c)
	return sandboxutil.SpawnSandbox("bwrap", args, os.Environ(), "Bubblewrap sandbox", cfg.Verbose)
}

// Command returns the full bwrap command (for display / terminal wrappers).
func (i *Isolator) Command(cfg *types.SandboxConfig, c types.Contribution) []string {
	return append([]string{"bwrap"}, i.buildArgs(cfg, c)...)
}

func (i *Isolator) buildArgs(cfg *types.SandboxConfig, c types.Contribution) []string {
	homeDir, _ := os.UserHomeDir()
	if cfg.HomeDir != "" {
		homeDir = cfg.HomeDir
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
		// Deny-default environment: start empty, add back an explicit allowlist.
		"--clearenv",
	}

	// Network — applied EXPLICITLY (regression guard: --unshare-net for none/proxy).
	args = append(args, sandboxutil.BwrapNetworkArgs(cfg.Network)...)

	// Sandbox-local HOME: a tmpfs at the home path (no host-$HOME bind), with only
	// the curated config paths bound back read-only.
	args = append(args, "--tmpfs", homeDir)
	for _, configPath := range sandbox.UserConfigPaths {
		src := filepath.Join(homeDir, configPath)
		if _, err := os.Stat(src); err == nil {
			args = append(args, "--ro-bind", src, src)
		}
	}

	// Workspace (rw) + source repo (ro for worktrees).
	args = append(args, "--bind", cfg.WorkDir, cfg.WorkDir)
	if cfg.RepoDir != "" && cfg.RepoDir != cfg.WorkDir {
		args = append(args, "--ro-bind", cfg.RepoDir, cfg.RepoDir)
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
		if _, err := os.Stat(p); err == nil {
			args = append(args, "--ro-bind", p, p)
		}
	}

	// Caller-supplied extra binds.
	for _, path := range cfg.BindPaths {
		if abs, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(abs); err == nil {
				args = append(args, "--bind", abs, abs)
			}
		}
	}
	for _, path := range cfg.RoBindPaths {
		if abs, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(abs); err == nil {
				args = append(args, "--ro-bind", abs, abs)
			}
		}
	}

	args = append(args, "--chdir", cfg.WorkDir)

	// Explicit setenv allowlist (since --clearenv wiped everything).
	for k, v := range i.sandboxEnv(cfg, c, homeDir) {
		args = append(args, "--setenv", k, v)
	}

	// Agent command, wrapped with the provisioner's init commands.
	args = append(args, "--")
	agentCmd := agentcmd.BuildAgentCommand(cfg, "")
	args = append(args, sandboxutil.WrapWithInitCommands(agentCmd, c.InitCommands)...)

	return args
}

// sandboxEnv builds the explicit environment allowlist for inside the sandbox:
// HOME/PATH/TMPDIR/SHELL, the contribution's env, the provider tokens, the
// caller's custom env, and the proxy egress env. API keys reach the sandbox ONLY
// through this allowlist.
func (i *Isolator) sandboxEnv(cfg *types.SandboxConfig, c types.Contribution, homeDir string) map[string]string {
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

	// Provider tokens + contribution env + custom env + proxy env.
	if providerEnv, err := agentcmd.BuildProviderEnv(cfg); err == nil {
		for k, v := range providerEnv {
			env[k] = v
		}
	}
	for k, v := range c.Env {
		env[k] = v
	}
	for k, v := range envutil.ParseCustomEnv(cfg.CustomEnv) {
		env[k] = v
	}
	for k, v := range sandboxutil.ProxyEnv(cfg.Network, sandboxutil.ProxyURL()) {
		env[k] = v
	}

	return env
}
