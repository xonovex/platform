# Agent Eval Go

Use this CLI to measure whether skills load when they should, whether they improve answers, whether an instruction rule changes what an agent does, and how a flow of agents performs, and to convert the legacy eval files of skills into eval documents.

## Installation

Install the command globally from npm. The package selects the binary for the platform; the supported platforms are Linux and macOS on x64 and arm64, and Windows on x64.

```bash
npm install -g @xonovex/agent-eval-go
```

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
| `agent-eval config show [--sources]` | Prints the merged eval section, and with `--sources` the layer and the file of each value |
| `agent-eval completion <shell>` | Prints a shell completion script |

### Run options

| Option | Effect |
| --- | --- |
| `--project FILE` | Project `agentic.json`, by default the nearest one from the working directory up to the repository root |
| `--approve-config-digest DIGEST` | Approves a project file with this digest: the eval section of the project `agentic.json`, or a harness override in `harnesses.d` beside it |
| `--profile NAME` | Profile of the eval section |
| `--harness claude\|codex` | Agent harness, from `SKILL_EVAL_HARNESS` or the profile by default |
| `--model`, `--judge-model` | Models that answer and that grade. Without them, a profile model or the variables `CLAUDE_MODEL`, `JUDGE_MODEL`, `CODEX_MODEL` and `CODEX_JUDGE_MODEL`, the run names no model and the harness uses its own default; the report and the results say `harness default` |
| `--runs N` | Runs per query or arm, at most 3 |
| `--concurrency N` | Parallel model runs, at most 2 |
| `--max-model-runs N`, `--max-cost-usd USD` | Budget of one invocation, up to the managed cap; the eval section, else 24 runs and 2 US dollars, by default |
| `--batch-size N` | Most units to run in one invocation |
| `--resume` | Continues the incomplete run with the same inputs |
| `--dry-run` | Plans the work and runs no model |
| `--plugin-dir DIR` | Plugin package, or directory of plugins, to search for skills (repeatable) |
| `--results-dir DIR` | Checkpoints and cost history, `.agent-eval/results` by default |
| `--format text\|json` | Output format |

The exit status is 0 on a pass, 1 on a fail, 2 when the input is invalid or an infrastructure fault voided the run, and 3 when a budget stopped the run with work left.

## Configuration

The `eval` section of `agentic.json` holds the plugin directories, the profiles, the protocol settings and the budgets. `shared-config-go` merges the managed, user and project files with the budget flags. A relative plugin directory resolves against the directory of the file that set it.

```json
{
  "$schema": "https://schemas.xonovex.com/agent/agentic.schema.json",
  "version": 1,
  "eval": {
    "pluginDirs": ["packages/runtime/plugin"],
    "defaultProfile": "ci",
    "profiles": { "ci": { "harness": "claude", "model": "claude-sonnet" } },
    "budgets": { "maxModelRuns": 12, "maxCostUsd": 1 }
  }
}
```

The eval section of a project file takes effect only after the user approves the digest of the file: at the terminal prompt, or with `--approve-config-digest`. The approval is kept in `$XDG_STATE_HOME/agentic/trust.json`, and a changed file needs a new approval. Without an approval the run stops with exit status 2. The budgets are caps: a user or managed file sets them, a project file may only lower them, and a flag may raise them up to the managed cap.

The program, the flags and the file layout of each harness come from the harness facts of `shared-agent-go`. A file `harnesses.d/<harness>.json` beside the `agentic.json` of a layer overrides the facts it names, and the schema `https://schemas.xonovex.com/agent/harness-override.schema.json` checks it. The user file applies first, then the project file, then the managed file. A project override can change the program a run starts, so it takes effect only after the user approves its own digest, like the eval section.

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

## Tasks

| Task | Runs |
| --- | --- |
| `npx moon run agent-eval-go:ci-check` | Unit tests, `arch` tests, per-package coverage, lint, vet and build |
| `npx moon run agent-eval-go:ci-integration` | The built binary against a fake `claude` and the eval files of the plugins; no model runs |
