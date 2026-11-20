package tmux

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "alphanumeric unchanged",
			input:    "myproject",
			expected: "myproject",
		},
		{
			name:     "dots replaced with hyphens",
			input:    "my.project.name",
			expected: "my-project-name",
		},
		{
			name:     "spaces replaced with hyphens",
			input:    "my project",
			expected: "my-project",
		},
		{
			name:     "special chars replaced",
			input:    "project@v1.0!test",
			expected: "project-v1-0-test",
		},
		{
			name:     "multiple hyphens collapsed",
			input:    "my...project",
			expected: "my-project",
		},
		{
			name:     "leading/trailing hyphens removed",
			input:    ".project.",
			expected: "project",
		},
		{
			name:     "underscores preserved",
			input:    "my_project_name",
			expected: "my_project_name",
		},
		{
			name:     "empty string returns agent",
			input:    "",
			expected: "agent",
		},
		{
			name:     "only special chars returns agent",
			input:    "...",
			expected: "agent",
		},
		{
			name:     "long names truncated",
			input:    "this-is-a-very-long-project-name-that-exceeds-thirty-characters",
			expected: "this-is-a-very-long-project-na",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateSessionName(t *testing.T) {
	info := &gitInfo{parentDir: "xonovex", repoName: "platform", branchName: "feature/x", shortCommit: "c89010f"}
	tests := []struct {
		name    string
		workDir string
		info    *gitInfo
		want    string
	}{
		{"git with parent", "/src/xonovex/platform", info, "xonovex-platform/feature-x"},
		{"git at root", "/platform", info, "platform/feature-x"},
		{"no git with parent", "/src/project", nil, "src-project"},
		{"no git at root", "/project", nil, "agent-project"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateSessionName(tt.workDir, tt.info); got != tt.want {
				t.Errorf("generateSessionName(%q) = %q, want %q", tt.workDir, got, tt.want)
			}
		})
	}
}

func TestGenerateWindowName(t *testing.T) {
	info := &gitInfo{branchName: "main", shortCommit: "c89010f"}
	if got := generateWindowName("/src/project", info); got != "main/c89010f" {
		t.Errorf("generateWindowName(git) = %q, want main/c89010f", got)
	}
	if got := generateWindowName("/src/my.project", nil); got != "my-project" {
		t.Errorf("generateWindowName(no git) = %q, want my-project", got)
	}
}

// gitHost records the git output of a repository at /src/xonovex/platform.
func gitHost(workDir string) *systemtest.Fake {
	host := systemtest.New()
	host.Outcomes[systemtest.InDir(workDir, "git rev-parse --show-toplevel")] = systemtest.Outcome{Stdout: "/src/xonovex/platform\n"}
	host.Outcomes[systemtest.InDir(workDir, "git rev-parse --abbrev-ref HEAD")] = systemtest.Outcome{Stdout: "main\n"}
	host.Outcomes[systemtest.InDir(workDir, "git rev-parse --short HEAD")] = systemtest.Outcome{Stdout: "c89010f\n"}
	return host
}

func TestGetGitInfo(t *testing.T) {
	t.Run("returns info for git directory", func(t *testing.T) {
		info := getGitInfo(gitHost("/work"), "/work")
		want := gitInfo{parentDir: "xonovex", repoName: "platform", branchName: "main", shortCommit: "c89010f"}
		if info == nil || *info != want {
			t.Fatalf("getGitInfo() = %+v, want %+v", info, want)
		}
	})

	for _, missing := range []string{"git rev-parse --show-toplevel", "git rev-parse --abbrev-ref HEAD", "git rev-parse --short HEAD"} {
		t.Run("returns nil without "+missing, func(t *testing.T) {
			host := gitHost("/work")
			host.Outcomes[systemtest.InDir("/work", missing)] = systemtest.Outcome{Err: errors.New("not a git repository")}
			if info := getGitInfo(host, "/work"); info != nil {
				t.Errorf("getGitInfo() = %+v, want nil", info)
			}
		})
	}
}

