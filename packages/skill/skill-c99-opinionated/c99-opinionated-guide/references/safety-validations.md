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
- Handle/index resolution: resolve a stored index/handle against the live `count` (and generation, if slots recycle) before dereferencing — see [references/handles-and-indices.md](./handles-and-indices.md).
