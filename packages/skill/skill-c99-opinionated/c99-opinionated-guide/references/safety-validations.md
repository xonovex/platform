# safety-validations: Safety Validations

Validate capacity, bounds, NULL, division, and multiplication overflow before use; return status codes rather than crashing.

```c
status_t buffer_write(buffer_t *buf, const void *data, size_t size) {
    if (size > SIZE_MAX - buf->used) return STATUS_OVERFLOW;
    if (buf->used + size > buf->capacity) return STATUS_FULL;
    memcpy(buf->data + buf->used, data, size);
    buf->used += size;
    return STATUS_OK;
}

vec3_t vec3_normalize(const vec3_t *v) {
    if (!v) return (vec3_t){0};
    float len = sqrtf(v->x*v->x + v->y*v->y + v->z*v->z);
    if (len < 1e-6f) return (vec3_t){0};        // epsilon zero-check for float division
    float inv = 1.0f / len;
    return (vec3_t){v->x*inv, v->y*inv, v->z*inv, 0};
}
```

- Multiplication overflow: test `count > SIZE_MAX / size` before `count * size`.
- Bounded containers: keep `data + count + capacity` together and gate every access through a checked accessor — a raw `T*` plus a separate, unenforced length is the bug surface this removes.
- Caller-provided storage: expose a `size_t *_req(...)` query beside the matching `*_init(T *state, void *memory, size_t memory_size)` layout calculation; include alignment padding, then validate pointer, size, and alignment before placement.
- Scratch buffers: pass `void *scratch, size_t scratch_size` explicitly and document whether input, output, and scratch may overlap; do not apply `restrict` assumptions when overlap is allowed.
- Handle/index resolution: check `index < capacity` before reading alive or generation arrays, then check liveness and generation; resolve to `T *` only within a scope where no grow, compact, recycle, or remove can run.
