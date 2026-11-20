package seatbelt

import (
	"io"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

var testLog = logging.New(io.Discard, io.Discard, false, false)

func newHost() *systemtest.Fake {
	host := systemtest.New().AddDir("/Users/dev").AddDir("/Users/dev/work")
	host.Env["PATH"] = "/usr/bin"
	return host.AddExecutable("/usr/bin/claude").AddExecutable("/usr/bin/sandbox-exec")
}

func claudeCfg(mode network.NetworkMethod) isoshared.RunConfig {
	agent, err := agents.GetAgent(types.AgentClaude)
	if err != nil {
		panic(err)
	}
	return isoshared.RunConfig{Agent: agent, Process: isoshared.Process{Argv: []string{"claude", "review"}, Dir: "/Users/dev/work"}, HomeDir: "/Users/dev", Network: mode}
}

func profileOf(t *testing.T, host *systemtest.Fake, cfg isoshared.RunConfig, c provision.Contribution) string {
	t.Helper()
	command, err := NewIsolator(host, testLog).Command(cfg, c)
	if err != nil {
		t.Fatalf("Command() error = %v", err)
	}
	if command[0] != "sandbox-exec" || command[1] != "-p" {
		t.Fatalf("command = %v, want sandbox-exec -p <profile> ...", command)
	}
	return command[2]
}

func TestCommandDeniesWritesAndHidesHome(t *testing.T) {
	profile := profileOf(t, newHost(), claudeCfg(network.NetworkHost), provision.Contribution{})
	for _, want := range []string{
		"(version 1)\n(allow default)\n(deny file-write*)",
		`(subpath "/Users/dev/work")`,
		`(literal "/dev/null")`,
		`(deny file-read-data (subpath "/Users/dev"))`,
		`(allow file-read-data (subpath "/Users/dev/work"))`,
	} {
		if !strings.Contains(profile, want) {
			t.Errorf("profile lacks %q:\n%s", want, profile)
		}
	}
	if strings.Contains(profile, "network") {
		t.Errorf("a host run must not restrict the network:\n%s", profile)
	}
	if strings.Index(profile, "(deny file-read-data") > strings.Index(profile, "(allow file-read-data") {
		t.Error("the allow rule must follow the deny rule: the last matching rule wins")
	}
}

func TestCommandRunsTheAgentAfterTheProfile(t *testing.T) {
	command, err := NewIsolator(newHost(), testLog).Command(claudeCfg(network.NetworkHost), provision.Contribution{})
	if err != nil {
		t.Fatal(err)
	}
	if command[3] != "claude" {
		t.Errorf("command = %v, want the agent binary after the profile", command)
	}
}

func TestCommandAddsConfigBindsAndRepo(t *testing.T) {
	host := newHost().AddDir("/Users/dev/.claude").AddFile("/Users/dev/.claude.json", "{}", 0o600).AddDir("/Users/dev/repo").AddDir("/Users/dev/extra").AddDir("/opt/tools")
	cfg := claudeCfg(network.NetworkHost)
	cfg.RepoDir = "/Users/dev/repo"
	cfg.BindPaths = []string{"/Users/dev/extra"}
	cfg.RoBindPaths = []string{"/opt/tools"}
	profile := profileOf(t, host, cfg, provision.Contribution{RoBindPaths: []string{"/opt/tools"}})
	for _, want := range []string{`(subpath "/Users/dev/.claude")`, `(subpath "/Users/dev/.claude.json")`, `(subpath "/Users/dev/extra")`, `(subpath "/Users/dev/repo")`, `(subpath "/opt/tools")`} {
		if !strings.Contains(profile, want) {
			t.Errorf("profile lacks %q:\n%s", want, profile)
		}
	}
	writeSection := profile[:strings.Index(profile, "(deny file-read-data")]
	if strings.Contains(writeSection, "/opt/tools") || strings.Contains(writeSection, "/Users/dev/repo") {
		t.Errorf("a read-only path must not be writable:\n%s", writeSection)
	}
}

func TestCommandNetworkModes(t *testing.T) {
	none := profileOf(t, newHost(), claudeCfg(network.NetworkNone), provision.Contribution{})
	if !strings.Contains(none, "(deny network*)") || strings.Contains(none, "allow network-outbound") {
		t.Errorf("none must deny all network:\n%s", none)
	}
	cfg := claudeCfg(network.NetworkManaged)
	cfg.Egress = isoshared.EgressRoute{Transport: network.TransportLoopback, Proxy: "127.0.0.1:40123"}
	managed := profileOf(t, newHost(), cfg, provision.Contribution{})
	if !strings.Contains(managed, `(allow network-outbound (remote ip "localhost:40123"))`) {
		t.Errorf("managed must allow only the proxy port:\n%s", managed)
	}
	if strings.Index(managed, "(deny network*)") > strings.Index(managed, "(allow network-outbound") {
		t.Error("the allow rule must follow the deny rule")
	}
}

func TestCommandSetsTheProxyEnvironmentWhenManaged(t *testing.T) {
	cfg := claudeCfg(network.NetworkManaged)
	cfg.Egress = isoshared.EgressRoute{Transport: network.TransportLoopback, Proxy: "127.0.0.1:40123"}
	_, env, err := NewIsolator(newHost(), testLog).TerminalCommand(cfg, provision.Contribution{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(env, "\n"), "HTTPS_PROXY=http://127.0.0.1:40123") {
		t.Errorf("env lacks the proxy variable: %v", env)
	}
}

func TestCommandRefusals(t *testing.T) {
	managed := claudeCfg(network.NetworkManaged)
	wrongTransport := managed
	wrongTransport.Egress = isoshared.EgressRoute{Transport: network.TransportUnixSocket, Socket: "/s", Bridge: "/b"}
	noRoute := managed
	badBind := claudeCfg(network.NetworkHost)
	badBind.BindPaths = []string{"relative"}
	badRoBind := claudeCfg(network.NetworkHost)
	badRoBind.RoBindPaths = []string{"/missing"}
	badRepo := claudeCfg(network.NetworkHost)
	badRepo.RepoDir = "/missing"
	badWork := claudeCfg(network.NetworkHost)
	badWork.Process.Dir = "/missing"
	badHome := claudeCfg(network.NetworkHost)
	badHome.HomeDir = "/missing"
	unknownMode := claudeCfg("other")
	cases := map[string]isoshared.RunConfig{
		"wrong transport": wrongTransport, "no route": noRoute, "relative bind": badBind,
		"missing read-only bind": badRoBind, "missing repo": badRepo, "missing work": badWork,
		"missing home": badHome, "unknown mode": unknownMode,
	}
	for name, cfg := range cases {
		if _, err := NewIsolator(newHost(), testLog).Command(cfg, provision.Contribution{}); err == nil {
			t.Errorf("%s: Command succeeded", name)
		}
	}
	if _, err := NewIsolator(newHost(), testLog).Command(claudeCfg(network.NetworkHost), provision.Contribution{RoBindPaths: []string{"/missing"}}); err == nil {
		t.Error("a missing provisioned bind must fail")
	}
	unsafe := newHost().AddDir("/Users/dev/we\"ird")
	cfg := claudeCfg(network.NetworkHost)
	cfg.BindPaths = []string{"/Users/dev/we\"ird"}
	if _, err := NewIsolator(unsafe, testLog).Command(cfg, provision.Contribution{}); err == nil {
		t.Error("a path with a quote must not enter the profile")
	}
	noCommand := claudeCfg(network.NetworkHost)
	noCommand.Process.Argv = nil
	if _, err := NewIsolator(newHost(), testLog).Command(noCommand, provision.Contribution{}); err == nil {
		t.Error("a process without a command must fail")
	}
}

func TestCapabilities(t *testing.T) {
	i := NewIsolator(newHost(), testLog)
	if !i.AcceptsHostBinds() {
		t.Error("the seatbelt profile grants access to a host path")
	}
	if i.HidesHost(false, "") || i.KernelIsolated("") || !i.ConfinesNetwork() || !i.Bounded() {
		t.Error("seatbelt hides no host, is no kernel boundary, confines the network and bounds the agent")
	}
	if i.EgressTransport() != network.TransportLoopback {
		t.Errorf("transport = %q", i.EgressTransport())
	}
	if !i.PinnedProvision(provision.ProvisionNix, true, "") || i.PinnedProvision(provision.ProvisionNone, false, "") {
		t.Error("pinned provision mirrors the provisioner")
	}
}

func TestAvailableAndCheckAgent(t *testing.T) {
	if ok, _ := NewIsolator(newHost(), testLog).Available(); !ok {
		t.Error("sandbox-exec is installed")
	}
	bare := systemtest.New()
	bare.Env["PATH"] = "/usr/bin"
	if ok, _ := NewIsolator(bare, testLog).Available(); ok {
		t.Error("sandbox-exec is missing")
	}
	if err := NewIsolator(newHost(), testLog).CheckAgent(claudeCfg(network.NetworkHost)); err != nil {
		t.Errorf("CheckAgent = %v", err)
	}
	if err := NewIsolator(bare, testLog).CheckAgent(claudeCfg(network.NetworkHost)); err == nil {
		t.Error("a missing agent must fail")
	}
}

func TestRunExecutesTheProfileCommand(t *testing.T) {
	host := newHost()
	isolator := NewIsolator(host, testLog)
	cfg := claudeCfg(network.NetworkNone)
	command, err := isolator.Command(cfg, provision.Contribution{})
	if err != nil {
		t.Fatal(err)
	}
	host.Outcomes[systemtest.CommandLine(command[0], command[1:]...)] = systemtest.Outcome{ExitCode: 3}
	code, err := isolator.Run(cfg, provision.Contribution{})
	if err != nil || code != 3 {
		t.Errorf("Run = %d, %v; want 3, nil", code, err)
	}
	if _, err := isolator.Run(claudeCfg("other"), provision.Contribution{}); err == nil {
		t.Error("a bad config must fail Run")
	}
	if _, _, err := isolator.TerminalCommand(claudeCfg("other"), provision.Contribution{}); err == nil {
		t.Error("a bad config must fail TerminalCommand")
	}
}

func TestQuote(t *testing.T) {
	if _, err := quote(""); err == nil {
		t.Error("an empty path must fail")
	}
	for _, bad := range []string{"rel", "/a\"b", "/a\\b", "/a\nb", "/a\x7fb"} {
		if _, err := quote(bad); err == nil {
			t.Errorf("quote(%q) succeeded", bad)
		}
	}
	if q, err := quote("/ok path"); err != nil || q != `"/ok path"` {
		t.Errorf("quote = %q, %v", q, err)
	}
}

func TestNetworkRulesRejectsABadProxy(t *testing.T) {
	for _, address := range []string{"", "nohost", "127.0.0.1:x", "127.0.0.1:0", "127.0.0.1:70000"} {
		if _, err := networkRules(profileSpec{Network: network.NetworkManaged, ProxyAddress: address}); err == nil {
			t.Errorf("proxy %q accepted", address)
		}
	}
}

func TestRenderProfileRejectsAnUnsafePath(t *testing.T) {
	if _, err := renderProfile(profileSpec{Writable: []string{"/a\"b"}, Network: network.NetworkHost}); err == nil {
		t.Error("an unsafe writable path must fail")
	}
	if _, err := renderProfile(profileSpec{Home: "/h\"", Network: network.NetworkHost}); err == nil {
		t.Error("an unsafe home must fail")
	}
	if _, err := renderProfile(profileSpec{Home: "/h", Readable: []string{"x"}, Network: network.NetworkHost}); err == nil {
		t.Error("an unsafe readable path must fail")
	}
	if _, err := renderProfile(profileSpec{Network: "other"}); err == nil {
		t.Error("an unknown network mode must fail")
	}
}
