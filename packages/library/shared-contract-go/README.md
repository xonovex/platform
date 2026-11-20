# Shared Contract Go

Use this Go library for the contracts that the agent validate, audit and eval commands share. It has no dependency on the other shared libraries.

## Packages

Import the package that owns the concern.

| Package | Description |
| --- | --- |
| `pkg/findings` | Finding model with the levels pass, warn, drift and fail, summaries and exit codes |
| `pkg/report` | Output formats text, JSON and SARIF, and the parser of a format flag |
| `pkg/sarif` | SARIF 2.1.0 writer with execution notifications |
| `pkg/fs` | File-system port |
| `pkg/fs/memfs` | In-memory file system for unit tests |
| `pkg/fs/osfs` | Host file system, the only package that touches a real disk |
| `pkg/discovery` | File walk that skips `dist`, `node_modules`, `eval-results` and `.git` |
| `pkg/condition` | The instruction conditions present, absent and ablated |
| `pkg/jsonschema` | Port that checks a JSON document against a named schema |
| `pkg/jsonschema/tekuri` | Draft 2020-12 adapter with format assertions that resolves every reference offline |
| `pkg/schemas` | Embedded schemas for sources, evals, results, baselines, cost history, `agentic.json`, provider routes and harness facts |
| `pkg/sources` | Typed source catalog, loader for `sources.json` and `sources.d`, and catalog checks |
| `pkg/wiring` | Composition root that binds the adapters |

## Schemas

Each schema has an `$id` under `https://schemas.xonovex.com/agent/`.

| Schema | Describes |
| --- | --- |
| `sources.schema.json` | Sources with kind, addresses, pins, digests, watched paths, review dates, supersession, claims and the prose of a legacy `SOURCES.md` |
| `eval-trigger.schema.json` | Prompts that should or should not load a skill |
| `eval-output.schema.json` | Cases that compare a run with and without a skill |
| `eval-instruction.schema.json` | Cases that run an instruction present, absent or ablated |
| `eval-orchestration.schema.json` | Flows of steps across several agents |
| `results.schema.json` | One eval run, resumable while it is incomplete |
| `baseline.schema.json` | Pass rates that a later run is compared with |
| `cost-history.schema.json` | Model cost of each eval run |
| `agentic.schema.json` | One layer of `agentic.json`: the version and one section for each tool |
| `agentic-run.schema.json` | The `run` section: agent, provider, sandbox axes, reachable paths and hosts, required guarantees |
| `agentic-providers.schema.json` | The `providers` section: provider routes by name, each checked against the route of `provider.schema.json` |
| `provider.schema.json` | One model provider route: agent, credential, environment, arguments and model slots |
| `harness.schema.json` | The facts of one agent harness: binary, nix package, model access, launch, sandbox, version and headless arguments, and the files it reads |
| `harness-override.schema.json` | A `harnesses.d/<name>.json` file that changes some facts of one harness |
| `run-record.schema.json` | The record a sandboxed run of agent-cli leaves: agent, axes, guarantees, overlay, configuration layers and outcome |
| `agentic-validate.schema.json` | The `validate` section: skipped directories and the paths, options and level cap of each check |
| `agentic-audit.schema.json` | The `audit` section: review ages and the directory that needs sources |
| `agentic-eval.schema.json` | The `eval` section: profiles, protocol settings, plugin directories and budgets |
| `kit.schema.json` | A Docker Sandbox kit v2 specification |
| `predicate.schema.json` | Filesystem, session and model checks on a run |
| `common.schema.json` | Definitions that the other schemas share |

A case holds either `prompt` or `promptFile`. `promptFile` names a Markdown file below the document for a long prompt.

## Usage

Bind the adapters once, then pass the ports to the code that needs them.

```go
loader, err := wiring.NewSourcesLoader(osfs.New())
if err != nil {
	return err
}
catalog, found, err := loader.Load("packages/runtime/plugin/plugin-agentic/skills/authoring-skills")
exit := findings.ExitCode(found, err, findings.DefaultPolicy())
```

## Tasks

| Task | Runs |
| --- | --- |
| `npx moon run shared-contract-go:ci-check` | Build, lint, vet, unit and architecture tests, coverage floors |
| `npx moon run shared-contract-go:ci-integration` | The os adapter against a temporary directory |
