# Review-refine — refine findings before posting

Refine the findings (in the shared contract) before they are published: resolve feedback, update them in place, and present the result for another pass. This is the middle step — nothing is posted here. The point is that comments land right the first time, with no edit-after-posting churn.

## Per-finding operations

For each finding:

- **keep** — leave it as is.
- **reword** — edit the `body` (stay self-contained; show the fix, not just the problem).
- **relabel** — change `label` and/or `decoration` + `blocking` (e.g. `issue (blocking)` → `suggestion (non-blocking)`).
- **re-anchor** — fix `path` / `line` / `lineType` to a real new-file diff line.
- **merge** — fold into another finding (combine bodies, keep one anchor).
- **split** — break one finding into two anchored at different lines.
- **drop** — remove it entirely.

Also editable: the top-level `summary` — keep it leading with positives and numbering the surviving blocking points so the poster can cross-link them.

## Discipline

- **Re-validate the anchor** after any re-anchor or split — the new `path` / `line` must be an `ADDED` / `CONTEXT` line in the diff, otherwise it orphans when posted.
- **Label discipline** — every finding keeps a known label and an explicit blocking / non-blocking decoration after editing. A relabel that drops `blocking` must also flip the decoration off `(blocking)`; the two must agree.
- **Re-derive the summary** — after edits, ensure it still reflects the surviving findings and numbers the blocking ones.
- **Idempotent** — re-running with no new feedback changes nothing.

## Stop after each pass

Refining is iterative. After applying a round of feedback, report the changes (kept / reworded / relabelled / re-anchored / merged / split / dropped) and **STOP** for the next review round. A final pass runs full validation — every finding has a known label, an explicit decoration, and a non-orphan anchor — and marks the findings ready to post.
