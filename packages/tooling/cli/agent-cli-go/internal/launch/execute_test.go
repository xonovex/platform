package launch

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

// executeFixture wires Execute to doubles: the bwrap and docker methods resolve
// to one isolator, a git checkout creates /work/created, and tmux is installed.
type executeFixture struct {
	host     *systemtest.Fake
	sandbox  testSandbox
	vcs      *fakeVCS
	terminal *fakeTerminal
	records  *fakeRecorder
	out      bytes.Buffer
	log      bytes.Buffer
}

func newExecuteFixture() *executeFixture {
	host := systemtest.New()
	host.WorkingDir = "/work/repo"
	return &executeFixture{
		host:     host,
		sandbox:  newTestSandbox(),
		vcs:      &fakeVCS{dir: "/work/created"},
		terminal: &fakeTerminal{},
		records:  &fakeRecorder{},
	}
}

func (f *executeFixture) deps() Dependencies {
	return Dependencies{
		Host:       f.host,
		Log:        logging.New(&f.log, &f.log, false, false),
		Out:        &f.out,
		Sandboxes:  f.sandbox.registry,
		Records:    f.records,
		Workspaces: map[wsp.VCSType]wsshared.VCS{wsp.VCSGit: f.vcs},
		Terminals:  map[termshared.TerminalType]termshared.TerminalExecutor{termshared.TerminalTmux: f.terminal},
	}
}

func bwrapOptions() Options {
	options := baseOptions()
	options.Isolation = "bwrap"
	return options
}

func TestExecuteRunsTheAgentInTheSourceDirectory(t *testing.T) {
	f := newExecuteFixture()

	if err := Execute(f.deps(), bwrapOptions()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	iso := f.sandbox.isolator
	if !iso.ran || iso.commandRun.WorkDir != "/work/repo" || iso.commandRun.HomeDir != "/home/user" {
		t.Errorf("isolator ran = %v with %+v", iso.ran, iso.commandRun)
	}
	if !iso.checked || len(f.vcs.configs) != 0 || f.terminal.executed {
		t.Errorf("checked, setups, terminal = %v %d %v", iso.checked, len(f.vcs.configs), f.terminal.executed)
	}
}

func TestExecuteDryRunWritesTheCommandWithoutACheckout(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.DryRun = true
	options.WorktreeBranch = "feature"
	options.Provider = "gemini"
	options.AgentArgs = []string{"review"}

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if f.out.String() != "agent run\n" {
		t.Errorf("Out = %q, want the command line", f.out.String())
	}
	if len(f.vcs.configs) != 0 || f.sandbox.isolator.ran {
		t.Errorf("dry run created a checkout (%d) or ran (%v)", len(f.vcs.configs), f.sandbox.isolator.ran)
	}
	for _, want := range []string{"Dry run", "Claude", "Gemini", "/work/repo-feature", "[review]"} {
		if !strings.Contains(f.log.String(), want) {
			t.Errorf("log = %q, want %q", f.log.String(), want)
		}
	}
}

func TestExecuteDryRunWithoutArgumentsOrCommand(t *testing.T) {
	f := newExecuteFixture()
	f.sandbox.isolator.command = nil
	options := bwrapOptions()
	options.DryRun = true

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if f.out.Len() != 0 || !strings.Contains(f.log.String(), "Arguments: (none)") {
		t.Errorf("Out = %q, log = %q", f.out.String(), f.log.String())
	}
}

func TestExecuteReportsADryRunWriteFailure(t *testing.T) {
	f := newExecuteFixture()
	deps := f.deps()
	deps.Out = failingWriter{}
	options := bwrapOptions()
	options.DryRun = true

	if err := Execute(deps, options); err == nil || !strings.Contains(err.Error(), "write dry run") {
		t.Fatalf("Execute() error = %v, want a write error", err)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestExecuteListsTheAvailableIsolationsWhenTheSelectedOneIsMissing(t *testing.T) {
	f := newExecuteFixture()
	f.sandbox.isolator.unavailable = true

	err := Execute(f.deps(), bwrapOptions())

	if err == nil || !strings.Contains(err.Error(), "isolation method bwrap is not available; available isolations: [none]") {
		t.Fatalf("Execute() error = %v, want the available isolations", err)
	}
}

func TestExecuteRejectsATerminalBeforeTheCheckout(t *testing.T) {
	tests := []struct {
		name     string
		terminal string
		missing  bool
		phrase   string
	}{
		{"unknown", "screen", false, "unknown terminal type: screen; available types: [tmux]"},
		{"unknown without installed", "screen", true, "available types: []"},
		{"not installed", "tmux", true, "terminal type tmux is not available"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := newExecuteFixture()
			f.terminal.missing = test.missing
			options := bwrapOptions()
			options.Terminal = test.terminal
			options.WorktreeBranch = "feature"

			err := Execute(f.deps(), options)

			if err == nil || !strings.Contains(err.Error(), test.phrase) {
				t.Fatalf("Execute() error = %v, want %q", err, test.phrase)
			}
			if len(f.vcs.configs) != 0 {
				t.Error("the checkout was created before the terminal check")
			}
		})
	}
}

func TestExecuteRejectsACheckoutWithoutItsVariant(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.WorktreeBranch = "feature"
	options.VCS = "jj"

	if err := Execute(f.deps(), options); err == nil || !strings.Contains(err.Error(), `unknown --vcs "jj"`) {
		t.Fatalf("Execute() error = %v, want an unknown variant error", err)
	}
}

func TestExecuteRunsInTheCreatedCheckout(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.WorktreeBranch = "feature"
	options.WorktreeSourceBranch = "main"

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(f.vcs.configs) != 1 || f.vcs.repos[0] != "/work/repo" || f.vcs.configs[0].Branch != "feature" || f.vcs.configs[0].SourceBranch != "main" {
		t.Fatalf("Setup calls = %+v in %v", f.vcs.configs, f.vcs.repos)
	}
	run := f.sandbox.isolator.commandRun
	if run.WorkDir != "/work/created" || run.RepoDir != "/work/repo" {
		t.Errorf("RunConfig directories = %q %q, want the checkout and the source", run.WorkDir, run.RepoDir)
	}
}

func TestExecuteStopsWhenTheCheckoutFails(t *testing.T) {
	f := newExecuteFixture()
	f.vcs.err = errors.New("worktree failed")
	options := bwrapOptions()
	options.WorktreeBranch = "feature"

	if err := Execute(f.deps(), options); err == nil || err.Error() != "worktree failed" || f.sandbox.isolator.ran {
		t.Fatalf("Execute() error = %v, ran = %v", err, f.sandbox.isolator.ran)
	}
}

func TestExecuteChecksTheAgentWhereTheProvisionerLeavesIt(t *testing.T) {
	f := newExecuteFixture()
	f.sandbox.provisioner.handled = true

	if err := Execute(f.deps(), bwrapOptions()); err != nil {
		t.Fatalf("Execute(handled) error = %v", err)
	}
	if f.sandbox.isolator.checked {
		t.Error("the isolator checked an agent the provisioner accounted for")
	}

	f = newExecuteFixture()
	f.sandbox.isolator.checkErr = errors.New("agent missing")
	if err := Execute(f.deps(), bwrapOptions()); err == nil || err.Error() != "agent missing" || !f.sandbox.isolator.checked {
		t.Fatalf("Execute(unhandled) error = %v, want the isolator check", err)
	}
}

func TestExecuteStopsOnAProvisionerFailure(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*fakeProvisioner)
		want   string
	}{
		{"contribute", func(p *fakeProvisioner) { p.contributeErr = errors.New("resolve failed") }, "resolve failed"},
		{"check", func(p *fakeProvisioner) { p.checkErr = errors.New("agent not supplied") }, "agent not supplied"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := newExecuteFixture()
			test.modify(f.sandbox.provisioner)
			if err := Execute(f.deps(), bwrapOptions()); err == nil || err.Error() != test.want || f.sandbox.isolator.ran {
				t.Fatalf("Execute() error = %v, ran = %v; want %q", err, f.sandbox.isolator.ran, test.want)
			}
		})
	}
}

