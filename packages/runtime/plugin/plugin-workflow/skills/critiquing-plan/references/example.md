# Critique Example

```text
Critique: order-import-backpressure (revision 2026-01-10, fresh context)

First pass, blind:
1. HIGH / likely / Current State: the plan assumes the reader is the
   only producer, but src/import/replay.ts also writes the queue, so
   backpressure on one producer protects nothing. Failure mode: replay
   floods during a live import. Suggested direction: name both
   producers in scope or exclude replay explicitly.
2. MEDIUM / possible / Success Criteria: "no dropped rows" names no
   counter and no fixture that would prove it.

Second pass, with supplied decisions: finding 1 confirmed (decision 3
covers only the reader); finding 2 unchanged; added by context:
3. LOW / confirmed / Validation: the cited task is `app:ci-check`, but
   the integration tier the plan relies on runs under
   `app:test-integration`.
```

Every finding carries severity, confidence, section, failure mode and a direction; the two passes stay distinguishable in the report.
