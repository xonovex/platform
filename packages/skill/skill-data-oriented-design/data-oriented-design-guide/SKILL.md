---
name: data-oriented-design-guide
description: "Use when designing or refactoring performance-critical data layouts for cache efficiency, in any language. Triggers on prompts about SoA/AoS/AoSoA, cache misses, hot/cold field splitting, data-oriented design, component-storage layout, batch/bulk processing, handles vs pointers, avoiding pointer chasing. Skip allocation/ownership mechanics (use memory-management-guide), ECS architecture / systems / change-tracking (use ecs-guide), language-specific style (use c99-opinionated-guide / c99-game-opinionated-guide), thread-synchronization & false-sharing-for-concurrency (use lock-free-guide), and pure algorithmic-complexity questions."
---

# Data-Oriented Design Guidelines

## Essentials

- **The problem is data movement** - Optimize cache traffic, not instruction count, see [references/cache-behavior.md](references/cache-behavior.md)
- **Design from the data** - Model bulk input→output transforms, not idealized objects, see [references/data-as-transforms.md](references/data-as-transforms.md)
- **Measure first** - Profile cache misses before changing layout, see [references/measurement-and-profiling.md](references/measurement-and-profiling.md)
- **Record cheaply, always on** - Per-frame accumulating counters recorded via a cached accumulator pointer (one store, no lookup/branch/lock), history allocated only for viewed counters, see [references/statistics-recording.md](references/statistics-recording.md)

## Layout

- **SoA / AoS / AoSoA** - Split structs into per-field arrays for hot loops, see [references/soa-aos-aosoa.md](references/soa-aos-aosoa.md)
- **Hot/cold splitting** - Keep frequently-touched fields out of cold bytes, see [references/hot-cold-splitting.md](references/hot-cold-splitting.md)
- **SIMD-friendly layout** - Contiguous columns, padding, alignment for vectorization, see [references/simd-friendly-layout.md](references/simd-friendly-layout.md)
- **Nested / variable-length arrays** - Hold child lists in bulk arrays + cache-line chunks, not per-object heap pointers, see [references/nested-arrays.md](references/nested-arrays.md)

## Processing

- **Sequential access** - Stream linearly, never chase pointers, see [references/access-patterns.md](references/access-patterns.md)
- **Existence-based processing** - Bucket by state so branches become loops, see [references/existence-based-processing.md](references/existence-based-processing.md)
- **Batch over one-at-a-time** - Transform N items per call, amortize overhead, see [references/data-as-transforms.md](references/data-as-transforms.md)
- **Default to zero** - Make all-zero a valid default; reserve 0 for none/neutral, not magic sentinels, see [references/zero-as-default.md](references/zero-as-default.md)

## Memory

- **Handles, not pointers** - Reference by index/generational handle for relocatable, stable arrays, see [references/handles-and-indices.md](references/handles-and-indices.md)
- **Allocation** - Contiguous storage matters for cache, but the allocators themselves (arenas/pools/lifetimes) are general — see **memory-management-guide**

## Gotchas

- SoA only wins when loops touch a subset of fields; full-record access can favor AoS — measure both.
- A generational handle prevents use-after-free dangling on swap-remove, but only if you actually compare the generation.
- The hardware prefetcher tracks linear strides; randomizing your index order silently disables it.
- Padding for alignment trades memory for throughput — on cache-bound loads, the smaller packed layout can still win.
- A counter you cache as a raw pointer dangles if its backing array reallocates — hand out indices into a stable block, or pointers into a non-relocating pool.

## Progressive Disclosure

- Read [references/cache-behavior.md](references/cache-behavior.md) - Load when reasoning about cache lines, latency, or miss types
- Read [references/data-as-transforms.md](references/data-as-transforms.md) - Load when modeling a system as bulk streams instead of objects
- Read [references/soa-aos-aosoa.md](references/soa-aos-aosoa.md) - Load when choosing or converting between AoS, SoA, and AoSoA layouts
- Read [references/hot-cold-splitting.md](references/hot-cold-splitting.md) - Load when a fat struct drags rarely-used fields through hot loops
- Read [references/access-patterns.md](references/access-patterns.md) - Load when iterating collections or replacing pointer-chasing structures
- Read [references/handles-and-indices.md](references/handles-and-indices.md) - Load when designing entity references, stable arrays, or free lists
- Read [references/existence-based-processing.md](references/existence-based-processing.md) - Load when removing per-item branches by sorting or bucketing
- Read [references/simd-friendly-layout.md](references/simd-friendly-layout.md) - Load when laying out data for vectorization or alignment
- Read [references/nested-arrays.md](references/nested-arrays.md) - Load when objects own variable-length child lists, or you're reaching for a per-object growable container
- Read [references/zero-as-default.md](references/zero-as-default.md) - Load when choosing sentinels, nil values, or default-initialized state
- Read [references/measurement-and-profiling.md](references/measurement-and-profiling.md) - Load when measuring before/after or reading hardware counters
- Read [references/statistics-recording.md](references/statistics-recording.md) - Load when building always-on counters/telemetry that must add near-zero overhead on hot paths
