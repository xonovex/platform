# Update: Plan Progress and Evidence

Refresh one explicit inline or provider-native plan from current implementation evidence without changing its planned scope. Return a traceable revised result inline.

## Core Workflow

1. Resolve the explicit plan, optional native revision, related children, canonical
   context, implementation evidence, and validation evidence.
2. Compare every task and success criterion with the exact code or resource revision and supporting evidence. Do not infer completion from conversation memory or one green check.
3. Record completed, pending, and blocked work plus validation results, limitations, and unavailable categories.
4. Reconcile child progress, carried-decision status, and cumulative completion
   without erasing independently stored results.
5. Treat status as optional descriptive metadata derived from evidence. It neither authorizes nor gates another operation.
6. Return the updated plan and remaining work inline. Use a separate Publish operation if the update must be persisted.

## Example

```text
validation:
  build: pass
  tests: pass
  integration: partial # staging soak pending, see Results

## Results

Landed 2026-01-14: the pull seam in order_reader.ts (tasks 1 to 5),
219 tests green. Criterion 4 unmet: the staging soak needs the staging
feed, owner named. Downstream: subplan 02 reads the pull API; the drop
counter is exported for its criterion 1.
```

The frontmatter states measured outcomes; the Results section records
what landed, what is unmet with its owner, and what downstream work
inherits. Scope stays untouched.

## Gotchas

- Updating evidence is not permission to change planned scope.
- Do not overwrite provider history when it supports native revisions.
- A descriptive completion value is only as trustworthy as its task and criterion evidence.
