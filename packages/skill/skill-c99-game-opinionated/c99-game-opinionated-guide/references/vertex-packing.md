# vertex-packing: Vertex Packing for GPU

Pack mesh data into GPU-ready formats named `vertex_{components}_{types}_t` (e.g. `vertex_3f_4u8_t`). Support interleaved (all attributes per vertex, single buffer) and streaming (position/normal/UV in separate buffers). Provide `*_pack_size` before allocation. The library never calls a graphics API — the caller uploads.

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
    switch (fmt) {
        case VERTEX_FORMAT_3F:     return m->vertex_count * 12;
        case VERTEX_FORMAT_3F_4U8: return m->vertex_count * 16;
    }
    return 0;
}
```
