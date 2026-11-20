// Package shared is the workspace axis core: the VCS port, the Host port the
// variants consume, and the git-repo plumbing both variants (git worktrees, jj
// workspaces layered on a git repo) build on. The leaves (git, jj) implement
// the VCS port; this core names no leaf. The composition root keys each leaf by
// the shared pkg/workspace VCSType.
package shared

import "io/fs"

// Config holds the checkout request shared by every VCS variant.
type Config struct {
	SourceBranch string
	Branch       string
	Dir          string
}

// VCS is the workspace port: create or reuse a checkout, returning its resolved
// directory. Each variant declares its own availability.
type VCS interface {
	Setup(config Config, repoDir string, verbose bool) (string, error)
	Available() bool
}

// Host is the workspace axis port to the machine: the git and jj commands the
// variants run, and the checkout files they inspect. A test supplies recorded
// output and an in-memory file tree instead of a real repository.
type Host interface {
	// Capture runs name with args in cwd and returns trimmed stdout.
	Capture(name string, args []string, cwd string) (string, error)
	// Stream runs name with args in cwd with its output on the terminal of the
	// CLI, which the worktree and workspace creation commands need.
	Stream(name string, args []string, cwd string) error
	LookPath(file string) (string, error)
	Stat(path string) (fs.FileInfo, error)
	ReadFile(path string) ([]byte, error)
	EvalSymlinks(path string) (string, error)
}

// ExecGit runs a git command in cwd and returns trimmed stdout.
func ExecGit(host Host, args []string, cwd string) (string, error) {
	return host.Capture("git", args, cwd)
}

// GetCurrentBranchSync returns the current git branch name, or "" if detached or
// on error. Both the git and jj variants use it to resolve a source revision.
func GetCurrentBranchSync(host Host, cwd string) string {
	result, err := ExecGit(host, []string{"rev-parse", "--abbrev-ref", "HEAD"}, cwd)
	if err != nil || result == "HEAD" {
		return ""
	}
	return result
}

// BuildBindPaths adds the worktree source repo to the bind paths so the sandbox
// can reach the shared .git directory.
func BuildBindPaths(basePaths []string, sourceRepoDir string) []string {
	if sourceRepoDir == "" {
		return basePaths
	}
	paths := make([]string, 0, len(basePaths)+1)
	paths = append(paths, basePaths...)
	paths = append(paths, sourceRepoDir)
	return paths
}
