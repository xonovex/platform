# change-notification: Making Mutations Observable

## Guideline

Route every mutation through the model so it can record what changed and tell interested parties, via change events, dirty flags, and dependency-driven recomputation.

## Rationale

A central model has many readers (UI, derived caches, renderer, save indicator) that must stay consistent without each polling everything. Observable mutations let readers update exactly what changed and recompute only its dependents. Batching per transaction stops one action touching fifty properties from emitting fifty notifications (a "notification storm") and stops observers seeing a half-edited intermediate state.

## Techniques

- **Change events (push)** - On mutation, record `{ object_id, property, change_kind }`. After the transaction commits, deliver the accumulated set to registered listeners. Listeners diff against their own state and update.
- **Version counters (pull)** - Keep a monotonically increasing version on the model (and optionally per object/property). A reader stores the last version it saw and, when convenient, compares: unchanged → skip, changed → re-read. No callbacks, no re-entrancy, cheap to poll.
- **Dirty flags** - Mark touched objects/properties dirty; consumers (save, layout, render) process and clear only the dirty set instead of rescanning everything.
- **Dependency tracking + recompute** - For derived data, record which inputs a computed value read; when an input changes, mark the dependent dirty and recompute lazily on next read. This recomputes the minimal set, in dependency order.
- **Push vs pull tradeoff** - Push gives immediate, fine-grained updates but risks re-entrancy and storms; pull is simple and storm-proof but adds latency until the next poll. Many models use pull for high-frequency readers (render) and push for discrete UI.

## How to Apply

1. Make all writes go through `set(model, id, prop, value)`; forbid direct field writes from outside.
2. Inside `set`, append a change record to the current transaction and bump the version; do not notify synchronously.
3. At transaction commit, coalesce the change set and deliver it (push) and/or expose the new version (pull).
4. For derived values, record input dependencies on compute; invalidate dependents when an input's change record commits.

## Example

```c
// Pull: a reader polls a version instead of subscribing.
if (model->version != view->last_seen_version) {
  rebuild_view(view, model);
  view->last_seen_version = model->version;
}
```

## Gotchas

- Without coalescing, a loop that sets one property N times emits N events; dedupe per `(object, property)`.
- Pull-based readers can miss an intermediate value entirely (it changed and changed back between polls): fine for "latest state" views, wrong if you need every transition.
- Dependency tracking is only correct if every input read is recorded; a hidden global read that isn't tracked yields stale derived values.

## Related

[references/undo-redo.md](./undo-redo.md), [references/object-model.md](./object-model.md), [references/snapshots-and-threading.md](./snapshots-and-threading.md)
