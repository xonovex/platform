package launch

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	sharednix "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision/nix"
	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

const testRevision = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func baseOptions() Options {
	return Options{
		Agent:     "claude",
		Harnesses: agents.Embedded(),
		Isolation: "none",
		Provision: "none",
		Network:   "host",
		NixSource: "packages",
		NixShell:  "default",
		VCS:       "git",
	}
}

// fold merges the run section as the user layer under the command line, the
// way resolve does, so a plan test drives the same options as a run. An empty
// section is no user layer.
func fold(t *testing.T, options Options, section string) Options {
	t.Helper()
	var documents []layer.Document
	if section != "" {
		document, err := layer.Parse(spec.User, "test", []byte(`{"version":1,"run":`+section+`}`))
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		documents = append(documents, document)
	}
	merged, err := runconfig.Merge(documents, flagInput(options))
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	return applyResolved(options, merged)
}

func plan(t *testing.T, options Options, section string) (Plan, error) {
	t.Helper()
	if len(options.Catalog.List()) == 0 {
		options.Catalog = testCatalog(t)
	}
	return NewPlan(fold(t, options, section), "/work/repo", "/home/user", newTestSandbox().registry)
}

func mustPlan(t *testing.T, options Options, section string) Plan {
	t.Helper()
	p, err := plan(t, options, section)
	if err != nil {
		t.Fatalf("NewPlan() error = %v", err)
	}
	return p
}

func TestNewPlanResolvesTheDefaultRun(t *testing.T) {
	options := baseOptions()
	options.AgentArgs = []string{"review"}
	options.DryRun = true
	options.Verbose = true

	p := mustPlan(t, options, "")

	if p.Agent.Binary != "claude" || p.Provider != nil {
		t.Errorf("agent, provider = %q, %v; want claude and no provider", p.Agent.Binary, p.Provider)
	}
	if p.IsolationName != isolation.IsolationNone || p.Isolator == nil || p.Provisioner == nil {
		t.Errorf("selection = %s %v %v, want none with both plugins", p.IsolationName, p.Isolator, p.Provisioner)
	}
	if p.SourceDir != "/work/repo" || p.Run.Process.Dir != "/work/repo" || p.Run.RepoDir != "/work/repo" {
		t.Errorf("directories = %q %q %q, want /work/repo", p.SourceDir, p.Run.Process.Dir, p.Run.RepoDir)
	}
	if p.Run.HomeDir != "/home/user" || p.Run.Network != network.NetworkHost {
		t.Errorf("home, network = %q, %q", p.Run.HomeDir, p.Run.Network)
	}
	if !slices.Equal(p.Run.BindPaths, []string{"/work/repo"}) || p.Run.RoBindPaths != nil {
		t.Errorf("binds = %v %v, want only the source directory", p.Run.BindPaths, p.Run.RoBindPaths)
	}
	if !slices.Equal(p.AgentArgs, []string{"review"}) || p.Checkout != nil || p.Terminal != nil {
		t.Errorf("args, checkout, terminal = %v %v %v", p.AgentArgs, p.Checkout, p.Terminal)
	}
	if !slices.Equal(p.Run.Process.Argv, []string{"claude", "review"}) || len(p.Run.Process.Env) != 0 {
		t.Errorf("process = %+v, want claude review with no custom environment", p.Run.Process)
	}
	if !p.DryRun || !p.Verbose || p.DisplayDir() != "/work/repo" {
		t.Errorf("dry run, verbose, display = %v %v %q", p.DryRun, p.Verbose, p.DisplayDir())
	}
}

func TestNewPlanRejectsAnUnknownAgentAndListsTheAgents(t *testing.T) {
	options := baseOptions()
	options.Agent = "unknown"

	_, err := plan(t, options, "")

	if err == nil || !strings.Contains(err.Error(), "available agents") || !strings.Contains(err.Error(), "claude") {
		t.Fatalf("NewPlan() error = %v, want the available agents", err)
	}
}

