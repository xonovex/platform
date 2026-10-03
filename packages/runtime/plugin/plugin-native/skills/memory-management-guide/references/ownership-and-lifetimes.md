# ownership-and-lifetimes: Ownership and Lifetimes

## Guideline

Every allocation has exactly one owner responsible for freeing it, and a clearly-bounded lifetime; everyone else borrows. Decide both explicitly, never leave ownership implicit.

## Rationale

Most memory bugs are ownership bugs: leak (nobody freed), double-free (two owners freed), use-after-free (freed while still borrowed). With no runtime to arbitrate, the discipline lives in the design: name the owner, tie the lifetime to a scope or phase, make borrows visibly shorter than the owner. Grouping by lifetime turns N free decisions into one arena/pool reset.

## How to Apply

1. For each allocation, answer two questions up front: _who frees it_ and _until when does it live_. Encode the answers in the API (owning vs borrowing parameters/returns).
2. Prefer **single ownership**: one owner, others hold borrowed pointers/handles valid only within a known scope.
3. **Tie lifetime to a phase** (frame, level, request, connection) and reclaim the whole phase at once via an arena/pool reset, rather than freeing objects individually.
4. When a borrow must outlive the owner's storage moving or being freed, reference by **index/handle**, not raw pointer (handle/relocatable-storage layout lives in data-oriented-design-guide).
5. Reference-count only when ownership is genuinely shared and lifetimes can't be made scope-bound; treat it as the exception, not the default.

## Example

```c
// Ownership in the signature: the pool OWNS slots; callers get a borrowing handle.
handle_t pool_alloc(pool_t *p);                 // pool owns the storage
entity_t *pool_borrow(pool_t *p, handle_t h);   // borrowed; valid only until the next reset/free
void      pool_free(pool_t *p, handle_t h);     // the one owner reclaims

// Lifetime by phase: everything for the frame dies together, no per-object frees.
arena_reset(&frame_arena);   // end of frame
```

## Gotchas

- A borrowed pointer cached across an allocation, reset, or free dangles: re-fetch it, or hold a handle.
- Reference counting has its own hazards (cycles leak; the inc/dec is a cost and, when shared across threads, a race, see lock-free-guide); don't reach for it to paper over unclear ownership.

## Related

[references/arenas-and-pools.md](./arenas-and-pools.md), [references/caller-owns-memory.md](./caller-owns-memory.md)
