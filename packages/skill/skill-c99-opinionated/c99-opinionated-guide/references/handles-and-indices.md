# Reference by index/handle, not raw pointer

Cross-object references are stored as a stable index or generational handle into a caller-owned array, never a raw `T *`; the pointer is recovered transiently at use, never persisted back into a struct.

## Generational handles guard use-after-free

When slots are recycled, a bare index can silently point at a _different_ object that reused the slot. A generation counter catches it:

```c
typedef struct { uint32_t idx; uint32_t gen; } handle_t;   /* INVALID = {UINT32_MAX, 0} */

T *resolve(pool_t *p, handle_t h) {
    if (h.idx >= p->count || p->slot_gen[h.idx] != h.gen) return NULL;  /* stale / out of range */
    return &p->items[h.idx];
}
```

Add a generation counter when slots are freed and reused; skip it for append-only pools. Validate the index against the live `count` before every dereference, see [references/safety-validations.md](./safety-validations.md).

## Related

[references/composability.md](./composability.md) (index-based currency between stages), [references/caller-owns-memory.md](./caller-owns-memory.md), **data-oriented-design-guide** (handles vs pointers, the layout rationale)
