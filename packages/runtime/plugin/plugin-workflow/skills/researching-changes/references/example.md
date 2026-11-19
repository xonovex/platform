# Research Report Example

```text
Research: order-import batching (for a future plan)

Stack: TypeScript 5.6, fastify 5; validation tasks `app:test`, `app:ci-check`
Locations: src/import/order_reader.ts (streaming reader, the seam),
  src/queue/batcher.ts (fixed 100-row batches)
Observed: the reader yields rows one at a time; the batcher drops rows
  on queue-full (src/queue/batcher.ts:41) with no retry
Options: 1) backpressure via async-iterator pull (recommended: no new
  dependency, matches the reader's shape) 2) bounded queue plus retry
Risks: the feed spikes 50x on promotion days; no staging fixture exists
Unresolved: whether duplicate rows are legal upstream
Skills to consult: writing-typescript, authoring-tests
Sources: src/queue/batcher.ts@a1b2c3; vendor feed spec v2 (2026-01)
```

Facts carry file anchors; options carry a recommendation; anything unverified stays under Unresolved rather than becoming a claim.
