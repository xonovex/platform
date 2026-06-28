// Package git is the workspace=git leaf: git-worktree checkouts.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	wsshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/workspace/shared"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/scriptlib"
)

// Worktree is the git VCS variant: it creates or reuses a git worktree.
type Worktree struct{}

// New creates the git worktree variant.
func New() *Worktree { return &Worktree{} }

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
func isWorktreeForRepo(worktreeDir, repoDir string) bool {
	worktreeGitDir, err := wsshared.ExecGit([]string{"rev-parse", "--git-dir"}, worktreeDir)
	if err != nil {
		return false
	}
	repoGitDir, err := wsshared.ExecGit([]string{"rev-parse", "--git-dir"}, repoDir)
	if err != nil {
		return false
	}

	resolvedWorktreeGitDir := worktreeGitDir
	if !filepath.IsAbs(worktreeGitDir) {
		resolvedWorktreeGitDir = filepath.Join(worktreeDir, worktreeGitDir)
	}
	resolvedRepoGitDir := repoGitDir
	if !filepath.IsAbs(repoGitDir) {
		resolvedRepoGitDir = filepath.Join(repoDir, repoGitDir)
	}
	return strings.HasPrefix(resolvedWorktreeGitDir, resolvedRepoGitDir)
}

// checkExisting inspects an existing directory's worktree status.
func checkExisting(dir, repoDir string) existingCheck {
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
	if !isWorktreeForRepo(resolvedDir, repoDir) {
		return existingCheck{Exists: true, IsWorktree: true, IsForThisRepo: false}
	}
	return existingCheck{
		Exists:        true,
		IsWorktree:    true,
		IsForThisRepo: true,
		CurrentBranch: wsshared.GetCurrentBranchSync(resolvedDir),
	}
}

// branchExists checks whether a local branch exists.
func branchExists(branch, cwd string) bool {
	_, err := wsshared.ExecGit([]string{"rev-parse", "--verify", "refs/heads/" + branch}, cwd)
	return err == nil
}

func createWorktreeForExistingBranch(dir, branch, cwd string) error {
	cmd := exec.Command("git", "worktree", "add", dir, branch)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createWorktreeWithNewBranch(dir, branch, sourceBranch, cwd string) error {
	cmd := exec.Command("git", "worktree", "add", dir, "-b", branch, sourceBranch)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func setMergeBackConfig(branch, sourceBranch, cwd string) error {
	_, err := wsshared.ExecGit([]string{"config", fmt.Sprintf("branch.%s.mergeBackTo", branch), sourceBranch}, cwd)
	return err
}

// Setup creates or reuses a git worktree. If the worktree already exists with the
// correct branch it is reused; otherwise a new worktree (and branch) is created
// and its mergeBackTo config recorded.
func (Worktree) Setup(config wsshared.Config, repoDir string, verbose bool) (string, error) {
	resolvedDir := config.Dir
	if !filepath.IsAbs(config.Dir) {
		resolvedDir = filepath.Join(repoDir, config.Dir)
	}

	existing := checkExisting(config.Dir, repoDir)
	if existing.Exists {
		if !existing.IsWorktree {
			scriptlib.LogError(fmt.Sprintf("Directory %s exists but is not a git worktree", config.Dir))
			return "", fmt.Errorf("directory exists but is not a worktree: %s", config.Dir)
		}
		if !existing.IsForThisRepo {
			scriptlib.LogError(fmt.Sprintf("Worktree %s exists but belongs to a different repository", config.Dir))
			return "", fmt.Errorf("worktree belongs to different repository: %s", config.Dir)
		}
		if existing.CurrentBranch != "" && existing.CurrentBranch != config.Branch {
			scriptlib.LogError(fmt.Sprintf("Worktree %s exists on branch '%s', expected '%s'",
				config.Dir, existing.CurrentBranch, config.Branch))
			return "", fmt.Errorf("worktree on wrong branch: expected '%s', found '%s'",
				config.Branch, existing.CurrentBranch)
		}
		if verbose {
			scriptlib.LogInfo(fmt.Sprintf("Reusing existing worktree at %s on branch %s", config.Dir, config.Branch))
		}
		return resolvedDir, nil
	}

	if branchExists(config.Branch, repoDir) {
		if verbose {
			scriptlib.LogInfo(fmt.Sprintf("Creating worktree at %s for existing branch %s", config.Dir, config.Branch))
		}
		if err := createWorktreeForExistingBranch(config.Dir, config.Branch, repoDir); err != nil {
			scriptlib.LogError(fmt.Sprintf("Failed to create worktree: %v", err))
			return "", err
		}
		if verbose {
			scriptlib.LogInfo("Worktree created successfully for existing branch")
			scriptlib.LogInfo(fmt.Sprintf("  Branch: %s", config.Branch))
		}
		return resolvedDir, nil
	}

	sourceBranch := config.SourceBranch
	if sourceBranch == "" {
		sourceBranch = wsshared.GetCurrentBranchSync(repoDir)
		if sourceBranch == "" {
			return "", fmt.Errorf("failed to determine source branch")
		}
	}

	if verbose {
		scriptlib.LogInfo(fmt.Sprintf("Creating worktree at %s on new branch %s from %s",
			config.Dir, config.Branch, sourceBranch))
	}
	if err := createWorktreeWithNewBranch(config.Dir, config.Branch, sourceBranch, repoDir); err != nil {
		scriptlib.LogError(fmt.Sprintf("Failed to create worktree: %v", err))
		return "", err
	}
	if err := setMergeBackConfig(config.Branch, sourceBranch, repoDir); err != nil {
		scriptlib.LogError(fmt.Sprintf("Failed to set mergeBackTo config: %v", err))
		return "", err
	}
	if verbose {
		scriptlib.LogSuccess("Worktree created successfully")
		scriptlib.LogInfo(fmt.Sprintf("  Branch: %s", config.Branch))
		scriptlib.LogInfo(fmt.Sprintf("  Source: %s", sourceBranch))
		scriptlib.LogInfo(fmt.Sprintf("  mergeBackTo: %s", sourceBranch))
	}
	return resolvedDir, nil
}
