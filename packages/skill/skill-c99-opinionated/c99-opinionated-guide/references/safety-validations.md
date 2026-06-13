# safety-validations: Safety Validations

**Guideline:** Validate capacity, bounds, NULL pointers, division-by-zero, overflow before use.

**Rationale:** Prevents undefined behavior, crashes, security vulnerabilities. Return status codes for graceful error handling.

**Example:**

```c
status_t buffer_write(buffer_t *buf, const void *data, size_t size) {
    if (buf->used + size > buf->capacity) return STATUS_FULL;
    if (size > SIZE_MAX - buf->used) return STATUS_OVERFLOW;
    memcpy(buf->data + buf->used, data, size);
    buf->used += size;
    return STATUS_OK;
}

vec3_t vec3_normalize(const vec3_t *v) {
    if (!v) return (vec3_t){0};
    float len = sqrtf(v->x*v->x + v->y*v->y + v->z*v->z);
    if (len < 1e-6f) return (vec3_t){0};
    float inv = 1.0f / len;
    return (vec3_t){v->x*inv, v->y*inv, v->z*inv, 0};
}
```

**Techniques:**

- Capacity checks: Validate `used + size <= capacity` before writing
- Bounds checks: Check `idx < count` before array access
- NULL checks: Test `if (!ptr)` at function entry
- Epsilon division: Use `len < 1e-6f` for zero checks in float operations
- Overflow checks: Test `count > SIZE_MAX / size` before multiplication
- Bounded containers: Keep `data + count + capacity` together and gate every access through a checked accessor — a raw `T*` paired with a separate, unenforced length is the bug surface this removes
- Handle/index resolution: Resolve a stored index/handle against the live `count` (and generation, if slots recycle) before dereferencing — see [references/handles-and-indices.md](./handles-and-indices.md)