func TestNewPlanPrefersTheProviderFlagOverTheFile(t *testing.T) {
	options := baseOptions()
	options.Provider = "alpha"

	p := mustPlan(t, options, `{"provider":"beta"}`)
	if p.Provider == nil || p.Provider.Name != "alpha" || p.Run.Provider != p.Provider {
		t.Fatalf("provider = %+v, want alpha", p.Provider)
	}

	fromFile := mustPlan(t, baseOptions(), `{"provider":"beta"}`)
	if fromFile.Provider == nil || fromFile.Provider.Name != "beta" {
		t.Fatalf("provider = %+v, want beta from the file", fromFile.Provider)
	}
}

func TestNewPlanRejectsAnUnknownProviderAndListsTheProviders(t *testing.T) {
	options := baseOptions()
	options.Provider = "unknown"

	_, err := plan(t, options, "")

	if err == nil || !strings.Contains(err.Error(), "provider unknown is not defined") || !strings.Contains(err.Error(), "alpha, beta") || !errors.Is(err, providers.ErrNotDefined) {
		t.Fatalf("NewPlan() error = %v, want the defined providers", err)
	}
}

func TestNewPlanResolvesTheWorkDirectoryAgainstTheWorkingDirectory(t *testing.T) {
	tests := []struct{ workDir, want string }{
		{"sub/dir", "/work/repo/sub/dir"},
		{"../other", "/work/other"},
		{"/abs/path/", "/abs/path"},
	}
	for _, test := range tests {
		options := baseOptions()
		options.WorkDir = test.workDir
		if p := mustPlan(t, options, ""); p.SourceDir != test.want {
			t.Errorf("SourceDir(%q) = %q, want %q", test.workDir, p.SourceDir, test.want)
		}
	}
}

func TestNewPlanValidatesTheCheckout(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*Options)
		phrase string
	}{
		{"branch", func(o *Options) { o.WorktreeBranch = "branch;rm -rf /" }, "invalid --worktree-branch"},
		{"source branch", func(o *Options) {
			o.WorktreeBranch = "feature"
			o.WorktreeSourceBranch = "main|evil"
		}, "invalid --worktree-source-branch"},
		{"vcs", func(o *Options) {
			o.WorktreeBranch = "feature"
			o.VCS = "svn"
		}, "unknown --vcs"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := baseOptions()
			test.modify(&options)
			if _, err := plan(t, options, ""); err == nil || !strings.Contains(err.Error(), test.phrase) {
				t.Fatalf("NewPlan() error = %v, want %q", err, test.phrase)
			}
		})
	}
}

func TestNewPlanPlacesTheDefaultCheckoutBesideTheSource(t *testing.T) {
	options := baseOptions()
	options.WorktreeBranch = "feature/x"
	options.WorktreeSourceBranch = "main"
	options.VCS = "jj"

	p := mustPlan(t, options, "")

	if p.Checkout == nil || p.Checkout.VCS != wsp.VCSJujutsu {
		t.Fatalf("Checkout = %+v, want a jj checkout", p.Checkout)
	}
	if p.Checkout.Config.Branch != "feature/x" || p.Checkout.Config.SourceBranch != "main" || !strings.HasPrefix(p.Checkout.Config.Dir, "../repo-") {
		t.Errorf("Checkout.Config = %+v", p.Checkout.Config)
	}
	if !strings.HasPrefix(p.Checkout.Dir, "/work/repo-") || p.DisplayDir() != p.Checkout.Dir {
		t.Errorf("Checkout.Dir = %q, DisplayDir = %q", p.Checkout.Dir, p.DisplayDir())
	}
	if p.Run.Process.Dir != "/work/repo" {
		t.Errorf("Run.Process.Dir = %q, want the source until the checkout exists", p.Run.Process.Dir)
	}
}

