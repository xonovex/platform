# Workspace Merge

## Goal

Validate one exact workspace and merge it into one exact destination while preserving the workspace after validation, update, or merge failure.

## Procedure

1. Resolve the exact source workspace, destination, revision, and selections. Infer a
   provider only when unambiguous and report the inference.
2. Load the selected workspace, validation, merge, and provider capabilities. Name and
   stop on any unavailable explicit capability.
3. Verify source and destination identity, source metadata, cleanliness, and required
   typecheck, lint, build, and tests.
4. Update the source against the current destination using the selected capability;
   never resolve conflicts by guesswork.
5. Preview the final effect. Merge only with exact-scope authorization, and remove the
   source only after success when removal was explicit.

For a Git worktree, verify the worktree and branch relationship, require
`branch.<branch>.mergeBackTo`, reject a mismatch with the selected destination, fetch,
rebase the feature branch onto the current destination, and then merge. Never remove
the worktree after a failed validation, rebase, or merge.

## Error handling

- Stop on a dirty, unvalidated, stale, or mismatched source.
- Leave the workspace intact and identify unresolved files after conflicts.
- Return a preview when exact authorization is absent.
- Never remove the workspace or native reference after a failed merge.
