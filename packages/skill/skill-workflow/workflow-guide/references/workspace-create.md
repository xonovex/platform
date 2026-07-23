# Workspace Create

## Goal

Preview or create one isolated workspace at an exact target from an exact source,
without executing domain work.

## Procedure

- [ ] Resolve source and target as independent bindings, including exact source
      revision and provider-native branch when applicable.
- [ ] Derive the workspace adapter and inspect target/reference collisions.
- [ ] Return the exact resources and metadata that preview would create.
- [ ] On apply, require runtime authorization and unchanged preconditions, then create
      only the previewed resources.
- [ ] Return every created locator, revision, receipt, and recovery action inline.

Workspace create owns creation only. It does not execute work, merge, abandon, or
remove. `preview` and `apply` are the only valid effect modes.

## Error handling

- Return blocked on broad targets, missing source revisions, provider ambiguity, or
  collisions.
- After partial creation, report each created resource and safe compensation without
  applying cleanup implicitly.