func TestExecutePassesTheProvisionInput(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.Provision = "command"
	options.InitCommands = []string{"npm ci"}
	f.sandbox.provisioner.contribution = provision.Contribution{PathEntries: []string{"/opt/bin"}}

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !slices.Equal(f.sandbox.provisioner.input.InitCommands, []string{"npm ci"}) {
		t.Errorf("Input = %+v", f.sandbox.provisioner.input)
	}
}

func TestExecuteWrapsACommandFailure(t *testing.T) {
	f := newExecuteFixture()
	f.sandbox.isolator.commandErr = errors.New("missing bind")
	options := bwrapOptions()
	options.DryRun = true

	err := Execute(f.deps(), options)

	if err == nil || err.Error() != "prepare sandbox command: missing bind" || f.out.Len() != 0 {
		t.Fatalf("Execute() error = %v, out = %q", err, f.out.String())
	}
}

func TestExecuteReturnsTheExitCodeOfTheAgent(t *testing.T) {
	f := newExecuteFixture()
	f.sandbox.isolator.runCode = 3

	err := Execute(f.deps(), bwrapOptions())

	var exit ExitError
	if !errors.As(err, &exit) || exit.Code != 3 || err.Error() != "agent exited with code 3" {
		t.Fatalf("Execute() error = %v, want ExitError{3}", err)
	}
}

