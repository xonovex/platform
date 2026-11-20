package bwrap

import (
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
)

// newHost returns an in-memory host with an empty home, a work directory and
// the host system directories that passthrough binds.
// registeredAgent returns the declared facts of an agent type, as the planner
// supplies them.
func registeredAgent(agentType types.AgentType) *types.AgentConfig {
	agent, err := agents.GetAgent(agentType)
	if err != nil {
		panic(err)
	}
	return agent
}

func newHost() *oshosttest.Fake {
	host := oshosttest.New().AddDir("/home/user").AddDir("/work").AddDir("/usr").AddDir("/bin").AddDir("/etc")
	host.Env["PATH"] = "/usr/bin"
	return host
}

// claudeCfg builds a run config rooted in the empty home of newHost, so the
// arguments depend on the config paths a test creates rather than on the host
// user's configuration.
func claudeCfg(net network.NetworkMethod, passthrough bool) isoshared.RunConfig {
	return isoshared.RunConfig{
		Agent:           registeredAgent(types.AgentClaude),
		Process:         isoshared.Process{Argv: []string{"claude", "review"}, Dir: "/work"},
		HomeDir:         "/home/user",
		Network:         net,
		HostPassthrough: passthrough,
	}
}

func bwrapCommand(t *testing.T, host *oshosttest.Fake, cfg isoshared.RunConfig, c provision.Contribution) []string {
	t.Helper()
	command, err := NewIsolator(host, testLog).Command(cfg, c)
	if err != nil {
		t.Fatalf("Command() error = %v", err)
	}
	return command
}

// argHas reports whether args contains s.
func argHas(args []string, s string) bool {
	for _, a := range args {
		if a == s {
			return true
		}
	}
	return false
}

// argHasPair reports whether a is immediately followed by b.
func argHasPair(args []string, a, b string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == a && args[i+1] == b {
			return true
		}
	}
	return false
}

// argHasTriple reports whether a,b,c appear consecutively.
func argHasTriple(args []string, a, b, c string) bool {
	for i := 0; i+2 < len(args); i++ {
		if args[i] == a && args[i+1] == b && args[i+2] == c {
			return true
		}
	}
	return false
}

func TestBwrap_DenyDefaultHardening(t *testing.T) {
	host := newHost()
	host.Env["BWRAP_UNRELATED_HOST_VALUE"] = "must-not-enter-sandbox"
	args := bwrapCommand(t, host, claudeCfg(network.NetworkNone, false), provision.Contribution{})

	if !argHasPair(args, "--dev", "/dev") {
		t.Error("missing minimal --dev /dev")
	}
	if argHas(args, "--dev-bind") {
		t.Error("must not use --dev-bind (exposes /dev/sda, /dev/mem)")
	}
	if !argHasPair(args, "--unsetenv", "BWRAP_UNRELATED_HOST_VALUE") {
		t.Error("unrelated host environment variable must be explicitly unset")
	}
	if argHasPair(args, "--unsetenv", "PATH") {
		t.Error("an allowlisted variable must not be unset")
	}
	if !argHasPair(args, "--tmpfs", "/home/user") {
		t.Error("HOME must be a sandbox-local tmpfs")
	}
	// No host-$HOME bind in deny-default mode.
	if argHasTriple(args, "--bind", "/home/user", "/home/user") {
		t.Error("must not bind host-$HOME in deny-default mode")
	}
	// No host system dirs bound.
	if argHasPair(args, "--ro-bind", "/usr") {
		t.Error("deny-default must not ro-bind host /usr")
	}
	if !argHasTriple(args, "--bind", "/work", "/work") || !argHasPair(args, "--chdir", "/work") {
		t.Error("the work directory must be bound read-write and be the working directory")
	}
}

// The sandbox-local HOME is a tmpfs, so the curated config paths are the only
// home content bound back into it, each read-only at its original path.
func TestBwrap_BindsExistingUserConfigPathsReadOnly(t *testing.T) {
	host := newHost().AddDir("/home/user/.claude").AddFile("/home/user/.claude.json", "{}\n", 0o600)

	args := bwrapCommand(t, host, claudeCfg(network.NetworkNone, false), provision.Contribution{})

	for _, configPath := range []string{"/home/user/.claude", "/home/user/.claude.json"} {
		if !argHasTriple(args, "--ro-bind", configPath, configPath) {
			t.Errorf("missing read-only user config bind for %q", configPath)
		}
	}
}

