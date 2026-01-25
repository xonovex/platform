# builder-pattern: Builder Pattern for Meshes

**Guideline:** Two-phase construction: `*_req()` queries memory, `*_build()` fills caller-provided buffer.

**Rationale:** Caller controls allocation (stack, pool, arena). Builder never allocates. Exact sizes prevent over-allocation.

**Example:**

```c
// Query
mesh_req_t req = sphere_mesh3d_req(1.0f, 32, 16);

// Allocate (caller decides strategy)
vertex3f_t vertices[req.vertex_count];
uint32_t indices[req.index_count];

// Build
mesh3d_t mesh = {.vertices = vertices, .indices = indices,
                 .vertex_capacity = req.vertex_count,
                 .index_capacity = req.index_count};
status_t status = sphere_mesh3d_build(&mesh, 1.0f, 32, 16);
```

**Techniques:**
- Query phase: Call `*_req()` to get exact vertex/index counts needed
- Caller allocation: Choose strategy (stack, malloc, arena) for complete control
- Build phase: Call `*_build()` to fill pre-allocated caller buffer
- Status checking: Always check return status for insufficient capacity errors
- Function naming: Use `{shape}_mesh{dim}_req` and `{shape}_mesh{dim}_build` pattern
