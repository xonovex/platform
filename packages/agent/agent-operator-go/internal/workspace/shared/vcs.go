// Package shared is the operator's workspace axis core: the VCS strategy port +
// registry, the workspace PVC builders, and the clone/worktree script builders.
// It provides pure data (scripts, strategies, PVCs, paths) that the isolation pod
// realizer composes into containers — it never builds pods or applies security.
package shared

import (
	"fmt"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/workspace/git"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/workspace/jj"
)

// VCSStrategy defines the per-VCS behavior for cloning and worktree creation.
type VCSStrategy interface {
	PostCloneScript() string
	WorktreeScript(path, branch, sourceBranch string) string
	InitContainerName() string
}

var vcsStrategies = map[agentv1alpha1.WorkspaceType]VCSStrategy{
	"":                                 &git.Strategy{},
	agentv1alpha1.WorkspaceTypeGit:     &git.Strategy{},
	agentv1alpha1.WorkspaceTypeJujutsu: &jj.Strategy{},
}

// GetVCSStrategy returns the VCS strategy for the given workspace type.
func GetVCSStrategy(wsType agentv1alpha1.WorkspaceType) (VCSStrategy, error) {
	s, ok := vcsStrategies[wsType]
	if !ok {
		return nil, fmt.Errorf("unsupported workspace type: %s", wsType)
	}
	return s, nil
}