// The isolator names every mounted path by its canonical form.
func TestBwrap_BindsCanonicalDirectories(t *testing.T) {
	host := newHost().AddDir("/data/home").AddLink("/home/link", "/data/home").AddDir("/data/home/.claude")
	cfg := claudeCfg(network.NetworkNone, false)
	cfg.HomeDir = "/home/link"

	args := bwrapCommand(t, host, cfg, provision.Contribution{})

	if !argHasPair(args, "--tmpfs", "/data/home") || !argHasTriple(args, "--ro-bind", "/data/home/.claude", "/data/home/.claude") {
		t.Errorf("args = %v, want the canonical home /data/home", args)
	}
}

func TestBwrap_RejectsUserConfigEscapingHome(t *testing.T) {
	host := newHost().AddDir("/outside").AddLink("/home/user/.claude", "/outside")

	if _, err := NewIsolator(host, testLog).Command(claudeCfg(network.NetworkNone, false), provision.Contribution{}); err == nil || !strings.Contains(err.Error(), "outside home") {
		t.Fatalf("Command() error = %v, want containment error", err)
	}
}

// An absent config path contributes no bind, so a home without agent
// configuration leaves the sandbox HOME an empty tmpfs.
func TestBwrap_SkipsAbsentUserConfigPaths(t *testing.T) {
	args := bwrapCommand(t, newHost(), claudeCfg(network.NetworkNone, false), provision.Contribution{})

	for _, configPath := range []string{"/home/user/.claude", "/home/user/.claude.json"} {
		if argHas(args, configPath) {
			t.Errorf("absent user config path %q must not be bound", configPath)
		}
	}
}

func TestBwrap_NetworkExplicit(t *testing.T) {
	host := newHost()
	args := bwrapCommand(t, host, claudeCfg(network.NetworkNone, false), provision.Contribution{})
	if !argHas(args, "--unshare-net") || argHas(args, "--share-net") {
		t.Error("network none must --unshare-net and not --share-net")
	}
	args = bwrapCommand(t, host, claudeCfg(network.NetworkHost, false), provision.Contribution{})
	if !argHas(args, "--share-net") || argHas(args, "--unshare-net") {
		t.Error("network host must --share-net and not --unshare-net")
	}
}

func TestBwrap_ManagedEgressHasNoNetworkAndABridge(t *testing.T) {
	host := newHost().AddDir("/run/egress").AddExecutable("/opt/agent-cli")
	cfg := claudeCfg(network.NetworkManaged, false)
	cfg.Egress = isoshared.EgressRoute{Transport: network.TransportUnixSocket, Socket: "/run/egress/proxy.sock", Bridge: "/opt/agent-cli"}
	args := bwrapCommand(t, host, cfg, provision.Contribution{})
	if !argHas(args, "--unshare-net") || argHas(args, "--share-net") {
		t.Error("managed must --unshare-net: the only route out is the proxy socket")
	}
	line := strings.Join(args, " ")
	for _, want := range []string{"--bind /run/egress /run/egress", "--ro-bind /opt/agent-cli /opt/agent-cli", "/opt/agent-cli egress-bridge --socket /run/egress/proxy.sock --listen 127.0.0.1:3128 --"} {
		if !strings.Contains(line, want) {
			t.Errorf("args lack %q:\n%s", want, line)
		}
	}
}

func TestBwrap_ManagedEgressNeedsItsRoute(t *testing.T) {
	cfg := claudeCfg(network.NetworkManaged, false)
	if _, err := NewIsolator(newHost(), testLog).Command(cfg, provision.Contribution{}); err == nil {
		t.Error("managed without a route must fail")
	}
	if NewIsolator(newHost(), testLog).EgressTransport() != network.TransportUnixSocket {
		t.Error("bwrap declares the unix-socket transport")
	}
}

func TestBwrap_AppliesContribution(t *testing.T) {
	host := newHost().AddDir("/nix/store/abc")
	c := provision.Contribution{
		RoBindPaths:  []string{"/nix/store/abc"},
		PathEntries:  []string{"/nix/store/abc/bin"},
		Env:          map[string]string{"FOO": "bar"},
		InitCommands: []string{"echo hi"},
	}
	isolator := NewIsolator(host, testLog)
	args := bwrapCommand(t, host, claudeCfg(network.NetworkNone, false), c)
	env, err := isolator.sandboxEnv(claudeCfg(network.NetworkNone, false), c, "/home/user")
	if err != nil {
		t.Fatalf("sandboxEnv() error = %v", err)
	}

	if !argHasTriple(args, "--ro-bind", "/nix/store/abc", "/nix/store/abc") {
		t.Error("contribution RoBindPaths must be ro-bound")
	}
	if path := env["PATH"]; path != "/nix/store/abc/bin:/usr/local/bin:/usr/bin:/bin" {
		t.Errorf("PATH = %q, want contribution entry prepended and no host PATH", path)
	}
	if env["FOO"] != "bar" {
		t.Error("contribution env not applied")
	}
	// Init command wraps the agent: the trailing command is sh -c '... echo hi ...'.
	if !argHas(args, "--") {
		t.Fatal("missing command separator")
	}
	tail := strings.Join(args, " ")
	if !strings.Contains(tail, "echo hi") || !strings.Contains(tail, "sh -c") {
		t.Errorf("init command not wrapped: %q", tail)
	}
}

