package plugins_test

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox/plugins"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

var updateGolden = flag.Bool("update-golden", false, "regenerate the launch golden file")

// goldenFiles holds the golden in the test binary, so a normal run reads no
// file; only -update-golden writes to testdata.
//
//go:embed testdata/launch.golden.txt
var goldenFiles embed.FS

const goldenPath = "testdata/launch.golden.txt"

const (
	goldenImage    = "registry.example/agent@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	goldenSocket   = "/run/egress/proxy.sock"
	goldenBridge   = "/opt/agent-cli/agent-cli"
	goldenProxy    = "127.0.0.1:41000"
	goldenOverlay  = "/tmp/overlay"
	goldenUser     = "/home/user/.config/agentic/providers.d"
	goldenCheckout = "/work/checkout"
)

// goldenValidator checks the configuration files against the schemas of the
// contract.
var goldenValidator = func() jsonschema.Validator {
	validator, err := wiring.NewValidator()
	if err != nil {
		panic(err)
	}
	return validator
}()

// goldenHost is a oshosttest.Fake whose every process succeeds, so a run of a
// real isolator reaches the launch of the agent and the test records it.
type goldenHost struct{ *oshosttest.Fake }

func (h goldenHost) Run(name string, args, env []string, dir string) (int, error) {
	h.Calls = append(h.Calls, oshosttest.Call{Kind: "run", Name: name, Args: args, Env: env, Dir: dir})
	return 0, nil
}

func (h goldenHost) Capture(name string, args []string, dir string) (string, error) {
	h.Calls = append(h.Calls, oshosttest.Call{Kind: "capture", Name: name, Args: args, Dir: dir})
	return "", nil
}

func newGoldenHost() goldenHost {
	fake := oshosttest.New()
	fake.WorkingDir = "/work/repo"
	fake.Env = map[string]string{
		"PATH":           "/usr/bin:/bin",
		"HOME":           "/home/user",
		"LANG":           "C.UTF-8",
		"ROUTE_TOKEN":    "route-token",
		"SECRET_SOURCE":  "s3cret",
		"UNRELATED_HOST": "host-only",
	}
	for _, dir := range []string{"/work/repo", "/home/user/.claude", "/usr", "/lib", "/bin", "/etc", "/run/egress", goldenOverlay, goldenCheckout} {
		fake.AddDir(dir)
	}
	for _, name := range []string{"bwrap", "docker", "sandbox-exec", "sbx", "openshell", "claude", "codex"} {
		fake.AddExecutable("/usr/bin/" + name)
	}
	fake.AddExecutable(goldenBridge)
	fake.AddFile(goldenUser+"/route.json", `{"name":"route","agent":"claude","models":{"main":{"id":"model-main"}},"environment":{"ROUTE_MODEL":"{model:main}"},"credential":{"sourceEnv":"ROUTE_TOKEN","targetEnv":"ROUTE_AUTH"}}`, 0o644)
	fake.AddFile(goldenUser+"/coder.json", `{"name":"coder","agent":"codex","models":{"main":{"id":"model-code"}},"args":["--model","{model:main}"]}`, 0o644)
	return goldenHost{fake}
}

// goldenEgress opens the route of each transport at fixed places.
type goldenEgress struct{}

func (goldenEgress) Open(transport network.EgressTransport, allow egress.Allowlist) (isoshared.EgressRoute, func(), error) {
	route := isoshared.EgressRoute{Transport: transport, Allow: allow}
	switch transport {
	case network.TransportUnixSocket:
		route.Socket, route.Bridge = goldenSocket, goldenBridge
	case network.TransportLoopback:
		route.Proxy = goldenProxy
	}
	return route, func() {}, nil
}

// goldenOverlays places a fixed overlay.
type goldenOverlays struct{}

func (goldenOverlays) Prepare(launch.OverlayRequest) (launch.OverlayResult, error) {
	return launch.OverlayResult{
		Root:        goldenOverlay,
		Environment: []string{"OVERLAY_HOME=" + goldenOverlay},
		Arguments:   []string{"--plugin-dir", goldenOverlay + "/plugins"},
		RoBindPaths: []string{goldenOverlay},
		Summary:     []string{"placed hello"},
		Cleanup:     func() {},
	}, nil
}

// goldenVCS creates every checkout at one place.
type goldenVCS struct{}

func (goldenVCS) Setup(wsshared.Config, string, bool) (string, error) { return goldenCheckout, nil }
func (goldenVCS) Available() bool                                     { return true }

// goldenTerminal records the command a terminal wrapper would start.
type goldenTerminal struct{ calls *[]string }

func (goldenTerminal) IsAvailable() bool { return true }
func (goldenTerminal) IsInside() bool    { return false }

func (g goldenTerminal) Execute(_ *termshared.TerminalConfig, command, env []string, workDir string) (termshared.Outcome, error) {
	*g.calls = append(*g.calls, "terminal: "+strings.Join(command, " "), "terminal env: "+strings.Join(sorted(env), " "), "terminal dir: "+workDir)
	return termshared.Outcome{}, nil
}

func sorted(values []string) []string {
	out := slices.Clone(values)
	slices.Sort(out)
	return out
}

type goldenVariant struct {
	name  string
	apply func(*launch.Options)
}

