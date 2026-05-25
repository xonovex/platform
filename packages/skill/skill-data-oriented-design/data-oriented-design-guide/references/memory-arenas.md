# memory-arenas: Arenas, Pools, and Contiguous Storage

**Guideline:** Allocate from arena/bump and pool allocators over contiguous blocks, and reclaim by lifetime reset, instead of per-object `malloc`/`free`.

**Rationale:** Per-object general-purpose allocation is slow (lock + free-list search), fragments the heap, scatters objects across random addresses (defeating the prefetcher), and adds per-allocation header overhead. Arenas allocate by bumping a pointer (a few instructions, no search) and place objects contiguously, which is exactly the layout streaming loops want. Reclaiming a whole lifetime with one reset is O(1) and leak-proof by construction. This ties directly to caller-owns-memory: the caller hands a system a block, the system never touches the global allocator.

**Techniques:**

- **Arena / bump / linear allocator** - One big block plus an offset; allocate by advancing the offset (with alignment); "free" the whole arena by setting offset back to zero. No individual frees.
- **Object pool** - Fixed-size slots for one type; a free list threads through dead slots. O(1) allocate/free, stable storage, no fragmentation. Pairs with generational handles.
- **Lifetime by reset** - Group allocations by lifetime (per-frame, per-level, per-request). Reset the arena at the lifetime boundary; everything in it dies at once.
- **Scratch / temp arena** - A per-frame arena for transient buffers; reset every frame, so temporaries never hit the global allocator.
- **Contiguous block storage** - Because everything is in one block, iteration is sequential and the data is trivially relocatable/serializable.
- **Virtual-memory reserve/commit** - Reserve a huge virtual address range once (cheap — no physical pages), then commit pages lazily as the arena grows. The block never moves and never reallocs, so pointers and back-pointers into it stay valid as it grows — a growable array with stable addresses. Decommit/reset returns the pages.

**How to Apply:**

1. Identify allocations that share a lifetime; route them through one arena.
2. Implement `arena_alloc(arena, size, align)` as an aligned pointer-bump; assert on overflow.
3. At the lifetime boundary call `arena_reset(arena)` (set used = 0) — no per-object free.
4. For long-lived same-type objects with individual lifetimes, use a pool with a free list and handles.

**Example:**

```c
typedef struct { uint8_t *base; size_t used, cap; } arena_t;

static void *arena_alloc(arena_t *a, size_t size, size_t align) {
  size_t off = (a->used + (align - 1)) & ~(align - 1); // align up
  if (off + size > a->cap) return NULL;                // caller checks
  a->used = off + size;
  return a->base + off;
}
static void arena_reset(arena_t *a) { a->used = 0; }    // free everything, O(1)

// Bad: per-object malloc — scattered addresses, fragmentation, per-call cost.
for (size_t i = 0; i < n; i++) node[i] = malloc(sizeof *node[i]);

// Good: one contiguous block, bump-allocated, freed in one reset.
node_t *nodes = arena_alloc(&frame, n * sizeof *nodes, _Alignof(node_t));
// ... use nodes[0..n) contiguously ...
arena_reset(&frame); // end of frame

// Virtual-memory arena: reserve a large range, commit pages on demand.
// The base never moves, so pointers into `items` survive growth (no realloc copy).
//   POSIX:   base = mmap(NULL, RESERVE, PROT_NONE, MAP_PRIVATE|MAP_ANON, -1, 0);
//            mprotect(base, commit_bytes, PROT_READ|PROT_WRITE); // commit as you grow
//   Windows: base = VirtualAlloc(NULL, RESERVE, MEM_RESERVE, PAGE_NOACCESS);
//            VirtualAlloc(base, commit_bytes, MEM_COMMIT, PAGE_READWRITE);
```

**Gotchas:**

- Arenas cannot free individual objects; an object outliving its arena's reset is a use-after-free — match lifetimes carefully.
- Bump allocation has no bounds safety unless you check capacity; always assert/return on overflow.
- Pointers into an arena become invalid after reset (and after a growing arena reallocates); prefer offsets/handles for anything stored across the boundary. (Exception: a virtual-memory reserve/commit arena never moves, so pointers survive growth — only reset invalidates them.)
- Reserve is free but not infinite (address space, and OS overcommit/page-table limits); pick a sane reservation ceiling and still assert on commit failure.

**Related:** [references/handles-and-indices.md](./handles-and-indices.md), [references/access-patterns.md](./access-patterns.md), [references/cache-behavior.md](./cache-behavior.md), [references/simd-friendly-layout.md](./simd-friendly-layout.md)