func TestExecuteReturnsARunFailure(t *testing.T) {
	f := newExecuteFixture()
	f.sandbox.isolator.runErr = errors.New("spawn failed")

	if err := Execute(f.deps(), bwrapOptions()); err == nil || err.Error() != "spawn failed" {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestExecuteRunsInTheTerminalInTheCheckout(t *testing.T) {
	f := newExecuteFixture()
	f.sandbox.isolator.terminalCmd = []string{"bwrap", "agent"}
	f.sandbox.isolator.terminalEnv = []string{"A=1"}
	options := bwrapOptions()
	options.Terminal = "tmux"
	options.TerminalSession = "work"
	options.WorktreeBranch = "feature"
	options.Verbose = true

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	term := f.terminal
	if !term.executed || term.workDir != "/work/created" || term.config.SessionName != "work" {
		t.Fatalf("terminal = %+v", term)
	}
	if !slices.Equal(term.command, []string{"bwrap", "agent"}) || !slices.Equal(term.env, []string{"A=1"}) {
		t.Errorf("terminal command, env = %v %v", term.command, term.env)
	}
	if f.sandbox.isolator.ran || !strings.Contains(f.log.String(), "Using terminal wrapper: tmux") {
		t.Errorf("ran = %v, log = %q", f.sandbox.isolator.ran, f.log.String())
	}
}

func TestExecuteReportsTerminalFailures(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*executeFixture)
		check  func(error) bool
	}{
		{"terminal command", func(f *executeFixture) { f.sandbox.isolator.terminalErr = errors.New("no env") }, func(err error) bool {
			return err != nil && err.Error() == "prepare terminal command: no env"
		}},
		{"execute", func(f *executeFixture) { f.terminal.err = errors.New("tmux failed") }, func(err error) bool {
			return err != nil && err.Error() == "tmux failed"
		}},
		{"exit code", func(f *executeFixture) { f.terminal.code = 2 }, func(err error) bool {
			var exit ExitError
			return errors.As(err, &exit) && exit.Code == 2
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := newExecuteFixture()
			test.modify(f)
			options := bwrapOptions()
			options.Terminal = "tmux"
			if err := Execute(f.deps(), options); !test.check(err) {
				t.Fatalf("Execute() error = %v", err)
			}
		})
	}
}

func TestExecuteStopsOnAnInvalidPlan(t *testing.T) {
	f := newExecuteFixture()
	options := bwrapOptions()
	options.Agent = "unknown"

	if err := Execute(f.deps(), options); err == nil || !strings.Contains(err.Error(), "available agents") {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestExecuteLoadsTheConfigurationFile(t *testing.T) {
	f := newExecuteFixture()
	f.host.AddFile("/etc/agent.yaml", "homeDir: /srv/home\nprovider: glm\n", 0o600)
	f.host.HomeErr = errors.New("no home")
	options := bwrapOptions()
	options.Config = "/etc/agent.yaml"

	if err := Execute(f.deps(), options); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	run := f.sandbox.isolator.commandRun
	if run.HomeDir != "/srv/home" || run.Provider == nil || run.Provider.Name != "glm" {
		t.Errorf("RunConfig = %+v, want the file home and provider", run)
	}
}

func TestExecuteReportsHostFailures(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*executeFixture, *Options)
		phrase string
	}{
		{"config", func(_ *executeFixture, o *Options) { o.Config = "/missing.yaml" }, "failed to load config"},
		{"invalid default config", func(f *executeFixture, _ *Options) {
			f.host.AddFile(config.DefaultPath("/home/user"), "providre=glm\n", 0o600)
		}, "failed to load config"},
		{"home", func(f *executeFixture, o *Options) {
			f.host.AddFile("/etc/agent.yaml", "{}\n", 0o600)
			f.host.HomeErr = errors.New("no home")
			o.Config = "/etc/agent.yaml"
		}, "resolve user home directory: no home"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := newExecuteFixture()
			options := bwrapOptions()
			test.modify(f, &options)
			err := Execute(f.deps(), options)
			if err == nil || !strings.Contains(err.Error(), test.phrase) {
				t.Fatalf("Execute() error = %v, want %q", err, test.phrase)
			}
		})
	}
}

// getwdFailure is a host whose working directory is gone.
type getwdFailure struct{ *systemtest.Fake }

func (getwdFailure) Getwd() (string, error) { return "", errors.New("deleted") }

func TestExecuteReportsAMissingWorkingDirectory(t *testing.T) {
	f := newExecuteFixture()
	deps := f.deps()
	deps.Host = getwdFailure{f.host}

	if err := Execute(deps, bwrapOptions()); err == nil || err.Error() != "resolve current working directory: deleted" {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestExecuteDryRunReportsTheDeclaredFacts(t *testing.T) {
	tests := []struct {
		agent string
		want  []string
	}{
		{"claude", []string{"Protocol: messages", "Model access: provider (environment)", "Version probe: claude --version"}},
		{"codex", []string{"Protocol: responses", "Model access: provider (environment, arguments)", "Version probe: codex --version"}},
		{"opencode", []string{"Protocol: none", "Model access: provider (arguments)", "Version probe: opencode --version"}},
		{"pi", []string{"Protocol: responses", "Version probe: pi --version"}},
		{"copilot", []string{"Protocol: none", "Model access: native login", "Version probe: copilot --version"}},
		{"kiro", []string{"Protocol: none", "Model access: native login", "Version probe: kiro-cli-chat --version"}},
	}
	for _, test := range tests {
		t.Run(test.agent, func(t *testing.T) {
			f := newExecuteFixture()
			options := bwrapOptions()
			options.DryRun = true
			options.Agent = test.agent

			if err := Execute(f.deps(), options); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			for _, want := range test.want {
				if !strings.Contains(f.log.String(), want) {
					t.Errorf("log = %q, want %q", f.log.String(), want)
				}
			}
		})
	}
}
