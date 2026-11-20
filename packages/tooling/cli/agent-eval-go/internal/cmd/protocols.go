package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	evalsection "github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/eval"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	contractreport "github.com/xonovex/platform/packages/library/shared-contract-go/pkg/report"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/budget"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/cost"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/report"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/run"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/store"
)

// DefaultResultsDir is where run checkpoints go, below the working directory.
const DefaultResultsDir = ".agent-eval/results"

// CostFile is the name of the cost history inside the results directory.
const CostFile = "cost-history.json"

// protocolSpec describes how a protocol command takes its target.
type protocolSpec struct {
	name    model.Protocol
	use     string
	short   string
	args    cobra.PositionalArgs
	catalog bool
}

var specs = []protocolSpec{
	{model.Trigger, "trigger <skill>", "Measure whether a skill loads when it should", cobra.ExactArgs(1), true},
	{model.Routing, "routing [plugin]", "Measure whether the model picks the right skill among several", cobra.MaximumNArgs(1), true},
	{model.Output, "output <skill>", "Measure whether a skill improves the answer, with and without it", cobra.ExactArgs(1), true},
	{model.Instruction, "instruction <eval.json>", "Measure whether a rule of an instruction file changes what an agent does", cobra.ExactArgs(1), false},
	{model.Orchestration, "orchestration <flow.json>", "Measure a flow in which several agents each run a step", cobra.ExactArgs(1), false},
}

// options are the flags every protocol command takes.
type options struct {
	project, approveDigest, profile, harness, model string
	judgeModel, split                               string
	runs, batchSize, concurrency, maxModelRuns      int
	threshold, maxCostUSD                           float64
	resultsDir, format                              string
	resume, dryRun                                  bool
	pluginDirs                                      []string
}

func (o *options) bind(c *cobra.Command) {
	f := c.Flags()
	bindConfig(c, &o.project, &o.approveDigest)
	f.StringVar(&o.profile, "profile", "", "profile of the eval section to run with")
	f.StringVar(&o.harness, "harness", "", "agent harness: claude or codex (env SKILL_EVAL_HARNESS)")
	f.StringVar(&o.model, "model", "", "model that answers the prompts")
	f.StringVar(&o.judgeModel, "judge-model", "", "model that grades the answers")
	f.StringVar(&o.split, "split", "", "query split to run: train, validation or test (default all)")
	f.IntVar(&o.runs, "runs", 0, fmt.Sprintf("runs per query or arm, at most %d", settings.MaxRuns))
	f.IntVar(&o.batchSize, "batch-size", 0, "most units to run in one invocation")
	f.IntVar(&o.concurrency, "concurrency", 0, fmt.Sprintf("parallel model runs, at most %d", settings.MaxConcurrency))
	f.IntVar(&o.maxModelRuns, "max-model-runs", 0, fmt.Sprintf("model runs the invocation may start, at most the managed cap (default the eval section, else %d)", evalsection.DefaultMaxModelRuns))
	f.Float64Var(&o.threshold, "threshold", 0, "share of runs that must load the skill for a query to count as triggered")
	f.Float64Var(&o.maxCostUSD, "max-cost-usd", 0, fmt.Sprintf("spend the invocation may reach in US dollars, at most the managed cap (default the eval section, else %.2f)", evalsection.DefaultMaxCostUSD))
	f.StringVar(&o.resultsDir, "results-dir", DefaultResultsDir, "directory for run checkpoints")
	f.StringVar(&o.format, "format", string(contractreport.Text), "output format: "+contractreport.Names(report.Formats()...))
	f.BoolVar(&o.resume, "resume", false, "continue the incomplete run with the same inputs")
	f.BoolVar(&o.dryRun, "dry-run", false, "plan the work and run no model")
	f.StringArrayVar(&o.pluginDirs, "plugin-dir", nil, "plugin package directory, or a directory of plugins (repeatable)")
}

func (o *options) flags(c *cobra.Command) settings.Flags {
	f := settings.Flags{
		Profile: o.profile, Harness: o.harness, Model: o.model, JudgeModel: o.judgeModel, Runs: o.runs,
		BatchSize: o.batchSize, Concurrency: o.concurrency, Split: o.split,
	}
	if c.Flags().Changed("threshold") {
		f.Threshold = &o.threshold
	}
	return f
}

