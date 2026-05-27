# statistics-recording: Minimal-Overhead Statistics Recording

**Guideline:** For always-on in-app counters (draw calls, allocations, subsystem timings), accumulate one value per counter per frame and let recording be a single `*ptr += value` through a cached accumulator pointer — so a hot path pays one add, the API surface stays tiny, and only counters someone is actually viewing cost any history memory.

**Rationale:** Telemetry that you want everywhere must cost almost nothing where it's recorded, or you won't leave it on — and a system you turn off doesn't catch the spike it was meant to catch. Per-event recording (timestamp every occurrence) is both higher overhead and a different data shape per source; accumulating into a per-frame counter instead unifies every source to one format, uses the frame as a free clock, and collapses N events into one number. The expensive part of recording is the `hash(name) → table lookup` on every call; hand the caller a direct pointer to the accumulator once and the steady-state cost drops to a single store — no call, no branch, no lock. Keeping the per-counter history (a ring of recent frames) allocated _only_ for counters currently displayed means thousands of registered counters cost almost nothing until inspected.

**How to Apply:**

1. Model a counter as `{ name, double *history /*ring, lazy*/, double frame_value }`; keep a name→index map and a `current_frame`.
2. Register a counter by name once and return a stable pointer to its `frame_value`; cache that pointer in a `static` at the call site.
3. On the hot path, record with `*counter += n;` — no lookup, no function call. Guard history writes with `if (history)` so unviewed counters skip work.
4. At the frame boundary, for each counter with allocated history: `history[frame % MAX_FRAMES] = frame_value; frame_value = 0;`.
5. Allocate the history ring lazily — only when a counter is first shown in the UI — so registration stays cheap for huge counter counts.
6. Read back as a graph (spikes, unbounded growth = leaks, anomalies) or a table (exact numbers); reuse the same counters for profiler-scope timings to find hitches.

**Example:**

```c
// Hot path: one store, no lookup/branch/lock after the first frame.
void submit_primitives(uint32_t n) {
    static double *primitive_count;                 // cached accumulator pointer
    if (!primitive_count)
        primitive_count = stat_counter("renderer/primitive-count");
    *primitive_count += n;                          // the entire recording cost
}

// Frame boundary: roll accumulators into the per-counter ring, then reset.
void stats_end_frame(stats_t *s) {
    for (uint32_t i = 0; i < array_count(s->sources); ++i) {
        source_t *src = &s->sources[i];
        if (src->history)                           // only viewed counters keep history
            src->history[s->current_frame % MAX_FRAMES] = src->frame_value;
        src->frame_value = 0.0;
    }
    s->current_frame++;
}
```

**Gotchas:**

- The cached-pointer trick relies on the accumulator never moving — if the source array can reallocate, hand out indices into a stable block or pointers into a non-relocating pool, not raw pointers into a growable array.
- Per-frame accumulation throws away intra-frame ordering; for "what happened in what order" you need event/scope tracing, not a counter.
- `*ptr += n` from multiple threads races — give each worker thread its own accumulator and sum at the frame boundary, or accept that only the owning thread records into a given counter.
- Lazy history means a counter shows no graph until first viewed; that's intended, but don't read history before it's allocated.
- A ring of `MAX_FRAMES` silently overwrites old samples; size it for the window you actually scrub, and remember plotted "leaks" are just monotonic growth in a counter.

**Related:** [references/measurement-and-profiling.md](./measurement-and-profiling.md), [references/handles-and-indices.md](./handles-and-indices.md), [references/access-patterns.md](./access-patterns.md)