func TestNewPlanKeepsAnExplicitCheckoutDirectory(t *testing.T) {
	options := baseOptions()
	options.WorktreeBranch = "feature"
	options.WorktreeDir = "trees/feature"

	p := mustPlan(t, options, "")

	if p.Checkout.Config.Dir != "trees/feature" || p.Checkout.Dir != "/work/repo/trees/feature" {
		t.Errorf("Checkout = %+v", p.Checkout)
	}
}

func TestNewPlanSelectsThePinnedPairOnlyWithoutAnExplicitAxis(t *testing.T) {
	options := baseOptions()
	options.NixRev = testRevision
	options.Policy.RequirePinnedProvision = true

	p := mustPlan(t, options, "")
	if p.IsolationName != isolation.IsolationBwrap || !p.Provisioner.Pinned() {
		t.Fatalf("pinned default = %s, pinned %v; want bwrap and nix", p.IsolationName, p.Provisioner.Pinned())
	}
	if p.Provision.NixSource.Kind != sharednix.NixSourcePackages {
		t.Errorf("NixSource = %+v, want a packages source", p.Provision.NixSource)
	}

	options.Explicit = runconfig.KeysOf(runconfig.KeyIsolation)
	if _, err := plan(t, options, ""); !errors.Is(err, policy.ErrPinnedProvisionUnmet) {
		t.Fatalf("explicit isolation error = %v, want %v", err, policy.ErrPinnedProvisionUnmet)
	}

	options.Explicit = runconfig.KeysOf(runconfig.KeyProvision)
	if _, err := plan(t, options, ""); !errors.Is(err, policy.ErrPinnedProvisionUnmet) {
		t.Fatalf("explicit provision error = %v, want %v", err, policy.ErrPinnedProvisionUnmet)
	}
}

func TestNewPlanRejectsTheRetiredProxyNetwork(t *testing.T) {
	options := baseOptions()
	options.Isolation = "bwrap"
	options.Network = "proxy"

	var invalid *network.InvalidModeError
	if _, err := plan(t, options, ""); !errors.As(err, &invalid) {
		t.Fatalf("NewPlan() error = %v, want an invalid network mode", err)
	}
}

func TestNewPlanBuildsTheManagedRoute(t *testing.T) {
	options := baseOptions()
	options.Isolation = "bwrap"
	options.Network = "managed"
	options.AllowHosts = []string{"api.example.com"}
	p := mustPlan(t, options, `{"allowHosts":["*.cdn.example.com:8443"]}`)
	if p.Run.Network != network.NetworkManaged || p.Run.Egress.Transport != network.TransportUnixSocket {
		t.Errorf("network, transport = %q, %q", p.Run.Network, p.Run.Egress.Transport)
	}
	if len(p.Allow.Rules()) != 2 || !p.Allow.Allows("api.example.com", 443) || !p.Allow.Allows("x.cdn.example.com", 8443) {
		t.Errorf("allowlist = %v", p.Allow.Rules())
	}
}

func TestNewPlanManagedNetworkFailsClosed(t *testing.T) {
	options := baseOptions()
	options.Isolation = "bwrap"
	options.Network = "managed"
	if _, err := plan(t, options, ""); err == nil || !strings.Contains(err.Error(), "--allow-host") {
		t.Errorf("managed without hosts error = %v, want a hint for --allow-host", err)
	}
	options.AllowHosts = []string{"https://api.example.com/path"}
	if _, err := plan(t, options, ""); err == nil {
		t.Error("an allowed host with a scheme must fail")
	}
	options.AllowHosts = []string{"api.example.com"}
	options.Isolation = "none"
	if _, err := plan(t, options, ""); err == nil || !strings.Contains(err.Error(), "cannot restrict network egress") {
		t.Errorf("managed on an isolation without a transport error = %v", err)
	}
}

