package tmux

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/shell"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

// gitInfo holds git repository information
type gitInfo struct {
	parentDir   string
	repoName    string
	branchName  string
	shortCommit string
}

// Host is the port the tmux executor consumes: the tmux and git commands it
// runs, the environment it reads, and the launch script it writes.
type Host interface {
	LookPath(file string) (string, error)
	Getenv(key string) string
	// Capture runs name in dir and returns its trimmed standard output.
	Capture(name string, args []string, dir string) (string, error)
	// Run starts name attached to the terminal of the CLI and returns its exit
	// code; a nil env inherits the environment of the CLI.
	Run(name string, args, env []string, dir string) (int, error)
	CreateTemp(pattern string) (string, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	Remove(path string) error
}

// Executor implements tmux terminal wrapper
type Executor struct {
	host Host
	log  logging.Logger
}

// NewExecutor creates a tmux executor over host, writing through log.
func NewExecutor(host Host, log logging.Logger) *Executor {
	return &Executor{host: host, log: log}
}

// IsAvailable checks if tmux is installed
func (e *Executor) IsAvailable() bool {
	_, err := e.host.LookPath("tmux")
	return err == nil
}

// IsInside checks if we're already inside a tmux session
func (e *Executor) IsInside() bool {
	return e.host.Getenv("TMUX") != ""
}

// Execute runs the command in a tmux session
func (e *Executor) Execute(config *termshared.TerminalConfig, command []string, env []string, workDir string) (int, error) {
	// Generate session and window names if not provided
	sessionName := config.SessionName
	if sessionName == "" {
		sessionName = generateSessionName(workDir, getGitInfo(e.host, workDir))
	}

	windowName := config.WindowName
	if windowName == "" {
		windowName = generateWindowName(workDir, getGitInfo(e.host, workDir))
	}

	launchScript, err := createLaunchScript(e.host, command, env)
	if err != nil {
		return 1, err
	}

	run := e.createSession
	if e.sessionExists(sessionName) {
		run = e.addWindow
	}
	exitCode, err := run(sessionName, windowName, workDir, launchScript, config.Detach)
	if err != nil || exitCode != 0 {
		_ = e.host.Remove(launchScript)
	}
	return exitCode, err
}

// sessionExists checks if a tmux session with the given name exists
func (e *Executor) sessionExists(name string) bool {
	_, err := e.host.Capture("tmux", []string{"has-session", "-t", name}, "")
	return err == nil
}

// getGitInfo retrieves git repository information for the given directory, or
// nil when workDir is not in a git repository with a commit.
func getGitInfo(host Host, workDir string) *gitInfo {
	repoRoot, err := host.Capture("git", []string{"rev-parse", "--show-toplevel"}, workDir)
	if err != nil {
		return nil
	}
	branchName, err := host.Capture("git", []string{"rev-parse", "--abbrev-ref", "HEAD"}, workDir)
	if err != nil {
		return nil
	}
	shortCommit, err := host.Capture("git", []string{"rev-parse", "--short", "HEAD"}, workDir)
	if err != nil {
		return nil
	}
	return &gitInfo{
		repoName:    filepath.Base(repoRoot),
		parentDir:   filepath.Base(filepath.Dir(repoRoot)),
		branchName:  branchName,
		shortCommit: shortCommit,
	}
}

// generateSessionName generates a session name from directory and git info
// Format: <parent>-<dir>/<branch> (e.g., "xonovex-platform/master")
func generateSessionName(workDir string, info *gitInfo) string {
	dirName := filepath.Base(workDir)
	parentDir := filepath.Base(filepath.Dir(workDir))

	if info != nil {
		if parentDir != "" && parentDir != "." && parentDir != "/" {
			return sanitizeName(parentDir) + "-" + sanitizeName(dirName) + "/" + sanitizeName(info.branchName)
		}
		return sanitizeName(dirName) + "/" + sanitizeName(info.branchName)
	}
	// Fallback to directory-based naming without branch
	if parentDir != "" && parentDir != "." && parentDir != "/" {
		return sanitizeName(parentDir) + "-" + sanitizeName(dirName)
	}
	return "agent-" + sanitizeName(dirName)
}

// generateWindowName generates a window name from git info or work directory
// Format: <branch>/<short-commit> (e.g., "master/c89010f")
func generateWindowName(workDir string, info *gitInfo) string {
	if info != nil {
		return sanitizeName(info.branchName) + "/" + info.shortCommit
	}
	// Fallback to directory-based naming
	return sanitizeName(filepath.Base(workDir))
}

// sanitizeName makes a name safe for use in tmux
func sanitizeName(name string) string {
	// Replace dots and other special characters with hyphens
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	sanitized := re.ReplaceAllString(name, "-")
	// Remove leading/trailing hyphens
	sanitized = strings.Trim(sanitized, "-")
	// Collapse multiple hyphens
	re = regexp.MustCompile(`-+`)
	sanitized = re.ReplaceAllString(sanitized, "-")
	// Limit length
	if len(sanitized) > 30 {
		sanitized = sanitized[:30]
	}
	if sanitized == "" {
		sanitized = "agent"
	}
	return sanitized
}

// createSession creates a new tmux session
func (e *Executor) createSession(sessionName, windowName, workDir, launchScript string, detach bool) (int, error) {
	args := []string{"new-session"}

	if detach {
		args = append(args, "-d")
	}

	args = append(args,
		"-s", sessionName,
		"-n", windowName,
		"-c", workDir,
		"sh", launchScript,
	)

	e.log.Debug("Creating tmux session: " + sessionName)

	if code, err := e.host.Run("tmux", args, nil, ""); err != nil || code != 0 {
		return code, err
	}

	// If detached, print session info
	if detach {
		e.log.Info("Started tmux session: " + sessionName)
		e.log.Info("Attach with: tmux attach-session -t " + sessionName)
	}

	return 0, nil
}

// addWindow adds a new window to an existing tmux session
func (e *Executor) addWindow(sessionName, windowName, workDir, launchScript string, detach bool) (int, error) {
	// Create new window in existing session
	args := []string{
		"new-window",
		"-t", sessionName,
		"-n", windowName,
		"-c", workDir,
		"sh", launchScript,
	}

	e.log.Debug("Adding window to tmux session: " + sessionName)

	if code, err := e.host.Run("tmux", args, nil, ""); err != nil || code != 0 {
		return code, err
	}

	// Attach to session if not detaching
	if !detach {
		return e.attachSession(sessionName)
	}

	e.log.Info("Added window to tmux session: " + sessionName)
	e.log.Info("Attach with: tmux attach-session -t " + sessionName)

	return 0, nil
}

// attachSession attaches to an existing tmux session
func (e *Executor) attachSession(sessionName string) (int, error) {
	args := []string{"attach-session", "-t", sessionName}

	e.log.Debug("Attaching to tmux session: " + sessionName)

	return e.host.Run("tmux", args, nil, "")
}

// buildShellCommand builds a shell command from arguments
func buildShellCommand(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shell.Quote(arg)
	}
	return strings.Join(quoted, " ")
}
