# worktree-abandon: Abandon Feature Worktree with Documented Reason

Abandon a feature: record the reason and recovery revision, optionally commit current
state, and keep the worktree by default. Remove it only when explicitly requested.

## Steps

1. Verify in a feature worktree (`*-feature-*`); get reason from user or prompt.
2. Inspect the worktree and branch for dirty or uncommitted state.
3. Capture the reason, current branch, HEAD revision, dirty-state summary, and the
   command needed to recover or resume the work.
4. Optional commit: stage only the intended partial state and commit it with a
   conventional message that explains the stopped attempt.
5. Optional removal: preview the exact worktree and branch effects, then use
   `git worktree remove <path>` only with explicit authorization.

## Gotchas

- Abandoning without recording _why_ the approach failed loses the learning — always capture it with the recovery revision
- `git worktree remove` doesn't delete the branch ref — pair with `git branch -D <feature-branch>` (or tag `abandoned/<name>` first) to preserve the work in history
