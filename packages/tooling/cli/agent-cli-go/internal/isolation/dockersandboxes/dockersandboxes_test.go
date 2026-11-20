package dockersandboxes

import (
	"bytes"
	"errors"
	"io"
	"slices"
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
	host := systemtest.New().AddDir("/home/dev").AddDir("/home/dev/work").AddDir("/home/dev/extra").AddDir("/opt/tools")
	host.Env["PATH"] = "/usr/bin"
	return host.AddExecutable("/usr/bin/sbx")
}

// script answers every sbx call of a run in the sandbox "agent-cli-1" with
// success, except those a test overrides afterwards.
func script(host *systemtest.Fake, run isoshared.RunConfig) *systemtest.Fake {
	i := NewIsolator(host, testLog)
	p, err := i.plan(run, provision.Contribution{}, "agent-cli-1")
	if err != nil {
		panic(err)
	}
	host.Outcomes["sbx ls -q"] = systemtest.Outcome{}
	host.Outcomes[systemtest.CommandLine(program, p.create...)] = systemtest.Outcome{}
	host.Outcomes[systemtest.CommandLine(program, p.exec...)] = systemtest.Outcome{}
	host.Outcomes["sbx stop agent-cli-1"] = systemtest.Outcome{}
	host.Outcomes["sbx rm --force agent-cli-1"] = systemtest.Outcome{}
	return host
}

func cfg() isoshared.RunConfig {
	agent, err := agents.GetAgent(types.AgentClaude)
	if err != nil {
		panic(err)
	}
	return isoshared.RunConfig{Agent: agent, Process: isoshared.Process{Argv: []string{"claude", "review"}, Dir: "/home/dev/work"}, HomeDir: "/home/dev", Network: network.NetworkHost}
}

func TestCommandNamesTheSandboxAndPassesEnvironmentByName(t *testing.T) {
	run := cfg()
	run.Secrets = map[string]string{"API_KEY": "s3cret"}
	run.Process.Env = []string{"MODE=fast"}
	run.ProviderEnv = map[string]string{"ROUTE_AUTH": "route-token"}
	command, err := NewIsolator(newHost(), testLog).Command(run, provision.Contribution{})
	if err != nil {
		t.Fatal(err)
	}
	line := strings.Join(command, " ")
	if line != "sbx exec -i -t -w /home/dev/work -e API_KEY -e MODE -e ROUTE_AUTH agent-cli-1 claude review" {
		t.Errorf("command = %q", line)
	}
	if strings.Contains(line, "s3cret") || strings.Contains(line, "route-token") {
		t.Error("a credential value must never appear in an argument")
	}
}

func TestRunCreatesExecutesAndRemoves(t *testing.T) {
	run := cfg()
	run.BindPaths = []string{"/home/dev/extra"}
	run.RoBindPaths = []string{"/opt/tools"}
	run.Secrets = map[string]string{"API_KEY": "s3cret"}
	host := newHost()
	p, err := NewIsolator(host, testLog).plan(run, provision.Contribution{}, "agent-cli-2")
	if err != nil {
		t.Fatal(err)
	}
	host.Outcomes["sbx ls -q"] = systemtest.Outcome{Stdout: "agent-cli-1\nother\n"}
	host.Outcomes[systemtest.CommandLine(program, p.create...)] = systemtest.Outcome{}
	host.Outcomes[systemtest.CommandLine(program, p.exec...)] = systemtest.Outcome{}
	host.Outcomes["sbx stop agent-cli-2"] = systemtest.Outcome{}
	host.Outcomes["sbx rm --force agent-cli-2"] = systemtest.Outcome{}
	code, err := NewIsolator(host, testLog).Run(run, provision.Contribution{})
	if err != nil || code != 0 {
		t.Fatalf("Run = %d, %v", code, err)
	}
	var lines []string
	for _, call := range host.Calls {
		lines = append(lines, call.Name+" "+strings.Join(call.Args, " "))
	}
	wantCreate := "sbx create --no-progress --name agent-cli-2 claude /home/dev/work /home/dev/extra /opt/tools:ro"
	if len(lines) != 5 || lines[1] != wantCreate || !strings.HasPrefix(lines[2], "sbx exec") || lines[3] != "sbx stop agent-cli-2" || lines[4] != "sbx rm --force agent-cli-2" {
		t.Fatalf("calls = %q", lines)
	}
	if !slices.Contains(host.Calls[2].Env, "API_KEY=s3cret") {
		t.Error("the credential must reach sbx exec through its environment")
	}
	for _, line := range lines {
		if strings.Contains(line, "s3cret") || strings.Contains(line, "route-token") {
			t.Errorf("a credential value reached an argument: %q", line)
		}
	}
}

func TestRunRemovesTheSandboxAfterAFailedExec(t *testing.T) {
	host := script(newHost(), cfg())
	command, err := NewIsolator(host, testLog).Command(cfg(), provision.Contribution{})
	if err != nil {
		t.Fatal(err)
	}
	host.Outcomes[systemtest.CommandLine(command[0], command[1:]...)] = systemtest.Outcome{ExitCode: 9}
	code, err := NewIsolator(host, testLog).Run(cfg(), provision.Contribution{})
	if err != nil || code != 9 {
		t.Errorf("Run = %d, %v; want 9, nil", code, err)
	}
	last := host.Calls[len(host.Calls)-1]
	if last.Name != "sbx" || last.Args[0] != "rm" {
		t.Errorf("the last call = %v, want sbx rm", last)
	}
}