func TestNewPlanRefusesAllowedHostsWithoutAManagedNetwork(t *testing.T) {
	options := baseOptions()
	options.AllowHosts = []string{"api.example.com"}
	if _, err := plan(t, options, ""); err == nil || !strings.Contains(err.Error(), "--network managed") {
		t.Errorf("error = %v, want a --network managed hint", err)
	}
}

func TestNewPlanParsesCredentialBindings(t *testing.T) {
	options := baseOptions()
	options.Credentials = []string{"API_KEY=env:SOURCE_KEY"}
	p := mustPlan(t, options, `{"credentials":["TOKEN=pass:work/token"]}`)
	if len(p.Credentials) != 2 || p.Credentials[0].Name != "TOKEN" || p.Credentials[1].Scheme != "env" {
		t.Errorf("credentials = %+v", p.Credentials)
	}
	options.Credentials = []string{"TOKEN=env:X"}
	if _, err := plan(t, options, `{"credentials":["TOKEN=pass:y"]}`); err == nil {
		t.Error("a variable bound by both the file and the flag must fail")
	}
	options.Credentials = []string{"not-a-binding"}
	if _, err := plan(t, options, ""); err == nil {
		t.Error("an invalid binding must fail")
	}
}

func TestNewPlanRejectsARestrictedNetworkForAnIsolationThatCannotConfineIt(t *testing.T) {
	options := baseOptions()
	options.Network = "none"

	if _, err := plan(t, options, ""); err == nil || !strings.Contains(err.Error(), "cannot restrict network egress") {
		t.Fatalf("NewPlan() error = %v, want a confinement error", err)
	}

	options.Isolation = "bwrap"
	if p := mustPlan(t, options, ""); p.Run.Network != network.NetworkNone {
		t.Errorf("Run.Network = %q, want none", p.Run.Network)
	}
}

func TestNewPlanPassesTheDockerRuntimeToTheKernelPolicy(t *testing.T) {
	options := baseOptions()
	options.Isolation = "docker"
	options.Policy.RequireKernelIsolation = true

	if _, err := plan(t, options, ""); !errors.Is(err, policy.ErrKernelIsolationUnmet) {
		t.Fatalf("NewPlan(runc) error = %v, want %v", err, policy.ErrKernelIsolationUnmet)
	}

	options.IsolationDockerRuntime = "runsc"
	p := mustPlan(t, options, "")
	if p.Run.Runtime != "runsc" || !p.Isolator.KernelIsolated(p.Run.Runtime) {
		t.Errorf("Run.Runtime = %q, want a kernel-isolated runsc run", p.Run.Runtime)
	}
}

func TestNewPlanRequiresADigestImageForAPinnedDockerRun(t *testing.T) {
	options := baseOptions()
	options.Isolation = "docker"
	options.Explicit = runconfig.KeysOf(runconfig.KeyIsolation, runconfig.KeyProvision)
	options.Policy.RequirePinnedProvision = true

	options.Image = "ghcr.io/xonovex/agent:latest"
	if _, err := plan(t, options, ""); !errors.Is(err, policy.ErrPinnedProvisionUnmet) {
		t.Fatalf("NewPlan(mutable image) error = %v, want %v", err, policy.ErrPinnedProvisionUnmet)
	}

	options.Image = "ghcr.io/xonovex/agent@sha256:" + strings.Repeat("a", 64)
	if p := mustPlan(t, options, ""); p.Run.Image != options.Image {
		t.Errorf("Run.Image = %q, want %q", p.Run.Image, options.Image)
	}
}

func TestNewPlanCountsFileBindsAsCustomBinds(t *testing.T) {
	options := baseOptions()
	options.Isolation = "bwrap"
	options.Policy.RequireHostToolsUnreachable = true

	mustPlan(t, options, "")
	if _, err := plan(t, options, `{"roBindPaths":["/data"]}`); !errors.Is(err, policy.ErrHostToolsReachable) {
		t.Fatalf("NewPlan(file bind) error = %v, want %v", err, policy.ErrHostToolsReachable)
	}
}

