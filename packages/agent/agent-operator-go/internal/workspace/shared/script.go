package shared

import (
	"fmt"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/shell"
)

const (
	// WorkspaceMountPath is where the workspace PVC is mounted in agent/init pods.
	WorkspaceMountPath = "/workspace"
	// WorktreeBasePath is the parent directory for per-run worktrees.
	WorktreeBasePath = "/workspace-wt"
	// WorkspaceVolumeName is the name of the workspace volume.
	WorkspaceVolumeName = "workspace"
)

// CloneScript builds the repository clone script for the workspace mount.
func CloneScript(repo agentv1alpha1.RepositorySpec, wsType agentv1alpha1.WorkspaceType) string {
	script := "set -e\n"
	script += "cd " + WorkspaceMountPath + "\n"
	script += "git clone"
	if repo.Branch != "" {
		script += " --branch " + shell.Quote(repo.Branch)
	}
	script += " --single-branch --depth 1"
	script += " -- " + shell.Quote(repo.URL) + " .\n"

	if repo.Commit != "" {
		script += "git fetch origin " + shell.Quote(repo.Commit) + "\n"
		script += "git checkout " + shell.Quote(repo.Commit) + "\n"
	}

	if vcs, err := GetVCSStrategy(wsType); err == nil {
		script += vcs.PostCloneScript()
	}

	return script
}

// WorktreePath returns the per-run worktree path under WorktreeBasePath.
func WorktreePath(runName string) string {
	return fmt.Sprintf("%s/%s", WorktreeBasePath, runName)
}

// WorktreeScriptAndName builds the worktree-creation script for worktreePath and
// returns the init-container name for the VCS type. An empty sourceBranch defaults
// to HEAD.
func WorktreeScriptAndName(wsType agentv1alpha1.WorkspaceType, worktreePath, branch, sourceBranch string) (script, name string) {
	if sourceBranch == "" {
		sourceBranch = "HEAD"
	}
	script = "set -e\n"
	script += "cd " + WorkspaceMountPath + "\n"
	name = "git-worktree"
	if vcs, err := GetVCSStrategy(wsType); err == nil {
		name = vcs.InitContainerName()
		script += vcs.WorktreeScript(worktreePath, branch, sourceBranch)
	}
	return script, name
}
