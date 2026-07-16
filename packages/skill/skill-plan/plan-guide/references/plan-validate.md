# plan-validate: Validate a Planning Result

Read-only validation confirms whether one Planning result's success criteria and Definition of Done are met at exact revisions. It reports evidence and gaps; it does not revise status or implementation.

## Core workflow

1. Resolve the provider context, opaque Planning reference, exact revision, child results, source results, profile, and evidence references.
2. Resolve the exact subject/workspace revisions being claimed. Reject stale, mismatched, missing, or unresolvable evidence.
3. Run applicable typecheck, lint, build, tests, integration, policy, and profile-specific checks through selected adapters. Record not-applicable/unavailable categories with rationale.
4. Check every success criterion, task, review/documentation requirement, non-functional requirement, unresolved finding, and cumulative completion rule independently.
5. Report PASS/FAIL/WARN per criterion with native evidence references, limitations, and the exact revisions checked.
6. Publish validation evidence if a provider was selected for the report, but do not mutate the Planning result or claim Acceptance.

Apply [early-lifecycle-contracts.md](early-lifecycle-contracts.md). Green tests alone do not satisfy the Definition of Done, and a runtime trace is not persistent identity.
