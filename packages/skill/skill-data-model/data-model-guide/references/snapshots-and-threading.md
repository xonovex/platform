# snapshots-and-threading: Consistent Reads While Writers Mutate

## Guideline

Give reader threads a consistent view of the model while a writer mutates it by publishing immutable snapshots (copy-on-write versions swapped in atomically) or, where that is too costly, a coarse lock; let a simulation/render thread take one read snapshot per frame, and reconcile any external mutations back into the model through the normal edit path.

## Rationale

An editor's model is read by threads that must not block the editing thread and must never observe a half-applied edit (a tree mid-relink, a buffer mid-resize). If readers and the writer share mutable state, every read needs a lock and every edit risks tearing. Publishing an immutable snapshot decouples them: readers hold a version that never changes under them, the writer builds the next version privately and swaps it in with one atomic store. Copy-on-write keeps this cheap by sharing the unchanged majority of the model between versions and copying only the touched objects.

## Techniques

- **Immutable snapshot** - A snapshot is a read-only view of the model at one version. Readers hold it for as long as they need; it is never mutated. Multiple readers share one snapshot safely with no locks.
- **Copy-on-write versions** - To produce the next version, copy only the objects being modified, leave the rest shared, then atomically publish the new root. Unchanged sub-trees are referenced, not copied, so a small edit to a large model is cheap. Old versions stay valid until their last reader releases them, then are reclaimed.
- **Atomic publish** - Swap the "current version" pointer with a single atomic store (release on the writer, acquire on readers). A reader either sees the whole old version or the whole new one — never a mix.
- **Per-frame read snapshot** - A simulation/render thread acquires the current snapshot once at frame start and reads from it for the entire frame, so the world cannot shift under it mid-frame. The editing thread keeps publishing new versions independently.
- **Coarse locking (the simple alternative)** - When snapshots are too expensive (huge mutable buffers, rare reads), a single reader-writer lock around the model is correct and far simpler. Prefer it until contention is measured; reach for COW when read concurrency actually matters.
- **Reconciling external mutations** - When a reader thread (or an external process) wants to feed changes back — physics moved an object, a tool edited a file on disk — apply them through the normal transactional edit path on the writer thread, so they go through notification, undo, and the next published version. Never let a reader mutate a snapshot in place.

## How to Apply

1. Keep a `current_version` pointer the writer publishes and readers acquire atomically.
2. To edit: copy the objects you will touch into a new version (sharing the rest), apply the transaction, then atomically store the new version pointer.
3. Readers acquire the current version once per logical read (e.g. per frame) and release it when done; reclaim a version after its last reader releases it.
4. Funnel any feedback from reader threads or external sources through the writer's normal edit/transaction path — never mutate a published snapshot.

## Example

```c
typedef struct version_s version_t;  // immutable once published
static _Atomic(version_t *) g_current;

// Reader: one consistent snapshot for the whole frame, lock-free.
static void render_frame(void) {
  version_t *snap = atomic_load_explicit(&g_current, memory_order_acquire);
  draw_world(snap);                  // snap never changes under us
  // (snap released when the frame's references drop; older versions reclaimed)
}

// Writer: build the next version privately, publish atomically.
static void apply_edit(model_t *m, txn_t *t) {
  version_t *next = version_clone_cow(m->current);  // copy only touched objects
  txn_apply(next, t);
  atomic_store_explicit(&g_current, next, memory_order_release); // single swap
}

// Bad: a reader mutates the shared model directly — tears edits, races the
// writer, and bypasses notification/undo. Feed changes back via apply_edit.
draw_world_and_move_objects(m); // reader writing shared state -> data race
```

## Gotchas

- A snapshot must be truly immutable; if any code mutates it in place, readers race and the COW sharing corrupts other versions.
- Old versions cannot be freed until their last reader releases them — track references (or use a reclamation scheme), or a slow reader pins unbounded memory.
- Copy-on-write only saves work if edits touch a small fraction; an edit that rewrites everything copies everything, so COW is no cheaper than a full copy there.
- Reader threads reconciling changes by mutating a snapshot bypass undo/notification and corrupt state — route them through the writer's transactional path.
- The atomic publish needs release/acquire ordering; a plain non-atomic pointer swap lets a reader see a torn or stale version on weakly-ordered hardware.

## Related

[references/change-notification.md](./change-notification.md), [references/undo-redo.md](./undo-redo.md), [references/serialization.md](./serialization.md); **lock-free-guide** for the snapshot-publish / atomic-version-swap and safe-reclamation mechanics
