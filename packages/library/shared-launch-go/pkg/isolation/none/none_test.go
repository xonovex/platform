package none

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
	host := oshosttest.New().AddDir("/work").AddDir("/home/user")
	host.Env["PATH"] = "/usr/bin"
	return host
}

func cfg(net network.NetworkMethod) isoshared.RunConfig {
	return isoshared.RunConfig{
		Agent:   registeredAgent(types.AgentClaude),
		HomeDir: "/home/user",
		Process: isoshared.Process{Argv: []string{"claude", "review"}, Dir: "/work"},
		Network: net,
	}
}

func TestNone_FailsClosedOnNetworkRestriction(t *testing.T) {
	for _, net := range []network.NetworkMethod{network.NetworkNone, network.NetworkManaged} {
		host := newHost()
		i := NewIsolator(host, testLog)
		c := provision.Contribution{}
		if code, err := i.Run(cfg(net), c); err == nil || code == 0 {
			t.Errorf("Run(network=%q) = (%d, %v), want fail-closed error (none cannot restrict egress)", net, code, err)
		}
		if _, err := i.Command(cfg(net), c); err == nil {
			t.Errorf("Command(network=%q) error = nil, want fail-closed error", net)
		}
		if _, _, err := i.TerminalCommand(cfg(net), c); err == nil {
			t.Errorf("TerminalCommand(network=%q) error = nil, want fail-closed error", net)
		}
		if len(host.Calls) != 0 {
			t.Errorf("network=%q started %v, want no process", net, host.Calls)
		}
	}
}

func TestNone_CommandIsTheProcessUnchanged(t *testing.T) {
	cmd, err := NewIsolator(newHost(), testLog).Command(cfg(network.NetworkHost), provision.Contribution{})
	if err != nil {
		t.Fatalf("Command() error = %v", err)
	}
	if !slices.Equal(cmd, []string{"claude", "review"}) {
		t.Errorf("command = %v, want the argument vector of the process", cmd)
	}
}

func TestNone_AppliesContribution(t *testing.T) {
	c := provision.Contribution{
		PathEntries:  []string{"/nix/store/abc/bin"},
		Env:          map[string]string{"FOO": "bar"},
		InitCommands: []string{"echo hi"},
	}
	cmd, env, err := NewIsolator(newHost(), testLog).hostCommand(cfg(network.NetworkHost), c)
	if err != nil {
		t.Fatalf("hostCommand() error = %v", err)
	}

	if cmd[0] != "sh" || cmd[1] != "-c" {
		t.Fatalf("init command must wrap the agent in sh -c, got %v", cmd)
	}
	if !strings.Contains(strings.Join(cmd, " "), "echo hi") {
		t.Error("init command missing")
	}
	if !slices.Contains(env, "PATH=/nix/store/abc/bin:/usr/bin") {
		t.Errorf("env = %v, want the contribution PathEntries before the host PATH", env)
	}
	if !slices.Contains(env, "FOO=bar") {
		t.Error("contribution env not applied")
	}
}

func TestNone_PathEntriesWithoutAHostPath(t *testing.T) {
	host := newHost()
	delete(host.Env, "PATH")

	_, env, err := NewIsolator(host, testLog).hostCommand(cfg(network.NetworkHost), provision.Contribution{PathEntries: []string{"/nix/store/abc/bin"}})

	if err != nil || !slices.Contains(env, "PATH=/nix/store/abc/bin") {
		t.Fatalf("hostCommand() = (%v, %v), want PATH of the contribution only", env, err)
	}
}

func TestNone_MergesHostProviderAndCustomEnvironment(t *testing.T) {
	host := newHost()
	host.Env["HOST_ONLY"] = "kept"
	config := cfg(network.NetworkHost)
	config.ProviderEnv = map[string]string{"AGENT_PROVIDER_TOKEN": "secret"}
	config.Process.Env = []string{"CUSTOM=1"}

	_, env, err := NewIsolator(host, testLog).TerminalCommand(config, provision.Contribution{})

	if err != nil {
		t.Fatalf("TerminalCommand() error = %v", err)
	}
	for _, want := range []string{"HOST_ONLY=kept", "AGENT_PROVIDER_TOKEN=secret", "CUSTOM=1"} {
		if !slices.Contains(env, want) {
			t.Errorf("env = %v, want %s", env, want)
		}
	}
}

