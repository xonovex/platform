# develop-run: Execute Development Assignments

## Core workflow

1. Resolve each exact Planning reference and revision, then load every required implementation skill before mutation.
2. Apply [development-contracts.md](development-contracts.md) to validate dependencies, execution groups, workspace isolation, profile requirements, executor permission, budgets, and failure behavior.
3. Baseline the actual toolchain in every target workspace. Record absent targets as unavailable rather than passing.
4. Select the least adaptive executor per assignment. Use deterministic tools for mechanical changes, a bounded model only for a narrow validated transform, and a bounded agent only for adaptive multi-step implementation.
5. Execute ready assignments up to the declared concurrency limit. Preserve isolation, cancellation, partial results, and provider-native evidence; never assume sibling ordering or rollback.
6. Run the assignment's typecheck, lint, build, test, and domain validation requirements. Fix regressions attributable to the assignment; report unrelated baseline failures separately.
7. Publish one Development result per assignment with exact start/result revisions, changes, validation, unresolved findings, executor origin/bounds, failures, and follow-up capabilities.

## Completion

An assignment completes only when its Planning success criteria and required validation pass. A completed assignment does not imply its execution group, parent Planning result, consolidated workspace, Deliverable Publication, or Acceptance is complete.
