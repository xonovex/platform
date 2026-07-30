# worktree-abandon: Abandon Feature Worktree with Documented Reason

Inspect an abandoned feature and return a recovery record. This operation is read-only:
it never commits, resets, removes, tags, pushes, or changes provider state.

## Steps

1. Verify the exact feature worktree (`*-feature-*`) and get the reason from the caller.
2. Inspect the worktree and branch for dirty or uncommitted state.
3. Capture the reason, current branch, HEAD revision, dirty-state summary, and the
   command needed to recover or resume the work.
4. Report any untracked work, unpushed revisions, detached state, or missing remote
   that would affect recovery.
5. Return the record inline. Use separate Commit, Publish, or Cleanup operations for
   any requested mutation.

## Gotchas

- Abandoning without recording _why_ the approach failed loses the learning, always capture it with the recovery revision
- Do not create a safety commit or tag: even a preservation-oriented write exceeds the Abandon operation
- Do not treat a request to abandon as authorization to discard or clean up work
