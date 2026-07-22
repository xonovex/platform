# Workspace Cleanup

## Goal

Inspect and clean only explicitly named workspaces after classifying their state and presenting a reviewable removal preview.

## Procedure

1. Resolve every exact target and selection; never discover additional removal targets
   implicitly. Infer a provider only when unambiguous and report the inference.
2. Load the selected workspace and provider capabilities. Name and stop on an
   unavailable explicit capability.
3. Classify each target as merged, stale, active, dirty, or unknown and show its
   recovery revision. Keep active, dirty, unmerged, and unknown targets by default.
4. Preview every workspace, reference, and administration record that would be
   removed. Reject broad patterns, unresolved variables, and implicit home or
   repository roots.
5. Remove only the confirmed set and report what was removed, retained, recoverable,
   and unrecoverable.

Provider-specific inventory, removal, reference, and administration commands belong
to the selected capability. This operation owns classification, preview, exact target
selection, and recovery reporting.

## Error handling

- Stop on a missing, broad, or unresolved target.
- Retain dirty or unmerged targets unless exact forced removal is selected.
- Return a preview when exact authorization is absent.
- Report remaining workspace, reference, and metadata separately after partial cleanup.