var goldenVariants = []goldenVariant{
	{"plain", func(*launch.Options) {}},
	{"provider and custom environment", func(o *launch.Options) {
		o.Provider = "route"
		o.CustomEnv = []string{"CUSTOM=1", "ROUTE_MODEL=custom-wins"}
	}},
	{"provider arguments", func(o *launch.Options) {
		o.Agent = "codex"
		o.Provider = "coder"
		o.Models = []string{"main=model-override"}
	}},
	{"credential", func(o *launch.Options) { o.Credentials = []string{"API_KEY=env:SECRET_SOURCE"} }},
	{"overlay", func(o *launch.Options) { o.Plugins = []string{"hello"} }},
	{"initialization commands", func(o *launch.Options) {
		o.Provision = "command"
		o.InitCommands = []string{"echo ready", "touch /tmp/marker"}
	}},
	{"binds", func(o *launch.Options) {
		o.BindPaths = []string{"/etc"}
		o.RoBindPaths = []string{"/lib"}
	}},
	{"checkout", func(o *launch.Options) { o.WorktreeBranch = "feature" }},
	{"network none", func(o *launch.Options) { o.Network = "none" }},
	{"network managed", func(o *launch.Options) {
		o.Network = "managed"
		o.AllowHosts = []string{"api.example.com"}
	}},
}

var goldenIsolations = []string{"none", "bwrap", "docker", "seatbelt", "docker-sandboxes", "openshell"}

type goldenMode int

const (
	modeDryRun goldenMode = iota
	modeRun
	modeTerminal
)

func (m goldenMode) String() string {
	return [...]string{"dry run", "run", "terminal"}[m]
}

// goldenCase runs one option set through the planner and the real isolator and
// returns what reached the host.
func goldenCase(isolation string, variant goldenVariant, mode goldenMode) string {
	host := newGoldenHost()
	var log, out bytes.Buffer
	var terminalCalls []string
	deps := launch.Dependencies{
		Host:        host,
		Log:         logging.New(&log, &log, false, false),
		Out:         &out,
		Sandboxes:   plugins.DefaultRegistry(host, logging.New(&log, &log, false, false)),
		Egress:      goldenEgress{},
		Credentials: plugins.DefaultCredentials(host),
		Overlays:    goldenOverlays{},
		Workspaces:  map[wsp.VCSType]wsshared.VCS{wsp.VCSGit: goldenVCS{}},
		Terminals:   map[termshared.TerminalType]termshared.TerminalExecutor{"tmux": goldenTerminal{&terminalCalls}},
		Validator:   goldenValidator,
	}
	options := launch.Options{
		Agent:                     "claude",
		Isolation:                 isolation,
		Provision:                 "none",
		Network:                   "host",
		Explicit:                  runconfig.KeysOf(runconfig.KeyAgent, runconfig.KeyIsolation, runconfig.KeyProvision, runconfig.KeyNetwork),
		IsolationBwrapPassthrough: isolation == "bwrap",
		Image:                     goldenImage,
		NixSource:                 "packages",
		NixShell:                  "default",
		VCS:                       "git",
		AgentArgs:                 []string{"review", "--depth", "2"},
	}
	variant.apply(&options)
	switch mode {
	case modeDryRun:
		options.DryRun = true
	case modeTerminal:
		options.Terminal = "tmux"
	}

	err := launch.Execute(deps, options)
	var lines []string
	lines = append(lines, fmt.Sprintf("## %s / %s / %s", isolation, variant.name, mode))
	if err != nil {
		lines = append(lines, "error: "+err.Error())
	}
	if mode == modeDryRun {
		lines = append(lines, "out: "+strings.TrimSpace(out.String()))
		for _, line := range strings.Split(strings.TrimSpace(log.String()), "\n") {
			lines = append(lines, "log: "+line)
		}
	}
	for _, call := range host.Calls {
		lines = append(lines, call.Kind+": "+call.Line())
		if call.Kind == "run" {
			lines = append(lines, "run env: "+strings.Join(sorted(call.Env), " "), "run dir: "+call.Dir)
		}
	}
	lines = append(lines, terminalCalls...)
	return strings.Join(lines, "\n") + "\n"
}

// TestLaunchGolden freezes, for every built-in isolator, what a dry run prints,
// what a run starts on the host and what a terminal wrapper receives, from the
// options of the command line down to the process.
func TestLaunchGolden(t *testing.T) {
	var got strings.Builder
	for _, isolation := range goldenIsolations {
		for _, variant := range goldenVariants {
			for _, mode := range []goldenMode{modeDryRun, modeRun, modeTerminal} {
				got.WriteString(goldenCase(isolation, variant, mode))
				got.WriteString("\n")
			}
		}
	}
	if *updateGolden {
		if err := os.WriteFile(goldenPath, []byte(got.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := goldenFiles.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s (run with -update-golden to generate): %v", goldenPath, err)
	}
	if got.String() != string(want) {
		t.Errorf("launch golden mismatch; run `go test -run TestLaunchGolden -update-golden` if the change is intended\n%s", firstDifference(string(want), got.String()))
	}
}

// firstDifference names the first line where got leaves want.
func firstDifference(want, got string) string {
	wantLines, gotLines := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := 0; i < len(wantLines) && i < len(gotLines); i++ {
		if wantLines[i] != gotLines[i] {
			return fmt.Sprintf("line %d:\nwant: %s\n got: %s", i+1, wantLines[i], gotLines[i])
		}
	}
	return fmt.Sprintf("line counts differ: want %d, got %d", len(wantLines), len(gotLines))
}
