# Workspace Abandon

## Goal

Return an inline record explaining why one exact workspace will no longer be used,
while preserving every workspace resource and recovery locator.

## Procedure

- [ ] Resolve the exact workspace, revision, reason, and known partial state.
- [ ] Inspect unsaved or untracked state without creating a snapshot.
- [ ] Report recovery references, unknown state, and the safe retry boundary.
- [ ] Preserve the workspace, branch, reference, and metadata.
- [ ] Return the abandonment `OperationResult` inline with `effect.mode: inspect`.

Workspace abandon records only. It does not snapshot, remove, prune, persist the
record, or update provider status. Administrative checkpointing remains a
policy-authorized runtime effect; explicit removal belongs only to Workspace cleanup.

## Error handling

- Return blocked on a missing target or reason.
- Mark uninspectable state unknown and preserve it.
- Never claim recovery is durable when no exact native recovery binding exists.
