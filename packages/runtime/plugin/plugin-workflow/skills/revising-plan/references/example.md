# Revision Example

```text
Revision: order-import-backpressure, 2026-01-10 to 2026-01-12

1. "Cover the replay producer too": applied. Scope, the group 2 note
   and criterion 1 updated; a dual-producer ordering risk added.
2. "Drop the metrics subplan": rejected. Criterion 4's counter needs
   it; kept with its scope halved instead, rationale in the plan.
3. "Ship this week": scope change, returned to the caller. The
   estimate says 8 days and no proposed cut restores the difference.

Revised plan follows; the source revision is preserved beside it.
```

Every feedback item gets a disposition with a rationale; nothing is silently dropped, and a conflict stops the revision rather than being averaged away.
