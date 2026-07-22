# evaluating-triggers: Eval-Driven Trigger-Rate Optimization

## Scope

This is about **whether the skill activates** for a given prompt — not whether its outputs are good once activated.

## Trigger Eval Set

- ~20 queries total: 8-10 should-trigger + 8-10 should-not-trigger
- **Vary** phrasing (formal/casual/typos), explicitness (named vs not), detail (terse vs context-heavy), complexity (1-step vs multi-step)
- **Strongest should-trigger queries:** skill helps but connection isn't obvious from the prompt
- **Strongest should-not-trigger queries:** near-misses — shared keywords but different intent (not "what's the weather")

## Making Eval Queries Realistic

These belong **in the test queries**, not in the description itself. Generic queries ("analyze this CSV") fail to surface description weaknesses that real prompts expose. Include:

- File paths (`~/Downloads/report_v2.xlsx`)
- Personal context ("my manager asked…")
- Specific names, column references, data values
- Casual language, abbreviations, occasional typos

## Eval Set Shape

```json
[
  {"query": "{realistic prompt}", "should_trigger": true},
  {"query": "{near-miss prompt}", "should_trigger": false}
]
```

## Measuring Trigger Rate

- Run each query ≥3 times (model is nondeterministic)
- `trigger_rate = triggers / runs`
- Should-trigger passes if rate ≥0.5; should-not-trigger passes if rate <0.5
- Detect triggering via the harness's tool-call log (varies by harness)
- Repository target runner: `npx moon-skill-eval-triggers <queries.json> <skill-name> --harness claude|codex`. It isolates one target for description recall and precision. Catalog routing runner: `npx moon-skill-eval-routing packages/skill --split all --harness claude|codex`. It finds queries with one positive owner and shared negative near-misses, loads those skills together, and passes only when the owner wins. The Claude adapter uses local plugins; the Codex adapter stages only the selected candidate guides under an isolated `.agents/skills`, ignores inherited configuration and rules, and uses an ephemeral read-only run. The bundled [scripts/eval-triggers.py](../scripts/eval-triggers.py) remains the portable target-only Claude reference implementation.
- **Cost and mutation control** — a run where the skill _doesn't_ fire would otherwise execute the whole task. The runner disables inherited settings, MCP, persistence, Chrome, and every tool except `Skill`; limits each run to one turn, 60 seconds, 2,000 response characters, and at most $0.05; limits one invocation to 24 model runs; and never retries. Use `--runs 1` on `train` while iterating.
- Treat timeout, cap exhaustion, output-limit, process failure, or a missing target skill as invalid infrastructure (`exit 2`), never as a negative routing result.

## Train / Validation Split

- ~60% train, ~40% validation, proportional positives/negatives
- Use **train** failures to guide edits; **validation** only as a generalization check
- Keep the split fixed across iterations

## Optimization Loop

1. Evaluate on both sets
2. Identify train-set failures
3. Revise the description:
   - Should-trigger failing → broaden or add context
   - Should-not-trigger false-positive → add positive intent or context that distinguishes the target; never add skip/out-of-scope clauses
   - Avoid copying failed-query keywords (overfit) — address the category
   - If stuck after a few iterations, try a structurally different framing
   - Stay ≤1024 chars
4. Repeat until train passes or improvement plateaus (~5 iterations typically enough)
5. Pick the iteration with the best **validation** pass rate (may not be the last)
6. Sanity-check with 5-10 fresh queries never seen during optimization
7. Run catalog routing so target-only gains do not create a sibling-skill ranking regression

## Gotchas

- Don't optimize against the validation set — that defeats the split
- A description that scores perfectly on train but poorly on validation is overfit; pick an earlier iteration instead
- Trigger rate isn't binary — a query that triggers 1/3 of the time still indicates instability; widen the eval set or tighten the description
- Stopping the run early once outcome is clear cuts cost — many harnesses let you abort once the skill is or isn't invoked
- Target-only perfection can still lose when sibling descriptions are present — catalog routing is the coexistence gate
