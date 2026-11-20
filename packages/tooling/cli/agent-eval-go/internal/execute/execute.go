// Package execute runs one harness request end to end: it makes a private
// workspace, stages what the harness leaf asks for, starts the program inside
// the sandbox of the run, feeds its output to the leaf's scanner, ends what the
// sandbox left running and removes the workspace. A run that fails
// for an infrastructure reason is tried again, and a run that keeps failing is
// an error, never a failed verdict.
package execute

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/scratch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
)

// Attempts is how many times a run is tried before its failure voids the batch.
const Attempts = 3

// InfrastructureError reports a run that failed on every attempt. It voids the
// whole batch: a partial result must never read as a pass.
type InfrastructureError struct {
	Harness  string
	Attempts int
	Detail   string
}

func (e *InfrastructureError) Error() string {
	return fmt.Sprintf("%s run failed after %d attempts: %s", e.Harness, e.Attempts, e.Detail)
}

// Executor is the harness.Executor built from its ports.
type Executor struct {
	Harnesses *harness.Registry
	// Confiner places every harness process inside the sandbox of the run.
	Confiner confine.Confiner
	Runner   process.Runner
	Scratch  scratch.Scratch
	FS       fs.FileSystem
	// Now is the clock. The composition root passes time.Now.
	Now func() time.Time
	// Pause waits between attempts. A test passes a no-op.
	Pause func(attempt int)
}

var _ harness.Executor = Executor{}

// Execute runs the request, retrying a transient failure.
func (e Executor) Execute(ctx context.Context, req harness.Request) (harness.Observation, error) {
	h, err := e.Harnesses.Get(req.Harness)
	if err != nil {
		return harness.Observation{}, err
	}
	var last harness.Observation
	for attempt := 1; attempt <= Attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return harness.Observation{}, err
		}
		last, err = e.once(ctx, h, req)
		if err != nil {
			return harness.Observation{}, err
		}
		if last.Failure == "" {
			return last, nil
		}
		if attempt < Attempts && e.Pause != nil {
			e.Pause(attempt)
		}
	}
	return harness.Observation{}, &InfrastructureError{Harness: req.Harness, Attempts: Attempts, Detail: last.Failure}
}

func (e Executor) once(ctx context.Context, h harness.Harness, req harness.Request) (harness.Observation, error) {
	workspace, err := e.Scratch.Make("agent-eval-")
	if err != nil {
		return harness.Observation{}, err
	}
	defer func() { _ = e.Scratch.Remove(workspace) }()

	plan, err := h.Plan(req, workspace)
	if err != nil {
		return harness.Observation{}, err
	}
	for _, step := range plan.Stage {
		if err := e.stage(step); err != nil {
			return harness.Observation{}, fmt.Errorf("stage %s: %w", step.Dst, err)
		}
	}
	for _, fixture := range req.Fixtures {
		if err := e.FS.WriteFile(path.Join(workspace, fixture.Path), []byte(fixture.Content)); err != nil {
			return harness.Observation{}, fmt.Errorf("fixture %s: %w", fixture.Path, err)
		}
	}

	scanner := h.Scan(req)
	started := e.Now()
	// The "--" keeps a variadic flag before it from taking the prompt.
	args := append(append([]string{}, plan.Args...), "--", plan.Prompt)
	confined, err := e.Confiner.Confine(confine.Request{Harness: req.Harness, Process: process.Request{
		Dir: plan.Dir, Name: h.Binary(), Args: args, Env: plan.Env, Timeout: req.Caps.Timeout,
	}})
	if err != nil {
		return harness.Observation{}, fmt.Errorf("confine %s: %w", req.Harness, err)
	}
	defer confined.Release()
	result, err := e.Runner.Run(ctx, confined.Process, process.Lines(scanner.Feed))
	if confined.Stop != nil && (err != nil || result.TimedOut || result.Stopped) {
		if stopErr := e.stop(ctx, *confined.Stop); stopErr != nil {
			return harness.Observation{}, errors.Join(stopErr, err)
		}
	}
	if err != nil {
		if ctx.Err() != nil {
			return harness.Observation{}, ctx.Err()
		}
		// A program that cannot start is retried like any other fault.
		return harness.Observation{Failure: err.Error()}, nil
	}
	obs := scanner.Finish(result, e.Now().Sub(started).Milliseconds())
	if req.Role == harness.RoleGenerate || req.Role == harness.RoleAgent {
		obs.Files = e.snapshot(workspace, plan.Stage)
	}
	return obs, nil
}

