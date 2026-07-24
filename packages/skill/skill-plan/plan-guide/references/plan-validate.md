# plan-validate: Validate a Plan Against Explicit Criteria

Validate one explicit inline or provider-native plan and its implementation evidence without changing the plan. Use an optional native revision to pin mutable state, and report a result for every explicit success criterion and Definition of Done requirement.

## Core Workflow

1. Resolve the explicit plan, optional native revision, criteria, related children, source references, and implementation evidence through selected providers where applicable.
2. Resolve the exact code or resource revisions being claimed. Mark stale, mismatched, missing, or unresolvable evidence as blocked or failed with rationale.
3. Run applicable typecheck, lint, build, tests, integration, policy, and non-functional checks. Record not-applicable or unavailable categories explicitly.
4. Evaluate every success criterion, task, review or documentation requirement, non-functional requirement, unresolved finding, and cumulative child condition independently.
5. Report pass, fail, or blocked per criterion with reproducible evidence, freshness, limitations, and exact revisions checked.
6. Return evidence inline. Do not mutate the plan or its status; use a separate Publish operation if the evidence must be persisted.

## Gotchas

- Green tests alone do not satisfy success criteria or the Definition of Done.
- Successful validation is not approval, publication, or authorization for later work.
- Status metadata is neither a criterion nor evidence unless the caller explicitly defines it as one.