func TestNewPlanKeepsInitCommandsWithoutNix(t *testing.T) {
	options := baseOptions()
	options.Provision = "command"
	options.InitCommands = []string{"npm install"}
	options.NixSource = "unknown"

	p := mustPlan(t, options, "")

	if !slices.Equal(p.Provision.InitCommands, []string{"npm install"}) || p.Provision.NixSource.Kind != "" {
		t.Errorf("Provision = %+v, want only the init commands", p.Provision)
	}
}

func TestNewPlanRejectsAnInvalidNixSource(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*Options)
		phrase string
	}{
		{"unknown kind", func(o *Options) { o.NixSource = "unknown" }, `unknown nix source "unknown"`},
		{"mutable revision", func(o *Options) { o.NixRev = "nixos-unstable" }, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := baseOptions()
			options.Provision = "nix"
			options.Isolation = "bwrap"
			options.NixRev = testRevision
			test.modify(&options)
			if _, err := plan(t, options, ""); err == nil || !strings.Contains(err.Error(), test.phrase) {
				t.Fatalf("NewPlan() error = %v, want %q", err, test.phrase)
			}
		})
	}
}

func TestNewPlanAddsTheAgentPackageOnce(t *testing.T) {
	tests := []struct {
		name  string
		agent string
		extra []string
		want  []string
	}{
		{"appended", "opencode", []string{"ripgrep"}, []string{"ripgrep", "opencode"}},
		{"already listed", "claude", []string{"claude-code"}, []string{"claude-code"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := baseOptions()
			options.Agent = test.agent
			options.Isolation = "bwrap"
			options.Provision = "nix"
			options.NixSource = ""
			options.NixRev = testRevision
			options.NixPackages = test.extra

			p := mustPlan(t, options, "")

			if !slices.Equal(p.Provision.NixSource.Packages, test.want) || p.Provision.NixSource.Rev != testRevision {
				t.Errorf("NixSource = %+v, want packages %v", p.Provision.NixSource, test.want)
			}
			if !slices.Equal(options.NixPackages, test.extra) {
				t.Errorf("options.NixPackages changed to %v", options.NixPackages)
			}
		})
	}
}

func TestNewPlanUsesTheSourceFlakeAndTheDefaultShell(t *testing.T) {
	options := baseOptions()
	options.Isolation = "bwrap"
	options.Provision = "nix"
	options.NixSource = "flake"
	options.NixShell = ""
	options.WorkDir = "project"

	p := mustPlan(t, options, "")

	source := p.Provision.NixSource
	if source.Kind != sharednix.NixSourceProjectFlake || source.FlakeRef != "/work/repo/project" || source.Shell != "default" {
		t.Errorf("NixSource = %+v, want the default shell of the source flake", source)
	}

	options.NixShell = "dev"
	if p := mustPlan(t, options, ""); p.Provision.NixSource.Shell != "dev" {
		t.Errorf("Shell = %q, want dev", p.Provision.NixSource.Shell)
	}
}

func TestNewPlanRejectsAMalformedCustomEnvironment(t *testing.T) {
	options := baseOptions()
	options.CustomEnv = []string{"MISSING_EQUALS"}

	if _, err := plan(t, options, ""); err == nil || !strings.Contains(err.Error(), "must use KEY=VALUE") {
		t.Fatalf("NewPlan() error = %v, want a custom environment error", err)
	}
}

func TestNewPlanMergesTheCustomEnvironment(t *testing.T) {
	options := baseOptions()
	options.CustomEnv = []string{"B=flag", "C=3"}

	p := mustPlan(t, options, `{"customEnv":["A=1","B=file"]}`)

	for _, want := range []string{"A=1", "B=flag", "C=3"} {
		if !slices.Contains(p.Run.Process.Env, want) {
			t.Errorf("Process.Env = %v, want %s", p.Run.Process.Env, want)
		}
	}
}

