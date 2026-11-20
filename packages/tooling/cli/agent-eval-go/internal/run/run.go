// Package run is the engine that carries out a protocol plan. It resumes an
// interrupted run, keeps the work inside the budget, checkpoints after every
// step, and turns an infrastructure fault into an invalid run, so a partial
// result is never mistaken for a pass.
package run

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/budget"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/store"
)

// Options configure one invocation of the engine.
type Options struct {
	Protocol model.Protocol
	Profile  string
	// Model and JudgeModel name the models of the run for its record.
	Model      string
	JudgeModel string
	// Resume continues the incomplete run that has the same inputs, and returns
	// the saved verdict of a complete one.
	Resume bool
	// DryRun plans the work and runs nothing.
	DryRun bool
	// Concurrency is how many units run at once.
	Concurrency int
	// BatchSize is the most units one invocation runs. Zero means no limit.
	BatchSize int
	Guard     *budget.Guard
	Store     store.Store
	// Now is the clock. The composition root passes time.Now.
	Now func() time.Time
}

// Report is the outcome of an invocation.
type Report struct {
	Results model.Results
	// Verdict is set when the run is complete.
	Verdict protocol.Verdict
	// Planned is the number of units of the plan.
	Planned int
	// Pending is the number of units still to run.
	Pending int
	// ModelRuns is the most model runs the pending units may start.
	ModelRuns int
	// Resumed is the number of cases carried over from the saved run.
	Resumed int
	// Dry reports a dry run.
	Dry bool
}

// Execute runs the plan.
func Execute(ctx context.Context, plan protocol.Plan, opts Options) (Report, error) {
	digest := store.Digest(append([]string{string(opts.Protocol), opts.Profile, plan.Subject}, plan.Inputs...)...)
	id := store.RunID(opts.Protocol, plan.Subject, digest)
	results := model.Results{
		Version: model.Version, RunID: id, Protocol: opts.Protocol, Profile: opts.Profile,
		InputsDigest: digest, StartedAt: opts.Now().UTC().Format(time.RFC3339), Status: model.Incomplete,
		Cases: []model.CaseResult{},
	}
	if opts.Resume {
		saved, found, err := opts.Store.Load(id)
		if err != nil {
			return Report{}, err
		}
		if found && saved.InputsDigest == digest && saved.Status != model.Invalid {
			results = saved
			results.Stop = ""
		}
	}
	results.Model, results.JudgeModel = opts.Model, opts.JudgeModel
	done := results.Done()
	var pending []protocol.Unit
	modelRuns := 0
	for _, unit := range plan.Units {
		if _, ok := done[unit.Key()]; !ok {
			pending = append(pending, unit)
			modelRuns += unit.ModelRuns
		}
	}
	report := Report{
		Results: results, Planned: len(plan.Units), Pending: len(pending), ModelRuns: modelRuns,
		Resumed: len(done), Dry: opts.DryRun,
	}
	if opts.DryRun {
		return report, nil
	}
	if results.Status == model.Complete && len(pending) == 0 {
		report.Verdict = plan.Decide(results.Cases)
		return report, nil
	}
	return drive(ctx, plan, opts, report, pending)
}

// drive runs the pending units in chunks of the concurrency, saving a
// checkpoint after each chunk.
func drive(ctx context.Context, plan protocol.Plan, opts Options, report Report, pending []protocol.Unit) (Report, error) {
	results := report.Results
	ran := 0
	for len(pending) > 0 {
		room := len(pending)
		if opts.BatchSize > 0 {
			room = opts.BatchSize - ran
		}
		if room <= 0 {
			results.Stop = model.StopBudget
			break
		}
		chunk := pending[:min(max(opts.Concurrency, 1), len(pending), room)]
		admitted, err := admit(opts.Guard, chunk)
		if admitted == 0 {
			if !errors.Is(err, budget.ErrExceeded) {
				return Report{}, err
			}
			results.Stop = model.StopBudget
			break
		}
		chunk = chunk[:admitted]
		failed := false
		for i, o := range runChunk(ctx, opts.Guard, chunk) {
			if o.err == nil {
				results.Cases = append(results.Cases, o.result)
				continue
			}
			if ctx.Err() != nil {
				return Report{}, ctx.Err()
			}
			failed = true
			results.Cases = append(results.Cases, model.CaseResult{
				ID: chunk[i].ID, Condition: chunk[i].Condition, Outcome: model.Errored, Error: o.err.Error(),
			})
		}
		if failed {
			results.Status, results.Stop = model.Invalid, model.StopInfrastructure
			results.FinishedAt = opts.Now().UTC().Format(time.RFC3339)
			return save(opts, report, results)
		}
		ran += len(chunk)
		pending = pending[len(chunk):]
		if err := opts.Store.Save(results); err != nil {
			return Report{}, err
		}
	}
	report.Pending = len(pending)
	if len(pending) == 0 {
		results.Status, results.Stop = model.Complete, ""
		results.FinishedAt = opts.Now().UTC().Format(time.RFC3339)
		report.Verdict = plan.Decide(results.Cases)
	}
	return save(opts, report, results)
}

// save totals the usage of the cases and writes the final checkpoint.
func save(opts Options, report Report, results model.Results) (Report, error) {
	var total model.Usage
	for _, c := range results.Cases {
		if c.Usage != nil {
			total = total.Add(*c.Usage)
		}
	}
	if total != (model.Usage{}) {
		results.Usage = &total
	}
	report.Results = results
	return report, opts.Store.Save(results)
}

// admit reserves the model runs of each unit of the chunk in order and returns
// how many units it reserved. It stops at the first unit that does not fit.
func admit(guard *budget.Guard, chunk []protocol.Unit) (int, error) {
	for i, unit := range chunk {
		if err := guard.Admit(unit.ModelRuns); err != nil {
			return i, err
		}
	}
	return len(chunk), nil
}

type outcome struct {
	result model.CaseResult
	err    error
}

func runChunk(ctx context.Context, guard *budget.Guard, chunk []protocol.Unit) []outcome {
	outcomes := make([]outcome, len(chunk))
	var wg sync.WaitGroup
	for i, unit := range chunk {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := unit.Run(ctx)
			if err != nil {
				outcomes[i] = outcome{err: fmt.Errorf("%s: %w", unit.Key(), err)}
				return
			}
			if result.Usage != nil {
				guard.Record(*result.Usage)
			}
			outcomes[i] = outcome{result: result}
		}()
	}
	wg.Wait()
	return outcomes
}
