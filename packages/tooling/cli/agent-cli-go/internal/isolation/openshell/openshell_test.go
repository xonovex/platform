package openshell

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

var testLog = logging.New(io.Discard, io.Discard, false, false)

const pinned = "registry.example/agent@sha256:95a34da32a840bd9b3b09a5b773591c16923e350174b1c50e1200c75bf15eaa9"

func newHost() *systemtest.Fake {
	host := systemtest.New().AddDir("/home/dev").AddDir("/home/dev/work")
	host.Env["PATH"] = "/usr/bin"
	return host.AddExecutable("/usr/bin/openshell")
}

func allowlist(t *testing.T, rules ...string) egress.Allowlist {
	t.Helper()
	list, err := egress.NewAllowlist(rules)
	if err != nil {
		t.Fatal(err)
	}
	return list
}

func cfg(t *testing.T, mode network.NetworkMethod) isoshared.RunConfig {
	t.Helper()
	agent, err := agents.GetAgent(types.AgentClaude)
	if err != nil {
		t.Fatal(err)
	}
	run := isoshared.RunConfig{Agent: agent, Process: isoshared.Process{Argv: []string{"claude", "review"}, Dir: "/home/dev/work"}, HomeDir: "/home/dev", Network: mode, Image: pinned}
	if mode == network.NetworkManaged {
		run.Egress = isoshared.EgressRoute{Transport: network.TransportPolicy, Allow: allowlist(t, "api.example.com", "*.cdn.example.com:8443")}
	}
	return run
}

// script answers every openshell call of a run in "agent-cli-1" with success.
func script(t *testing.T, host *systemtest.Fake, run isoshared.RunConfig) *plan {
	t.Helper()
	p, err := NewIsolator(host, testLog).plan(run, provision.Contribution{}, "agent-cli-1")
	if err != nil {
		t.Fatal(err)
	}
	host.Outcomes["openshell sandbox list --names"] = systemtest.Outcome{}
	host.Outcomes[systemtest.CommandLine(program, p.create("/tmp/agent-cli-openshell-1/policy.json")...)] = systemtest.Outcome{}
	for _, args := range [][]string{p.upload, p.exec, p.download, p.remove} {
		host.Outcomes[systemtest.CommandLine(program, args...)] = systemtest.Outcome{}
	}
	return &p
}

func TestPolicyForANoneRunDeniesTheNetwork(t *testing.T) {
	data, err := renderPolicy(network.NetworkNone, egress.Allowlist{})
	if err != nil {
		t.Fatal(err)
	}
	var document policyDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.NetworkPolicies) != 0 || document.Landlock.Compatibility != "hard_requirement" || !document.Filesystem.IncludeWorkdir {
		t.Errorf("document = %+v", document)
	}
	if len(document.Filesystem.ReadOnly) == 0 || document.Filesystem.ReadWrite[0] != "/dev/null" {
		t.Errorf("filesystem = %+v: a policy with no network rule needs the system paths", document.Filesystem)
	}
}

func TestPolicyForAManagedRunListsTheAllowlist(t *testing.T) {
	data, err := renderPolicy(network.NetworkManaged, allowlist(t, "api.example.com", "*.cdn.example.com:8443"))
	if err != nil {
		t.Fatal(err)
	}
	var document policyDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	rule := document.NetworkPolicies["egress"]
	for _, e := range rule.Endpoints {
		if e.Enforcement != "enforce" {
			t.Errorf("endpoint %+v must enforce, not audit", e)
		}
	}
	hosts := map[string]int{}
	for _, e := range rule.Endpoints {
		hosts[e.Host] = e.Port
	}
	if hosts["api.example.com"] != 443 || hosts["*.cdn.example.com"] != 8443 || len(hosts) != 2 {
		t.Errorf("endpoints = %v", hosts)
	}
	if len(rule.Binaries) != 1 || rule.Binaries[0].Path != "/**" {
		t.Errorf("binaries = %v", rule.Binaries)
	}
}

func TestPolicyRefusals(t *testing.T) {
	if _, err := renderPolicy(network.NetworkManaged, egress.Allowlist{}); err == nil {
		t.Error("a managed policy without hosts must fail")
	}
	if _, err := renderPolicy(network.NetworkHost, egress.Allowlist{}); !errors.Is(err, ErrUnsupported) {
		t.Errorf("host mode error = %v, want ErrUnsupported", err)
	}
}

func TestCommandExecsInTheUploadedDirectory(t *testing.T) {
	run := cfg(t, network.NetworkNone)
	run.Process.Env = []string{"MODE=fast"}
	command, err := NewIsolator(newHost(), testLog).Command(run, provision.Contribution{})
	if err != nil {
		t.Fatal(err)
	}
	line := strings.Join(command, " ")
	want := "openshell sandbox exec --name agent-cli-1 --workdir /sandbox/work --no-tty --env MODE=fast -- claude review"
	if line != want {
		t.Errorf("command = %q, want %q", line, want)
	}
}