func TestNewPlanPrefersTheFileHomeAndMakesItAbsolute(t *testing.T) {
	if p := mustPlan(t, baseOptions(), `{"homeDir":"/srv/home"}`); p.Run.HomeDir != "/srv/home" {
		t.Errorf("HomeDir = %q, want /srv/home", p.Run.HomeDir)
	}
	if p := mustPlan(t, baseOptions(), `{"homeDir":"sandbox-home"}`); p.Run.HomeDir != "/work/repo/sandbox-home" {
		t.Errorf("HomeDir = %q, want /work/repo/sandbox-home", p.Run.HomeDir)
	}
}

func TestNewPlanTakesTheImageFromTheRunSection(t *testing.T) {
	const fileImage, flagImage = "registry.example/file@sha256:1111", "registry.example/flag@sha256:2222"
	options := baseOptions()
	options.Isolation = "openshell"

	if p := mustPlan(t, options, `{"image":"`+fileImage+`"}`); p.Run.Image != fileImage {
		t.Errorf("Image = %q, want the image of the file", p.Run.Image)
	}
	options.Image = flagImage
	if p := mustPlan(t, options, `{"image":"`+fileImage+`"}`); p.Run.Image != flagImage {
		t.Errorf("Image = %q, want the image of the command line", p.Run.Image)
	}
}

func TestNewPlanNamesTheImageSettingOfAnImageIsolationWithoutAnImage(t *testing.T) {
	options := baseOptions()
	options.Isolation = "openshell"

	_, err := plan(t, options, "")
	if !errors.Is(err, isoshared.ErrNoImage) || !strings.Contains(err.Error(), "set run.image in the user or managed agentic.json, or pass --image") {
		t.Fatalf("NewPlan() error = %v, want ErrNoImage naming run.image and --image", err)
	}
}

func TestNewPlanMakesEveryBindAbsolute(t *testing.T) {
	options := baseOptions()
	options.BindPaths = []string{"rw-flag"}
	options.RoBindPaths = []string{"/ro/flag", "ro-flag/"}

	p := mustPlan(t, options, `{"bindPaths":["rw-file"],"roBindPaths":["ro-file"]}`)

	wantRW := []string{"/work/repo/rw-file", "/work/repo/rw-flag", "/work/repo"}
	wantRO := []string{"/work/repo/ro-file", "/ro/flag", "/work/repo/ro-flag"}
	if !slices.Equal(p.Run.BindPaths, wantRW) || !slices.Equal(p.Run.RoBindPaths, wantRO) {
		t.Errorf("binds = %v %v, want %v %v", p.Run.BindPaths, p.Run.RoBindPaths, wantRW, wantRO)
	}
}

func TestNewPlanConfiguresTheTerminal(t *testing.T) {
	options := baseOptions()
	options.Terminal = "tmux"
	options.TerminalSession = "session"
	options.TerminalWindow = "window"
	options.TerminalDetach = true

	p := mustPlan(t, options, "")

	want := termshared.TerminalConfig{Type: termshared.TerminalTmux, SessionName: "session", WindowName: "window", Detach: true}
	if p.Terminal == nil || *p.Terminal != want {
		t.Errorf("Terminal = %+v, want %+v", p.Terminal, want)
	}
}

func TestNewPlanRejectsAnUnregisteredAxis(t *testing.T) {
	options := baseOptions()
	options.Provision = "apt"

	if _, err := plan(t, options, ""); err == nil || !strings.Contains(err.Error(), "apt") {
		t.Fatalf("NewPlan() error = %v, want an unregistered provision error", err)
	}
}

