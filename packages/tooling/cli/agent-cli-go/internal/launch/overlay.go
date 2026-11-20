package launch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
)

// OverlayRequest asks for the artifact overlay of one run: the skills, agents,
// commands and instructions of installed plugins and of directories, laid out the
// way the agent reads them.
type OverlayRequest struct {
	Agent *types.AgentConfig
	// WorkDir is the directory whose project instruction files the overlay yields to.
	WorkDir string
	Plugins []string
	// Directories are absolute directories in the standard artifact layout.
	Directories []string
	// Store is the plugin store, empty for the default one.
	Store string
	// DryRun asks for the plan of the overlay without writing it.
	DryRun bool
}

// OverlayResult is the overlay of a run.
type OverlayResult struct {
	// Root is the directory that holds the overlay, empty when nothing was placed.
	Root        string
	Environment []string
	Arguments   []string
	BindPaths   []string
	RoBindPaths []string
	// Summary describes what was placed and what was skipped, one line each.
	Summary []string
	// Sources are the plugins and directories the overlay drew on, each with the
	// digest of its content.
	Sources []OverlaySource
	// Cleanup removes the overlay. It is never nil.
	Cleanup func()
}

// Overlay source kinds.
const (
	SourcePlugin    = "plugin"
	SourceDirectory = "directory"
)

// OverlaySource is one source of the overlay and the digest of its content.
type OverlaySource struct {
	Kind   string
	Name   string
	Digest string
}

// Overlays is the port that builds the overlay of a run.
type Overlays interface {
	Prepare(request OverlayRequest) (OverlayResult, error)
}

// planOverlay validates the overlay options and returns the request, or nil when
// the run has none. It fails closed on an agent whose overlay would cut it off
// from its model access.
func planOverlay(options Options, agent *types.AgentConfig, cwd, sourceDir string) (*OverlayRequest, error) {
	if len(options.Plugins)+len(options.OverlayDirs) == 0 {
		return nil, nil
	}
	for _, name := range options.Plugins {
		if !artifact.ValidName(name) {
			return nil, fmt.Errorf("invalid --plugin %q: a plugin name is a kebab-case name", name)
		}
	}
	for _, dir := range options.OverlayDirs {
		if dir == "" {
			return nil, errors.New("invalid --overlay-dir: the directory is empty")
		}
	}
	if err := overlay.Check(agent); err != nil {
		return nil, err
	}
	request := &OverlayRequest{
		Agent:       agent,
		WorkDir:     sourceDir,
		Plugins:     options.Plugins,
		Directories: absoluteAll(cwd, options.OverlayDirs),
		DryRun:      options.DryRun,
	}
	if options.Store != "" {
		request.Store = absolute(cwd, options.Store)
	}
	return request, nil
}

// prepareOverlay builds the overlay of the plan. A run without one gets an empty
// result whose Cleanup does nothing.
func prepareOverlay(overlays Overlays, plan Plan) (OverlayResult, error) {
	if plan.Overlay == nil {
		return OverlayResult{Cleanup: func() {}}, nil
	}
	if overlays == nil {
		return OverlayResult{}, errors.New("the overlay is not available in this build")
	}
	return overlays.Prepare(*plan.Overlay)
}

// applyOverlay adds the environment, the arguments and the mounts of the overlay
// to run. The overlay arguments come right before userArgs, the arguments of the
// user that end the argument vector of the process.
func applyOverlay(run isoshared.RunConfig, userArgs []string, result OverlayResult) isoshared.RunConfig {
	run.Process.Env = append(append([]string{}, run.Process.Env...), result.Environment...)
	run.Process.Argv = insertBefore(run.Process.Argv, len(userArgs), result.Arguments)
	run.BindPaths = append(append([]string{}, run.BindPaths...), result.BindPaths...)
	run.RoBindPaths = append(append([]string{}, run.RoBindPaths...), result.RoBindPaths...)
	return run
}

// insertBefore returns a copy of argv with extra placed before its last tail
// elements.
func insertBefore(argv []string, tail int, extra []string) []string {
	at := len(argv) - tail
	out := make([]string, 0, len(argv)+len(extra))
	out = append(out, argv[:at]...)
	out = append(out, extra...)
	return append(out, argv[at:]...)
}

// logOverlay reports the overlay of a dry run.
func logOverlay(log logging.Logger, result OverlayResult) {
	if len(result.Summary) == 0 {
		return
	}
	log.Info("  Overlay:")
	for _, line := range result.Summary {
		log.Info("    " + strings.TrimSpace(line))
	}
}
