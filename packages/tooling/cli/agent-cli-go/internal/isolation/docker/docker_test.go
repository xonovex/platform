package docker

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

// newHost returns an in-memory host with an empty home and a work directory,
// uid and gid 1000 and 4 CPUs.
func newHost() *systemtest.Fake {
	return systemtest.New().AddDir("/home/user").AddDir("/work")
}

// dockerCfg builds a run config rooted in the empty home of newHost, so the
// arguments depend on the config paths a test creates rather than on the host
// user's configuration.
func dockerCfg(net netshared.Mode) isoshared.RunConfig {
	return isoshared.RunConfig{
		Agent:   &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"},
		WorkDir: "/work",
		HomeDir: "/home/user",
		Network: net,
	}
}

func dockerCommand(t *testing.T, host *systemtest.Fake, cfg isoshared.RunConfig, c provision.Contribution) []string {
	t.Helper()
	command, err := NewIsolator(host, testLog).Command(cfg, c)
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
	args := dockerCommand(t, newHost(), dockerCfg(netshared.ModeNone), provision.Contribution{})

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
	if !argHasPair(args, "--cpus", "4") {
		t.Error("missing --cpus limit of the host CPU count")
	}
	if !argHasPair(args, "-u", "1000:1000") {
		t.Error("the container must run as the host user")
	}
	// Never disable seccomp.
	for _, a := range args {
		if strings.Contains(a, "seccomp=unconfined") {
			t.Error("must never set seccomp=unconfined")
		}
	}
	// Only the curated config paths enter the container, never the home as a whole.
	if argHasPair(args, "-v", "/home/user:/home/user") {
		t.Error("must not mount the whole host home directory")
	}
}

// The curated user config paths are the only home content the container sees,
// and each is mounted read-only into the synthetic HOME.
func TestDocker_MountsExistingUserConfigPathsReadOnly(t *testing.T) {
	host := newHost().AddDir("/home/user/.claude").AddFile("/home/user/.claude.json", "{}\n", 0o600)

	args := dockerCommand(t, host, dockerCfg(netshared.ModeNone), provision.Contribution{})

	for _, configPath := range []string{".claude", ".claude.json"} {
		bind := filepath.Join("/home/user", configPath) + ":" + filepath.Join(containerHome, configPath) + ":ro"
		if !argHasPair(args, "-v", bind) {
			t.Errorf("missing read-only user config bind %q", bind)
		}
	}
}

// An absent config path contributes no bind, so a home without agent
// configuration produces a container that sees none of it.
func TestDocker_SkipsAbsentUserConfigPaths(t *testing.T) {
	args := dockerCommand(t, newHost(), dockerCfg(netshared.ModeNone), provision.Contribution{})

	for _, configPath := range []string{".claude", ".claude.json"} {
		if strings.Contains(strings.Join(args, " "), filepath.Join(containerHome, configPath)) {
			t.Errorf("absent user config path %q must not be mounted", configPath)
		}
	}
}

func TestDocker_NetworkExplicit(t *testing.T) {
	host := newHost()
	if args := dockerCommand(t, host, dockerCfg(netshared.ModeNone), provision.Contribution{}); !argHasPair(args, "--network", "none") {
		t.Error("network none must emit --network none")
	}
	if args := dockerCommand(t, host, dockerCfg(netshared.ModeHost), provision.Contribution{}); !argHasPair(args, "--network", "host") {
		t.Error("network host must emit --network host")
	}
	if _, err := NewIsolator(host, testLog).Command(dockerCfg(netshared.ModeProxy), provision.Contribution{}); !errors.Is(err, netshared.ErrProxyEnforcementUnavailable) {
		t.Errorf("network proxy error = %v, want %v", err, netshared.ErrProxyEnforcementUnavailable)
	}
}

func TestDocker_RuntimeWired(t *testing.T) {
	host := newHost()
	cfg := dockerCfg(netshared.ModeNone)
	cfg.Runtime = "runsc"
	if args := dockerCommand(t, host, cfg, provision.Contribution{}); !argHasPair(args, "--runtime", "runsc") {
		t.Error("RunConfig.Runtime must emit --runtime <runtime>")
	}
	// Default runc emits no --runtime flag.
	if args := dockerCommand(t, host, dockerCfg(netshared.ModeNone), provision.Contribution{}); argHas(args, "--runtime") {
		t.Error("empty Runtime must not emit --runtime")
	}
}

