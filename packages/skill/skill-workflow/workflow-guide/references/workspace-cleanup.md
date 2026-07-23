# Workspace Cleanup

## Goal

Exclusively preview or remove explicitly named workspace resources after classifying
their state and recovery boundary.

## Procedure

- [ ] Resolve every target independently; never discover additional removal targets.
- [ ] Classify each workspace, native reference, and attached metadata as merged,
      stale, active, dirty, unmerged, or unknown.
- [ ] Return an exact preview of each proposed removal and retained resource.
- [ ] On apply, require runtime authorization for the unchanged preview and remove
      only its exact targets.
- [ ] Report applied, failed, skipped, and unknown effects plus remaining recovery.

Workspace cleanup is the only workspace transaction that removes resources. Merge
and abandon never call it implicitly. Pruning is limited to metadata attached to the
explicit targets; broad patterns, unresolved variables, implicit roots, and
discovered neighbors are invalid.

`preview` and `apply` are the only valid effect modes.

## Error handling

- Preserve dirty, active, unmerged, unknown, or uninspectable targets unless exact
  forced removal is separately authorized by policy.
- After partial removal, report workspace, native reference, and metadata status
  independently.
- Reconcile unknown outcomes before any retry.
