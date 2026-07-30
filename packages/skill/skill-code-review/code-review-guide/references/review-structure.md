# Structuring a review: summary, inline, links, verification

Host-independent assembly of a whole review.

- **One top-level summary + line-anchored inline comments.** Summary = verdict on line 1, then the headline issues; it is the map. Detail, reasoning, and suggested fix live inline on the exact line. For a non-trivial issue, naming the problem and letting the author pick the fix teaches more than prescribing one, but show a concrete fix when it removes ambiguity.
- **Cross-link, do not renumber.** Link the summary to the inline thread, never "see comment 3". Link mutually-dependent comments both ways. No comment-linking on the host → make each comment fully self-contained.
- **Verify before you assert.** Read the code **on the branch under review** (not the base branch, not memory); confirm API/component signatures against source before suggesting a call; cite `file:line`. Unsure a thing is real → use `question`, not `issue`.
- **Match the repo's prose conventions**: commit/PR prose rules (e.g. no em-dash/semicolon/ellipsis) usually apply to review comments too; check the project instructions.
- **Re-check dependents** when you edit a comment others link to.
