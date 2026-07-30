# hot-reloading-content: Watching, Recompiling, and Swapping Live Runtime Data

## Guideline

Watch source files for changes, recompile only the affected dependency closure in the background, then publish the new runtime resource and atomically repoint references: retiring the old version only once no in-flight work still reads it, so content updates while the running application stays live.

## Rationale

The hard part of hot-reload is the swap: runtime data may be referenced by code already mid-frame, so freeing the old resource the instant the new one is ready can crash work submitted against the old handle. The safe pattern is publish-then-retire: build the new version fully off to the side (never mutate the live resource in place), repoint references atomically (handle swap or generation bump), and defer freeing the old version until in-flight work referencing it completes, gated by the same fence/frame-in-flight boundary the renderer already tracks.

## How to Apply

1. Run a file watcher over the source tree; on a change event, map the changed file to the assets that depend on it via the dependency graph.
2. Queue the affected closure onto the same async compile path used offline; content hashing means unchanged outputs are no-ops and only genuinely stale data recompiles.
3. Build the new runtime resource fully off to the side before exposing it, never mutate the live resource in place.
4. Publish atomically: swap the handle (or bump a generation/version) so new readers pick up the new resource on their next access, while readers already holding the old pointer finish safely.
5. Retire the old version behind the in-flight boundary: only free it once the last frame/command that could reference it has completed (e.g. past the renderer's frames-in-flight fence), not at the moment of swap.
6. Coalesce rapid successive saves (debounce the watcher) so a burst of edits triggers one recompile-and-swap rather than a storm.

## Example

```c
// New data is built off to the side, then published atomically; old data retired late.
typedef struct live_resource_t {
    _Atomic(runtime_texture_t *) current; // readers load this; swap is a single store
} live_resource_t;

void on_texture_saved(pipeline_t *p, asset_id_t src) {
    id_array_t closure = reverse_closure(&p->graph, src);
    for (size_t i = 0; i < closure.len; ++i) {
        queue_async(p->compile_pool, closure.ids[i], on_cooked); // background, non-blocking
    }
}

void on_cooked(pipeline_t *p, asset_id_t id, runtime_texture_t *fresh) {
    live_resource_t *lr = live_lookup(p, id);
    runtime_texture_t *old = atomic_exchange(&lr->current, fresh); // publish atomically
    retire_after_in_flight(p, old); // free only past the frames-in-flight fence
}
```

## Gotchas

- Freeing the old resource on swap can crash a frame already submitted with the old handle; defer the free until past the in-flight fence the renderer tracks.
- Mutating the live resource in place exposes half-written data to a reader mid-frame; always build the new version separately and swap a reference.
- A non-debounced watcher fires several events per save (editors write temp-then-rename), causing redundant recompiles and visible churn; coalesce events before queuing.
- Blocking the frame loop on the recompile defeats the purpose; the cook must run on a background pool, and the swap must be a cheap atomic publish.
- Hot-reload that bypasses the cache/dependency path can recook the world on every save; route live edits through the same closure-and-hash machinery as offline builds.
- Reload safety also depends on the consuming code tolerating a swapped pointer between accesses: the discipline of writing such reloadable code lives in c99-opinionated-guide.

## Related

[references/dependency-tracking.md](./dependency-tracking.md), [references/content-hash-and-cache.md](./content-hash-and-cache.md), **c99-opinionated-guide**, **data-oriented-design-guide**
