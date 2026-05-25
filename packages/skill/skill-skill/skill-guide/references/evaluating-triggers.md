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
- Reference implementation: [scripts/eval-triggers.py](../scripts/eval-triggers.py) — PEP 723 self-contained Python. Streams the tool-call log and terminates the run on first match. Targets Claude Code's `claude` CLI; adapt the `check_triggered` function for other harnesses. Run with `uv run scripts/eval-triggers.py <queries.json> <skill-name>`.
- **Cost control** — a run where the skill _doesn't_ fire would otherwise execute the whole task, which dominates token spend. Keep it cheap: `--model haiku`, the default tool-blocking (everything but `Skill`, so non-triggering runs stay short), a `--max-budget-usd` per-run ceiling, and `--runs 1` on the `train` split while iterating.

## Train / Validation Split

- ~60% train, ~40% validation, proportional positives/negatives
- Use **train** failures to guide edits; **validation** only as a generalization check
- Keep the split fixed across iterations

## Optimization Loop

1. Evaluate on both sets
2. Identify train-set failures
3. Revise the description:
   - Should-trigger failing → broaden or add context
   - Should-not-trigger false-positive → add specificity / "skip" clauses
   - Avoid copying failed-query keywords (overfit) — address the category
   - If stuck after a few iterations, try a structurally different framing
   - Stay ≤1024 chars
4. Repeat until train passes or improvement plateaus (~5 iterations typically enough)
5. Pick the iteration with the best **validation** pass rate (may not be the last)
6. Sanity-check with 5-10 fresh queries never seen during optimization

## Gotchas

- Don't optimize against the validation set — that defeats the split
- A description that scores perfectly on train but poorly on validation is overfit; pick an earlier iteration instead
- Trigger rate isn't binary — a query that triggers 1/3 of the time still indicates instability; widen the eval set or tighten the description
- Stopping the run early once outcome is clear cuts cost — many harnesses let you abort once the skill is or isn't invoked
