# Agent Eval Go

Use this CLI to measure whether skills load when they should, whether they improve answers, whether an instruction rule changes what an agent does, and how a flow of agents performs, and to convert the legacy eval files of skills into eval documents.

## Commands

| Command | Measures |
| --- | --- |
| `agent-eval trigger <skill>` | Whether the skill loads for the queries that should load it, and not for the others |
| `agent-eval routing [plugin]` | Whether the model picks the right skill when several are on offer |
| `agent-eval output <skill>` | Whether the answers with the skill beat the answers without it, by tier |
| `agent-eval instruction <eval.json>` | Whether a rule of an instruction file changes behavior, by comparing present, absent and ablated runs |
| `agent-eval orchestration <flow.json>` | Whether a flow of agents, each running one step, meets its predicates within its budget |
| `agent-eval baseline record\|compare <results.json>` | Records the pass rate of each case, and fails a later run that dropped by more than the tolerance |
| `agent-eval cost` | Sums what the recorded runs cost |
| `agent-eval migrate <path>...` | Converts `evals.json` and `eval-queries.json` into `eval-output.json` and `eval-trigger.json` |
| `agent-eval completion <shell>` | Prints a shell completion script |

### Run options

| Option | Effect |
| --- | --- |
| `--project FILE` | Project file, `agentic.eval.json` in the working directory by default |
| `--profile NAME` | Profile of the project file |
| `--harness claude\|codex` | Agent harness, from `SKILL_EVAL_HARNESS` or the profile by default |
| `--model`, `--judge-model` | Models that answer and that grade |
| `--runs N` | Runs per query or arm, at most 3 |
| `--concurrency N` | Parallel model runs, at most 2 |
| `--max-model-runs N`, `--max-cost-usd USD` | Budget of one invocation, 24 runs and 2 US dollars by default |
| `--batch-size N` | Most units to run in one invocation |
| `--resume` | Continues the incomplete run with the same inputs |
| `--dry-run` | Plans the work and runs no model |
| `--plugin-dir DIR` | Plugin package, or directory of plugins, to search for skills (repeatable) |
| `--results-dir DIR` | Checkpoints and cost history, `.agent-eval/results` by default |
| `--format text\|json` | Output format |

The exit status is 0 on a pass, 1 on a fail, 2 when the input is invalid or an infrastructure fault voided the run, and 3 when a budget stopped the run with work left.

## Layout

| Package | Concern |
| --- | --- |
| `cmd/agent-eval` | Composition root |
| `internal/cmd` | Command tree, the driving adapter |
| `internal/registry` | The one file that names the concrete harnesses, protocols and predicates |
| `internal/run` | Drives the units of a plan under a budget, with checkpoints and resume |
| `internal/protocol/{trigger,routing,output,instruction,orchestration}` | One eval protocol each |
| `internal/harness`, `internal/harness/{claude,codex}` | Harness port and its leaves |
| `internal/predicate`, `internal/predicate/{filesystem,session,llm}` | Predicate port and its leaves |
| `internal/execute`, `internal/judge` | Runs one request with retries, and grades with a model |
| `internal/document`, `internal/model`, `internal/catalog`, `internal/settings`, `internal/budget` | Documents, plugin discovery, settings and caps |
| `internal/store`, `internal/baseline`, `internal/cost`, `internal/report`, `internal/migrate`, `internal/ablation` | Checkpoints, baselines, cost history, output, legacy conversion, rule ablation |
| `internal/process`, `internal/scratch` | Ports for programs and throwaway directories, with `osexec`, `osscratch` and test doubles |

## Tasks

| Task | Runs |
| --- | --- |
| `npx moon run agent-eval-go:ci-check` | Unit tests, `arch` tests, per-package coverage, lint, vet and build |
| `npx moon run agent-eval-go:ci-integration` | The built binary against a fake `claude` and the eval files of the plugins; no model runs |
