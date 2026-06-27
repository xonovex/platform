package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandboxutil"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agentcmd"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/sandbox"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/envutil"
)

// containerHome is the synthetic, writable HOME inside the container. The host
// $HOME is never mounted; only the curated config paths are bound read-only here.
const containerHome = "/home/agent"

// DoS ceilings applied as defaults. They are deliberately generous (containment,
// not quotas); --cpus is clamped to the host CPU count since docker rejects a
// value above the available cores.
const (
	dockerPidsLimit = "4096"
	dockerMemory    = "8g"
)

// Isolator confines the agent in a container with deny-default security flags
// applied as DEFAULTS (not opt-in): --security-opt=no-new-privileges,
// --cap-drop ALL, the default seccomp + apparmor=docker-default profiles,
// --read-only rootfs + tmpfs /tmp + a writable workdir bind (the only writable
// host bind), and --pids-limit/--memory/--cpus. The curated config paths are
// mounted read-only into a synthetic HOME; the host $HOME is never mounted, so
// host tools are unreachable when a pinned image is used.
//
// Default runc is attack-surface reduction; a sandboxed runtime (runsc/gVisor)
// is the kernel boundary (see KernelIsolated), wired separately.
type Isolator struct{}

// NewIsolator creates a docker isolator.
func NewIsolator() *Isolator { return &Isolator{} }

// Available reports whether the docker daemon is reachable.
func (i *Isolator) Available() (bool, error) {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil, nil
}

// HidesHost reports whether host tools are unreachable. The host filesystem is
// never mounted, so the actual host is always hidden; an image-less container,
// however, resolves host-equivalent tools, so a pinned image is required.
func (i *Isolator) HidesHost(_ bool, image string) bool { return image != "" }

// KernelIsolated reports whether the runtime gives a kernel boundary (gVisor
// runsc); default runc is attack-surface reduction, not a kernel boundary.
func (i *Isolator) KernelIsolated(runtime string) bool {
	return runtime == "runsc" || runtime == "gvisor"
}

// Run executes the agent in a docker container.
func (i *Isolator) Run(cfg *types.SandboxConfig, c types.Contribution) (int, error) {
	return sandboxutil.SpawnSandbox("docker", i.buildArgs(cfg, c), os.Environ(), "Docker sandbox", cfg.Verbose)
}

// Command returns the full docker command (for display / terminal wrappers).
func (i *Isolator) Command(cfg *types.SandboxConfig, c types.Contribution) []string {
	return append([]string{"docker"}, i.buildArgs(cfg, c)...)
}

func (i *Isolator) buildArgs(cfg *types.SandboxConfig, c types.Contribution) []string {
	homeDir, _ := os.UserHomeDir()
	if cfg.HomeDir != "" {
		homeDir = cfg.HomeDir
	}

	args := []string{"run", "--rm", "-it"}

	// Security defaults (deny-default; never seccomp=unconfined).
	args = append(args,
		"--security-opt", "no-new-privileges",
		"--security-opt", "apparmor=docker-default",
		"--cap-drop", "ALL",
		"--read-only",
		"--tmpfs", "/tmp:rw,noexec,nosuid",
		// Writable synthetic HOME (mode 1777 so the mapped uid can write).
		"--tmpfs", containerHome+":rw,mode=1777",
		"--pids-limit", dockerPidsLimit,
		"--memory", dockerMemory,
		"--cpus", fmt.Sprintf("%d", runtime.NumCPU()),
	)

	// Network — applied EXPLICITLY.
	args = append(args, sandboxutil.DockerNetworkArgs(cfg.Network)...)

	// Run as the current user so written files are owned by the caller.
	args = append(args, "-u", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()))

	// Workspace (rw — the only writable host bind) + working directory.
	args = append(args, "-w", cfg.WorkDir, "-v", fmt.Sprintf("%s:%s", cfg.WorkDir, cfg.WorkDir))
	if cfg.RepoDir != "" && cfg.RepoDir != cfg.WorkDir {
		args = append(args, "-v", fmt.Sprintf("%s:%s:ro", cfg.RepoDir, cfg.RepoDir))
	}

	// Curated config paths, read-only, into the synthetic HOME.
	for _, configPath := range sandbox.UserConfigPaths {
		src := filepath.Join(homeDir, configPath)
		if _, err := os.Stat(src); err == nil {
			args = append(args, "-v", fmt.Sprintf("%s:%s:ro", src, filepath.Join(containerHome, configPath)))
		}
	}

	// Contribution: read-only binds of the provisioned closure's requisites.
	for _, p := range c.RoBindPaths {
		if _, err := os.Stat(p); err == nil {
			args = append(args, "-v", fmt.Sprintf("%s:%s:ro", p, p))
		}
	}

	// Caller-supplied extra binds.
	for _, path := range cfg.BindPaths {
		if abs, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(abs); err == nil {
				args = append(args, "-v", fmt.Sprintf("%s:%s", abs, abs))
			}
		}
	}
	for _, path := range cfg.RoBindPaths {
		if abs, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(abs); err == nil {
				args = append(args, "-v", fmt.Sprintf("%s:%s:ro", abs, abs))
			}
		}
	}

	// Environment.
	for k, v := range i.containerEnv(cfg, c) {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Image (pinned by the caller; falls back to the shared default).
	image := cfg.Image
	if image == "" {
		image = sandbox.DefaultContainerImage
	}
	args = append(args, image)

	// Agent command, wrapped with the provisioner's init commands.
	agentCmd := agentcmd.BuildAgentCommand(cfg, "")
	args = append(args, sandboxutil.WrapWithInitCommands(agentCmd, c.InitCommands)...)

	return args
}

// containerEnv builds the container environment: HOME/TMPDIR/SHELL/PATH, the
// contribution's env, the provider tokens, the caller's custom env, and the
// proxy egress env.
func (i *Isolator) containerEnv(cfg *types.SandboxConfig, c types.Contribution) map[string]string {
	env := map[string]string{
		"HOME":   containerHome,
		"TMPDIR": "/tmp",
		"SHELL":  "/bin/bash",
	}

	// PATH: the contribution's tool dirs first, then the image's standard dirs
	// (the pinned image, not the host). HostPassthrough exposes the base-image
	// tools, which already live on these dirs.
	pathEntries := append([]string{}, c.PathEntries...)
	pathEntries = append(pathEntries, "/usr/local/bin", "/usr/bin", "/bin")
	env["PATH"] = strings.Join(pathEntries, ":")

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