func TestDocker_AppliesContribution(t *testing.T) {
	host := newHost().AddDir("/nix/store/abc")
	closure := "/nix/store/abc"
	c := provision.Contribution{
		RoBindPaths: []string{closure},
		PathEntries: []string{"/nix/store/abc/bin"},
		Env:         map[string]string{"FOO": "bar"},
	}
	args := dockerCommand(t, host, dockerCfg(netshared.ModeNone), c)
	env, err := NewIsolator(host, testLog).containerEnv(dockerCfg(netshared.ModeNone), c)
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
	i := NewIsolator(newHost(), testLog)
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
	if !i.ConfinesNetwork() {
		t.Error("docker applies --network none")
	}
}

func TestDocker_CommandRejectsMissingBindAndProviderToken(t *testing.T) {
	host := newHost()
	cfg := dockerCfg(netshared.ModeNone)
	cfg.BindPaths = []string{"/work/missing"}
	if _, err := NewIsolator(host, testLog).Command(cfg, provision.Contribution{}); err == nil {
		t.Error("Command() error = nil, want missing bind error")
	}

	cfg.BindPaths = nil
	cfg.Provider = &types.ModelProvider{
		CredentialSourceEnv: "MISSING_DOCKER_TEST_TOKEN",
		CredentialTargetEnv: "AGENT_PROVIDER_TOKEN",
	}
	if _, err := NewIsolator(host, testLog).Command(cfg, provision.Contribution{}); err == nil {
		t.Error("Command() error = nil, want missing provider token error")
	}
}

func TestDocker_CommandKeepsProviderSecretOutOfArguments(t *testing.T) {
	const secret = "docker-secret-must-not-appear"
	host := newHost()
	host.Env["DOCKER_TEST_TOKEN"] = secret
	cfg := dockerCfg(netshared.ModeNone)
	cfg.Provider = &types.ModelProvider{
		CredentialSourceEnv: "DOCKER_TEST_TOKEN",
		CredentialTargetEnv: "ANTHROPIC_AUTH_TOKEN",
		Environment:         map[string]string{"ANTHROPIC_BASE_URL": "http://proxy.example"},
	}

	command := dockerCommand(t, host, cfg, provision.Contribution{})

	if strings.Contains(strings.Join(command, " "), secret) {
		t.Fatal("docker command arguments contain the provider secret")
	}
	env, err := NewIsolator(host, testLog).processEnv(cfg, provision.Contribution{})
	if err != nil {
		t.Fatalf("processEnv() error = %v", err)
	}
	if got, _ := envValue(env, "ANTHROPIC_AUTH_TOKEN"); got != secret {
		t.Error("the provider secret must reach docker through its environment")
	}
}

func TestDocker_CommandRejectsUnknownAgentType(t *testing.T) {
	cfg := dockerCfg(netshared.ModeNone)
	cfg.Agent = &types.AgentConfig{Type: "unknown", Binary: "unknown"}

	if _, err := NewIsolator(newHost(), testLog).Command(cfg, provision.Contribution{}); err == nil {
		t.Fatal("Command() error = nil, want unknown agent type error")
	}
}

func TestDocker_MountsTheSourceRepositoryAndCallerPaths(t *testing.T) {
	host := newHost().AddDir("/repo").AddDir("/data/rw").AddFile("/data/ro.txt", "x", 0o600)
	cfg := dockerCfg(netshared.ModeNone)
	cfg.RepoDir = "/repo"
	cfg.BindPaths = []string{"/data/rw"}
	cfg.RoBindPaths = []string{"/data/ro.txt"}
	cfg.Image = digestPinnedImage
	cfg.AgentArgs = []string{"review"}

	args := dockerCommand(t, host, cfg, provision.Contribution{InitCommands: []string{"echo hi"}})

	for _, bind := range []string{"/work:/work", "/repo:/repo:ro", "/data/rw:/data/rw", "/data/ro.txt:/data/ro.txt:ro"} {
		if !argHasPair(args, "-v", bind) {
			t.Errorf("args = %v, want -v %s", args, bind)
		}
	}
	if !argHasPair(args, "-w", "/work") {
		t.Error("the work directory must be the container working directory")
	}
	if !argHas(args, digestPinnedImage) || !strings.Contains(strings.Join(args, " "), "echo hi") {
		t.Errorf("args = %v, want the image and the wrapped agent command", args)
	}
}

func TestDocker_UsesTheDefaultImageWithoutAnImage(t *testing.T) {
	if args := dockerCommand(t, newHost(), dockerCfg(netshared.ModeNone), provision.Contribution{}); !argHas(args, isolation.DefaultContainerImage) {
		t.Errorf("args = %v, want the default image", args)
	}
}

