// Package jj is the workspace=jj leaf: Jujutsu workspaces layered on a git repo.
package jj

import (
	"fmt"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

// Workspace is the jj VCS variant: it creates or reuses a jj workspace.
type Workspace struct {
	host wsshared.Host
	log  logging.Logger
}

// New creates the jj workspace variant over host, writing through log.
func New(host wsshared.Host, log logging.Logger) *Workspace {
	return &Workspace{host: host, log: log}
}

// Available reports whether the jj binary is on PATH.
func (w Workspace) Available() bool {
	_, err := w.host.LookPath("jj")
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
	if _, err := w.host.Stat(resolvedDir); err == nil {
		if verbose {
			w.log.Info(fmt.Sprintf("Reusing existing jj workspace at %s", config.Dir))
		}
		return resolvedDir, nil
	}

	if !w.Available() {
		return "", fmt.Errorf("jj is not installed or not on PATH; install from https://martinvonz.github.io/jj/")
	}

	sourceBranch := config.SourceBranch
	if sourceBranch == "" {
		sourceBranch = wsshared.GetCurrentBranchSync(w.host, repoDir)
		if sourceBranch == "" {
			return "", fmt.Errorf("failed to determine source revision")
		}
	}

	if verbose {
		w.log.Info(fmt.Sprintf("Creating jj workspace at %s from %s", config.Dir, sourceBranch))
	}

	if err := w.host.Stream("jj", []string{"workspace", "add", resolvedDir, "--revision", sourceBranch}, repoDir); err != nil {
		return "", fmt.Errorf("jj workspace add failed: %w", err)
	}

	if verbose {
		w.log.Success("jj workspace created successfully")
	}
	return resolvedDir, nil
}
