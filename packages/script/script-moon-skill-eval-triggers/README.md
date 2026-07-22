# moon-skill-eval-triggers

TRIGGER eval runs the queries in an `eval-queries.json` and reports whether the target **Skill fires** for each, killing the `claude` run on first match so non-triggering runs stay cheap.

## Usage

```bash
npx moon-skill-eval-triggers [eval-queries.json] [skill-name] [train|validation|all] [--harness claude|codex] [--runs N] [--threshold F] [--model M] [--max-budget-usd N] [--plugin-dir PATH]
# queries defaults to ./eval-queries.json; skill-name defaults to the name in ./SKILL.md
```

Requires the selected CLI on PATH and its API credential. Claude is the default harness.

Each run loads only the target local plugin and its declared local plugin dependencies, with empty user/project/local settings and MCP configuration. Plugin preflight requires skill-only manifests and rejects commands, agents, hooks, MCP, LSP, and settings components. `Skill` is the only exposed tool; `Read`, shell, file mutation, web, browser, task, and undeclared plugins are unavailable. The process is killed as soon as the target Skill call appears.

Negative runs are limited to one turn, 2,000 streamed response characters, 60 seconds, and $0.05. The spend cap can only be lowered, runs per query are limited to three, and one invocation may schedule at most 24 model runs. Larger query sets must be split into bounded batches. Any timeout, cap, output limit, process failure, or unavailable target skill exits `2` immediately instead of being scored as a routing result. There are no automatic retries.

The Codex adapter copies only the target guide into an isolated user skill directory, ignores user configuration and execution rules, uses ephemeral JSONL output in a read-only sandbox, and detects a marker appended only to the staged copy. Its CLI has no dollar-budget flag, so the run-count, batch, timeout, and output ceilings are the enforced bounds.

## Catalog routing

```bash
npx moon-skill-eval-routing packages/skill --split all --harness claude --runs 3
```

Catalog routing finds exact queries that one skill marks positive and one or more other skills mark negative. It loads the expected owner and all of those competing skills in the same isolated run, then passes only when the expected owner is selected. This tests ownership ranking and multi-skill coexistence; the target-only evaluator remains useful for measuring an individual description's recall and precision.

Use `--owners alpha-guide,beta-guide` to restrict scenarios by expected owner. Pull requests use a one-run changed-owner smoke; scheduled and full manual runs use three runs per scenario so the routing rate reflects nondeterminism.
