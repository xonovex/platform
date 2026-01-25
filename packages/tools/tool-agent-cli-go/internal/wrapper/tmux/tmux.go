package tmux

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/tools/tool-agent-cli-go/internal/types"
	"github.com/xonovex/platform/packages/tools/tool-lib-go/pkg/scriptlib"
)

// gitInfo holds git repository information
type gitInfo struct {
	parentDir   string
	repoName    string
	branchName  string
	shortCommit string
}

// Executor implements tmux terminal wrapper
type Executor struct{}

// NewExecutor creates a new tmux executor
func NewExecutor() *Executor {
	return &Executor{}
}

// IsAvailable checks if tmux is installed
func (e *Executor) IsAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// IsInside checks if we're already inside a tmux session
func (e *Executor) IsInside() bool {
	return os.Getenv("TMUX") != ""
}

// Execute runs the command in a tmux session
func (e *Executor) Execute(config *types.TerminalConfig, command []string, env []string, workDir string, verbose bool) (int, error) {
	// Generate session and window names if not provided
	sessionName := config.SessionName
	if sessionName == "" {
		sessionName = generateSessionName(workDir)
	}

	windowName := config.WindowName
	if windowName == "" {
		windowName = generateWindowName(workDir)
	}

	// Build the shell command with environment exports
	envExports := BuildEnvExports(env)
	shellCommand := envExports + buildShellCommand(command)

	// Check if session already exists
	sessionExists := e.sessionExists(sessionName)

	if sessionExists {
		// Add new window to existing session
		return e.addWindow(sessionName, windowName, workDir, shellCommand, config.Detach, verbose)
	}

	// Create new session
	return e.createSession(sessionName, windowName, workDir, shellCommand, config.Detach, verbose)
}

// sessionExists checks if a tmux session with the given name exists
func (e *Executor) sessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	err := cmd.Run()
	return err == nil
}

// getGitInfo retrieves git repository information for the given directory
func getGitInfo(workDir string) *gitInfo {
	info := &gitInfo{}

	// Get repository root and name
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	repoRoot := strings.TrimSpace(string(output))
	info.repoName = filepath.Base(repoRoot)
	info.parentDir = filepath.Base(filepath.Dir(repoRoot))

	// Get current branch name
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = workDir
	output, err = cmd.Output()
	if err != nil {
		return nil
	}
	info.branchName = strings.TrimSpace(string(output))

	// Get short commit hash
	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = workDir
	output, err = cmd.Output()
	if err != nil {
		return nil
	}
	info.shortCommit = strings.TrimSpace(string(output))

	return info
}

// generateSessionName generates a session name from directory and git info
// Format: <parent>-<dir>/<branch> (e.g., "xonovex-platform/master")
func generateSessionName(workDir string) string {
	dirName := filepath.Base(workDir)
	parentDir := filepath.Base(filepath.Dir(workDir))

	if info := getGitInfo(workDir); info != nil {
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
func generateWindowName(workDir string) string {
	if info := getGitInfo(workDir); info != nil {
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
func (e *Executor) createSession(sessionName, windowName, workDir, shellCommand string, detach bool, verbose bool) (int, error) {
	args := []string{"new-session"}

	if detach {
		args = append(args, "-d")
	}

	args = append(args,
		"-s", sessionName,
		"-n", windowName,
		"-c", workDir,
		"sh", "-c", shellCommand,
	)

	if verbose {
		scriptlib.LogDebug(verbose, "Creating tmux session: "+sessionName)
		scriptlib.LogDebug(verbose, "Executing: tmux "+strings.Join(args, " "))
	}

	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}

	// If detached, print session info
	if detach {
		scriptlib.LogInfo("Started tmux session: " + sessionName)
		scriptlib.LogInfo("Attach with: tmux attach-session -t " + sessionName)
	}

	return 0, nil
}

// addWindow adds a new window to an existing tmux session
func (e *Executor) addWindow(sessionName, windowName, workDir, shellCommand string, detach bool, verbose bool) (int, error) {
	// Create new window in existing session
	args := []string{
		"new-window",
		"-t", sessionName,
		"-n", windowName,
		"-c", workDir,
		"sh", "-c", shellCommand,
	}

	if verbose {
		scriptlib.LogDebug(verbose, "Adding window to tmux session: "+sessionName)
		scriptlib.LogDebug(verbose, "Executing: tmux "+strings.Join(args, " "))
	}

	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}

	// Attach to session if not detaching
	if !detach {
		return e.attachSession(sessionName, verbose)
	}

	scriptlib.LogInfo("Added window to tmux session: " + sessionName)
	scriptlib.LogInfo("Attach with: tmux attach-session -t " + sessionName)

	return 0, nil
}

// attachSession attaches to an existing tmux session
func (e *Executor) attachSession(sessionName string, verbose bool) (int, error) {
	args := []string{"attach-session", "-t", sessionName}

	if verbose {
		scriptlib.LogDebug(verbose, "Attaching to tmux session: "+sessionName)
	}

	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}

	return 0, nil
}

// buildShellCommand builds a shell command from arguments
func buildShellCommand(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuote(arg)
	}
	return strings.Join(quoted, " ")
}

// shellQuote quotes a string for safe use in a shell command
func shellQuote(s string) string {
	// If the string contains no special characters, return as-is
	safe := true
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '.' || c == '/' || c == ':' || c == '=' || c == '-' || c == '_') {
			safe = false
			break
		}
	}
	if safe {
		return s
	}
	// Wrap in single quotes and escape any single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
