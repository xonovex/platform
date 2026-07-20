# Workspace Create

## Goal

Create one isolated workspace at an explicit target from an exact source, preserving the relationship needed for a later merge.

## Procedure

1. Resolve the exact target, source, branch when applicable, and explicit capability or
   provider. Infer a provider only when context makes it unambiguous and report the
   inference.
2. Load only the selected or unambiguous workspace capability. Name and stop on an
   unavailable explicit capability.
3. Verify the source exists and that neither target nor native reference would be
   overwritten or used elsewhere.
4. Preview the exact resources, then create them only within the named target.
5. Record the source relationship in provider-native metadata when supported and
   return the workspace locator and revision.

For a Git worktree, require an explicit branch, verify the source and branch resolve,
reject an existing target or branch checked out elsewhere, use `git worktree add`, and
record `branch.<branch>.mergeBackTo`. Do not associate a plan, infer a lifecycle stage,
or derive status from either.

## Error handling

- Stop on a missing or broad target, source, or required branch.
- Report target or branch collisions without overwriting.
- Stop on provider ambiguity or an unavailable explicit capability.
- Report every created resource and its safe recovery action after partial creation.
