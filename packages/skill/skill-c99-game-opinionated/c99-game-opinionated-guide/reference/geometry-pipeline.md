# geometry-pipeline: Geometry Pipeline

**Guideline:** Three-layer pipeline: Analytic (logic) → Discrete (mesh) → Packing (GPU).

**Rationale:** Each layer handles specific concerns: math/physics, rendering data, GPU format optimization.

**Example:**

```c
// 1. Analytic
sphere3d_t sphere = {.center = {0, 1, 0}, .radius = 2.0f};

// 2. Discrete: Build mesh
mesh_req_t req = sphere_mesh3d_req(&sphere, 32, 16);
mesh3d_t mesh = {.vertices = verts, .indices = idx,
                 .vertex_capacity = req.vertex_count};
sphere_mesh3d_build(&mesh, &sphere, 32, 16);

// 3. Packing: GPU format
size_t size = mesh3d_pack_size(&mesh, VERTEX_FORMAT_3F_4U8);
mesh3d_pack_3f_4u8(packed, &mesh, color);
```

**Techniques:**

- Analytic layer: Define shapes like `sphere3d_t`, `circle2d_t` for physics/logic
- Discrete layer: Build indexed meshes from analytic shapes with builder pattern
- Packing layer: Convert mesh to GPU-ready vertex formats (interleaved/streaming)
- Separation of concerns: Keep physics, rendering, and GPU separate
- Conversion functions: Provide `*_req` and `*_build` for shape-to-mesh conversion
