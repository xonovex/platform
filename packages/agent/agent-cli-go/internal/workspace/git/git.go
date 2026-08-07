// Package git is the workspace=git leaf: git-worktree checkouts.
package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	wsshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/workspace/shared"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/logging"
)

// Worktree is the git VCS variant: it creates or reuses a git worktree.
type Worktree struct{ runner wsshared.Runner }

// New creates the git worktree variant over the given process runner.
func New(runner wsshared.Runner) *Worktree { return &Worktree{runner: runner} }

// Available reports true: git worktrees are always usable in a git repo.
func (Worktree) Available() bool { return true }

// existingCheck holds the result of inspecting an existing directory.
type existingCheck struct {
	Exists        bool
	IsWorktree    bool
	IsForThisRepo bool
	CurrentBranch string
}

// isWorktreeDirectory checks whether dir is a git worktree (a .git file, not dir).
func isWorktreeDirectory(dir string) bool {
	gitPath := filepath.Join(dir, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	content, err := os.ReadFile(gitPath)
	if err != nil {
		return false
	}
	return strings.HasPrefix(string(content), "gitdir:")
}

// isWorktreeForRepo checks whether worktreeDir belongs to repoDir.
func isWorktreeForRepo(runner wsshared.Runner, worktreeDir, repoDir string) bool {
	worktreeGitDir, err := commonGitDir(runner, worktreeDir)
	if err != nil {
		return false
	}
	repoGitDir, err := commonGitDir(runner, repoDir)
	if err != nil {
		return false
	}
	return worktreeGitDir == repoGitDir
}

func commonGitDir(runner wsshared.Runner, dir string) (string, error) {
	gitDir, err := wsshared.ExecGit(runner, []string{"rev-parse", "--git-common-dir"}, dir)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(dir, gitDir)
	}
	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(gitDir)
	if err != nil {
		return "", err
	}
	return filepath.Clean(resolved), nil
}

// checkExisting inspects an existing directory's worktree status.
func checkExisting(runner wsshared.Runner, dir, repoDir string) existingCheck {
	resolvedDir := dir
	if !filepath.IsAbs(dir) {
		resolvedDir = filepath.Join(repoDir, dir)
	}

	if _, err := os.Stat(resolvedDir); os.IsNotExist(err) {
		return existingCheck{Exists: false}
	}
	if !isWorktreeDirectory(resolvedDir) {
		return existingCheck{Exists: true, IsWorktree: false}
	}
	if !isWorktreeForRepo(runner, resolvedDir, repoDir) {
		return existingCheck{Exists: true, IsWorktree: true, IsForThisRepo: false}
	}
	return existingCheck{
		Exists:        true,
		IsWorktree:    true,
		IsForThisRepo: true,
		CurrentBranch: wsshared.GetCurrentBranchSync(runner, resolvedDir),
	}
}

// branchExists checks whether a local branch exists.
func branchExists(runner wsshared.Runner, branch, cwd string) bool {
	_, err := wsshared.ExecGit(runner, []string{"rev-parse", "--verify", "refs/heads/" + branch}, cwd)
	return err == nil
}

func createWorktreeForExistingBranch(runner wsshared.Runner, dir, branch, cwd string) error {
	return runner.Stream("git", []string{"worktree", "add", dir, branch}, cwd)
}

func createWorktreeWithNewBranch(runner wsshared.Runner, dir, branch, sourceBranch, cwd string) error {
	return runner.Stream("git", []string{"worktree", "add", dir, "-b", branch, sourceBranch}, cwd)
}

func setMergeBackConfig(runner wsshared.Runner, branch, sourceBranch, cwd string) error {
	_, err := wsshared.ExecGit(runner, []string{"config", fmt.Sprintf("branch.%s.mergeBackTo", branch), sourceBranch}, cwd)
	return err
}

// Setup creates or reuses a git worktree. If the worktree already exists with the
// correct branch it is reused; otherwise a new worktree (and branch) is created
// and its mergeBackTo config recorded.
func (w Worktree) Setup(config wsshared.Config, repoDir string, verbose bool) (string, error) {
	resolvedDir := config.Dir
	if !filepath.IsAbs(config.Dir) {
		resolvedDir = filepath.Join(repoDir, config.Dir)
	}

	existing := checkExisting(w.runner, config.Dir, repoDir)
	if existing.Exists {
		if !existing.IsWorktree {
			logging.LogError(fmt.Sprintf("Directory %s exists but is not a git worktree", config.Dir))
			return "", fmt.Errorf("directory exists but is not a worktree: %s", config.Dir)
		}
		if !existing.IsForThisRepo {
			logging.LogError(fmt.Sprintf("Worktree %s exists but belongs to a different repository", config.Dir))
			return "", fmt.Errorf("worktree belongs to different repository: %s", config.Dir)
		}
		if existing.CurrentBranch != "" && existing.CurrentBranch != config.Branch {
			logging.LogError(fmt.Sprintf("Worktree %s exists on branch '%s', expected '%s'",
				config.Dir, existing.CurrentBranch, config.Branch))
			return "", fmt.Errorf("worktree on wrong branch: expected '%s', found '%s'",
				config.Branch, existing.CurrentBranch)
		}
		if verbose {
			logging.LogInfo(fmt.Sprintf("Reusing existing worktree at %s on branch %s", config.Dir, config.Branch))
		}
		return resolvedDir, nil
	}

	if branchExists(w.runner, config.Branch, repoDir) {
		if verbose {
			logging.LogInfo(fmt.Sprintf("Creating worktree at %s for existing branch %s", config.Dir, config.Branch))
		}
		if err := createWorktreeForExistingBranch(w.runner, config.Dir, config.Branch, repoDir); err != nil {
			logging.LogError(fmt.Sprintf("Failed to create worktree: %v", err))
			return "", err
		}
		if verbose {
			logging.LogInfo("Worktree created successfully for existing branch")
			logging.LogInfo(fmt.Sprintf("  Branch: %s", config.Branch))
		}
		return resolvedDir, nil
	}

	sourceBranch := config.SourceBranch
	if sourceBranch == "" {
		sourceBranch = wsshared.GetCurrentBranchSync(w.runner, repoDir)
		if sourceBranch == "" {
			return "", fmt.Errorf("failed to determine source branch")
		}
	}

	if verbose {
		logging.LogInfo(fmt.Sprintf("Creating worktree at %s on new branch %s from %s",
			config.Dir, config.Branch, sourceBranch))
	}
	if err := createWorktreeWithNewBranch(w.runner, config.Dir, config.Branch, sourceBranch, repoDir); err != nil {
		logging.LogError(fmt.Sprintf("Failed to create worktree: %v", err))
		return "", err
	}
	if err := setMergeBackConfig(w.runner, config.Branch, sourceBranch, repoDir); err != nil {
		logging.LogError(fmt.Sprintf("Failed to set mergeBackTo config: %v", err))
		return "", err
	}
	if verbose {
		logging.LogSuccess("Worktree created successfully")
		logging.LogInfo(fmt.Sprintf("  Branch: %s", config.Branch))
		logging.LogInfo(fmt.Sprintf("  Source: %s", sourceBranch))
		logging.LogInfo(fmt.Sprintf("  mergeBackTo: %s", sourceBranch))
	}
	return resolvedDir, nil
}
