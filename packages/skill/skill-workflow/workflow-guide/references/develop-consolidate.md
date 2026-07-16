# develop-consolidate: Consolidate Development Results

## Core workflow

1. Resolve every source Development reference at the exact requested revision and verify its Planning relationship, outcome, change identity, and validation evidence.
2. Resolve the selected target development workspace and exact starting revision through its workspace provider.
3. Reject stale sources, overlapping changes without an explicit conflict policy, failed mandatory assignments, a protected integration target, or a request that would silently change providers.
4. Apply source changes in deterministic dependency order. Record conflicts and resolutions; preserve each constituent Development result and do not rewrite its evidence.
5. Run the consolidated validation set and compare it with every source result's validation scope. A source pass is not evidence that the combination passes.
6. Publish a new Development result containing source references/revisions, target start/result revisions, consolidation order, conflicts, validation, unresolved findings, and partial-failure behavior.

## Boundary

Consolidation prepares a development candidate. It does not record Acceptance, authorize a protected target change, or perform Integration. If the selected operation changes an accepted target, stop and route it to the Integration capability.
