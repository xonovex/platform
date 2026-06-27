# arenas-and-pools: Arenas, Pools, and Bulk Allocators

## Guideline

Allocate from arena/bump and pool allocators over contiguous blocks and reclaim by lifetime reset, instead of per-object general-purpose `malloc`/`free`.

## Rationale

Per-object general-purpose allocation is slow (lock + free-list search), fragments the heap, scatters objects across random addresses, and adds a per-allocation header. An arena allocates by bumping a pointer (a few instructions, no search) and places objects contiguously; reclaiming a whole lifetime is one O(1) reset and leak-proof by construction. Pools give O(1) allocate/free of same-size objects with stable addresses and no fragmentation. (Contiguity also helps cache behavior — see data-oriented-design-guide — but the allocation strategy is the concern here.)

## Techniques

- **Arena / bump / linear** - One block + an offset; allocate by advancing the offset (aligned); "free" by resetting the offset to zero. No individual frees.
- **Object pool** - Fixed-size slots for one type; a free list threads through dead slots. O(1) alloc/free, stable storage, no fragmentation.
- **Lifetime by reset** - Group allocations by lifetime (per-frame, per-level, per-request); reset the arena at the boundary so everything dies at once.
- **Scratch / temp arena** - A per-cycle arena for transient buffers, reset every cycle, so temporaries never hit the global allocator.
- **Virtual-memory reserve/commit** - Reserve a huge virtual range once (cheap — no physical pages), commit pages lazily as the arena grows. The block never moves and never reallocs, so pointers/back-pointers into it stay valid as it grows — a growable array with stable addresses. Reset/decommit returns the pages.

## How to Apply

1. Identify allocations that share a lifetime; route them through one arena.
2. Implement `arena_alloc(a, size, align)` as an aligned pointer-bump; assert/return on overflow.
3. Reset at the lifetime boundary (`used = 0`) — no per-object free.
4. For long-lived same-type objects with individual lifetimes, use a pool with a free list.

## Example

```c
typedef struct { uint8_t *base; size_t used, cap; } arena_t;

static void *arena_alloc(arena_t *a, size_t size, size_t align) {
  size_t off = (a->used + (align - 1)) & ~(align - 1); // align up
  if (off + size > a->cap) return NULL;                // caller checks
  a->used = off + size;
  return a->base + off;
}
static void arena_reset(arena_t *a) { a->used = 0; }   // free everything, O(1)

// Virtual-memory arena: reserve large, commit on demand; base never moves.
//   POSIX:   base = mmap(NULL, RESERVE, PROT_NONE, MAP_PRIVATE|MAP_ANON, -1, 0);
//            mprotect(base, commit_bytes, PROT_READ|PROT_WRITE);   // commit as you grow
//   Windows: base = VirtualAlloc(NULL, RESERVE, MEM_RESERVE, PAGE_NOACCESS);
//            VirtualAlloc(base, commit_bytes, MEM_COMMIT, PAGE_READWRITE);
```

## Gotchas

- Arenas cannot free individual objects; an object outliving its arena's reset is a use-after-free — match lifetimes.
- Bump allocation has no bounds safety unless you check capacity; always assert/return on overflow.
- Pointers into a growing/reset arena become invalid — prefer offsets/handles across a boundary. (Exception: a reserve/commit arena never moves, so pointers survive growth; only reset invalidates them.)
- Reserve is cheap but not infinite (address space, overcommit/page-table limits); pick a sane ceiling and still assert on commit failure.

## Related

[references/caller-owns-memory.md](./caller-owns-memory.md), [references/ownership-and-lifetimes.md](./ownership-and-lifetimes.md)