// config names what the command reads of agentic.json. A budget flag joins the
// merge only when the user sets it.
func (o *options) config(c *cobra.Command) configRequest {
	r := configRequest{project: o.project, approveDigest: o.approveDigest}
	if c.Flags().Changed("max-model-runs") {
		r.maxModelRuns = &o.maxModelRuns
	}
	if c.Flags().Changed("max-cost-usd") {
		r.maxCostUSD = &o.maxCostUSD
	}
	return r
}

func newProtocolCommands(deps Dependencies) []*cobra.Command {
	commands := make([]*cobra.Command, len(specs))
	for i, spec := range specs {
		o := &options{}
		c := &cobra.Command{
			Use: spec.use, Short: spec.short, Args: spec.args,
			RunE: func(c *cobra.Command, args []string) error { return runProtocol(c, deps, spec, o, args) },
		}
		o.bind(c)
		commands[i] = c
	}
	return commands
}

func runProtocol(c *cobra.Command, deps Dependencies, spec protocolSpec, o *options, args []string) error {
	format, err := contractreport.Parse(o.format, report.Formats()...)
	if err != nil {
		return err
	}
	project, projectDir, err := deps.loadProject(o.config(c))
	if err != nil {
		return err
	}
	resolved, err := settings.Resolve(project, o.flags(c), deps.Env, spec.name)
	if err != nil {
		return err
	}
	proto, err := deps.Protocols.Get(spec.name)
	if err != nil {
		return err
	}
	facts, err := deps.loadHarnesses(o.config(c))
	if err != nil {
		return err
	}
	exec, predicates, err := deps.Engine(facts)
	if err != nil {
		return err
	}
	env := protocol.Env{FS: deps.FS, Reader: deps.reader(), Exec: exec, Predicates: predicates}
	if spec.catalog {
		if env.Catalog, err = deps.loadCatalog(facts, project, projectDir, o.pluginDirs); err != nil {
			return err
		}
	}
	target := ""
	if len(args) > 0 {
		target = args[0]
		if !spec.catalog {
			target = resolvePath(deps.Dir, target)
		}
	}
	plan, err := proto.Plan(protocol.Input{Target: target, Settings: resolved, Env: env})
	if err != nil {
		return err
	}
	guard := budget.New(resolved.Limits)
	resultsDir := resolvePath(deps.Dir, o.resultsDir)
	st := store.Store{FS: deps.FS, Reader: deps.reader(), Dir: resultsDir}
	rep, err := run.Execute(context.Background(), plan, run.Options{
		Protocol: spec.name, Profile: resolved.ProfileName, Model: resolved.ModelLabel(), JudgeModel: resolved.JudgeModelLabel(),
		Resume: o.resume, DryRun: o.dryRun,
		Concurrency: resolved.Concurrency, BatchSize: resolved.BatchSize, Guard: guard, Store: st, Now: deps.Now,
	})
	if err != nil {
		return err
	}
	out, err := report.Render(rep, format)
	if err != nil {
		return err
	}
	if _, err := c.OutOrStdout().Write(out); err != nil {
		return err
	}
	if !rep.Dry {
		if err := recordCost(deps, resultsDir, rep, resolved.ModelLabel(), guard, deps.Now()); err != nil {
			return err
		}
	}
	if code := report.ExitCode(rep); code != findings.ExitClean {
		return ExitError{Code: code}
	}
	return nil
}

// recordCost appends the run to the cost history. A run that spent nothing is
// not recorded.
func recordCost(deps Dependencies, resultsDir string, rep run.Report, modelName string, guard *budget.Guard, now time.Time) error {
	if rep.Results.Usage == nil {
		return nil
	}
	runs, _ := guard.Spent()
	file := resultsDir + "/" + CostFile
	s := cost.Store{FS: deps.FS, Reader: deps.reader()}
	history, err := s.Load(file)
	if err != nil {
		return err
	}
	history = cost.Append(history, model.CostEntry{
		RunID: rep.Results.RunID, RecordedAt: now.UTC().Format(time.RFC3339), Protocol: rep.Results.Protocol,
		Profile: rep.Results.Profile, Model: modelName, ModelRuns: runs, Usage: *rep.Results.Usage,
	})
	return s.Save(file, history)
}
