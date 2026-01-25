package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/tools/tool-lib-go/pkg/scriptlib"
)

// Config holds worktree configuration
type Config struct {
	SourceBranch string
	Branch       string
	Dir          string
}

// ExistingWorktreeCheck holds the result of checking an existing worktree
type ExistingWorktreeCheck struct {
	Exists        bool
	IsWorktree    bool
	IsForThisRepo bool
	CurrentBranch string
}

// execGit runs a git command and returns stdout
func execGit(args []string, cwd string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetGitRootSync returns the git repository root directory
func GetGitRootSync(cwd string) string {
	result, err := execGit([]string{"rev-parse", "--show-toplevel"}, cwd)
	if err != nil {
		return ""
	}
	return result
}

// GetCurrentBranchSync returns the current git branch name
func GetCurrentBranchSync(cwd string) string {
	result, err := execGit([]string{"rev-parse", "--abbrev-ref", "HEAD"}, cwd)
	if err != nil || result == "HEAD" {
		return ""
	}
	return result
}

// SanitizeBranchName converts a branch name to a valid directory name
func SanitizeBranchName(branch string) string {
	// Replace slashes with hyphens
	result := regexp.MustCompile(`[/\\]`).ReplaceAllString(branch, "-")
	// Replace other special chars with hyphens
	result = regexp.MustCompile(`[^\w-]`).ReplaceAllString(result, "-")
	// Collapse multiple hyphens
	result = regexp.MustCompile(`-+`).ReplaceAllString(result, "-")
	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")
	return result
}

// GetDefaultDir returns the default worktree directory path
// Returns a path like "../<repo-name>-<sanitized-branch-name>"
func GetDefaultDir(branch, repoName string) string {
	sanitizedRepo := SanitizeBranchName(repoName)
	sanitizedBranch := SanitizeBranchName(branch)
	return fmt.Sprintf("../%s-%s", sanitizedRepo, sanitizedBranch)
}

// IsWorktreeDirectory checks if a directory is a git worktree
func IsWorktreeDirectory(dir string) bool {
	gitPath := filepath.Join(dir, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}

	// Worktrees have a .git file, not a directory
	if info.IsDir() {
		return false
	}

	// Check if it contains a gitdir pointer
	content, err := os.ReadFile(gitPath)
	if err != nil {
		return false
	}

	return strings.HasPrefix(string(content), "gitdir:")
}

// IsWorktreeForRepo checks if a worktree belongs to a specific repository
func IsWorktreeForRepo(worktreeDir, repoDir string) bool {
	// Get the git directory of the worktree
	worktreeGitDir, err := execGit([]string{"rev-parse", "--git-dir"}, worktreeDir)
	if err != nil {
		return false
	}

	// Get the git directory of the main repo
	repoGitDir, err := execGit([]string{"rev-parse", "--git-dir"}, repoDir)
	if err != nil {
		return false
	}

	// Resolve to absolute paths
	resolvedWorktreeGitDir := worktreeGitDir
	if !filepath.IsAbs(worktreeGitDir) {
		resolvedWorktreeGitDir = filepath.Join(worktreeDir, worktreeGitDir)
	}

	resolvedRepoGitDir := repoGitDir
	if !filepath.IsAbs(repoGitDir) {
		resolvedRepoGitDir = filepath.Join(repoDir, repoGitDir)
	}

	// Worktree git dir should be inside the repo's .git/worktrees/
	return strings.HasPrefix(resolvedWorktreeGitDir, resolvedRepoGitDir)
}

// CheckExistingWorktree checks the status of an existing directory
func CheckExistingWorktree(dir, repoDir string) ExistingWorktreeCheck {
	resolvedDir := dir
	if !filepath.IsAbs(dir) {
		resolvedDir = filepath.Join(repoDir, dir)
	}

	if _, err := os.Stat(resolvedDir); os.IsNotExist(err) {
		return ExistingWorktreeCheck{Exists: false}
	}

	if !IsWorktreeDirectory(resolvedDir) {
		return ExistingWorktreeCheck{Exists: true, IsWorktree: false}
	}

	if !IsWorktreeForRepo(resolvedDir, repoDir) {
		return ExistingWorktreeCheck{Exists: true, IsWorktree: true, IsForThisRepo: false}
	}

	currentBranch := GetCurrentBranchSync(resolvedDir)
	return ExistingWorktreeCheck{
		Exists:        true,
		IsWorktree:    true,
		IsForThisRepo: true,
		CurrentBranch: currentBranch,
	}
}

// branchExists checks if a branch exists
func branchExists(branch, cwd string) bool {
	_, err := execGit([]string{"rev-parse", "--verify", "refs/heads/" + branch}, cwd)
	return err == nil
}

// createWorktreeForExistingBranch creates a worktree for an existing branch
func createWorktreeForExistingBranch(dir, branch, cwd string) error {
	cmd := exec.Command("git", "worktree", "add", dir, branch)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// createWorktreeWithNewBranch creates a worktree with a new branch
func createWorktreeWithNewBranch(dir, branch, sourceBranch, cwd string) error {
	cmd := exec.Command("git", "worktree", "add", dir, "-b", branch, sourceBranch)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// setMergeBackConfig sets the mergeBackTo config for a branch
func setMergeBackConfig(branch, sourceBranch, cwd string) error {
	_, err := execGit([]string{"config", fmt.Sprintf("branch.%s.mergeBackTo", branch), sourceBranch}, cwd)
	return err
}

// Setup creates or reuses a git worktree
// If the worktree already exists with the correct branch, it will be reused.
// Otherwise, creates a new worktree and sets the mergeBackTo config.
func Setup(config Config, repoDir string, verbose bool) (string, error) {
	resolvedDir := config.Dir
	if !filepath.IsAbs(config.Dir) {
		resolvedDir = filepath.Join(repoDir, config.Dir)
	}

	// Check if worktree already exists
	existing := CheckExistingWorktree(config.Dir, repoDir)

	if existing.Exists {
		// Directory exists - check if we can reuse it
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

		// Worktree exists and is valid - reuse it
		if verbose {
			scriptlib.LogInfo(fmt.Sprintf("Reusing existing worktree at %s on branch %s",
				config.Dir, config.Branch))
		}
		return resolvedDir, nil
	}

	// Worktree doesn't exist - check if branch exists
	branchAlreadyExists := branchExists(config.Branch, repoDir)

	if branchAlreadyExists {
		// Branch exists - create worktree for existing branch
		if verbose {
			scriptlib.LogInfo(fmt.Sprintf("Creating worktree at %s for existing branch %s",
				config.Dir, config.Branch))
		}

		if err := createWorktreeForExistingBranch(config.Dir, config.Branch, repoDir); err != nil {
			scriptlib.LogError(fmt.Sprintf("Failed to create worktree: %v", err))
			return "", err
		}

		if verbose {
			scriptlib.LogInfo("Worktree created successfully for existing branch")
			scriptlib.LogInfo(fmt.Sprintf("  Branch: %s", config.Branch))
		}
	} else {
		// Branch doesn't exist - create new branch and worktree
		sourceBranch := config.SourceBranch
		if sourceBranch == "" {
			sourceBranch = GetCurrentBranchSync(repoDir)
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

		// Set the mergeBackTo config only for new branches
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
	}

	return resolvedDir, nil
}

// BuildBindPaths adds worktree source repo to bind paths
func BuildBindPaths(basePaths []string, sourceRepoDir string) []string {
	if sourceRepoDir == "" {
		return basePaths
	}

	// Add source repo for .git access
	paths := make([]string, 0, len(basePaths)+1)
	paths = append(paths, basePaths...)
	paths = append(paths, sourceRepoDir)

	return paths
}
