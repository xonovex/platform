package docker

import (
	"strings"
	"testing"

	isoshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func dockerCfg(net netshared.Mode, workDir string) isoshared.RunConfig {
	return isoshared.RunConfig{
		Agent:   &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"},
		HomeDir: "/home/testuser",
		WorkDir: workDir,
		Network: net,
	}
}

func argHas(args []string, s string) bool {
	for _, a := range args {
		if a == s {
			return true
		}
	}
	return false
}

func argHasPair(args []string, a, b string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == a && args[i+1] == b {
			return true
		}
	}
	return false
}

func envValue(args []string, key string) (string, bool) {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "-e" && strings.HasPrefix(args[i+1], key+"=") {
			return strings.TrimPrefix(args[i+1], key+"="), true
		}
	}
	return "", false
}

func TestDocker_SecurityDefaults(t *testing.T) {
	work := t.TempDir()
	args := NewIsolator().Command(dockerCfg(netshared.ModeNone, work), provision.Contribution{})

	if !argHas(args, "--read-only") {
		t.Error("missing --read-only rootfs")
	}
	if !argHasPair(args, "--cap-drop", "ALL") {
		t.Error("missing --cap-drop ALL")
	}
	if !argHasPair(args, "--security-opt", "no-new-privileges") {
		t.Error("missing no-new-privileges")
	}
	if !argHasPair(args, "--security-opt", "apparmor=docker-default") {
		t.Error("missing apparmor=docker-default")
	}
	if !argHasPair(args, "--tmpfs", "/tmp:rw,noexec,nosuid") {
		t.Error("missing hardened /tmp tmpfs")
	}
	if !argHasPair(args, "--pids-limit", dockerPidsLimit) {
		t.Error("missing --pids-limit")
	}
	if !argHasPair(args, "--memory", dockerMemory) {
		t.Error("missing --memory limit")
	}
	if !argHas(args, "--cpus") {
		t.Error("missing --cpus limit")
	}
	// Never disable seccomp.
	for _, a := range args {
		if strings.Contains(a, "seccomp=unconfined") {
			t.Error("must never set seccomp=unconfined")
		}
	}
	// No whole host-$HOME mount.
	if argHasPair(args, "-v", "/home/testuser:/home/testuser") {
		t.Error("must not mount the whole host $HOME")
	}
}

func TestDocker_NetworkExplicit(t *testing.T) {
	work := t.TempDir()
	if args := NewIsolator().Command(dockerCfg(netshared.ModeNone, work), provision.Contribution{}); !argHasPair(args, "--network", "none") {
		t.Error("network none must emit --network none")
	}
	if args := NewIsolator().Command(dockerCfg(netshared.ModeHost, work), provision.Contribution{}); !argHasPair(args, "--network", "host") {
		t.Error("network host must emit --network host")
	}
	if args := NewIsolator().Command(dockerCfg(netshared.ModeProxy, work), provision.Contribution{}); !argHasPair(args, "--network", "bridge") {
		t.Error("network proxy must keep a reachable bridge")
	}
}

func TestDocker_RuntimeWired(t *testing.T) {
	work := t.TempDir()
	cfg := dockerCfg(netshared.ModeNone, work)
	cfg.Runtime = "runsc"
	if args := NewIsolator().Command(cfg, provision.Contribution{}); !argHasPair(args, "--runtime", "runsc") {
		t.Error("RunConfig.Runtime must emit --runtime <runtime>")
	}
	// Default runc emits no --runtime flag.
	if args := NewIsolator().Command(dockerCfg(netshared.ModeNone, work), provision.Contribution{}); argHas(args, "--runtime") {
		t.Error("empty Runtime must not emit --runtime")
	}
}

func TestDocker_AppliesContribution(t *testing.T) {
	work := t.TempDir()
	closure := t.TempDir()
	c := provision.Contribution{
		RoBindPaths: []string{closure},
		PathEntries: []string{"/nix/store/abc/bin"},
		Env:         map[string]string{"FOO": "bar"},
	}
	args := NewIsolator().Command(dockerCfg(netshared.ModeNone, work), c)

	if !argHasPair(args, "-v", closure+":"+closure+":ro") {
		t.Error("contribution RoBindPaths must be mounted read-only")
	}
	if path, ok := envValue(args, "PATH"); !ok || !strings.HasPrefix(path, "/nix/store/abc/bin:") {
		t.Errorf("PATH = %q, want contribution entry prepended", path)
	}
	if v, ok := envValue(args, "FOO"); !ok || v != "bar" {
		t.Error("contribution env not applied")
	}
	if home, _ := envValue(args, "HOME"); home != containerHome {
		t.Errorf("HOME = %q, want synthetic %q", home, containerHome)
	}
}

func TestDocker_Capabilities(t *testing.T) {
	i := NewIsolator()
	if i.HidesHost(false, "") {
		t.Error("image-less docker resolves host-equivalent tools (does not hide host)")
	}
	if !i.HidesHost(false, "alpine:3.20") {
		t.Error("docker with a pinned image hides the host")
	}
	if !i.KernelIsolated("runsc") || !i.KernelIsolated("gvisor") {
		t.Error("docker + runsc/gvisor is a kernel boundary")
	}
	if i.KernelIsolated("") {
		t.Error("docker default runc is not a kernel boundary")
	}
}