func TestBwrap_HostPassthrough(t *testing.T) {
	host := newHost()
	isolator := NewIsolator(host, testLog)
	args := bwrapCommand(t, host, claudeCfg(network.NetworkHost, true), provision.Contribution{})
	for _, dir := range []string{"/usr", "/bin", "/etc"} {
		if !argHasTriple(args, "--ro-bind", dir, dir) {
			t.Errorf("HostPassthrough must ro-bind host %s", dir)
		}
	}
	if argHasPair(args, "--ro-bind", "/lib") {
		t.Error("an absent host directory must not be bound")
	}
	env, err := isolator.sandboxEnv(claudeCfg(network.NetworkHost, true), provision.Contribution{}, "/home/user")
	if err != nil || env["PATH"] != "/usr/local/bin:/usr/bin:/bin:/usr/bin" {
		t.Errorf("sandboxEnv() PATH = %q, %v, want the host PATH appended", env["PATH"], err)
	}
}

func TestBwrap_BindsCallerPathsAndTheSourceRepository(t *testing.T) {
	host := newHost().AddDir("/repo").AddDir("/data/rw").AddFile("/data/ro.txt", "x", 0o600)
	cfg := claudeCfg(network.NetworkNone, false)
	cfg.RepoDir = "/repo"
	cfg.BindPaths = []string{"/data/rw"}
	cfg.RoBindPaths = []string{"/data/ro.txt"}

	args := bwrapCommand(t, host, cfg, provision.Contribution{})

	if !argHasTriple(args, "--ro-bind", "/repo", "/repo") {
		t.Error("source repository must be ro-bound")
	}
	if !argHasTriple(args, "--bind", "/data/rw", "/data/rw") || !argHasTriple(args, "--ro-bind", "/data/ro.txt", "/data/ro.txt") {
		t.Errorf("args = %v, want the caller binds", args)
	}
}

func TestBwrap_CommandRejectsUnresolvablePaths(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*isoshared.RunConfig, *provision.Contribution)
	}{
		{"home", func(c *isoshared.RunConfig, _ *provision.Contribution) { c.HomeDir = "/missing" }},
		{"work", func(c *isoshared.RunConfig, _ *provision.Contribution) { c.Process.Dir = "/missing" }},
		{"repo", func(c *isoshared.RunConfig, _ *provision.Contribution) { c.RepoDir = "/missing" }},
		{"closure", func(_ *isoshared.RunConfig, p *provision.Contribution) {
			p.RoBindPaths = []string{"/nix/store/missing"}
		}},
		{"read-write bind", func(c *isoshared.RunConfig, _ *provision.Contribution) { c.BindPaths = []string{"/missing"} }},
		{"read-only bind", func(c *isoshared.RunConfig, _ *provision.Contribution) { c.RoBindPaths = []string{"/missing"} }},
		{"custom env", func(c *isoshared.RunConfig, _ *provision.Contribution) { c.Process.Env = []string{"MISSING_EQUALS"} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := claudeCfg(network.NetworkNone, false)
			contribution := provision.Contribution{}
			test.mutate(&cfg, &contribution)
			if _, err := NewIsolator(newHost(), testLog).Command(cfg, contribution); err == nil {
				t.Error("Command() error = nil, want an error")
			}
		})
	}
}

func TestBwrap_CommandKeepsProviderSecretOutOfArguments(t *testing.T) {
	const secret = "bwrap-secret-must-not-appear"
	host := newHost()
	cfg := claudeCfg(network.NetworkNone, false)
	cfg.ProviderEnv = map[string]string{"ANTHROPIC_AUTH_TOKEN": secret, "ANTHROPIC_BASE_URL": "http://proxy.example"}

	command := bwrapCommand(t, host, cfg, provision.Contribution{})

	if strings.Contains(strings.Join(command, " "), secret) {
		t.Fatal("bubblewrap command arguments contain the provider secret")
	}
	env, err := NewIsolator(host, testLog).processEnv(cfg, provision.Contribution{})
	if err != nil {
		t.Fatalf("processEnv() error = %v", err)
	}
	if got, _ := envValue(env, "ANTHROPIC_AUTH_TOKEN"); got != secret {
		t.Error("the provider secret must reach bwrap through its environment")
	}
}

