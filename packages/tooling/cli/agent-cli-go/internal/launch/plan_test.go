package launch

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	sharednix "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision/nix"
	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

const testRevision = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func baseOptions() Options {
	return Options{
		Agent:     "claude",
		Isolation: "none",
		Provision: "none",
		Network:   "host",
		NixSource: "packages",
		NixShell:  "default",
		VCS:       "git",
	}
}

// fold merges the file as the user layer under the command line, the way resolve
// does, so a plan test drives the same options as a run.
func fold(t *testing.T, options Options, file config.FileConfig) Options {
	t.Helper()
	merged, err := config.Merge([]config.Input{
		{Layer: config.LayerUser, Origin: "test", Entries: file.Entries()},
		flagInput(options),
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	return applyResolved(options, merged)
}

func plan(t *testing.T, options Options, file config.FileConfig) (Plan, error) {
	t.Helper()
	return NewPlan(fold(t, options, file), "/work/repo", "/home/user", newTestSandbox().registry)
}

func mustPlan(t *testing.T, options Options, file config.FileConfig) Plan {
	t.Helper()
	p, err := plan(t, options, file)
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

	p := mustPlan(t, options, config.FileConfig{})

	if p.Agent.Binary != "claude" || p.Provider != nil {
		t.Errorf("agent, provider = %q, %v; want claude and no provider", p.Agent.Binary, p.Provider)
	}
	if p.IsolationName != isolation.IsolationNone || p.Isolator == nil || p.Provisioner == nil {
		t.Errorf("selection = %s %v %v, want none with both plugins", p.IsolationName, p.Isolator, p.Provisioner)
	}
	if p.SourceDir != "/work/repo" || p.Run.WorkDir != "/work/repo" || p.Run.RepoDir != "/work/repo" {
		t.Errorf("directories = %q %q %q, want /work/repo", p.SourceDir, p.Run.WorkDir, p.Run.RepoDir)
	}
	if p.Run.HomeDir != "/home/user" || p.Run.Network != netshared.ModeHost {
		t.Errorf("home, network = %q, %q", p.Run.HomeDir, p.Run.Network)
	}
	if !slices.Equal(p.Run.BindPaths, []string{"/work/repo"}) || p.Run.RoBindPaths != nil {
		t.Errorf("binds = %v %v, want only the source directory", p.Run.BindPaths, p.Run.RoBindPaths)
	}
	if !slices.Equal(p.Run.AgentArgs, []string{"review"}) || p.Checkout != nil || p.Terminal != nil {
		t.Errorf("args, checkout, terminal = %v %v %v", p.Run.AgentArgs, p.Checkout, p.Terminal)
	}
	if !p.DryRun || !p.Verbose || p.DisplayDir() != "/work/repo" {
		t.Errorf("dry run, verbose, display = %v %v %q", p.DryRun, p.Verbose, p.DisplayDir())
	}
}

func TestNewPlanRejectsAnUnknownAgentAndListsTheAgents(t *testing.T) {
	options := baseOptions()
	options.Agent = "unknown"

	_, err := plan(t, options, config.FileConfig{})

	if err == nil || !strings.Contains(err.Error(), "available agents") || !strings.Contains(err.Error(), "claude") {
		t.Fatalf("NewPlan() error = %v, want the available agents", err)
	}
}

func TestNewPlanPrefersTheProviderFlagOverTheFile(t *testing.T) {
	options := baseOptions()
	options.Provider = "gemini"

	p := mustPlan(t, options, config.FileConfig{Provider: "glm"})
	if p.Provider == nil || p.Provider.Name != "gemini" || p.Run.Provider != p.Provider {
		t.Fatalf("provider = %+v, want gemini", p.Provider)
	}

	fromFile := mustPlan(t, baseOptions(), config.FileConfig{Provider: "glm"})
	if fromFile.Provider == nil || fromFile.Provider.Name != "glm" {
		t.Fatalf("provider = %+v, want glm from the file", fromFile.Provider)
	}
}

func TestNewPlanRejectsAnUnknownProviderAndListsTheProviders(t *testing.T) {
	options := baseOptions()
	options.Provider = "unknown"

	_, err := plan(t, options, config.FileConfig{})

	if err == nil || !strings.Contains(err.Error(), "available providers") || !strings.Contains(err.Error(), "gemini") {
		t.Fatalf("NewPlan() error = %v, want the available providers", err)
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
		if p := mustPlan(t, options, config.FileConfig{}); p.SourceDir != test.want {
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
			if _, err := plan(t, options, config.FileConfig{}); err == nil || !strings.Contains(err.Error(), test.phrase) {
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

	p := mustPlan(t, options, config.FileConfig{})

	if p.Checkout == nil || p.Checkout.VCS != wsp.VCSJujutsu {
		t.Fatalf("Checkout = %+v, want a jj checkout", p.Checkout)
	}
	if p.Checkout.Config.Branch != "feature/x" || p.Checkout.Config.SourceBranch != "main" || !strings.HasPrefix(p.Checkout.Config.Dir, "../repo-") {
		t.Errorf("Checkout.Config = %+v", p.Checkout.Config)
	}
	if !strings.HasPrefix(p.Checkout.Dir, "/work/repo-") || p.DisplayDir() != p.Checkout.Dir {
		t.Errorf("Checkout.Dir = %q, DisplayDir = %q", p.Checkout.Dir, p.DisplayDir())
	}
	if p.Run.WorkDir != "/work/repo" {
		t.Errorf("Run.WorkDir = %q, want the source until the checkout exists", p.Run.WorkDir)
	}
}

func TestNewPlanKeepsAnExplicitCheckoutDirectory(t *testing.T) {
	options := baseOptions()
	options.WorktreeBranch = "feature"
	options.WorktreeDir = "trees/feature"

	p := mustPlan(t, options, config.FileConfig{})

	if p.Checkout.Config.Dir != "trees/feature" || p.Checkout.Dir != "/work/repo/trees/feature" {
		t.Errorf("Checkout = %+v", p.Checkout)
	}
}

func TestNewPlanSelectsThePinnedPairOnlyWithoutAnExplicitAxis(t *testing.T) {
	options := baseOptions()
	options.NixRev = testRevision
	options.Policy.RequirePinnedProvision = true

	p := mustPlan(t, options, config.FileConfig{})
	if p.IsolationName != isolation.IsolationBwrap || !p.Provisioner.Pinned() {
		t.Fatalf("pinned default = %s, pinned %v; want bwrap and nix", p.IsolationName, p.Provisioner.Pinned())
	}
	if p.Provision.NixSource.Kind != sharednix.NixSourcePackages {
		t.Errorf("NixSource = %+v, want a packages source", p.Provision.NixSource)
	}

	options.Explicit = options.Explicit.With(config.KeyIsolation)
	if _, err := plan(t, options, config.FileConfig{}); !errors.Is(err, policy.ErrPinnedProvisionUnmet) {
		t.Fatalf("explicit isolation error = %v, want %v", err, policy.ErrPinnedProvisionUnmet)
	}

	options.Explicit = nil
	options.Explicit = options.Explicit.With(config.KeyProvision)
	if _, err := plan(t, options, config.FileConfig{}); !errors.Is(err, policy.ErrPinnedProvisionUnmet) {
		t.Fatalf("explicit provision error = %v, want %v", err, policy.ErrPinnedProvisionUnmet)
	}
}

func TestNewPlanRejectsTheRetiredProxyNetwork(t *testing.T) {
	options := baseOptions()
	options.Isolation = "bwrap"
	options.Network = "proxy"

	var invalid *netshared.InvalidModeError
	if _, err := plan(t, options, config.FileConfig{}); !errors.As(err, &invalid) {
		t.Fatalf("NewPlan() error = %v, want an invalid network mode", err)
	}
}

func TestNewPlanBuildsTheManagedRoute(t *testing.T) {
	options := baseOptions()
	options.Isolation = "bwrap"
	options.Network = "managed"
	options.AllowHosts = []string{"api.example.com"}
	p := mustPlan(t, options, config.FileConfig{AllowHosts: []string{"*.cdn.example.com:8443"}})
	if p.Run.Network != netshared.ModeManaged || p.Run.Egress.Transport != netshared.TransportUnixSocket {
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
	if _, err := plan(t, options, config.FileConfig{}); err == nil || !strings.Contains(err.Error(), "--allow-host") {
		t.Errorf("managed without hosts error = %v, want a hint for --allow-host", err)
	}
	options.AllowHosts = []string{"https://api.example.com/path"}
	if _, err := plan(t, options, config.FileConfig{}); err == nil {
		t.Error("an allowed host with a scheme must fail")
	}
	options.AllowHosts = []string{"api.example.com"}
	options.Isolation = "none"
	if _, err := plan(t, options, config.FileConfig{}); err == nil || !strings.Contains(err.Error(), "cannot restrict network egress") {
		t.Errorf("managed on an isolation without a transport error = %v", err)
	}
}

func TestNewPlanRefusesAllowedHostsWithoutAManagedNetwork(t *testing.T) {
	options := baseOptions()
	options.AllowHosts = []string{"api.example.com"}
	if _, err := plan(t, options, config.FileConfig{}); err == nil || !strings.Contains(err.Error(), "--network managed") {
		t.Errorf("error = %v, want a --network managed hint", err)
	}
}

func TestNewPlanParsesCredentialBindings(t *testing.T) {
	options := baseOptions()
	options.Credentials = []string{"API_KEY=env:SOURCE_KEY"}
	p := mustPlan(t, options, config.FileConfig{Credentials: []string{"TOKEN=pass:work/token"}})
	if len(p.Credentials) != 2 || p.Credentials[0].Name != "TOKEN" || p.Credentials[1].Scheme != "env" {
		t.Errorf("credentials = %+v", p.Credentials)
	}
	options.Credentials = []string{"TOKEN=env:X"}
	if _, err := plan(t, options, config.FileConfig{Credentials: []string{"TOKEN=pass:y"}}); err == nil {
		t.Error("a variable bound by both the file and the flag must fail")
	}
	options.Credentials = []string{"not-a-binding"}
	if _, err := plan(t, options, config.FileConfig{}); err == nil {
		t.Error("an invalid binding must fail")
	}
}

func TestNewPlanRejectsARestrictedNetworkForAnIsolationThatCannotConfineIt(t *testing.T) {
	options := baseOptions()
	options.Network = "none"

	if _, err := plan(t, options, config.FileConfig{}); err == nil || !strings.Contains(err.Error(), "cannot restrict network egress") {
		t.Fatalf("NewPlan() error = %v, want a confinement error", err)
	}

	options.Isolation = "bwrap"
	if p := mustPlan(t, options, config.FileConfig{}); p.Run.Network != netshared.ModeNone {
		t.Errorf("Run.Network = %q, want none", p.Run.Network)
	}
}

func TestNewPlanPassesTheDockerRuntimeToTheKernelPolicy(t *testing.T) {
	options := baseOptions()
	options.Isolation = "docker"
	options.Policy.RequireKernelIsolation = true

	if _, err := plan(t, options, config.FileConfig{}); !errors.Is(err, policy.ErrKernelIsolationUnmet) {
		t.Fatalf("NewPlan(runc) error = %v, want %v", err, policy.ErrKernelIsolationUnmet)
	}

	options.IsolationDockerRuntime = "runsc"
	p := mustPlan(t, options, config.FileConfig{})
	if p.Run.Runtime != "runsc" || !p.Isolator.KernelIsolated(p.Run.Runtime) {
		t.Errorf("Run.Runtime = %q, want a kernel-isolated runsc run", p.Run.Runtime)
	}
}

func TestNewPlanRequiresADigestImageForAPinnedDockerRun(t *testing.T) {
	options := baseOptions()
	options.Isolation = "docker"
	options.Explicit = options.Explicit.With(config.KeyIsolation)
	options.Explicit = options.Explicit.With(config.KeyProvision)
	options.Policy.RequirePinnedProvision = true

	options.Image = "ghcr.io/xonovex/agent:latest"
	if _, err := plan(t, options, config.FileConfig{}); !errors.Is(err, policy.ErrPinnedProvisionUnmet) {
		t.Fatalf("NewPlan(mutable image) error = %v, want %v", err, policy.ErrPinnedProvisionUnmet)
	}

	options.Image = "ghcr.io/xonovex/agent@sha256:" + strings.Repeat("a", 64)
	if p := mustPlan(t, options, config.FileConfig{}); p.Run.Image != options.Image {
		t.Errorf("Run.Image = %q, want %q", p.Run.Image, options.Image)
	}
}

func TestNewPlanCountsFileBindsAsCustomBinds(t *testing.T) {
	options := baseOptions()
	options.Isolation = "bwrap"
	options.Policy.RequireHostToolsUnreachable = true

	mustPlan(t, options, config.FileConfig{})
	if _, err := plan(t, options, config.FileConfig{RoBindPaths: []string{"/data"}}); !errors.Is(err, policy.ErrHostToolsReachable) {
		t.Fatalf("NewPlan(file bind) error = %v, want %v", err, policy.ErrHostToolsReachable)
	}
}

func TestNewPlanKeepsInitCommandsWithoutNix(t *testing.T) {
	options := baseOptions()
	options.Provision = "command"
	options.InitCommands = []string{"npm install"}
	options.NixSource = "unknown"

	p := mustPlan(t, options, config.FileConfig{})

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
			if _, err := plan(t, options, config.FileConfig{}); err == nil || !strings.Contains(err.Error(), test.phrase) {
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

			p := mustPlan(t, options, config.FileConfig{})

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

	p := mustPlan(t, options, config.FileConfig{})

	source := p.Provision.NixSource
	if source.Kind != sharednix.NixSourceProjectFlake || source.FlakeRef != "/work/repo/project" || source.Shell != "default" {
		t.Errorf("NixSource = %+v, want the default shell of the source flake", source)
	}

	options.NixShell = "dev"
	if p := mustPlan(t, options, config.FileConfig{}); p.Provision.NixSource.Shell != "dev" {
		t.Errorf("Shell = %q, want dev", p.Provision.NixSource.Shell)
	}
}

func TestNewPlanRejectsAMalformedCustomEnvironment(t *testing.T) {
	options := baseOptions()
	options.CustomEnv = []string{"MISSING_EQUALS"}

	if _, err := plan(t, options, config.FileConfig{}); err == nil || !strings.Contains(err.Error(), "must use KEY=VALUE") {
		t.Fatalf("NewPlan() error = %v, want a custom environment error", err)
	}
}

func TestNewPlanMergesTheCustomEnvironment(t *testing.T) {
	options := baseOptions()
	options.CustomEnv = []string{"B=flag", "C=3"}

	p := mustPlan(t, options, config.FileConfig{CustomEnv: []string{"A=1", "B=file"}})

	for _, want := range []string{"A=1", "B=flag", "C=3"} {
		if !slices.Contains(p.Run.CustomEnv, want) {
			t.Errorf("CustomEnv = %v, want %s", p.Run.CustomEnv, want)
		}
	}
}

func TestNewPlanPrefersTheFileHomeAndMakesItAbsolute(t *testing.T) {
	if p := mustPlan(t, baseOptions(), config.FileConfig{HomeDir: "/srv/home"}); p.Run.HomeDir != "/srv/home" {
		t.Errorf("HomeDir = %q, want /srv/home", p.Run.HomeDir)
	}
	if p := mustPlan(t, baseOptions(), config.FileConfig{HomeDir: "sandbox-home"}); p.Run.HomeDir != "/work/repo/sandbox-home" {
		t.Errorf("HomeDir = %q, want /work/repo/sandbox-home", p.Run.HomeDir)
	}
}

func TestNewPlanMakesEveryBindAbsolute(t *testing.T) {
	options := baseOptions()
	options.BindPaths = []string{"rw-flag"}
	options.RoBindPaths = []string{"/ro/flag", "ro-flag/"}

	p := mustPlan(t, options, config.FileConfig{BindPaths: []string{"rw-file"}, RoBindPaths: []string{"ro-file"}})

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

	p := mustPlan(t, options, config.FileConfig{})

	want := termshared.TerminalConfig{Type: termshared.TerminalTmux, SessionName: "session", WindowName: "window", Detach: true}
	if p.Terminal == nil || *p.Terminal != want {
		t.Errorf("Terminal = %+v, want %+v", p.Terminal, want)
	}
}

func TestNewPlanRejectsAnUnregisteredAxis(t *testing.T) {
	options := baseOptions()
	options.Provision = "apt"

	if _, err := plan(t, options, config.FileConfig{}); err == nil || !strings.Contains(err.Error(), "apt") {
		t.Fatalf("NewPlan() error = %v, want an unregistered provision error", err)
	}
}

func TestNewPlanDefaultsEmptyAxes(t *testing.T) {
	p := mustPlan(t, Options{Agent: "claude", VCS: "git"}, config.FileConfig{})

	if p.IsolationName != isolation.IsolationNone || p.Run.Network != netshared.ModeHost {
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

			p := mustPlan(t, options, config.FileConfig{})

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

	_, err := plan(t, options, config.FileConfig{})

	for _, want := range []string{"claude", "codex", "copilot", "kiro", "opencode", "pi"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("NewPlan() error = %v, want it to list %q", err, want)
		}
	}
}

func TestNewPlanResolvesTheCodexProvider(t *testing.T) {
	options := baseOptions()
	options.Agent = "codex"
	options.Provider = "gpt5-codex"

	p := mustPlan(t, options, config.FileConfig{})

	if p.Provider == nil || p.Provider.Name != "gpt5-codex" || p.Provider.AgentType != "codex" {
		t.Fatalf("provider = %+v, want the codex gpt5-codex preset", p.Provider)
	}
}

func TestNewPlanRejectsAProviderForAnAgentThatUsesItsOwnLogin(t *testing.T) {
	for _, agent := range []string{"copilot", "kiro"} {
		for name, fromFlag := range map[string]bool{"flag": true, "file": false} {
			t.Run(agent+"/"+name, func(t *testing.T) {
				options := baseOptions()
				options.Agent = agent
				file := config.FileConfig{}
				if fromFlag {
					options.Provider = "gemini"
				} else {
					file.Provider = "gemini"
				}

				_, err := plan(t, options, file)

				if err == nil || !strings.Contains(err.Error(), "uses its own login") {
					t.Fatalf("NewPlan() error = %v, want a native login rejection", err)
				}
			})
		}
	}
}

func TestNewPlanRejectsAProviderPresetThatAnotherAgentOwns(t *testing.T) {
	options := baseOptions()
	options.Agent = "pi"
	options.Provider = "gemini"

	_, err := plan(t, options, config.FileConfig{})

	if err == nil || !strings.Contains(err.Error(), "unknown provider") {
		t.Fatalf("NewPlan() error = %v, want an unknown provider error", err)
	}
}

func TestNewPlanKeepsTheRequirePoliciesForEveryHarness(t *testing.T) {
	for _, agent := range []string{"codex", "copilot", "kiro", "pi"} {
		t.Run(agent, func(t *testing.T) {
			options := baseOptions()
			options.Agent = agent
			options.Policy.RequireKernelIsolation = true

			if _, err := plan(t, options, config.FileConfig{}); err == nil {
				t.Fatal("NewPlan() error = nil, want the kernel isolation guarantee refused on isolation none")
			}
		})
	}
}
