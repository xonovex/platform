# cache-behavior: The Memory Wall and Cache Behavior

## Guideline

Treat data movement through the cache hierarchy — not arithmetic — as the dominant cost, and lay out data so the bytes you touch arrive together in cache lines.

## Rationale

CPU compute throughput has grown far faster than memory latency (the "memory wall"). A modern core retires several instructions per cycle but stalls hundreds of cycles on a main-memory miss. The first DOD premise is therefore: the problem is almost never the instructions, it is moving the right bytes to the core at the right time. A cache miss can cost more than the entire arithmetic body of a hot loop.

## Key facts

- **Cache line** - Memory moves between levels in fixed blocks, typically 64 bytes (some ARM use 128). Touching one byte loads the whole line. Useful work per line = (bytes you read) / 64.
- **Latency orders of magnitude** - Rough cycle costs on a typical desktop core:
  - L1 hit: ~4 cycles
  - L2 hit: ~12 cycles
  - L3 hit: ~40 cycles
  - Main memory (RAM): ~200+ cycles (a TLB miss adds more)
- **Working set** - The set of bytes actively reused. If it fits in L1/L2 you stay fast; exceed a level's capacity and you spill to the next, slower level.
- **Locality** - Spatial: bytes near each other are likely needed together (favors contiguous arrays). Temporal: recently used bytes are likely reused soon (favors keeping hot state small).

## Cache-miss types (the "three Cs")

- **Compulsory** - First-ever touch of a line; unavoidable, but a tight layout reduces how many lines you must touch.
- **Capacity** - Working set exceeds cache size; lines get evicted before reuse. Fix by shrinking the working set (hot/cold split, smaller records).
- **Conflict** - Multiple addresses map to the same set; evict each other despite free capacity. Fix by avoiding power-of-two strides and pathological alignment.

## How to Apply

1. Estimate bytes-touched-per-line for your hot loop; if you read 8 of 64 bytes per record, ~87% of memory traffic is wasted.
2. Shrink records and split cold fields so more useful elements fit per line and the working set drops a cache level.
3. Make access linear so the hardware prefetcher hides compulsory-miss latency.
4. Confirm with a profiler — cache-miss counts, not instruction counts.

## Example

```c
// Bad: 64-byte record, loop reads only 4 bytes of each.
// 1 useful float per cache line -> ~94% of bandwidth wasted.
typedef struct { float hp; char name[44]; void *owner; double extra; } unit_t;
unit_t units[N];
for (size_t i = 0; i < N; i++) total += units[i].hp;

// Good: pack the hot field contiguously.
// 16 useful floats per cache line, prefetcher streams ahead.
float hp[N];
for (size_t i = 0; i < N; i++) total += hp[i];
```

## Gotchas

- Reasoning about "fast code" by counting instructions is misleading; a single L3/RAM miss dwarfs dozens of adds.
- Cache sizes and line sizes vary by target; verify on the actual hardware, not assumptions.

## Related

[references/data-as-transforms.md](./data-as-transforms.md), [references/soa-aos-aosoa.md](./soa-aos-aosoa.md), [references/access-patterns.md](./access-patterns.md), [references/measurement-and-profiling.md](./measurement-and-profiling.md)
