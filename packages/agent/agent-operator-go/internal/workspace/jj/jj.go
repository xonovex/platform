// Package jj is the workspace=jj leaf: Jujutsu workspace VCS strategy.
package jj

import (
	"fmt"

	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/shell"
)

// Strategy implements the workspace VCS strategy for Jujutsu (jj).
type Strategy struct{}

// PostCloneScript colocates a jj repo onto the cloned git repo.
func (j *Strategy) PostCloneScript() string { return "jj git init --colocate\n" }

// WorktreeScript returns the script that creates a jj workspace.
func (j *Strategy) WorktreeScript(path, _, sourceBranch string) string {
	return fmt.Sprintf("jj workspace add %s --revision %s\n", shell.Quote(path), shell.Quote(sourceBranch))
}

// InitContainerName returns the workspace init-container name.
func (j *Strategy) InitContainerName() string { return "jj-workspace" }
