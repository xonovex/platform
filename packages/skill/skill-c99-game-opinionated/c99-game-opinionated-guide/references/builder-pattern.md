# builder-pattern: Builder Pattern for Meshes

Two-phase construction: `*_req()` returns exact vertex/index counts, `*_build()` fills a caller-provided buffer and returns a status (never allocates; caller picks stack/pool/arena). Name the pair `{shape}_mesh{dim}_req` / `{shape}_mesh{dim}_build`. Always check the build status for insufficient-capacity errors.

```c
mesh_req_t req = sphere_mesh3d_req(1.0f, 32, 16);
vertex3f_t vertices[req.vertex_count];
uint32_t indices[req.index_count];
mesh3d_t mesh = {.vertices = vertices, .indices = indices,
                 .vertex_capacity = req.vertex_count,
                 .index_capacity = req.index_count};
status_t status = sphere_mesh3d_build(&mesh, 1.0f, 32, 16);
```
