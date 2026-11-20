// Package git is the workspace=git leaf: git-worktree checkouts.
package git

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

// Worktree is the git VCS variant: it creates or reuses a git worktree.
type Worktree struct {
	host wsshared.Host
	log  logging.Logger
}

// New creates the git worktree variant over host, writing through log.
func New(host wsshared.Host, log logging.Logger) *Worktree {
	return &Worktree{host: host, log: log}
}

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
func isWorktreeDirectory(host wsshared.Host, dir string) bool {
	gitPath := filepath.Join(dir, ".git")
	info, err := host.Stat(gitPath)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	content, err := host.ReadFile(gitPath)
	if err != nil {
		return false
	}
	return strings.HasPrefix(string(content), "gitdir:")
}

// isWorktreeForRepo checks whether worktreeDir belongs to repoDir.
func isWorktreeForRepo(host wsshared.Host, worktreeDir, repoDir string) bool {
	worktreeGitDir, err := commonGitDir(host, worktreeDir)
	if err != nil {
		return false
	}
	repoGitDir, err := commonGitDir(host, repoDir)
	if err != nil {
		return false
	}
	return worktreeGitDir == repoGitDir
}

func commonGitDir(host wsshared.Host, dir string) (string, error) {
	gitDir, err := wsshared.ExecGit(host, []string{"rev-parse", "--git-common-dir"}, dir)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(dir, gitDir)
	}
	resolved, err := host.EvalSymlinks(filepath.Clean(gitDir))
	if err != nil {
		return "", err
	}
	return filepath.Clean(resolved), nil
}

// checkExisting inspects an existing directory's worktree status.
func checkExisting(host wsshared.Host, dir, repoDir string) existingCheck {
	resolvedDir := dir
	if !filepath.IsAbs(dir) {
		resolvedDir = filepath.Join(repoDir, dir)
	}

	if _, err := host.Stat(resolvedDir); errors.Is(err, fs.ErrNotExist) {
		return existingCheck{Exists: false}
	}
	if !isWorktreeDirectory(host, resolvedDir) {
		return existingCheck{Exists: true, IsWorktree: false}
	}
	if !isWorktreeForRepo(host, resolvedDir, repoDir) {
		return existingCheck{Exists: true, IsWorktree: true, IsForThisRepo: false}
	}
	return existingCheck{
		Exists:        true,
		IsWorktree:    true,
		IsForThisRepo: true,
		CurrentBranch: wsshared.GetCurrentBranchSync(host, resolvedDir),
	}
}

// branchExists checks whether a local branch exists.
func branchExists(host wsshared.Host, branch, cwd string) bool {
	_, err := wsshared.ExecGit(host, []string{"rev-parse", "--verify", "refs/heads/" + branch}, cwd)
	return err == nil
}

func createWorktreeForExistingBranch(host wsshared.Host, dir, branch, cwd string) error {
	return host.Stream("git", []string{"worktree", "add", dir, branch}, cwd)
}

func createWorktreeWithNewBranch(host wsshared.Host, dir, branch, sourceBranch, cwd string) error {
	return host.Stream("git", []string{"worktree", "add", dir, "-b", branch, sourceBranch}, cwd)
}

func setMergeBackConfig(host wsshared.Host, branch, sourceBranch, cwd string) error {
	_, err := wsshared.ExecGit(host, []string{"config", fmt.Sprintf("branch.%s.mergeBackTo", branch), sourceBranch}, cwd)
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

	existing := checkExisting(w.host, config.Dir, repoDir)
	if existing.Exists {
		if !existing.IsWorktree {
			w.log.Error(fmt.Sprintf("Directory %s exists but is not a git worktree", config.Dir))
			return "", fmt.Errorf("directory exists but is not a worktree: %s", config.Dir)
		}
		if !existing.IsForThisRepo {
			w.log.Error(fmt.Sprintf("Worktree %s exists but belongs to a different repository", config.Dir))
			return "", fmt.Errorf("worktree belongs to different repository: %s", config.Dir)
		}
		if existing.CurrentBranch != "" && existing.CurrentBranch != config.Branch {
			w.log.Error(fmt.Sprintf("Worktree %s exists on branch '%s', expected '%s'",
				config.Dir, existing.CurrentBranch, config.Branch))
			return "", fmt.Errorf("worktree on wrong branch: expected '%s', found '%s'",
				config.Branch, existing.CurrentBranch)
		}
		if verbose {
			w.log.Info(fmt.Sprintf("Reusing existing worktree at %s on branch %s", config.Dir, config.Branch))
		}
		return resolvedDir, nil
	}

	if branchExists(w.host, config.Branch, repoDir) {
		if verbose {
			w.log.Info(fmt.Sprintf("Creating worktree at %s for existing branch %s", config.Dir, config.Branch))
		}
		if err := createWorktreeForExistingBranch(w.host, config.Dir, config.Branch, repoDir); err != nil {
			w.log.Error(fmt.Sprintf("Failed to create worktree: %v", err))
			return "", err
		}
		if verbose {
			w.log.Info("Worktree created successfully for existing branch")
			w.log.Info(fmt.Sprintf("  Branch: %s", config.Branch))
		}
		return resolvedDir, nil
	}

	sourceBranch := config.SourceBranch
	if sourceBranch == "" {
		sourceBranch = wsshared.GetCurrentBranchSync(w.host, repoDir)
		if sourceBranch == "" {
			return "", fmt.Errorf("failed to determine source branch")
		}
	}

	if verbose {
		w.log.Info(fmt.Sprintf("Creating worktree at %s on new branch %s from %s",
			config.Dir, config.Branch, sourceBranch))
	}
	if err := createWorktreeWithNewBranch(w.host, config.Dir, config.Branch, sourceBranch, repoDir); err != nil {
		w.log.Error(fmt.Sprintf("Failed to create worktree: %v", err))
		return "", err
	}
	if err := setMergeBackConfig(w.host, config.Branch, sourceBranch, repoDir); err != nil {
		w.log.Error(fmt.Sprintf("Failed to set mergeBackTo config: %v", err))
		return "", err
	}
	if verbose {
		w.log.Success("Worktree created successfully")
		w.log.Info(fmt.Sprintf("  Branch: %s", config.Branch))
		w.log.Info(fmt.Sprintf("  Source: %s", sourceBranch))
		w.log.Info(fmt.Sprintf("  mergeBackTo: %s", sourceBranch))
	}
	return resolvedDir, nil
}
