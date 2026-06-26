---
description: >-
  Refine review findings one at a time — reword, relabel, re-anchor, merge,
  split, or drop — before publishing (in session, or a findings file)
allowed-tools:
  - Read
  - Write
  - Edit
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "[findings-file] [--walk] [--final]"
---

# /xonovex-workflow:pr-review-refine – Refine Review Findings Before Posting

Resolves feedback on the findings from `/xonovex-workflow:pr-review-analyze` before `/xonovex-workflow:pr-review-post` publishes them — the middle stage of the review pipeline, the analogue of `plan-refine`. It refines whatever findings are in play: the in-session findings from a same-session analyze run (no file needed), or a `findings.json` passed as the argument for a cross-session or hand-edited review.

## Arguments

- `findings-file` (optional): Path to a findings JSON. Omit it to refine the findings already in the session context
- `--walk` (optional): Step through every finding interactively, even when no annotations are present
- `--final` (optional): Treat this as the last pass, validate fully and mark the findings ready to post

## Delegation

Load `code-review-guide` (plugin `xonovex-skill-code-review`) via the `Skill` tool and perform its **`review-refine`** operation under the **`findings-schema`** contract. They are the source of truth — the per-finding operations (keep / reword / relabel / re-anchor / merge / split / drop), label discipline, anchor re-validation, summary re-derivation, idempotency, safe `python3`+`json` editing, the stop-after-each-pass rhythm, and the `--final` full validation — do not restate them.

This command only supplies the orchestration around that op:

- **Findings source / write-back**: take the in-session findings, or read/write the `findings-file` if given; with no findings and no file, point the user at `/xonovex-workflow:pr-review-analyze` first.
- **Feedback collection**: gather feedback from any mix of `--walk` (step each finding, showing its body, `path:line` anchor, label, and blocking flag), inline `_action`/`note` annotations on a finding (`drop` | `reword` | `relabel` | `merge:<id>` | `split`), and free-text prompt instructions — then map each to its `review-refine` operation.

## Examples

```bash
# Walk the in-session findings and decide each
/xonovex-workflow:pr-review-refine --walk

# Process annotations / prompt instructions on a findings file
/xonovex-workflow:pr-review-refine review.json

# Final pass: validate and mark ready to post
/xonovex-workflow:pr-review-refine review.json --final

# Cross-session pipeline via a file
/xonovex-workflow:pr-review-analyze feat/x --out review.json
/xonovex-workflow:pr-review-refine review.json --walk
/xonovex-workflow:pr-review-post feat/x --findings review.json
```