func TestExecutorIsAvailableWhenTmuxResolves(t *testing.T) {
	host := systemtest.New().AddExecutable("/usr/bin/tmux")
	host.Env["PATH"] = "/usr/bin"

	if !NewExecutor(host, testLog).IsAvailable() {
		t.Error("IsAvailable() = false, want true when tmux resolves on PATH")
	}
}

func TestExecutorIsUnavailableWhenTmuxDoesNotResolve(t *testing.T) {
	host := systemtest.New()
	host.Env["PATH"] = "/usr/bin"

	if NewExecutor(host, testLog).IsAvailable() {
		t.Error("IsAvailable() = true, want false when tmux does not resolve on PATH")
	}
}

func TestExecutorIsInsideReadsTheTmuxEnvironment(t *testing.T) {
	host := systemtest.New()
	if NewExecutor(host, testLog).IsInside() {
		t.Error("IsInside() = true, want false when TMUX is empty")
	}

	host.Env["TMUX"] = "/tmp/tmux-1000/default,123,0"
	if !NewExecutor(host, testLog).IsInside() {
		t.Error("IsInside() = false, want true when TMUX is set")
	}
}

const launchScript = "/tmp/xonovex-agent-terminal-1.sh"

// tmuxHost records a tmux whose session named "session" exists or not.
func tmuxHost(sessionExists bool) *systemtest.Fake {
	host := systemtest.New()
	hasSession := systemtest.Outcome{}
	if !sessionExists {
		hasSession.Err = errors.New("no session")
	}
	host.Outcomes["tmux has-session -t session"] = hasSession
	return host
}

func runCalls(host *systemtest.Fake) []string {
	var lines []string
	for _, call := range host.Calls {
		if call.Kind == "run" {
			lines = append(lines, call.Line())
		}
	}
	return lines
}

func TestExecutorExecuteCreatesADetachedSessionWithoutTheEnvironmentInItsArguments(t *testing.T) {
	host := tmuxHost(false)
	newSession := "tmux new-session -d -s session -n window -c /work sh " + launchScript
	host.Outcomes[newSession] = systemtest.Outcome{}
	secret := "provider-secret-value"
	config := &termshared.TerminalConfig{SessionName: "session", WindowName: "window", Detach: true}

	exitCode, err := NewExecutor(host, testLog).Execute(config, []string{"claude"}, []string{"PROVIDER_TOKEN=" + secret}, "/work")

	if err != nil || exitCode != 0 {
		t.Fatalf("Execute() exitCode = %d, error = %v", exitCode, err)
	}
	if got := runCalls(host); !slices.Equal(got, []string{newSession}) {
		t.Fatalf("tmux runs = %v, want %q", got, newSession)
	}
	for _, call := range host.Calls {
		if call.Env != nil || strings.Contains(call.Line(), secret) || strings.Contains(call.Line(), "PROVIDER_TOKEN") {
			t.Fatalf("tmux call %q carries the environment", call.Line())
		}
	}
	content, ok := host.File(launchScript)
	if !ok || !strings.Contains(content, "export PROVIDER_TOKEN='"+secret+"'") {
		t.Errorf("launch script = %q, want the exported secret", content)
	}
}

func TestExecutorExecuteCreatesAnAttachedSession(t *testing.T) {
	host := tmuxHost(false)
	newSession := "tmux new-session -s session -n window -c /work sh " + launchScript
	host.Outcomes[newSession] = systemtest.Outcome{}
	config := &termshared.TerminalConfig{SessionName: "session", WindowName: "window"}

	exitCode, err := NewExecutor(host, testLog).Execute(config, []string{"claude"}, nil, "/work")

	if err != nil || exitCode != 0 {
		t.Fatalf("Execute() exitCode = %d, error = %v", exitCode, err)
	}
	if got := runCalls(host); !slices.Equal(got, []string{newSession}) {
		t.Fatalf("tmux runs = %v, want %q", got, newSession)
	}
}

