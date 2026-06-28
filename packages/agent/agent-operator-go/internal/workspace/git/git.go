// Package git is the workspace=git leaf: git-worktree VCS strategy.
package git

import (
	"fmt"

	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/shell"
)

// Strategy implements the workspace VCS strategy for git.
type Strategy struct{}

// PostCloneScript returns the script to run after cloning (none for git).
func (g *Strategy) PostCloneScript() string { return "" }

// WorktreeScript returns the script that creates a git worktree.
func (g *Strategy) WorktreeScript(path, branch, sourceBranch string) string {
	return fmt.Sprintf("git worktree add %s -b %s %s\n", shell.Quote(path), shell.Quote(branch), shell.Quote(sourceBranch))
}

// InitContainerName returns the worktree init-container name.
func (g *Strategy) InitContainerName() string { return "git-worktree" }
