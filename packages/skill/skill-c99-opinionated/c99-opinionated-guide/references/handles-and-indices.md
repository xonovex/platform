# Reference by index/handle, not raw pointer

## Guideline

In this style, a cross-object reference is stored as a stable index or a generational handle into a caller-owned array — never a raw `T *`. The pointer is recovered transiently, at the point of use, and never persisted back into a struct.

## Rationale

Raw pointers into a caller-owned array are fragile and non-portable; an index is neither.

- **Survives relocation.** The backing array can be re-bound, compacted, or grown into a bigger block; every stored pointer would dangle, every stored index stays valid.
- **Trivially serializable.** An index is position-independent — dump the arrays and the indices verbatim and the graph reloads identically. Pointers are addresses; they mean nothing in the next process.
- **Smaller and cache-friendly.** A `uint32_t`/`int16_t` reference packs tighter than an 8-byte pointer, so more of the graph fits in a cache line.
- **Deterministic.** Two runs build byte-identical structures; pointer values (ASLR, allocator state) differ run to run and break reproducibility and hashing.

## Example

```c
typedef struct {
    int32_t  parent;        /* index into the node pool, -1 for root — not node_t* */
    int32_t  first_child;
    uint16_t child_count;
} node_t;

/* Recover the pointer transiently, at use; never store it back into a struct. */
node_t *n = &pool->nodes[idx];
```

## Generational handles guard use-after-free

when slots are recycled, a bare index can silently point at a _different_ object that reused the slot. A generation counter catches it:

```c
typedef struct { uint32_t idx; uint32_t gen; } handle_t;   /* INVALID = {UINT32_MAX, 0} */

T *resolve(pool_t *p, handle_t h) {
    if (h.idx >= p->count || p->slot_gen[h.idx] != h.gen) return NULL;  /* stale / out of range */
    return &p->items[h.idx];
}
```

## How to Apply

1. Stored references are indices/handles; raw pointers are local, recovered per use, never persisted into a struct that outlives the call.
2. Use `-1` / `UINT32_MAX` as the explicit null sentinel, never a null pointer, so the "no link" state serializes too.
3. Validate the index against the live `count` before every dereference, see [references/safety-validations.md](./safety-validations.md).
4. Add a generation counter when slots are freed and reused; skip it for append-only pools.

## Related

[references/composability.md](./composability.md) (index-based currency between stages), [references/caller-owns-memory.md](./caller-owns-memory.md), **data-oriented-design-guide** (handles vs pointers, the layout rationale)
