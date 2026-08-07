// Package jj is the workspace=jj leaf: Jujutsu workspaces layered on a git repo.
package jj

import (
	"fmt"
	"os"
	"path/filepath"

	wsshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/workspace/shared"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/logging"
)

// Workspace is the jj VCS variant: it creates or reuses a jj workspace.
type Workspace struct{ runner wsshared.Runner }

// New creates the jj workspace variant over the given process runner.
func New(runner wsshared.Runner) *Workspace { return &Workspace{runner: runner} }

// Available reports whether the jj binary is on PATH.
func (w Workspace) Available() bool { return w.runner.Available("jj") }

// Setup creates or reuses a jj workspace at config.Dir, branching from the source
// revision (defaulting to the repo's current git branch).
func (w Workspace) Setup(config wsshared.Config, repoDir string, verbose bool) (string, error) {
	resolvedDir := config.Dir
	if !filepath.IsAbs(config.Dir) {
		resolvedDir = filepath.Join(repoDir, config.Dir)
	}

	// Reuse an existing workspace directory.
	if _, err := os.Stat(resolvedDir); err == nil {
		if verbose {
			logging.LogInfo(fmt.Sprintf("Reusing existing jj workspace at %s", config.Dir))
		}
		return resolvedDir, nil
	}

	if !w.Available() {
		return "", fmt.Errorf("jj is not installed or not on PATH; install from https://martinvonz.github.io/jj/")
	}

	sourceBranch := config.SourceBranch
	if sourceBranch == "" {
		sourceBranch = wsshared.GetCurrentBranchSync(w.runner, repoDir)
		if sourceBranch == "" {
			return "", fmt.Errorf("failed to determine source revision")
		}
	}

	if verbose {
		logging.LogInfo(fmt.Sprintf("Creating jj workspace at %s from %s", config.Dir, sourceBranch))
	}

	if err := w.runner.Stream("jj", []string{"workspace", "add", resolvedDir, "--revision", sourceBranch}, repoDir); err != nil {
		return "", fmt.Errorf("jj workspace add failed: %w", err)
	}

	if verbose {
		logging.LogSuccess("jj workspace created successfully")
	}
	return resolvedDir, nil
}