func TestExecutorExecuteAddsWindowToExistingSession(t *testing.T) {
	tests := []struct {
		name   string
		detach bool
		want   []string
	}{
		{"detached", true, []string{"tmux new-window -t session -n window -c /work sh " + launchScript}},
		{"attached", false, []string{"tmux new-window -t session -n window -c /work sh " + launchScript, "tmux attach-session -t session"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host := tmuxHost(true)
			for _, line := range tt.want {
				host.Outcomes[line] = systemtest.Outcome{}
			}
			config := &termshared.TerminalConfig{SessionName: "session", WindowName: "window", Detach: tt.detach}

			exitCode, err := NewExecutor(host, testLog).Execute(config, []string{"claude"}, nil, "/work")

			if err != nil || exitCode != 0 {
				t.Fatalf("Execute() exitCode = %d, error = %v", exitCode, err)
			}
			if got := runCalls(host); !slices.Equal(got, tt.want) {
				t.Fatalf("tmux runs = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutorExecuteRemovesTheScriptWhenTmuxFails(t *testing.T) {
	tests := []struct {
		name     string
		outcome  systemtest.Outcome
		wantCode int
	}{
		{"non-zero exit", systemtest.Outcome{ExitCode: 3}, 3},
		{"start failure", systemtest.Outcome{Err: errors.New("exec failed")}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host := tmuxHost(false)
			host.Outcomes["tmux new-session -s session -n window -c /work sh "+launchScript] = tt.outcome
			config := &termshared.TerminalConfig{SessionName: "session", WindowName: "window"}

			exitCode, err := NewExecutor(host, testLog).Execute(config, []string{"claude"}, nil, "/work")

			if exitCode != tt.wantCode || (err != nil) != (tt.outcome.Err != nil) {
				t.Fatalf("Execute() = (%d, %v), want code %d", exitCode, err, tt.wantCode)
			}
			if host.Exists(launchScript) {
				t.Error("launch script remains after tmux failed")
			}
		})
	}
}

func TestExecutorExecuteReportsNonZeroWindowAndAttachCodes(t *testing.T) {
	newWindow := "tmux new-window -t session -n window -c /work sh " + launchScript
	attach := "tmux attach-session -t session"
	tests := []struct {
		name     string
		outcomes map[string]systemtest.Outcome
		wantCode int
	}{
		{"window fails", map[string]systemtest.Outcome{newWindow: {ExitCode: 2}}, 2},
		{"attach fails", map[string]systemtest.Outcome{newWindow: {}, attach: {ExitCode: 4}}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host := tmuxHost(true)
			for line, outcome := range tt.outcomes {
				host.Outcomes[line] = outcome
			}
			config := &termshared.TerminalConfig{SessionName: "session", WindowName: "window"}

			exitCode, err := NewExecutor(host, testLog).Execute(config, []string{"claude"}, nil, "/work")

			if err != nil || exitCode != tt.wantCode {
				t.Fatalf("Execute() = (%d, %v), want code %d", exitCode, err, tt.wantCode)
			}
		})
	}
}

func TestExecutorExecuteNamesTheSessionAndWindowFromGit(t *testing.T) {
	host := gitHost("/src/xonovex/platform")
	host.Outcomes["tmux has-session -t xonovex-platform/main"] = systemtest.Outcome{Err: errors.New("no session")}
	newSession := "tmux new-session -s xonovex-platform/main -n main/c89010f -c /src/xonovex/platform sh " + launchScript
	host.Outcomes[newSession] = systemtest.Outcome{}

	exitCode, err := NewExecutor(host, testLog).Execute(&termshared.TerminalConfig{}, []string{"claude"}, nil, "/src/xonovex/platform")

	if err != nil || exitCode != 0 {
		t.Fatalf("Execute() exitCode = %d, error = %v", exitCode, err)
	}
	if got := runCalls(host); !slices.Equal(got, []string{newSession}) {
		t.Fatalf("tmux runs = %v, want %q", got, newSession)
	}
}

func TestExecutorExecuteReportsALaunchScriptFailure(t *testing.T) {
	host := tmuxHost(false)
	config := &termshared.TerminalConfig{SessionName: "session", WindowName: "window"}

	exitCode, err := NewExecutor(host, testLog).Execute(config, nil, nil, "/work")

	if err == nil || exitCode != 1 {
		t.Fatalf("Execute(no command) = (%d, %v), want code 1 and an error", exitCode, err)
	}
}
