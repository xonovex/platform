// Package launch is the application core of agent-cli: a pure planner that
// turns the run options into a resolved Plan, and an executor that carries the
// Plan out through the ports it is given. It names no concrete isolator,
// provisioner, workspace variant or terminal, and it does no I/O of its own.
package launch

import (
	"fmt"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/validation"
	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	provshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runspec"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

// Checkout is a worktree or workspace the run creates before it launches.
type Checkout struct {
	VCS    wsp.VCSType
	Config wsshared.Config
	// Dir is the absolute directory the checkout resolves to.
	Dir string
}

// Plan is the portable Spec of a run plus what only this CLI does around it:
// the isolator and provisioner it launches through, the checkout it creates,
// the overlay it builds and the terminal it runs in.
type Plan struct {
	runspec.Spec
	Isolator    isoshared.Isolator
	Provisioner provshared.Provisioner
	// SourceDir is the absolute directory the run starts from; Checkout is nil
	// when the run uses SourceDir itself.
	SourceDir string
	Checkout  *Checkout
	// Overlay is the artifact overlay of the run, nil when the run has none.
	Overlay *OverlayRequest
	// AgentArgs are the arguments of the user, which end the argument vector of
	// the agent process; the overlay places its arguments before them.
	AgentArgs []string
	Terminal  *termshared.TerminalConfig
	DryRun    bool
	Verbose   bool
}

// NewPlan resolves options, which already carry every configuration layer,
// against the working directory and the home directory of the CLI, and selects
// the isolator and the provisioner from registry under the requested policy. It
// fails closed on any invalid choice or unmet guarantee and performs no I/O.
func NewPlan(options Options, cwd, home string, registry *sandbox.Registry[isoshared.Isolator]) (Plan, error) {
	return buildPlan(options, cwd, home, registry, true)
}

// evaluatePolicy evaluates each guarantee of the policy against the capabilities
// of the selected sandbox.
func evaluatePolicy(plan Plan) []policy.Outcome {
	return policy.Evaluate(plan.Capabilities, plan.Policy)
}

// advisoryPlan is NewPlan without the demand: it resolves the same plan whether
// or not the sandbox meets the policy, and leaves the guarantees to the caller,
// which reports each of them. Every other invalid choice still fails.
func advisoryPlan(options Options, cwd, home string, registry *sandbox.Registry[isoshared.Isolator]) (Plan, error) {
	return buildPlan(options, cwd, home, registry, false)
}

func buildPlan(options Options, cwd, home string, registry *sandbox.Registry[isoshared.Isolator], enforce bool) (Plan, error) {
	agent, provider, err := planAgent(options)
	if err != nil {
		return Plan{}, err
	}
	sourceDir := absolute(cwd, options.WorkDir)
	checkout, err := planCheckout(options, sourceDir)
	if err != nil {
		return Plan{}, err
	}
	overlayRequest, err := planOverlay(options, agent, cwd, sourceDir)
	if err != nil {
		return Plan{}, err
	}
	spec, choice, err := planSpec(options, agent, provider, sourceDir, cwd, home, registry, enforce)
	if err != nil {
		return Plan{}, err
	}
	return Plan{
		Spec:        spec,
		Isolator:    choice.Isolator,
		Provisioner: choice.Provisioner,
		SourceDir:   sourceDir,
		Checkout:    checkout,
		Overlay:     overlayRequest,
		AgentArgs:   options.AgentArgs,
		Terminal:    planTerminal(options),
		DryRun:      options.DryRun,
		Verbose:     options.Verbose,
	}, nil
}

// DisplayDir is the directory a dry run reports: the checkout it would create,
// or the source directory.
func (p Plan) DisplayDir() string {
	if p.Checkout != nil {
		return p.Checkout.Dir
	}
	return p.SourceDir
}

// planCheckout validates the worktree options and resolves the checkout
// directory, or returns nil when no worktree branch is requested.
func planCheckout(options Options, sourceDir string) (*Checkout, error) {
	if options.WorktreeBranch == "" {
		return nil, nil
	}
	if err := validation.ValidateBranch(options.WorktreeBranch); err != nil {
		return nil, fmt.Errorf("invalid --worktree-branch: %w", err)
	}
	if options.WorktreeSourceBranch != "" {
		if err := validation.ValidateBranch(options.WorktreeSourceBranch); err != nil {
			return nil, fmt.Errorf("invalid --worktree-source-branch: %w", err)
		}
	}
	vcs := wsp.VCSType(options.VCS)
	if !vcs.IsValid() {
		return nil, fmt.Errorf("unknown --vcs %q; valid values: git, jj", options.VCS)
	}
	dir := options.WorktreeDir
	if dir == "" {
		dir = wsshared.GetDefaultDir(options.WorktreeBranch, filepath.Base(sourceDir))
	}
	return &Checkout{
		VCS: vcs,
		Config: wsshared.Config{
			SourceBranch: options.WorktreeSourceBranch,
			Branch:       options.WorktreeBranch,
			Dir:          dir,
		},
		Dir: absolute(sourceDir, dir),
	}, nil
}

// planTerminal returns the terminal wrapper configuration, or nil when the run
// uses no wrapper.
func planTerminal(options Options) *termshared.TerminalConfig {
	if options.Terminal == "" {
		return nil
	}
	return &termshared.TerminalConfig{
		Type:        termshared.TerminalType(options.Terminal),
		SessionName: options.TerminalSession,
		WindowName:  options.TerminalWindow,
		Detach:      options.TerminalDetach,
	}
}
