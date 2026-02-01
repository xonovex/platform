# mesh-types: Mesh Data Structures

**Guideline:** Use `mesh2d_t`/`mesh3d_t` with capacity tracking. Builders generate from analytic shapes.

**Rationale:** Separate count/capacity enables efficient reuse; builder pattern decouples construction from memory strategy.

**Example:**

```c
typedef struct {
    vertex2f_t *vertices;
    uint32_t *indices;
    size_t vertex_count, index_count;
    size_t vertex_capacity, index_capacity;
} mesh2d_t;

typedef struct {
    vertex3f_t *vertices;
    vec3f_t *normals;     // Optional
    vec2f_t *uvs;         // Optional
    uint32_t *indices;
    size_t vertex_count, index_count;
    size_t vertex_capacity, index_capacity;
} mesh3d_t;
```

**Techniques:**

- Capacity tracking: Maintain separate count and capacity for dynamic growth
- Builder pattern: Use `{shape}_mesh{dim}_{req|build}` naming convention
- Optional data: Include normals/UVs in mesh3d for advanced rendering
- Operations: Provide bounds calculation, normal computation, merging
- Vertex types: Use aligned types like `vertex3f_t` with padding for SIMD
