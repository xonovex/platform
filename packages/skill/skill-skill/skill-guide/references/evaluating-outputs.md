# evaluating-outputs: Eval-Driven Iteration on Output Quality

## Contents

[Scope](#scope) · [Test Case Shape](#test-case-shape) · [Automated Runner](#automated-runner) · [Designing Prompts](#designing-prompts) · [Workspace Layout](#workspace-layout) · [Running Eval Pairs](#running-eval-pairs) · [Writing Assertions](#writing-assertions) · [Grading](#grading) · [Aggregating](#aggregating-benchmarkjson) · [Pattern Analysis](#pattern-analysis) · [Human Review](#human-review) · [Iteration Signals](#iteration-signals) · [Iteration Principles](#iteration-principles) · [Gotchas](#gotchas)

## Scope

This is about **whether the outputs are good once the skill activates** — not about whether the skill triggers on the right prompts.

## Test Case Shape

- **Prompt** — realistic user message (not "process this data")
- **Expected output** — human-readable reference for what success looks like
- **Assertions** — specific, binary PASS/FAIL checks the judge grades against
- **Files** (optional) — input artifacts the skill needs

Store in `evals/evals.json` inside the skill directory:

```json
{
  "skill_name": "{name}",
  "evals": [
    {
      "id": 1,
      "prompt": "{realistic prompt}",
      "expected_output": "{success criterion}",
      "assertions": ["{verifiable check}", "{another check}"],
      "files": ["evals/files/{input}"]
    }
  ]
}
```

## Automated Runner

`scripts/eval-outputs.py <evals.json> <skill-name>` runs the whole loop below: each eval runs in an isolated `claude -p` context **with the skill** and **without it** (Skill tool blocked), in parallel; outputs are graded by a reference-guided LLM-as-judge (binary per assertion, evidence required); results land in the workspace layout and `benchmark.json`. Env: `RUNS`, `CONCURRENCY`, `CLAUDE_MODEL`, `JUDGE_MODEL`. Do the manual loop when you need human grading or file-producing skills the runner doesn't capture.

## Designing Prompts

- **Start with 2-3 cases** — don't over-invest before first results
- **Vary** phrasing (casual / precise), file paths, column names, context
- **Cover at least one edge case** — malformed input, ambiguous request
- Defer assertions until you've seen actual outputs

## Workspace Layout

```
{skill}-workspace/iteration-N/
├── eval-{name}/
│   ├── with_skill/    (outputs/, timing.json, grading.json)
│   └── without_skill/ (outputs/, timing.json, grading.json)
└── benchmark.json
```

Each pass through the loop gets its own `iteration-N/` directory.

## Running Eval Pairs

- Each test runs **twice**: with the skill, and without (or vs. previous version snapshot)
- **Fresh context per run** — no leftover state; isolated subagent or new session
- Capture per run: outputs, `timing.json` (`total_tokens`, `duration_ms`), `grading.json`

```json
// timing.json
{"total_tokens": 84852, "duration_ms": 23332}
```

## Writing Assertions

### Good

- "Output is valid JSON"
- "Bar chart has labeled axes"
- "Report includes at least 3 recommendations"

### Weak

- "Output is good" (vague)
- "Output uses exactly the phrase X" (brittle)

Rule: programmatically verifiable, specific, observable. Not everything needs an assertion — reserve human review for style/feel.

## Grading

- Per assertion: `PASS` / `FAIL` + **concrete evidence** (quote / file reference)
- Use scripts for mechanical checks (valid JSON, row count, file dims); LLM for the rest
- **No benefit of the doubt** — a "Summary" heading with one vague sentence is FAIL
- **Review the assertions themselves** — fix ones that always pass, always fail, or aren't verifiable

```json
// grading.json
{
  "assertion_results": [
    {
      "text": "Both axes are labeled",
      "passed": false,
      "evidence": "Y-axis labeled 'Revenue ($)' but X-axis has no label"
    }
  ],
  "summary": {"passed": 3, "failed": 1, "total": 4, "pass_rate": 0.75}
}
```

## Aggregating: benchmark.json

```json
{
  "run_summary": {
    "with_skill": {"pass_rate": {"mean": 0.83}, "tokens": {"mean": 3800}},
    "without_skill": {"pass_rate": {"mean": 0.33}, "tokens": {"mean": 2100}},
    "delta": {"pass_rate": 0.5, "tokens": 1700}
  }
}
```

`delta` shows cost (extra time/tokens) vs benefit (pass-rate lift).

## Pattern Analysis

- **Always-pass in both** — assertion useless, drop or replace
- **Always-fail in both** — broken assertion, too-hard test, or wrong check
- **Passes with skill, fails without** — where the skill earns its keep; understand why
- **High variance across runs** — instructions ambiguous; tighten with examples
- **Time/token outliers** — read the transcript to find the bottleneck

## Human Review

- Catches what assertions miss (style, intent, holistic quality)
- Record actionable per-case feedback in `feedback.json`; empty string = looked fine
- "Missing axis labels" is actionable; "looks bad" is not

## Iteration Signals

Feed all three into an LLM (or yourself) along with current `SKILL.md`:

- **Failed assertions** — specific gaps
- **Human feedback** — broader quality issues
- **Execution transcripts** — _why_ things went wrong

Then revise. Loop: propose → apply → re-run → grade → review.

## Iteration Principles

- **Generalize** from feedback — fixes address categories, not single prompts
- **Stay lean** — if pass rates plateau, try removing rules rather than adding
- **Explain why** — reasoned instructions ("X because Y") beat rigid directives
- **Bundle repeated work** — if every run reinvents the same script, move it to `scripts/`
- **Blind-compare** for holistic quality (LLM judge with version labels hidden)

## Gotchas

- `stddev` only meaningful with multiple runs per case — ignore in tiny test sets
- Token-savings count as a regression even with the same pass rate — track both
- Test cases written before first run often test the wrong thing; revise after seeing outputs
- Don't add narrow patches for specific failing prompts — that's overfit
