package worktree

import (
	"fmt"
	"regexp"
	"strings"
)

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
