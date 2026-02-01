# vertex-packing: Vertex Packing for GPU

**Guideline:** Pack mesh data to GPU-ready formats. Support interleaved (single buffer) and non-interleaved (streaming).

**Rationale:** GPU-ready vertex formats maximize performance; flexible layouts support different rendering pipelines and optimization strategies.

**Example:**

```c
typedef struct {
    float x, y, z;
    uint8_t r, g, b, a;
} vertex_3f_4u8_t;

void mesh3d_pack_3f_4u8(vertex_3f_4u8_t *out, const mesh3d_t *m, color4u8_t c) {
    for (size_t i = 0; i < m->vertex_count; i++) {
        out[i] = (vertex_3f_4u8_t){
            m->vertices[i].x, m->vertices[i].y, m->vertices[i].z,
            c.r, c.g, c.b, c.a
        };
    }
}

size_t mesh3d_pack_size(const mesh3d_t *m, vertex_format_t fmt) {
    switch(fmt) {
        case VERTEX_FORMAT_3F: return m->vertex_count * 12;
        case VERTEX_FORMAT_3F_4U8: return m->vertex_count * 16;
    }
    return 0;
}
```

**Techniques:**

- Vertex format naming: Use `vertex_{components}_{types}_t` (e.g., `vertex_3f_4u8_t`)
- Interleaved layout: Pack all attributes per-vertex in single buffer
- Streaming layout: Separate position/normal/UV into different buffers
- Size queries: Provide `*_pack_size` functions before allocation
- No GPU calls: Library never touches graphics APIs; caller handles upload