func TestRunReportsFailures(t *testing.T) {
	boom := errors.New("boom")
	cases := map[string]func(*systemtest.Fake){
		"list fails": func(h *systemtest.Fake) { h.Outcomes["sbx ls -q"] = systemtest.Outcome{Err: boom} },
		"create fails": func(h *systemtest.Fake) {
			h.Outcomes["sbx create --no-progress --name agent-cli-1 claude /home/dev/work"] = systemtest.Outcome{Err: boom}
		},
	}
	for name, setup := range cases {
		host := script(newHost(), cfg())
		setup(host)
		if _, err := NewIsolator(host, testLog).Run(cfg(), provision.Contribution{}); err == nil {
			t.Errorf("%s: Run succeeded", name)
		}
		if name == "create fails" && host.Calls[len(host.Calls)-1].Args[0] != "rm" {
			t.Errorf("%s: a failed create must still try to remove the sandbox", name)
		}
	}
	if _, err := NewIsolator(newHost(), testLog).Run(isoshared.RunConfig{}, provision.Contribution{}); err == nil {
		t.Error("a bad config must fail Run")
	}
}

func TestRunWarnsWhenTheSandboxCannotBeRemoved(t *testing.T) {
	host := script(newHost(), cfg())
	host.Outcomes["sbx rm --force agent-cli-1"] = systemtest.Outcome{Err: errors.New("busy")}
	var out bytes.Buffer
	if _, err := NewIsolator(host, logging.New(&out, &out, false, false)).Run(cfg(), provision.Contribution{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "agent-cli-1") {
		t.Errorf("the log lacks the leftover sandbox name: %q", out.String())
	}
}

func TestRunFailsWhenEveryNameIsTaken(t *testing.T) {
	host := newHost()
	var names []string
	for n := 1; n <= maxNames; n++ {
		names = append(names, namePrefix+itoa(n))
	}
	host.Outcomes["sbx ls -q"] = systemtest.Outcome{Stdout: strings.Join(names, "\n")}
	if _, err := NewIsolator(host, testLog).Run(cfg(), provision.Contribution{}); err == nil {
		t.Error("Run must fail without a free name")
	}
}

func itoa(n int) string {
	var digits []byte
	for ; n > 0; n /= 10 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
	}
	return string(digits)
}

func TestRefusals(t *testing.T) {
	none := cfg()
	none.Network = network.NetworkNone
	managed := cfg()
	managed.Network = network.NetworkManaged
	badBind := cfg()
	badBind.BindPaths = []string{"/missing"}
	badRoBind := cfg()
	badRoBind.RoBindPaths = []string{"/missing"}
	badWork := cfg()
	badWork.Process.Dir = "/missing"
	badEnv := cfg()
	badEnv.Process.Env = []string{"not-an-assignment"}
	colon := cfg()
	colon.BindPaths = []string{"/home/dev/a:b"}
	for name, run := range map[string]isoshared.RunConfig{
		"none": none, "managed": managed, "missing bind": badBind, "missing ro bind": badRoBind,
		"missing work": badWork, "bad env": badEnv, "colon": colon,
	} {
		host := newHost().AddDir("/home/dev/a:b")
		if _, err := NewIsolator(host, testLog).Command(run, provision.Contribution{}); err == nil {
			t.Errorf("%s: Command succeeded", name)
		}
	}
	for name, c := range map[string]provision.Contribution{
		"init commands": {InitCommands: []string{"setup"}},
		"path entries":  {PathEntries: []string{"/opt/bin"}},
	} {
		if _, err := NewIsolator(newHost(), testLog).Command(cfg(), c); !errors.Is(err, ErrUnsupported) {
			t.Errorf("%s: error = %v, want ErrUnsupported", name, err)
		}
	}
	if _, _, err := NewIsolator(newHost(), testLog).TerminalCommand(cfg(), provision.Contribution{}); !errors.Is(err, ErrUnsupported) {
		t.Errorf("TerminalCommand error = %v", err)
	}
}

func TestMountable(t *testing.T) {
	for _, bad := range []string{"-x", "/a:b", "/a\nb"} {
		if mountable(bad) == nil {
			t.Errorf("mountable(%q) succeeded", bad)
		}
	}
	if mountable("/ok") != nil {
		t.Error("an ordinary path must pass")
	}
}

func TestCapabilitiesFailClosed(t *testing.T) {
	i := NewIsolator(newHost(), testLog)
	if i.HidesHost(true, "img") || i.KernelIsolated("runsc") || i.ConfinesNetwork() || i.PinnedProvision(provision.ProvisionNix, true, "img") || !i.Bounded() {
		t.Error("docker-sandboxes must claim none of the guarantees it cannot verify")
	}
	if i.EgressTransport() != network.TransportNone {
		t.Errorf("transport = %q, want none", i.EgressTransport())
	}
	if err := i.CheckAgent(cfg()); err != nil {
		t.Errorf("CheckAgent = %v", err)
	}
	if ok, _ := i.Available(); !ok {
		t.Error("sbx is installed")
	}
	bare := systemtest.New()
	bare.Env["PATH"] = "/usr/bin"
	if ok, _ := NewIsolator(bare, testLog).Available(); ok {
		t.Error("sbx is missing")
	}
}