func TestRunLifecycle(t *testing.T) {
	run := cfg(t, network.NetworkManaged)
	host := newHost()
	script(t, host, run)
	code, err := NewIsolator(host, testLog).Run(run, provision.Contribution{})
	if err != nil || code != 0 {
		t.Fatalf("Run = %d, %v", code, err)
	}
	var verbs []string
	for _, call := range host.Calls {
		verbs = append(verbs, call.Args[0]+" "+call.Args[1])
	}
	want := "sandbox list,sandbox create,sandbox upload,sandbox exec,sandbox download,sandbox delete"
	if strings.Join(verbs, ",") != want {
		t.Errorf("calls = %v, want %s", verbs, want)
	}
	create := strings.Join(host.Calls[1].Args, " ")
	if !strings.Contains(create, "--from "+pinned) || !strings.Contains(create, "--policy ") || !strings.Contains(create, "--no-auto-providers") {
		t.Errorf("create = %q", create)
	}
	if host.Exists("/tmp/agent-cli-openshell-1") {
		t.Error("the policy directory must be removed")
	}
}

func TestRunDownloadsAndDeletesAfterAFailedAgent(t *testing.T) {
	run := cfg(t, network.NetworkNone)
	host := newHost()
	p := script(t, host, run)
	host.Outcomes[systemtest.CommandLine(program, p.exec...)] = systemtest.Outcome{ExitCode: 5}
	code, err := NewIsolator(host, testLog).Run(run, provision.Contribution{})
	if err != nil || code != 5 {
		t.Fatalf("Run = %d, %v; want 5, nil", code, err)
	}
	last := host.Calls[len(host.Calls)-1]
	prev := host.Calls[len(host.Calls)-2]
	if last.Args[1] != "delete" || prev.Args[1] != "download" {
		t.Errorf("the run must download then delete, got %v then %v", prev.Args, last.Args)
	}
}

func TestRunFailures(t *testing.T) {
	boom := errors.New("boom")
	run := cfg(t, network.NetworkNone)
	cases := map[string]func(*systemtest.Fake, *plan){
		"list": func(h *systemtest.Fake, _ *plan) {
			h.Outcomes["openshell sandbox list --names"] = systemtest.Outcome{Err: boom}
		},
		"create": func(h *systemtest.Fake, p *plan) {
			h.Outcomes[systemtest.CommandLine(program, p.create("/tmp/agent-cli-openshell-1/policy.json")...)] = systemtest.Outcome{Err: boom}
		},
		"upload": func(h *systemtest.Fake, p *plan) {
			h.Outcomes[systemtest.CommandLine(program, p.upload...)] = systemtest.Outcome{Err: boom}
		},
		"download": func(h *systemtest.Fake, p *plan) {
			h.Outcomes[systemtest.CommandLine(program, p.download...)] = systemtest.Outcome{Err: boom}
		},
		"exec": func(h *systemtest.Fake, p *plan) {
			h.Outcomes[systemtest.CommandLine(program, p.exec...)] = systemtest.Outcome{Err: boom}
		},
	}
	for name, setup := range cases {
		host := newHost()
		p := script(t, host, run)
		setup(host, p)
		if _, err := NewIsolator(host, testLog).Run(run, provision.Contribution{}); err == nil {
			t.Errorf("%s: Run succeeded", name)
		}
		if name != "list" && host.Calls[len(host.Calls)-1].Args[1] != "delete" {
			t.Errorf("%s: the sandbox must be deleted after a failure", name)
		}
	}
}

func TestRunWarnsWhenTheSandboxCannotBeDeleted(t *testing.T) {
	run := cfg(t, network.NetworkNone)
	host := newHost()
	p := script(t, host, run)
	host.Outcomes[systemtest.CommandLine(program, p.remove...)] = systemtest.Outcome{Err: errors.New("busy")}
	var out strings.Builder
	if _, err := NewIsolator(host, logging.New(&out, &out, false, false)).Run(run, provision.Contribution{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "agent-cli-1") {
		t.Errorf("the log lacks the leftover sandbox: %q", out.String())
	}
}

type failingHost struct {
	*systemtest.Fake
	mkdirErr, writeErr error
}

func (f failingHost) MkdirTemp(pattern string) (string, error) {
	if f.mkdirErr != nil {
		return "", f.mkdirErr
	}
	return f.Fake.MkdirTemp(pattern)
}

func (f failingHost) WriteFile(path string, data []byte, perm fs.FileMode) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	return f.Fake.WriteFile(path, data, perm)
}