func TestDocker_CommandRejectsUnresolvablePaths(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*isoshared.RunConfig, *provision.Contribution)
	}{
		{"home", func(c *isoshared.RunConfig, _ *provision.Contribution) { c.HomeDir = "/missing" }},
		{"work", func(c *isoshared.RunConfig, _ *provision.Contribution) { c.WorkDir = "/missing" }},
		{"repo", func(c *isoshared.RunConfig, _ *provision.Contribution) { c.RepoDir = "/missing" }},
		{"closure", func(_ *isoshared.RunConfig, p *provision.Contribution) {
			p.RoBindPaths = []string{"/nix/store/missing"}
		}},
		{"read-only bind", func(c *isoshared.RunConfig, _ *provision.Contribution) { c.RoBindPaths = []string{"/missing"} }},
		{"user config escape", func(_ *isoshared.RunConfig, _ *provision.Contribution) {}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			host := newHost()
			if test.name == "user config escape" {
				host.AddDir("/outside").AddLink("/home/user/.claude", "/outside")
			}
			cfg := dockerCfg(netshared.ModeNone)
			contribution := provision.Contribution{}
			test.mutate(&cfg, &contribution)
			if _, err := NewIsolator(host, testLog).Command(cfg, contribution); err == nil {
				t.Error("Command() error = nil, want an error")
			}
		})
	}
}

func TestDocker_RunSpawnsDockerWithTheProcessEnvironment(t *testing.T) {
	host := newHost()
	host.Env["HOST_VALUE"] = "kept"
	command := dockerCommand(t, host, dockerCfg(netshared.ModeNone), provision.Contribution{})
	host.Outcomes[systemtest.CommandLine(command[0], command[1:]...)] = systemtest.Outcome{ExitCode: 5}

	code, err := NewIsolator(host, testLog).Run(dockerCfg(netshared.ModeNone), provision.Contribution{})

	if err != nil || code != 5 {
		t.Fatalf("Run() = (%d, %v), want (5, nil)", code, err)
	}
	if len(host.Calls) != 1 {
		t.Fatalf("calls = %+v, want one docker run", host.Calls)
	}
	if got, _ := envValue(host.Calls[0].Env, "HOST_VALUE"); got != "kept" {
		t.Error("the docker client must inherit the host environment")
	}
}

func TestDocker_RunReportsPreparationErrors(t *testing.T) {
	for name, mutate := range map[string]func(*isoshared.RunConfig){
		"arguments":   func(c *isoshared.RunConfig) { c.WorkDir = "/missing" },
		"environment": func(c *isoshared.RunConfig) { c.CustomEnv = []string{"MISSING_EQUALS"} },
	} {
		t.Run(name, func(t *testing.T) {
			host := newHost()
			cfg := dockerCfg(netshared.ModeNone)
			mutate(&cfg)
			if code, err := NewIsolator(host, testLog).Run(cfg, provision.Contribution{}); err == nil || code != 1 {
				t.Errorf("Run() = (%d, %v), want (1, error)", code, err)
			}
			if len(host.Calls) != 0 {
				t.Errorf("calls = %+v, want no process", host.Calls)
			}
		})
	}
}

func TestDocker_AvailableTracksTheDaemon(t *testing.T) {
	host := newHost()
	isolator := NewIsolator(host, testLog)

	if available, err := isolator.Available(); available || err != nil {
		t.Errorf("Available() without a daemon = (%v, %v), want (false, nil)", available, err)
	}
	host.Outcomes["docker info"] = systemtest.Outcome{Stdout: "Server Version: 27"}
	if available, err := isolator.Available(); !available || err != nil {
		t.Errorf("Available() with a daemon = (%v, %v), want (true, nil)", available, err)
	}
}

func TestDocker_CheckAgent(t *testing.T) {
	isolator := NewIsolator(newHost(), testLog)
	for _, image := range []string{"", isolation.DefaultContainerImage} {
		cfg := dockerCfg(netshared.ModeNone)
		cfg.Image = image
		if err := isolator.CheckAgent(cfg); err == nil || !strings.Contains(err.Error(), "default docker image") {
			t.Errorf("CheckAgent(image=%q) error = %v, want default-image error", image, err)
		}
	}
	cfg := dockerCfg(netshared.ModeNone)
	cfg.Image = digestPinnedImage
	if err := isolator.CheckAgent(cfg); err != nil {
		t.Errorf("CheckAgent(custom image) error = %v, want nil", err)
	}
}