func TestBwrap_RunSpawnsBwrapWithTheProcessEnvironment(t *testing.T) {
	host := newHost()
	isolator := NewIsolator(host, testLog)
	command := bwrapCommand(t, host, claudeCfg(network.NetworkNone, false), provision.Contribution{})
	host.Outcomes[oshosttest.CommandLine(command[0], command[1:]...)] = oshosttest.Outcome{ExitCode: 4}

	code, err := isolator.Run(claudeCfg(network.NetworkNone, false), provision.Contribution{})

	if err != nil || code != 4 {
		t.Fatalf("Run() = (%d, %v), want (4, nil)", code, err)
	}
	if len(host.Calls) != 1 || host.Calls[0].Dir != "" {
		t.Fatalf("calls = %+v, want one bwrap run", host.Calls)
	}
	if got, _ := envValue(host.Calls[0].Env, "HOME"); got != "/home/user" {
		t.Errorf("bwrap HOME = %q, want /home/user", got)
	}
}

func TestBwrap_RunReportsPreparationErrors(t *testing.T) {
	host := newHost()
	cfg := claudeCfg(network.NetworkNone, false)
	cfg.Process.Dir = "/missing"
	if code, err := NewIsolator(host, testLog).Run(cfg, provision.Contribution{}); err == nil || code != 1 {
		t.Errorf("Run(missing work) = (%d, %v), want (1, error)", code, err)
	}
	if len(host.Calls) != 0 {
		t.Errorf("calls = %+v, want no process", host.Calls)
	}
}

func TestBwrap_Capabilities(t *testing.T) {
	i := NewIsolator(newHost(), testLog)
	if !i.HidesHost(false, "") {
		t.Error("bwrap deny-default must hide host tools")
	}
	if i.HidesHost(true, "") {
		t.Error("bwrap with HostPassthrough must not hide host tools")
	}
	if i.KernelIsolated("runsc") {
		t.Error("bwrap is never a kernel boundary")
	}
	if !i.ConfinesNetwork() {
		t.Error("bwrap unshares the network namespace")
	}
	if i.RequiresImage() {
		t.Error("bwrap confines the host file system and takes no image")
	}
	if !i.AcceptsHostBinds() {
		t.Error("bwrap binds a host path into the sandbox")
	}
	if !i.Bounded() {
		t.Error("bwrap bounds the agent")
	}
}

func TestBwrap_CheckAgent(t *testing.T) {
	host := newHost()
	i := NewIsolator(host, testLog)

	if err := i.CheckAgent(claudeCfg(network.NetworkNone, false)); err == nil || !strings.Contains(err.Error(), "host passthrough is disabled") {
		t.Fatalf("CheckAgent(no passthrough) error = %v, want passthrough error", err)
	}
	if err := i.CheckAgent(claudeCfg(network.NetworkNone, true)); err == nil || !strings.Contains(err.Error(), "not available on passthrough PATH") {
		t.Fatalf("CheckAgent(passthrough, missing) error = %v, want PATH error", err)
	}
	host.AddExecutable("/usr/bin/claude")
	if err := i.CheckAgent(claudeCfg(network.NetworkNone, true)); err != nil {
		t.Fatalf("CheckAgent(passthrough, present) error = %v", err)
	}
}

func TestBwrap_CommandEndsWithTheProcess(t *testing.T) {
	c := claudeCfg(network.NetworkHost, false)
	c.Process.Argv = []string{"codex", "--sandbox", "danger-full-access", "go"}

	command := bwrapCommand(t, newHost(), c, provision.Contribution{})

	if tail := command[len(command)-len(c.Process.Argv)-1:]; !slices.Equal(tail, append([]string{"--"}, c.Process.Argv...)) {
		t.Errorf("command tail = %v, want -- and the argument vector of the process", tail)
	}
}

func TestStopCommandHasNothingToStop(t *testing.T) {
	if stop := NewIsolator(oshosttest.New(), testLog).StopCommand(isoshared.RunConfig{Name: "named"}); stop != nil {
		t.Errorf("StopCommand = %v, want nil: the end of the process ends the sandbox", stop)
	}
}