// stop ends what a confined process left running. It runs after a cancelled
// context too, so a cancelled eval leaves no sandbox behind. A sandbox that is
// tearing itself down can refuse a stop until it is gone, so a failed stop is
// tried again. A stop that keeps failing leaves a sandbox running, which is a
// fault and not a result.
func (e Executor) stop(ctx context.Context, request process.Request) error {
	var err error
	for attempt := 1; attempt <= Attempts; attempt++ {
		if err = e.stopOnce(context.WithoutCancel(ctx), request); err == nil {
			return nil
		}
		if attempt < Attempts && e.Pause != nil {
			e.Pause(attempt)
		}
	}
	return err
}

func (e Executor) stopOnce(ctx context.Context, request process.Request) error {
	result, err := e.Runner.Run(ctx, request, nil)
	switch {
	case err != nil:
		return fmt.Errorf("stop sandbox: %w", err)
	case !result.Succeeded():
		return fmt.Errorf("stop sandbox: %s exited with status %d: %s", request.Name, result.ExitCode, strings.TrimSpace(result.Stderr))
	}
	return nil
}

// maxSnapshotBytes bounds one file of the workspace snapshot, so a run that
// writes a huge file cannot exhaust memory.
const maxSnapshotBytes = 1 << 20

// snapshot reads the files of the workspace by relative path. What the plan
// staged (skills, plugin manifests, the isolated home) is the harness's own
// setup and not a result of the run, so the snapshot leaves it out.
func (e Executor) snapshot(root string, staged []harness.Stage) map[string]string {
	skip := map[string]bool{}
	for _, step := range staged {
		skip[strings.TrimPrefix(strings.TrimPrefix(step.Dst, root), "/")] = true
	}
	files := map[string]string{}
	var walk func(dir, rel string)
	walk = func(dir, rel string) {
		entries, err := e.FS.ReadDir(dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			child, relChild := path.Join(dir, entry.Name), path.Join(rel, entry.Name)
			if skip[relChild] {
				continue
			}
			if entry.IsDir {
				walk(child, relChild)
				continue
			}
			if info, err := e.FS.Stat(child); err != nil || info.Size > maxSnapshotBytes {
				continue
			}
			if data, err := e.FS.ReadFile(child); err == nil {
				files[relChild] = string(data)
			}
		}
	}
	walk(root, "")
	return files
}

func (e Executor) stage(step harness.Stage) error {
	switch step.Kind {
	case harness.StageDir:
		return e.FS.MkdirAll(step.Dst)
	case harness.StageWrite:
		return e.FS.WriteFile(step.Dst, []byte(step.Content))
	case harness.StageAppend:
		current, err := e.FS.ReadFile(step.Dst)
		if err != nil {
			return err
		}
		return e.FS.WriteFile(step.Dst, append(current, step.Content...))
	case harness.StageCopy:
		return e.copyTree(step.Src, step.Dst)
	default:
		return errors.New("unknown stage kind " + string(step.Kind))
	}
}

func (e Executor) copyTree(src, dst string) error {
	info, err := e.FS.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir {
		data, err := e.FS.ReadFile(src)
		if err != nil {
			return err
		}
		return e.FS.WriteFile(dst, data)
	}
	if err := e.FS.MkdirAll(dst); err != nil {
		return err
	}
	entries, err := e.FS.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := e.copyTree(path.Join(src, entry.Name), path.Join(dst, entry.Name)); err != nil {
			return err
		}
	}
	return nil
}
