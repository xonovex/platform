# Review-analyze — turn a diff into structured findings

Read a branch diff and produce findings in the shared contract — a top-level summary plus line-anchored findings, each with a Conventional Comments label and a blocking flag. This is the analysis step; it posts nothing. The findings shape lives in [findings-schema.md](findings-schema.md).

## Procedure

1. **Compute the diff** — `git diff <base>...<branch>` (or read a pre-saved unified diff). Track new-file line numbers from each hunk header so findings can anchor to `ADDED` / `CONTEXT` lines (see [findings-schema.md](findings-schema.md)).
2. **Read for correctness first** — bugs, regressions, broken edge cases, accessibility and contract violations. Re-read the surrounding code to confirm each claim; do not assert from the diff alone.
3. **Then quality** — reuse, simplification, efficiency, naming, consistency; for the dimension vocabulary (robustness, duplication, dead code, over-abstraction) see **code-quality-guide**. Right-size to the change — a small diff does not need a long list.
4. **Label each finding** — pick a label and a decoration; set `blocking: true` only for must-fix-before-merge items. If unsure a thing is real, use `question`, not `issue`.
5. **Anchor each finding** — a specific new-file `path` + `line` + `lineType` that exists in the diff.
6. **Draft the summary** — lead with what works, then number the priority (blocking) points so cross-links can attach later.

## Effort dial

- `low` / `medium` favour fewer, high-confidence findings.
- `high` widens coverage and may include uncertain ones — lean on `question` for those rather than asserting `issue`.

## Comparing against prior findings

When a prior findings set is given, diff against it to track recurrence — comparing **findings only, never the PR**, so the comparison stays platform-independent:

- Match each fresh finding to the prior set by `path` plus body similarity, **not** by line number (which shifts after edits).
- Tag each fresh finding `new` or `recurring`; report prior findings with no match as `gone`.
- Carry the prior `commentId` onto a recurring finding so it stays linked to its existing thread.

## Notes

- Anchors must be **new-file** line numbers — otherwise the poster orphans them.
- Do not invent findings to fill space. Zero issues on a clean diff is a valid result.
- Lead the summary with what works; an all-negatives review hides the signal.
- Match the repo's prose conventions (commit / PR rules often apply to review comments too) — check the project instructions.
