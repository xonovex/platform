# false-sharing: False Sharing

## Guideline

Give each hot per-thread or independently-written atomic field its own cache line; pad and align to `CACHE_LINE_SIZE`.

## Rationale

Coherence works at cache-line granularity, not per-variable. When two threads write _distinct_ variables that happen to share one line, the line ping-pongs between cores: each write invalidates the other core's copy even though there is no logical sharing. The result is silent, correctness-irrelevant serialization that can erase all expected scaling — a per-thread counter array can run slower than a single global one.

## How to Apply

1. Identify fields written by different threads: per-thread counters, the head and tail of a queue, adjacent atomics in a struct.
2. Pad/align each to a full cache line. Define a `CACHE_LINE_SIZE` constant — or detect it per target (64 B on x86-64; 128 B on Apple silicon and some ARM/POWER, where the coherence/prefetch unit is two 64 B lines). Pad to the larger value when targeting both.
3. Verify with a profiler (perf c2c, VTune) that the line is no longer shared; don't pad blindly — padding costs cache footprint.

## Example

```c
#include <stdatomic.h>
#include <stddef.h>

#ifndef CACHE_LINE_SIZE
#define CACHE_LINE_SIZE 64
#endif

// Each worker's counter sits alone on its own line: no cross-core invalidation.
typedef struct {
    _Alignas(CACHE_LINE_SIZE) atomic_uint_fast64_t value;
    char pad[CACHE_LINE_SIZE - sizeof(atomic_uint_fast64_t)];
} padded_counter_t;

padded_counter_t worker_hits[MAX_WORKERS];  // index by thread id

// Same idea inside one struct: keep producer-written head off the line
// that holds the consumer-written tail (see the SPSC ring buffer).
typedef struct {
    _Alignas(CACHE_LINE_SIZE) atomic_size_t head;
    _Alignas(CACHE_LINE_SIZE) atomic_size_t tail;
} ring_indices_t;
```

## Gotchas

- `_Alignas` on a struct field aligns the field, but an _array_ of such structs only stays line-aligned if the struct's size is a whole multiple of the line — the explicit `pad` guarantees that (and aligns the whole struct via the `_Alignas` member).
- `malloc` does not guarantee cache-line alignment; for line-aligned dynamic buffers use `aligned_alloc(CACHE_LINE_SIZE, n)`.
- Many runtimes already line-align their per-worker queues and per-CPU counters; follow the existing padding pattern in your codebase rather than inventing a new one.

## Related

[references/spsc-ring-buffer.md](./spsc-ring-buffer.md), [references/scalable-architecture.md](./scalable-architecture.md)
