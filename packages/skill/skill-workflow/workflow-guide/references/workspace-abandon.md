# Workspace Abandon

## Goal

Stop using one explicit workspace, record why, and preserve partial work and recovery information before any optional removal.

## Procedure

1. Resolve the exact target, revision, reason, and explicit selections. Infer a provider
   only when unambiguous and report the inference.
2. Load the selected workspace and provider capabilities. Name and stop on an
   unavailable explicit capability.
3. Inspect unsaved state and report how it can be recovered. Create a snapshot only
   when explicitly selected.
4. Produce the abandonment record with reason, partial state, revision, and recovery
   locator. Persist it only to an explicit destination.
5. Keep the workspace by default. Remove the workspace or native reference only after
   exact-scope authorization; never broaden the target.

Provider-specific snapshot, removal, and reference commands belong to the selected
capability. This operation owns the abandonment record, recovery boundary, and exact
authorization for optional removal.

## Error handling

- Stop on a missing target or reason.
- Preserve dirty state when removal lacks a recoverable snapshot.
- Return a preview when exact authorization is absent.
- Report remaining path, metadata, and reference state after partial removal.
