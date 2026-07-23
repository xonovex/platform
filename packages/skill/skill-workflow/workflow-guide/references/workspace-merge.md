# Workspace Merge

## Goal

Preview or integrate one exact workspace into one exact destination while preserving
the source workspace and its native reference.

## Procedure

- [ ] Resolve source workspace and integration destination as independent bindings.
- [ ] Verify identity, cleanliness, source metadata, destination freshness, conflicts,
      and any exact required validation-result bindings.
- [ ] Derive the integration adapter and return the exact preview and strategy.
- [ ] On apply, require runtime authorization and unchanged revisions, then integrate
      only the previewed change.
- [ ] Verify the native receipt and return integration evidence inline.

Workspace merge owns integration only. It does not remove a workspace, branch,
reference, or metadata. Semantic typecheck, lint, build, and test evidence belongs to
separate Validate results when policy requires it; merge consumes those exact results
as preconditions.

`preview` and `apply` are the only valid effect modes.

## Error handling

- Return blocked on a dirty, stale, mismatched, conflicting, or insufficiently
  validated source.
- Preserve the workspace after every failure, including a successful merge whose
  receipt is not yet reconciled.
- Use Workspace cleanup as a later separately authorized transaction.
