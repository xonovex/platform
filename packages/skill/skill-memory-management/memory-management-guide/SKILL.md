---
name: memory-management-guide
description: "Use when deciding how memory is allocated, owned, and freed in manual-memory or buffer-passing code, in any language. Triggers on prompts about caller-owns-memory / caller-provided buffers, arena/bump/linear allocators, object pools and free lists, virtual-memory reserve/commit, scratch/temp allocators, object lifetimes, ownership (single-owner/borrow), leaks/use-after-free/double-free, even when the user doesn't say 'memory' or 'allocator'. Skip cache/layout optimization (use data-oriented-design-guide), lock-free reclamation of shared nodes (use lock-free-guide), GPU device-memory specifics (use gpu-rendering-vulkan-guide), and garbage-collected languages where the runtime owns lifetimes."
---

# Memory Management Guidelines

General principles for allocating, owning, and freeing memory in manual-memory or buffer-passing code. Language-agnostic; examples are in C. Domain skills (e.g. C99, GPU device memory) keep their own specifics and link here for the underlying principle.

## Essentials

- **Caller owns memory** - APIs take caller-provided storage; never allocate internally, see [references/caller-owns-memory.md](references/caller-owns-memory.md)
- **Group by lifetime** - Allocate together what dies together; free as a unit, see [references/ownership-and-lifetimes.md](references/ownership-and-lifetimes.md)
- **One owner** - Exactly one owner frees; everyone else borrows, see [references/ownership-and-lifetimes.md](references/ownership-and-lifetimes.md)

## Allocation strategy

- **Arena / bump** - Pointer-bump from a block; reset frees everything at once, see [references/arenas-and-pools.md](references/arenas-and-pools.md)
- **Object pool** - Fixed-size slots + free list; O(1) alloc/free, stable, see [references/arenas-and-pools.md](references/arenas-and-pools.md)
- **Scratch / temp** - Per-frame/per-request arena reset each cycle, see [references/arenas-and-pools.md](references/arenas-and-pools.md)
- **Virtual-memory reserve/commit** - Reserve a large range, commit pages on demand; stable addresses, no realloc copy, see [references/arenas-and-pools.md](references/arenas-and-pools.md)
- **Virtual-memory tricks** - Cap-free arrays, page-aligned growth, gapless ring buffers, end-of-page bounds checks, see [references/virtual-memory.md](references/virtual-memory.md)

## Ownership

- **Lifetime by scope** - Tie allocation lifetime to a clear scope/phase, see [references/ownership-and-lifetimes.md](references/ownership-and-lifetimes.md)
- **Handles over pointers** - Reference relocatable storage by index/handle (layout detail in data-oriented-design-guide), see [references/ownership-and-lifetimes.md](references/ownership-and-lifetimes.md)

## Gotchas

- An arena cannot free individual objects — anything outliving its reset is a use-after-free; match lifetimes.
- Bump allocation has no bounds safety unless you check capacity; always assert/return on overflow.
- Two owners means double-free or leak; decide ownership explicitly, don't infer it.
- Pointers into a growing/reset allocation dangle — prefer offsets/handles across a boundary (a virtual-memory reserve/commit block is the exception: it never moves).

## Progressive Disclosure

- Read [references/caller-owns-memory.md](references/caller-owns-memory.md) - Load when designing an API's allocation boundary or a non-allocating library
- Read [references/arenas-and-pools.md](references/arenas-and-pools.md) - Load when choosing an allocator: arena, pool, scratch, or virtual-memory backed
- Read [references/virtual-memory.md](references/virtual-memory.md) - Load when using address-space reservation for cap-free arrays, ring buffers, page-aligned growth, or end-of-page bounds checking
- Read [references/ownership-and-lifetimes.md](references/ownership-and-lifetimes.md) - Load when deciding who owns/frees memory and how long it lives
