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

Refines the findings produced by `/xonovex-workflow:pr-review-analyze` before they are published by `/xonovex-workflow:pr-review-post`. It is the middle stage of the review pipeline, the analogue of `plan-refine` for a plan: it resolves feedback on the findings, updates them in place, and presents the result for another pass. Refining here means the comments land right the first time, with no edit-after-posting churn.

It refines whatever findings are in play: the in-session findings from a same-session `pr-review-analyze` run (no file needed), or a `findings.json` passed as the argument for a cross-session or hand-edited review.

It composes with the **code review skill** so relabelling and rewording stay consistent with the label vocabulary, decorations, and self-contained-comment rules.

## Prerequisites

- Findings exist for this review, in the session context or as a findings file in the shared schema
- Feedback is available as a walk-through, inline annotations, prompt instructions, or any mix

## Goal

- Resolve every piece of feedback on the findings, one finding at a time
- Keep each finding self-contained and correctly labelled after edits
- Update the findings file in place, preserving the schema and valid anchors
- STOP after each pass so the user can review, annotate again, or approve

## Arguments

- `findings-file` (optional): Path to a findings JSON. Omit it to refine the findings already in the session context
- `--walk` (optional): Step through every finding interactively, even when no annotations are present
- `--final` (optional): Treat this as the last pass, validate fully and mark the findings ready to post

## Feedback Sources

- **Walk-through**: for each finding, show its rendered body, anchor (`path:line`), label, and blocking flag, then capture the call.
- **Inline annotations**: a `note` or `_action` field the user adds to a finding in the JSON (`drop`, `reword`, `relabel`, `merge:<id>`, `split`).
- **Prompt instructions**: free-text in the invocation ("make 4 non-blocking", "drop the skeleton nit").

## Per-Finding Operations

For each finding, the available actions:

- **keep** — leave as is
- **reword** — edit the `body` (stay self-contained, show the fix not just the problem)
- **relabel** — change `label` and/or `decoration` + `blocking` (e.g. `issue (blocking)` to `suggestion (non-blocking)`)
- **re-anchor** — fix `path` / `line` / `lineType` to a real new-file diff line
- **merge** — fold into another finding (combine bodies, keep one anchor)
- **split** — break one finding into two anchored at different lines
- **drop** — remove it entirely

Also editable: the top-level `summary` (lead with positives, keep the numbered priority points so the poster can cross-link them).

## Core Workflow

1. **Load craft + findings**: invoke the code review skill, take the in-context findings or read the given file, validate they parse.
2. **Collect feedback**: gather walk-through answers, inline annotations, and prompt instructions.
3. **Resolve one by one**: apply the chosen operation to each finding. After a relabel, re-check the decoration matches intent, after a re-anchor, confirm the line exists in the diff.
4. **Re-derive the summary**: ensure it still reflects the surviving findings and numbers the blocking ones.
5. **Write back**: update the refined findings in place, the session context, or the file if one was given, preserving the schema.
6. **Report and STOP**: print a table of the changes (kept / reworded / relabelled / merged / split / dropped) and stop for the next review round. With `--final`, run full validation and mark it ready.

## Implementation Details

- **Edit JSON safely**: when a file is involved, read, mutate, and re-serialize with `python3` + `json` so bodies with backticks and newlines stay valid, do not hand-patch it.
- **Anchor check on re-anchor / split**: the new `path`/`line` must be an `ADDED` / `CONTEXT` line in the diff, otherwise it orphans when posted.
- **Label discipline**: every finding keeps a known label and an explicit blocking / non-blocking decoration after editing.
- **Idempotent**: re-running with no new feedback changes nothing.

## Error Handling

- No findings in context and no file given → run `/xonovex-workflow:pr-review-analyze` first, or pass a findings file
- Given findings file missing or unparseable → report the path and the parse error
- An annotation references an unknown finding → list the current ones and ask
- A merge / split leaves a finding without an anchor → require a valid `path`/`line` before writing
- `--final` with an invalid label, missing decoration, or orphan-prone anchor → block and report which finding

## Gotchas

- This stage edits the findings, not the PR — nothing is posted here. That is the point: refine before publishing so the live comments are clean.
- Keep bodies self-contained through every edit, merged or reworded comments still must stand alone.
- A relabel that drops `blocking` should also flip the decoration, the two must agree.
- Re-anchoring is where orphans creep in, re-validate the line against the diff each time.

## Examples

```bash
# Walk the in-session findings and decide each
/xonovex-workflow:pr-review-refine --walk

# Process inline annotations / prompt instructions on a findings file
/xonovex-workflow:pr-review-refine review.json

# Final pass: validate and mark ready to post
/xonovex-workflow:pr-review-refine review.json --final

# Same-session pipeline, no file
/xonovex-workflow:pr-review-analyze feat/x
/xonovex-workflow:pr-review-refine --walk
/xonovex-workflow:pr-review-post feat/x

# Cross-session pipeline via a file
/xonovex-workflow:pr-review-analyze feat/x --out review.json
/xonovex-workflow:pr-review-refine review.json --walk
/xonovex-workflow:pr-review-post feat/x --findings review.json
```
