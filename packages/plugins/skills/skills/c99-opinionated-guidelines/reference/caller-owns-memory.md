# caller-owns-memory: Caller-Owns-Memory Pattern

**Guideline:** Libraries never allocate. Caller provides all storage; library operates on it.

**Rationale:** No hidden allocations. Caller controls memory strategy (stack, pool, arena). No leaks. Batch-friendly.

**Example:**

```c
typedef struct {
    float *pos_x, *pos_y;
    uint32_t count, capacity;
} entity_system_t;

status_t entity_system_init(entity_system_t *sys,
    float *x, float *y, uint32_t cap) {
    sys->pos_x = x;
    sys->pos_y = y;
    sys->count = 0;
    sys->capacity = cap;
    return STATUS_OK;
}

status_t entity_add(entity_system_t *sys, float x, float y) {
    if (sys->count >= sys->capacity) return STATUS_FULL;
    sys->pos_x[sys->count] = x;
    sys->pos_y[sys->count] = y;
    sys->count++;
    return STATUS_OK;
}
```

**Techniques:**
- Functions receive pointers: Accept caller-provided storage instead of allocating
- Init functions: Bind library state to caller-owned arrays with capacity tracking
- Status returns: Return status codes, never allocated pointers
- Work buffers: Accept scratch space from caller for temporary operations
- Batch-friendly: Enable processing arrays without individual allocations
