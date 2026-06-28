// Package jj is the workspace=jj leaf: Jujutsu workspaces layered on a git repo.
package jj

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	wsshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/workspace/shared"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/scriptlib"
)

// Workspace is the jj VCS variant: it creates or reuses a jj workspace.
type Workspace struct{}

// New creates the jj workspace variant.
func New() *Workspace { return &Workspace{} }

// Available reports whether the jj binary is on PATH.
func (Workspace) Available() bool {
	_, err := exec.LookPath("jj")
	return err == nil
}

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
			scriptlib.LogInfo(fmt.Sprintf("Reusing existing jj workspace at %s", config.Dir))
		}
		return resolvedDir, nil
	}

	if !w.Available() {
		return "", fmt.Errorf("jj is not installed or not on PATH; install from https://martinvonz.github.io/jj/")
	}

	sourceBranch := config.SourceBranch
	if sourceBranch == "" {
		sourceBranch = wsshared.GetCurrentBranchSync(repoDir)
		if sourceBranch == "" {
			return "", fmt.Errorf("failed to determine source revision")
		}
	}

	if verbose {
		scriptlib.LogInfo(fmt.Sprintf("Creating jj workspace at %s from %s", config.Dir, sourceBranch))
	}

	cmd := exec.Command("jj", "workspace", "add", resolvedDir, "--revision", sourceBranch)
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("jj workspace add failed: %w", err)
	}

	if verbose {
		scriptlib.LogSuccess("jj workspace created successfully")
	}
	return resolvedDir, nil
}
