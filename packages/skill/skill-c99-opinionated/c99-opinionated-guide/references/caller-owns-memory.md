# caller-owns-memory: Caller-Owns-Memory (C)

## Guideline

In this style, libraries never allocate — the caller provides all storage and the library operates on it through pointers + capacities, returning status codes (never an allocated pointer).

## Rationale

This is the defining memory choice of the opinionated style: no hidden `malloc`, nothing to leak, embeddable with no allocator dependency, and batch-friendly. The general principle — ownership, lifetimes, and the allocators the caller plugs in (arena/pool/virtual-memory) — is language-agnostic; see **memory-management-guide**. This doc is just the C mechanics.

## How to Apply (C specifics)

1. Init functions bind library state to caller arrays + a capacity, never to internally-malloc'd memory.
2. Operations check capacity/bounds and return a status code on overflow (no `realloc`).
3. Complex operations take a caller-provided scratch buffer, sized via a `*_req()` query.
4. Return status codes or indices/handles — never ownership of a fresh allocation.

## Example

```c
typedef struct { float *x, *y; uint32_t count, capacity; } entity_system_t;

status_t entity_system_init(entity_system_t *s, float *x, float *y, uint32_t cap) {
  s->x = x; s->y = y; s->count = 0; s->capacity = cap;  // caller owns x[]/y[]
  return STATUS_OK;
}
status_t entity_add(entity_system_t *s, float x, float y) {
  if (s->count >= s->capacity) return STATUS_FULL;      // never grows/allocates
  s->x[s->count] = x; s->y[s->count] = y; s->count++;
  return STATUS_OK;
}
```

## Related

**memory-management-guide** (general allocation, ownership, arenas/pools), [references/safety-validations.md](./safety-validations.md)
