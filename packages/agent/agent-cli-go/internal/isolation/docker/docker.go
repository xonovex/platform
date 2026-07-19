// Package docker is the isolation=docker leaf: container confinement with
// deny-default security flags.
package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	isoshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agentcmd"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
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
// is the kernel boundary (see KernelIsolated), wired via RunConfig.Runtime.
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
func (i *Isolator) HidesHost(_ bool, image string) bool {
	return isolation.IsDigestPinnedImage(resolveImage(image))
}

// PinnedProvision requires the container image itself to be digest-pinned. A
// separate provisioner must also be pinned unless the image is the only source
// of tools.
func (i *Isolator) PinnedProvision(method provision.ProvisionMethod, provisionerPinned bool, image string) bool {
	if !isolation.IsDigestPinnedImage(resolveImage(image)) {
		return false
	}
	return method == provision.ProvisionNone || provisionerPinned
}

// KernelIsolated reports whether the runtime gives a kernel boundary (gVisor
// runsc); default runc is attack-surface reduction, not a kernel boundary.
func (i *Isolator) KernelIsolated(runtime string) bool {
	return runtime == "runsc" || runtime == "gvisor"
}

// Run executes the agent in a docker container.
func (i *Isolator) Run(cfg isoshared.RunConfig, c provision.Contribution) (int, error) {
	args, err := i.buildArgs(cfg, c)
	if err != nil {
		return 1, err
	}
	return isoshared.SpawnSandbox("docker", args, os.Environ(), "Docker sandbox", cfg.Verbose)
}

// Command returns the full docker command (for display / terminal wrappers).
func (i *Isolator) Command(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	args, err := i.buildArgs(cfg, c)
	if err != nil {
		return nil, err
	}
	return append([]string{"docker"}, args...), nil
}

// TerminalCommand returns the docker command plus the host env to launch it; the
// container environment is baked into the command via -e, so the wrapper needs
// only the host env.
func (i *Isolator) TerminalCommand(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string, error) {
	command, err := i.Command(cfg, c)
	return command, os.Environ(), err
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

	args := []string{"run", "--rm", "-it"}

	// Kernel-isolating runtime (e.g. runsc/gVisor); empty means default runc.
	if cfg.Runtime != "" {
		args = append(args, "--runtime", cfg.Runtime)
	}

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

	// Network — applied EXPLICITLY via the network bridge.
	args = append(args, networkArgs(cfg.Network)...)

	// Run as the current user so written files are owned by the caller.
	args = append(args, "-u", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()))

	// Workspace (rw — the only writable host bind) + working directory.
	args = append(args, "-w", workDir, "-v", fmt.Sprintf("%s:%s", workDir, workDir))
	if repoDir != "" && repoDir != workDir {
		args = append(args, "-v", fmt.Sprintf("%s:%s:ro", repoDir, repoDir))
	}

	// Curated config paths, read-only, into the synthetic HOME.
	for _, configPath := range isolation.UserConfigPaths {
		src := filepath.Join(homeDir, configPath)
		exists, err := isoshared.OptionalPath(src, "user config path")
		if err != nil {
			return nil, err
		}
		if exists {
			args = append(args, "-v", fmt.Sprintf("%s:%s:ro", src, filepath.Join(containerHome, configPath)))
		}
	}

	// Contribution: read-only binds of the provisioned closure's requisites.
	for _, p := range c.RoBindPaths {
		resolved, err := isoshared.ResolveExistingPath(p, "provisioned read-only bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "-v", fmt.Sprintf("%s:%s:ro", resolved, resolved))
	}

	// Caller-supplied extra binds.
	for _, path := range cfg.BindPaths {
		resolved, err := isoshared.ResolveExistingPath(path, "read-write bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "-v", fmt.Sprintf("%s:%s", resolved, resolved))
	}
	for _, path := range cfg.RoBindPaths {
		resolved, err := isoshared.ResolveExistingPath(path, "read-only bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "-v", fmt.Sprintf("%s:%s:ro", resolved, resolved))
	}

	// Environment.
	containerEnv, err := i.containerEnv(cfg, c)
	if err != nil {
		return nil, err
	}
	for k, v := range containerEnv {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Image (pinned by the caller; falls back to the shared default).
	args = append(args, resolveImage(cfg.Image))

	// Agent command, wrapped with the provisioner's init commands.
	agentCmd := agentcmd.BuildAgentCommand(cfg.Agent, cfg.Provider, cfg.AgentArgs, "")
	args = append(args, isoshared.WrapWithInitCommands(agentCmd, c.InitCommands)...)

	return args, nil
}

func resolveImage(image string) string {
	if image == "" {
		return isolation.DefaultContainerImage
	}
	return image
}

// containerEnv builds the container environment: HOME/TMPDIR/SHELL/PATH, the
// contribution's env, the provider tokens, the caller's custom env, and the
// proxy egress env.
func (i *Isolator) containerEnv(cfg isoshared.RunConfig, c provision.Contribution) (map[string]string, error) {
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
	for k, v := range envutil.ParseCustomEnv(cfg.CustomEnv) {
		env[k] = v
	}
	for k, v := range cfg.ProxyEnv {
		env[k] = v
	}

	return env, nil
}
