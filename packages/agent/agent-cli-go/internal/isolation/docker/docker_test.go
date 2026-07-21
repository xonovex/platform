package docker

import (
	"errors"
	"os"
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
		WorkDir: workDir,
		Network: net,
	}
}

func dockerCommand(t *testing.T, cfg isoshared.RunConfig, c provision.Contribution) []string {
	t.Helper()
	command, err := NewIsolator().Command(cfg, c)
	if err != nil {
		t.Fatalf("Command() error = %v", err)
	}
	return command
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

func TestDocker_SecurityDefaults(t *testing.T) {
	work := t.TempDir()
	args := dockerCommand(t, dockerCfg(netshared.ModeNone, work), provision.Contribution{})

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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}
	if argHasPair(args, "-v", homeDir+":"+homeDir) {
		t.Error("must not mount the whole host $HOME")
	}
}

func TestDocker_NetworkExplicit(t *testing.T) {
	work := t.TempDir()
	if args := dockerCommand(t, dockerCfg(netshared.ModeNone, work), provision.Contribution{}); !argHasPair(args, "--network", "none") {
		t.Error("network none must emit --network none")
	}
	if args := dockerCommand(t, dockerCfg(netshared.ModeHost, work), provision.Contribution{}); !argHasPair(args, "--network", "host") {
		t.Error("network host must emit --network host")
	}
	if _, err := NewIsolator().Command(dockerCfg(netshared.ModeProxy, work), provision.Contribution{}); !errors.Is(err, netshared.ErrProxyEnforcementUnavailable) {
		t.Errorf("network proxy error = %v, want %v", err, netshared.ErrProxyEnforcementUnavailable)
	}
}

func TestDocker_RuntimeWired(t *testing.T) {
	work := t.TempDir()
	cfg := dockerCfg(netshared.ModeNone, work)
	cfg.Runtime = "runsc"
	if args := dockerCommand(t, cfg, provision.Contribution{}); !argHasPair(args, "--runtime", "runsc") {
		t.Error("RunConfig.Runtime must emit --runtime <runtime>")
	}
	// Default runc emits no --runtime flag.
	if args := dockerCommand(t, dockerCfg(netshared.ModeNone, work), provision.Contribution{}); argHas(args, "--runtime") {
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
	args := dockerCommand(t, dockerCfg(netshared.ModeNone, work), c)
	env, err := NewIsolator().containerEnv(dockerCfg(netshared.ModeNone, work), c)
	if err != nil {
		t.Fatalf("containerEnv() error = %v", err)
	}

	if !argHasPair(args, "-v", closure+":"+closure+":ro") {
		t.Error("contribution RoBindPaths must be mounted read-only")
	}
	if path := env["PATH"]; !strings.HasPrefix(path, "/nix/store/abc/bin:") {
		t.Errorf("PATH = %q, want contribution entry prepended", path)
	}
	if env["FOO"] != "bar" {
		t.Error("contribution env not applied")
	}
	if home := env["HOME"]; home != containerHome {
		t.Errorf("HOME = %q, want synthetic %q", home, containerHome)
	}
	for _, key := range []string{"PATH", "FOO", "HOME"} {
		if !argHasPair(args, "-e", key) {
			t.Errorf("docker command missing inherited environment name %q", key)
		}
	}
}

func TestDocker_Capabilities(t *testing.T) {
	i := NewIsolator()
	if !i.HidesHost(false, "") {
		t.Error("docker with the digest-pinned default image hides host-equivalent tools")
	}
	digest := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if i.HidesHost(false, "alpine:3.20") {
		t.Error("docker with a mutable tag must not claim to hide host-equivalent image tools")
	}
	if !i.HidesHost(false, "alpine@sha256:"+digest) {
		t.Error("docker with a digest-pinned image hides host-equivalent tools")
	}
	if !i.KernelIsolated("runsc") || !i.KernelIsolated("gvisor") {
		t.Error("docker + runsc/gvisor is a kernel boundary")
	}
	if i.KernelIsolated("") {
		t.Error("docker default runc is not a kernel boundary")
	}
}

func TestDocker_CommandRejectsMissingBindAndProviderToken(t *testing.T) {
	work := t.TempDir()
	cfg := dockerCfg(netshared.ModeNone, work)
	cfg.BindPaths = []string{work + "/missing"}
	if _, err := NewIsolator().Command(cfg, provision.Contribution{}); err == nil {
		t.Error("Command() error = nil, want missing bind error")
	}

	cfg.BindPaths = nil
	cfg.Provider = &types.ModelProvider{AuthTokenEnv: "MISSING_DOCKER_TEST_TOKEN"}
	if _, err := NewIsolator().Command(cfg, provision.Contribution{}); err == nil {
		t.Error("Command() error = nil, want missing provider token error")
	}
}

func TestDocker_CommandKeepsProviderSecretOutOfArguments(t *testing.T) {
	const secret = "docker-secret-must-not-appear"
	t.Setenv("DOCKER_TEST_TOKEN", secret)
	cfg := dockerCfg(netshared.ModeNone, t.TempDir())
	cfg.Provider = &types.ModelProvider{
		AuthTokenEnv: "DOCKER_TEST_TOKEN",
		Environment:  map[string]string{"ANTHROPIC_BASE_URL": "http://proxy.example"},
	}

	command := dockerCommand(t, cfg, provision.Contribution{})

	if strings.Contains(strings.Join(command, " "), secret) {
		t.Fatal("docker command arguments contain the provider secret")
	}
}
