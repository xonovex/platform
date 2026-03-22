package builder

import "fmt"

// GitStrategy implements VCSStrategy for git
type GitStrategy struct{}

func (g *GitStrategy) PostCloneScript() string {
	return ""
}

func (g *GitStrategy) WorktreeScript(path, branch, sourceBranch string) string {
	return fmt.Sprintf("git worktree add %s -b %s %s\n", shellQuote(path), shellQuote(branch), shellQuote(sourceBranch))
}

func (g *GitStrategy) InitContainerName() string {
	return "git-worktree"
}