func TestNewPlanDefaultsEmptyAxes(t *testing.T) {
	p := mustPlan(t, Options{Agent: "claude", Harnesses: agents.Embedded(), VCS: "git"}, "")

	if p.IsolationName != isolation.IsolationNone || p.Run.Network != network.NetworkHost {
		t.Errorf("defaults = %s %s, want none and host", p.IsolationName, p.Run.Network)
	}
	if _, ok := any(p.Provisioner).(*fakeProvisioner); !ok || p.Provisioner.Pinned() {
		t.Errorf("Provisioner = %v, want the unpinned none provisioner", p.Provisioner)
	}
}

func TestNewPlanSupportsEveryHarness(t *testing.T) {
	tests := []struct {
		agent  string
		binary string
		nix    string
	}{
		{"claude", "claude", "claude-code"},
		{"codex", "codex", "codex"},
		{"copilot", "copilot", "github-copilot-cli"},
		{"kiro", "kiro-cli-chat", "kiro-cli-unwrapped"},
		{"opencode", "opencode", "opencode"},
		{"pi", "pi", "pi-coding-agent"},
	}
	for _, test := range tests {
		t.Run(test.agent, func(t *testing.T) {
			options := baseOptions()
			options.Agent = test.agent

			p := mustPlan(t, options, "")

			if p.Agent.Binary != test.binary || p.Agent.NixPackage != test.nix {
				t.Fatalf("agent = %+v, want binary %q and nix package %q", p.Agent, test.binary, test.nix)
			}
			if p.Provider != nil {
				t.Fatalf("provider = %+v, want none", p.Provider)
			}
		})
	}
}

func TestNewPlanListsEveryHarnessForAnUnknownAgent(t *testing.T) {
	options := baseOptions()
	options.Agent = "unknown"

	_, err := plan(t, options, "")

	for _, want := range []string{"claude", "codex", "copilot", "kiro", "opencode", "pi"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("NewPlan() error = %v, want it to list %q", err, want)
		}
	}
}

func TestNewPlanResolvesTheCodexProvider(t *testing.T) {
	options := baseOptions()
	options.Agent = "codex"
	options.Provider = "gamma"

	p := mustPlan(t, options, "")

	if p.Provider == nil || p.Provider.Name != "gamma" || p.Provider.AgentType != "codex" {
		t.Fatalf("provider = %+v, want the codex provider gamma", p.Provider)
	}
}

func TestNewPlanRejectsAProviderForAnAgentThatUsesItsOwnLogin(t *testing.T) {
	for _, agent := range []string{"copilot", "kiro"} {
		for name, fromFlag := range map[string]bool{"flag": true, "file": false} {
			t.Run(agent+"/"+name, func(t *testing.T) {
				options := baseOptions()
				options.Agent = agent
				section := ""
				if fromFlag {
					options.Provider = "alpha"
				} else {
					section = `{"provider":"alpha"}`
				}

				_, err := plan(t, options, section)

				if err == nil || !strings.Contains(err.Error(), "uses its own login") {
					t.Fatalf("NewPlan() error = %v, want a native login rejection", err)
				}
			})
		}
	}
}

func TestNewPlanRejectsAProviderThatIsDefinedForAnotherAgent(t *testing.T) {
	options := baseOptions()
	options.Agent = "pi"
	options.Provider = "alpha"

	_, err := plan(t, options, "")

	if err == nil || !strings.Contains(err.Error(), "provider alpha is not defined") {
		t.Fatalf("NewPlan() error = %v, want an unknown provider error", err)
	}
}

func TestNewPlanKeepsTheRequirePoliciesForEveryHarness(t *testing.T) {
	for _, agent := range []string{"codex", "copilot", "kiro", "pi"} {
		t.Run(agent, func(t *testing.T) {
			options := baseOptions()
			options.Agent = agent
			options.Policy.RequireKernelIsolation = true

			if _, err := plan(t, options, ""); err == nil {
				t.Fatal("NewPlan() error = nil, want the kernel isolation guarantee refused on isolation none")
			}
		})
	}
}
