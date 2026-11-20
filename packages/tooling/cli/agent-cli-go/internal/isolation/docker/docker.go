// Package docker is the isolation=docker leaf: container confinement with
// deny-default security flags.
package docker

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	isohost "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
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
type Isolator struct {
	host Host
	log  logging.Logger
}

// Host is the port the docker isolator consumes: the isolation host plus the
// identity the container runs as and the CPU count it is clamped to.
type Host interface {
	isohost.Host
	Getuid() int
	Getgid() int
	NumCPU() int
}

// NewIsolator creates a docker isolator over host, writing through log.
func NewIsolator(host Host, log logging.Logger) *Isolator {
	return &Isolator{host: host, log: log}
}

// Available reports whether the docker daemon is reachable.
func (i *Isolator) Available() (bool, error) {
	_, err := i.host.Capture("docker", []string{"info"}, "")
	return err == nil, nil
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

// ConfinesNetwork reports true: docker applies --network none.
func (i *Isolator) ConfinesNetwork() bool { return true }

// EgressTransport declares the unix socket: a managed run has --network none,
// and reaches the allowlist proxy through a bound socket.
func (i *Isolator) EgressTransport() network.EgressTransport { return network.TransportUnixSocket }

// Bounded reports true: the container confines what the agent writes and
// reaches.
func (i *Isolator) Bounded() bool { return true }

// CheckAgent rejects the default image, which does not contain an agent; a
// caller-supplied image is trusted to contain it.
func (i *Isolator) CheckAgent(cfg isoshared.RunConfig) error {
	if cfg.Image == "" || cfg.Image == isolation.DefaultContainerImage {
		return fmt.Errorf("the default docker image does not contain agent executable %q; use provision=nix or a digest-pinned --image that contains it", cfg.Agent.Binary)
	}
	return nil
}

// Run executes the agent in a docker container.
func (i *Isolator) Run(cfg isoshared.RunConfig, c provision.Contribution) (int, error) {
	args, err := i.buildArgs(cfg, c)
	if err != nil {
		return 1, err
	}
	env, err := i.processEnv(cfg, c)
	if err != nil {
		return 1, err
	}
	return isohost.SpawnSandbox(i.host, i.log, "docker", args, env, "", "Docker sandbox")
}

// Command returns the full docker command (for display / terminal wrappers).
func (i *Isolator) Command(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	args, err := i.buildArgs(cfg, c)
	if err != nil {
		return nil, err
	}
	return append([]string{"docker"}, args...), nil
}

// TerminalCommand returns the docker command plus the environment used to launch it.
func (i *Isolator) TerminalCommand(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string, error) {
	command, err := i.Command(cfg, c)
	if err != nil {
		return nil, nil, err
	}
	env, err := i.processEnv(cfg, c)
	return command, env, err
}

func (i *Isolator) buildArgs(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	homeDir, err := isohost.ResolveDirectory(i.host, cfg.HomeDir, "home directory")
	if err != nil {
		return nil, err
	}
	workDir, err := isohost.ResolveDirectory(i.host, cfg.Process.Dir, "work directory")
	if err != nil {
		return nil, err
	}
	repoDir := ""
	if cfg.RepoDir != "" {
		repoDir, err = isohost.ResolveDirectory(i.host, cfg.RepoDir, "repository directory")
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
		"--cpus", fmt.Sprintf("%d", i.host.NumCPU()),
	)

	// Network — applied EXPLICITLY via the network bridge.
	args = append(args, networkArgs(cfg.Network)...)

	// Run as the current user so written files are owned by the caller.
	args = append(args, "-u", fmt.Sprintf("%d:%d", i.host.Getuid(), i.host.Getgid()))

	// Workspace (rw — the only writable host bind) + working directory.
	args = append(args, "-w", workDir, "-v", fmt.Sprintf("%s:%s", workDir, workDir))
	if repoDir != "" && repoDir != workDir {
		args = append(args, "-v", fmt.Sprintf("%s:%s:ro", repoDir, repoDir))
	}

	// Curated config paths, read-only, into the synthetic HOME.
	for _, configPath := range cfg.Agent.UserConfigPaths {
		src, exists, err := isohost.ResolveContainedOptionalPath(i.host, homeDir, configPath, "user config path")
		if err != nil {
			return nil, err
		}
		if exists {
			args = append(args, "-v", fmt.Sprintf("%s:%s:ro", src, filepath.Join(containerHome, configPath)))
		}
	}

	// Contribution: read-only binds of the provisioned closure's requisites.
	for _, p := range c.RoBindPaths {
		resolved, err := isohost.ResolveExistingPath(i.host, p, "provisioned read-only bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "-v", fmt.Sprintf("%s:%s:ro", resolved, resolved))
	}

	// Caller-supplied extra binds.
	for _, path := range cfg.BindPaths {
		resolved, err := isohost.ResolveExistingPath(i.host, path, "read-write bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "-v", fmt.Sprintf("%s:%s", resolved, resolved))
	}
	for _, path := range cfg.RoBindPaths {
		resolved, err := isohost.ResolveExistingPath(i.host, path, "read-only bind")
		if err != nil {
			return nil, err
		}
		args = append(args, "-v", fmt.Sprintf("%s:%s:ro", resolved, resolved))
	}

	// Managed egress: bind the proxy socket directory and the bridge executable;
	// the command runs under the bridge below.
	var route isoshared.EgressRoute
	if cfg.Network == network.NetworkManaged {
		route, err = isohost.ResolveBridge(i.host, cfg.Egress)
		if err != nil {
			return nil, err
		}
		socketDir, bridge := isohost.BridgeMounts(route)
		args = append(args, "-v", fmt.Sprintf("%s:%s", socketDir, socketDir), "-v", fmt.Sprintf("%s:%s:ro", bridge, bridge))
	}

	// Environment.
	containerEnv, err := i.containerEnv(cfg, c)
	if err != nil {
		return nil, err
	}
	envNames := make([]string, 0, len(containerEnv))
	for key := range containerEnv {
		envNames = append(envNames, key)
	}
	sort.Strings(envNames)
	for _, key := range envNames {
		args = append(args, "-e", key)
	}

	// Image (pinned by the caller; falls back to the shared default).
	args = append(args, resolveImage(cfg.Image))

	// Agent process, wrapped with the provisioner's init commands.
	command := isohost.WrapWithInitCommands(cfg.Process.Argv, c.InitCommands)
	if cfg.Network == network.NetworkManaged {
		command = isohost.WrapWithBridge(route, command)
	}
	args = append(args, command...)

	return args, nil
}

func (i *Isolator) processEnv(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	containerEnv, err := i.containerEnv(cfg, c)
	if err != nil {
		return nil, err
	}
	return envutil.EnvMapToSlice(envutil.MergeEnvMaps(envutil.ParseEnv(i.host.Environ()), containerEnv)), nil
}

func resolveImage(image string) string {
	if image == "" {
		return isolation.DefaultContainerImage
	}
	return image
}

// containerEnv builds the container environment: HOME/TMPDIR/SHELL/PATH, the
// provider tokens, the contribution's env, and the environment of the process.
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

	for k, v := range cfg.ProviderEnv {
		env[k] = v
	}
	for k, v := range c.Env {
		env[k] = v
	}
	customEnv, err := isohost.RunEnv(cfg)
	if err != nil {
		return nil, err
	}
	for k, v := range customEnv {
		env[k] = v
	}
	return env, nil
}
