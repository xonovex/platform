# Review-refine: refine findings before posting

Return a refined copy of findings in the shared contract so a later Publish operation
can deliver them without edit-after-posting churn. Do not mutate an input file or
provider resource, and post nothing.

## Per-finding operations

- **keep**: leave as is.
- **reword**: edit `body` (stay self-contained; show the fix).
- **relabel**: change `label` and/or `decoration` + `blocking`.
- **re-anchor**: fix `path` / `line` / `lineType` to a real new-file diff line.
- **merge**: fold into another finding (combine bodies, keep one anchor).
- **split**: break one finding into two at different lines.
- **drop**: remove entirely.

Also edit the top-level `summary` to track the surviving findings and their blocking numbering.

## Discipline

- **Re-validate the anchor** after any re-anchor/split: must be an ADDED/CONTEXT diff line or it orphans.
- **Label discipline**: a relabel dropping `blocking` must also flip the decoration off `(blocking)`; the two must agree.
- **Idempotent**: re-running with no new feedback changes nothing.

## Stop after each pass

After applying a round of feedback, report the changes (kept / reworded / relabelled / re-anchored / merged / split / dropped) and **STOP** for the next review round. A final pass runs full validation (every finding has a known label, explicit decoration, non-orphan anchor) and marks the findings ready to post.
