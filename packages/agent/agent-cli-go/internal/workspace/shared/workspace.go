// Package shared is the workspace axis core: the VCS port plus the git-repo
// plumbing both variants (git worktrees, jj workspaces layered on a git repo)
// build on. The leaves (git, jj) implement the port; this core names no leaf. The
// composition root (cmd) selects a leaf by the shared pkg/workspace VCSType.
package shared

import (
	"os/exec"
	"strings"
)

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

// ExecGit runs a git command in cwd and returns trimmed stdout.
func ExecGit(args []string, cwd string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranchSync returns the current git branch name, or "" if detached or
// on error. Both the git and jj variants use it to resolve a source revision.
func GetCurrentBranchSync(cwd string) string {
	result, err := ExecGit([]string{"rev-parse", "--abbrev-ref", "HEAD"}, cwd)
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
