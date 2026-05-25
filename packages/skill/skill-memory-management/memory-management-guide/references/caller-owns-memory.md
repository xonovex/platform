# caller-owns-memory: Caller-Owns-Memory

**Guideline:** A library never allocates; the caller provides all storage and the library operates on it. Functions take pointers + capacities and return status codes, never freshly allocated pointers.

**Rationale:** Hidden allocations are the root of leaks, ownership confusion, and surprise allocator cost on a hot path. When the caller owns the memory, it picks the strategy (stack, pool, arena, virtual-memory block), there is nothing to leak, and batch processing needs no per-item allocation. It also makes the library trivially testable and embeddable (no allocator dependency), and serialization/relocation become the caller's clear responsibility.

**How to Apply:**

1. Init functions bind library state to caller-provided arrays + a capacity, not to internally-malloc'd memory.
2. Operations check capacity/bounds and return a status code on overflow instead of growing.
3. Complex operations that need scratch space take a caller-provided work buffer (often sized via a `*_req()` query).
4. Return values are status codes or indices/handles — never ownership of a new allocation the caller must remember to free.

**Example:**

```c
typedef struct { float *x, *y; uint32_t count, capacity; } entity_system_t;

// Caller owns x[]/y[]; the system just borrows and tracks them.
status_t entity_system_init(entity_system_t *s, float *x, float *y, uint32_t cap) {
  s->x = x; s->y = y; s->count = 0; s->capacity = cap;
  return STATUS_OK;
}
status_t entity_add(entity_system_t *s, float x, float y) {
  if (s->count >= s->capacity) return STATUS_FULL;   // never grows/allocates
  s->x[s->count] = x; s->y[s->count] = y; s->count++;
  return STATUS_OK;
}

// Bad: hidden allocation — who frees it? leaks on error paths, allocator on hot path.
float *make_buffer(size_t n) { return malloc(n * sizeof(float)); }
```

**Gotchas:**

- A `*_req()` size query must stay in lockstep with the code that consumes the buffer; a mismatch is a buffer overflow.
- "Caller owns" is a whole-API contract — one function that secretly allocates breaks the leak-free guarantee for everyone.

**Related:** [references/arenas-and-pools.md](./arenas-and-pools.md), [references/ownership-and-lifetimes.md](./ownership-and-lifetimes.md)
