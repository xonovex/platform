# self-review: Review Your Own Diff First

**Guideline:** Read your own complete diff and get CI green before assigning reviewers.

**Rationale:** If the PR feels too large or unclear to you, it will to the reviewer too. A self-review trims debug code, stray files, and unclear names, and a description written while reviewing your own diff is sharper. Reviewers should not start on a red CI.

**How to Apply:**

1. Read the full diff file by file as if you were the reviewer.
2. Remove debug logging, commented-out code, unrelated changes, and stray files.
3. Confirm the diff matches the description, and update the description if it drifted.
4. Run the project checks (typecheck / lint / build / test) and wait for CI to pass.
5. Keep the PR a draft until it is genuinely ready, then flip to ready and assign.

**Example:**

```text
// Bad
Open the PR straight from the editor, CI red, reviewers tagged.

// Good
Self-review the diff, drop a leftover debug log, push, wait for green CI, then request review.
```

**Counter-Example:** None.

**Related:** [size-and-atomicity.md](./size-and-atomicity.md)
