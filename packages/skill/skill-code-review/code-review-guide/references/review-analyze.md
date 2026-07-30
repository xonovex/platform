# Review-analyze: turn a diff into structured findings

Read a branch diff and produce findings in the shared contract ([findings-schema.md](findings-schema.md)). Posts nothing.

## Procedure

1. **Compute the diff**: `git diff <base>...<branch>` (or a pre-saved unified diff). Track new-file line numbers from each hunk header for anchoring.
2. **Correctness first**: bugs, regressions, broken edge cases, accessibility and contract violations. Re-read surrounding code to confirm each claim; do not assert from the diff alone.
3. **Then quality**: reuse, simplification, efficiency, naming, consistency; for the dimension vocabulary see **code-quality-guide**.
4. **Label + anchor**: pick label + decoration, `blocking: true` only for must-fix items; unsure → `question`, not `issue`. Anchor to a real new-file `path` + `line` + `lineType`.
5. **Draft the summary**: lead with what works, number the blocking points.

Zero issues on a clean diff is a valid result: do not invent findings to fill space.

## Effort dial

Effort scales coverage, not what gets reported. Report every finding at every setting and let [review-refine.md](review-refine.md) drop what should not survive; withholding one here loses it, because refine only sees what analysis returned.

- `low` / `medium`: a fast pass over the highest-risk hunks.
- `high`: wide coverage across the diff and the code surrounding it.

## Comparing against prior findings

Compare **findings only, never the PR** (keeps it platform-independent):

- Match each fresh finding to the prior set by `path` + body similarity, **not** line number (shifts after edits).
- Tag fresh findings `new` / `recurring`; report unmatched prior findings as `gone`.
- Carry the prior `commentId` onto a recurring finding so it stays linked to its thread.
