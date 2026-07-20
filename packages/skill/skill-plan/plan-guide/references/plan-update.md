# plan-update: Update Plan Progress and Evidence

Refresh one explicit inline or provider-native plan from current implementation evidence without changing its planned scope. Return a traceable revised result inline or through an explicitly selected provider.

## Core Workflow

1. Resolve the explicit plan, optional native revision, related children, implementation evidence, and validation evidence.
2. Compare every task and success criterion with the exact code or resource revision and supporting evidence. Do not infer completion from conversation memory or one green check.
3. Record completed, pending, and blocked work plus validation results, limitations, and unavailable categories.
4. Reconcile child progress and cumulative completion without erasing independently stored results.
5. Treat status as optional descriptive metadata derived from evidence. It neither authorizes nor gates another operation.
6. Return the updated plan and remaining work inline, or ask the explicitly selected provider to persist a new native revision.

## Gotchas

- Updating evidence is not permission to change planned scope.
- Do not overwrite provider history when it supports native revisions.
- A descriptive completion value is only as trustworthy as its task and criterion evidence.
