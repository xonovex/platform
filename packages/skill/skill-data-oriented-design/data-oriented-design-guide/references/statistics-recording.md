# statistics-recording: Minimal-Overhead Statistics Recording

## Guideline

For always-on in-app counters (draw calls, allocations, subsystem timings), accumulate one value per counter per frame and let recording be a single `*ptr += value` through a cached accumulator pointer — so a hot path pays one add, the API surface stays tiny, and only counters someone is actually viewing cost any history memory.

## How to Apply

1. Register a counter by name once and return a stable pointer to its accumulator; cache that pointer in a `static` at the call site.
2. On the hot path, record with `*counter += n;` — no lookup, no function call, no branch, no lock.
3. At the frame boundary, roll each counter's accumulator into its history ring and reset it: `history[frame % MAX_FRAMES] = frame_value; frame_value = 0;`.
4. Allocate the history ring lazily — only when a counter is first shown in the UI — so thousands of registered-but-unviewed counters cost almost nothing.

## Example

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

## Gotchas

- The cached-pointer trick relies on the accumulator never moving — if the source array can reallocate, hand out indices into a stable block or pointers into a non-relocating pool, not raw pointers into a growable array.
- Per-frame accumulation throws away intra-frame ordering; for "what happened in what order" you need event/scope tracing, not a counter.
- `*ptr += n` from multiple threads races — give each worker thread its own accumulator and sum at the frame boundary, or accept that only the owning thread records into a given counter.
- Lazy history means a counter shows no graph until first viewed; that's intended, but don't read history before it's allocated.

## Related

[references/measurement-and-profiling.md](./measurement-and-profiling.md), [references/handles-and-indices.md](./handles-and-indices.md), [references/access-patterns.md](./access-patterns.md)
