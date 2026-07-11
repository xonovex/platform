# mesh-types: Mesh Data Structures

`mesh2d_t`/`mesh3d_t` carry separate `count` and `capacity` per array so buffers can be reused; builders fill them from analytic shapes (`{shape}_mesh{dim}_{req|build}`). Normals and UVs on `mesh3d_t` are optional pointers.

```c
typedef struct {
    vertex2f_t *vertices;
    uint32_t *indices;
    size_t vertex_count, index_count;
    size_t vertex_capacity, index_capacity;
} mesh2d_t;

typedef struct {
    vertex3f_t *vertices;
    vec3f_t *normals;     // optional
    vec2f_t *uvs;         // optional
    uint32_t *indices;
    size_t vertex_count, index_count;
    size_t vertex_capacity, index_capacity;
} mesh3d_t;
```
