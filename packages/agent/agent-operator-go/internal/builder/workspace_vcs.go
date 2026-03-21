package builder

import (
	"fmt"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// VCSStrategy defines the behavior for a version control system
type VCSStrategy interface {
	PostCloneScript() string
	WorktreeScript(path, branch, sourceBranch string) string
	InitContainerName() string
}

var vcsStrategies = map[agentv1alpha1.WorkspaceType]VCSStrategy{
	"":                                &GitStrategy{},
	agentv1alpha1.WorkspaceTypeGit:     &GitStrategy{},
	agentv1alpha1.WorkspaceTypeJujutsu: &JujutsuStrategy{},
}

// GetVCSStrategy returns the VCS strategy for the given workspace type
func GetVCSStrategy(wsType agentv1alpha1.WorkspaceType) (VCSStrategy, error) {
	s, ok := vcsStrategies[wsType]
	if !ok {
		return nil, fmt.Errorf("unsupported workspace type: %s", wsType)
	}
	return s, nil
}
