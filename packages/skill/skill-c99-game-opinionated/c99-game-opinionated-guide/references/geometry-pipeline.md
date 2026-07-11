# geometry-pipeline: Geometry Pipeline

Three separated layers: Analytic (shapes for physics/logic, e.g. `sphere3d_t`, `circle2d_t`) → Discrete (indexed mesh via builder pattern) → Packing (GPU-ready vertex format). Keep physics, rendering, and GPU format concerns apart.

```c
// 1. Analytic
sphere3d_t sphere = {.center = {0, 1, 0}, .radius = 2.0f};

// 2. Discrete: build mesh
mesh_req_t req = sphere_mesh3d_req(&sphere, 32, 16);
mesh3d_t mesh = {.vertices = verts, .indices = idx,
                 .vertex_capacity = req.vertex_count};
sphere_mesh3d_build(&mesh, &sphere, 32, 16);

// 3. Packing: GPU format
size_t size = mesh3d_pack_size(&mesh, VERTEX_FORMAT_3F_4U8);
mesh3d_pack_3f_4u8(packed, &mesh, color);
```
