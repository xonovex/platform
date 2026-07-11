# worktree-abandon: Abandon Feature Worktree with Documented Reason

Abandon a feature: record the reason in its plan, optionally commit current state, and keep the worktree by default (remove only if requested) so insights can be extracted later.

## Steps

1. Verify in a feature worktree (`*-feature-*`); get reason from user or prompt.
2. Read plan path from `git config branch.<branch>.plan`.
3. Optional commit: `git add . && git commit -m "wip: abandoned work on <feature>"`.
4. Update the plan (unless asked to skip):

```yaml
status: "abandoned"
abandoned_reason: "Superseded by OAuth 2.0"
abandoned_date: "2026-05-13"
```

Add an `## Abandonment Notes` section with the detailed explanation and learnings.

5. Optional: `git worktree remove <path>`.

## Gotchas

- Abandoning without recording _why_ the approach failed loses the learning — always capture it
- A plan marked `abandoned` is the canonical signal for downstream ops — don't reuse the same plan file for a fresh attempt; bump a new path
- `git worktree remove` doesn't delete the branch ref — pair with `git branch -D <feature-branch>` (or tag `abandoned/<name>` first) to preserve the work in history
