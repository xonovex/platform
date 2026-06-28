// Package shared is the operator's workspace axis core: the leaf-free VCS-strategy
// port, the workspace PVC builders, and the clone/worktree script builders (which
// take a VCSStrategy rather than resolving one). The strategy registry that wires
// concrete leaves (git, jj) lives in the composition root (internal/plugins). This
// core provides pure data that the isolation pod realizer composes into containers.
package shared

// VCSStrategy defines the per-VCS behavior for cloning and worktree creation.
type VCSStrategy interface {
	PostCloneScript() string
	WorktreeScript(path, branch, sourceBranch string) string
	InitContainerName() string
}