func TestRunFailsWhenThePolicyCannotBeWritten(t *testing.T) {
	run := cfg(t, network.NetworkNone)
	host := newHost()
	script(t, host, run)
	for name, h := range map[string]failingHost{
		"mkdir": {Fake: host, mkdirErr: errors.New("disk full")},
		"write": {Fake: host, writeErr: errors.New("read only")},
	} {
		if _, err := NewIsolator(h, testLog).Run(run, provision.Contribution{}); err == nil {
			t.Errorf("%s: Run succeeded", name)
		}
	}
}

func TestRunFailsWhenEveryNameIsTaken(t *testing.T) {
	host := newHost()
	var names []string
	for n := 1; n <= maxNames; n++ {
		names = append(names, "agent-cli-"+strings.TrimLeft(strings.Repeat("0", 4)+itoa(n), "0"))
	}
	host.Outcomes["openshell sandbox list --names"] = systemtest.Outcome{Stdout: strings.Join(names, "\n")}
	if _, err := NewIsolator(host, testLog).Run(cfg(t, network.NetworkNone), provision.Contribution{}); err == nil {
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
	noImage := cfg(t, network.NetworkNone)
	noImage.Image = ""
	dashImage := cfg(t, network.NetworkNone)
	dashImage.Image = "-rm"
	secrets := cfg(t, network.NetworkNone)
	secrets.Secrets = map[string]string{"K": "v"}
	provider := cfg(t, network.NetworkNone)
	provider.Provider = &types.ModelProvider{}
	bind := cfg(t, network.NetworkNone)
	bind.BindPaths = []string{"/home/dev"}
	roBind := cfg(t, network.NetworkNone)
	roBind.RoBindPaths = []string{"/home/dev"}
	host := cfg(t, network.NetworkHost)
	missingWork := cfg(t, network.NetworkNone)
	missingWork.Process.Dir = "/missing"
	badEnv := cfg(t, network.NetworkNone)
	badEnv.Process.Env = []string{"not-an-assignment"}
	for name, run := range map[string]isoshared.RunConfig{
		"no image": noImage, "image with hyphen": dashImage, "secrets": secrets, "provider": provider,
		"bind": bind, "read-only bind": roBind, "host mode": host, "missing work": missingWork,
		"bad env": badEnv,
	} {
		if _, err := NewIsolator(newHost(), testLog).Command(run, provision.Contribution{}); err == nil {
			t.Errorf("%s: Command succeeded", name)
		}
	}
	for name, c := range map[string]provision.Contribution{
		"path entries": {PathEntries: []string{"/opt/bin"}},
		"ro bind":      {RoBindPaths: []string{"/opt"}},
	} {
		if _, err := NewIsolator(newHost(), testLog).Command(cfg(t, network.NetworkNone), c); !errors.Is(err, ErrUnsupported) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
	if _, err := NewIsolator(newHost(), testLog).Run(noImage, provision.Contribution{}); err == nil {
		t.Error("Run without an image must fail")
	}
	if _, _, err := NewIsolator(newHost(), testLog).TerminalCommand(cfg(t, network.NetworkNone), provision.Contribution{}); !errors.Is(err, ErrUnsupported) {
		t.Errorf("TerminalCommand error = %v", err)
	}
}

func TestInitCommandsRunBeforeTheAgent(t *testing.T) {
	command, err := NewIsolator(newHost(), testLog).Command(cfg(t, network.NetworkNone), provision.Contribution{InitCommands: []string{"prepare"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(command, " "), "prepare") {
		t.Errorf("command = %v", command)
	}
}

func TestCapabilities(t *testing.T) {
	i := NewIsolator(newHost(), testLog)
	if !i.HidesHost(false, pinned) || i.HidesHost(false, "node:latest") {
		t.Error("only a digest-pinned image hides the host")
	}
	if !i.PinnedProvision(provision.ProvisionNone, false, pinned) || i.PinnedProvision(provision.ProvisionNix, false, pinned) || i.PinnedProvision(provision.ProvisionNone, true, "node:latest") {
		t.Error("pinned provision needs a pinned image and a pinned provisioner")
	}
	if !i.PinnedProvision(provision.ProvisionNix, true, pinned) {
		t.Error("a pinned image with a pinned provisioner is pinned")
	}
	if i.KernelIsolated("runsc") || !i.ConfinesNetwork() || i.EgressTransport() != network.TransportPolicy || !i.Bounded() {
		t.Error("openshell declares a policy transport and no verified kernel boundary")
	}
	if err := i.CheckAgent(cfg(t, network.NetworkNone)); err != nil {
		t.Errorf("CheckAgent = %v", err)
	}
	if ok, _ := i.Available(); !ok {
		t.Error("openshell is installed")
	}
	bare := systemtest.New()
	bare.Env["PATH"] = "/usr/bin"
	if ok, _ := NewIsolator(bare, testLog).Available(); ok {
		t.Error("openshell is missing")
	}
}