func TestNone_CommandRejectsMalformedCustomEnvironment(t *testing.T) {
	config := cfg(network.NetworkHost)
	config.Process.Env = []string{"MISSING_EQUALS"}

	if _, err := NewIsolator(newHost(), testLog).Command(config, provision.Contribution{}); err == nil {
		t.Fatal("Command() error = nil, want custom environment error")
	}
}

func TestNone_CommandRejectsMissingWorkDirectory(t *testing.T) {
	config := cfg(network.NetworkHost)
	config.Process.Dir = "/missing"

	if _, err := NewIsolator(newHost(), testLog).Command(config, provision.Contribution{}); err == nil {
		t.Fatal("Command() error = nil, want work directory error")
	}
}

func TestNone_RunSpawnsTheAgentInTheWorkDirectory(t *testing.T) {
	host := newHost()
	host.Outcomes["claude review"] = oshosttest.Outcome{ExitCode: 3}

	code, err := NewIsolator(host, testLog).Run(cfg(network.NetworkHost), provision.Contribution{})

	if err != nil || code != 3 {
		t.Fatalf("Run() = (%d, %v), want (3, nil)", code, err)
	}
	if len(host.Calls) != 1 || host.Calls[0].Dir != "/work" || !slices.Contains(host.Calls[0].Env, "PATH=/usr/bin") {
		t.Errorf("calls = %+v, want one run of claude in /work with the host environment", host.Calls)
	}
}

func TestNone_Capabilities(t *testing.T) {
	i := NewIsolator(newHost(), testLog)
	if i.HidesHost(false, "img") {
		t.Error("none always exposes the host")
	}
	if i.KernelIsolated("runsc") {
		t.Error("none has no kernel boundary")
	}
	if i.ConfinesNetwork() {
		t.Error("none has no network namespace")
	}
	if !i.AcceptsHostBinds() {
		t.Error("none reaches every host path")
	}
	if i.Bounded() {
		t.Error("none sets no boundary around the agent")
	}
}

func TestNone_CheckAgentLooksUpTheHostPath(t *testing.T) {
	host := newHost()
	i := NewIsolator(host, testLog)

	if err := i.CheckAgent(cfg(network.NetworkHost)); err == nil || !strings.Contains(err.Error(), "not available on host PATH") {
		t.Fatalf("CheckAgent(missing) error = %v, want host PATH error", err)
	}
	host.AddExecutable("/usr/bin/claude")
	if err := i.CheckAgent(cfg(network.NetworkHost)); err != nil {
		t.Fatalf("CheckAgent(present) error = %v", err)
	}
}

// Host execution has no image and no boundary, so availability is constant and
// immutability is exactly the provisioner's.
func TestNone_AvailableAndPinnedProvision(t *testing.T) {
	isolator := NewIsolator(newHost(), testLog)
	if available, err := isolator.Available(); !available || err != nil {
		t.Errorf("Available() = (%v, %v), want (true, nil)", available, err)
	}
	for _, pinned := range []bool{true, false} {
		if got := isolator.PinnedProvision(provision.ProvisionNix, pinned, ""); got != pinned {
			t.Errorf("PinnedProvision(pinned=%v) = %v, want %v", pinned, got, pinned)
		}
	}
}

func TestNone_DeclaresNoEgressTransport(t *testing.T) {
	if got := NewIsolator(newHost(), testLog).EgressTransport(); got != network.TransportNone {
		t.Errorf("EgressTransport() = %q, want none: host execution has no boundary to route traffic through", got)
	}
}

func TestStopCommandHasNothingToStop(t *testing.T) {
	if stop := NewIsolator(oshosttest.New(), testLog).StopCommand(isoshared.RunConfig{Name: "named"}); stop != nil {
		t.Errorf("StopCommand = %v, want nil: the end of the process ends the sandbox", stop)
	}
}
