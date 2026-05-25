---
name: data-oriented-design-guide
description: "Use when designing or refactoring performance-critical data layouts for cache efficiency, in any language. Triggers on prompts about SoA/AoS/AoSoA, cache misses, hot/cold field splitting, data-oriented design, ECS/component storage, batch/bulk processing, handles vs pointers, avoiding pointer chasing, arena/pool allocation. Skip language-specific style (use c99-opinionated-guide / c99-game-opinionated-guide), thread-synchronization & false-sharing-for-concurrency (use lock-free-guide), and pure algorithmic-complexity questions."
---

# Data-Oriented Design Guidelines

## Essentials

- **The problem is data movement** - Optimize cache traffic, not instruction count, see [references/cache-behavior.md](references/cache-behavior.md)
- **Design from the data** - Model bulk input→output transforms, not idealized objects, see [references/data-as-transforms.md](references/data-as-transforms.md)
- **Measure first** - Profile cache misses before changing layout, see [references/measurement-and-profiling.md](references/measurement-and-profiling.md)

## Layout

- **SoA / AoS / AoSoA** - Split structs into per-field arrays for hot loops, see [references/soa-aos-aosoa.md](references/soa-aos-aosoa.md)
- **Hot/cold splitting** - Keep frequently-touched fields out of cold bytes, see [references/hot-cold-splitting.md](references/hot-cold-splitting.md)
- **SIMD-friendly layout** - Contiguous columns, padding, alignment for vectorization, see [references/simd-friendly-layout.md](references/simd-friendly-layout.md)

## Processing

- **Sequential access** - Stream linearly, never chase pointers, see [references/access-patterns.md](references/access-patterns.md)
- **Existence-based processing** - Bucket by state so branches become loops, see [references/existence-based-processing.md](references/existence-based-processing.md)
- **Batch over one-at-a-time** - Transform N items per call, amortize overhead, see [references/data-as-transforms.md](references/data-as-transforms.md)

## Memory

- **Handles, not pointers** - Reference by index/generational handle for relocatable, stable arrays, see [references/handles-and-indices.md](references/handles-and-indices.md)
- **Arenas and pools** - Bump/pool allocators over per-object malloc, see [references/memory-arenas.md](references/memory-arenas.md)

## Gotchas

- SoA only wins when loops touch a subset of fields; full-record access can favor AoS — measure both.
- A generational handle prevents use-after-free dangling on swap-remove, but only if you actually compare the generation.
- Bump allocators cannot free individual objects; lifetimes must be batched and reset together or you leak.
- The hardware prefetcher tracks linear strides; randomizing your index order silently disables it.
- Padding for alignment trades memory for throughput — on cache-bound loads, the smaller packed layout can still win.

## Progressive Disclosure

- Read [references/cache-behavior.md](references/cache-behavior.md) - Load when reasoning about cache lines, latency, or miss types
- Read [references/data-as-transforms.md](references/data-as-transforms.md) - Load when modeling a system as bulk streams instead of objects
- Read [references/soa-aos-aosoa.md](references/soa-aos-aosoa.md) - Load when choosing or converting between AoS, SoA, and AoSoA layouts
- Read [references/hot-cold-splitting.md](references/hot-cold-splitting.md) - Load when a fat struct drags rarely-used fields through hot loops
- Read [references/access-patterns.md](references/access-patterns.md) - Load when iterating collections or replacing pointer-chasing structures
- Read [references/handles-and-indices.md](references/handles-and-indices.md) - Load when designing entity references, stable arrays, or free lists
- Read [references/existence-based-processing.md](references/existence-based-processing.md) - Load when removing per-item branches by sorting or bucketing
- Read [references/memory-arenas.md](references/memory-arenas.md) - Load when replacing per-object allocation with arenas or pools
- Read [references/simd-friendly-layout.md](references/simd-friendly-layout.md) - Load when laying out data for vectorization or alignment
- Read [references/measurement-and-profiling.md](references/measurement-and-profiling.md) - Load when measuring before/after or reading hardware counters
