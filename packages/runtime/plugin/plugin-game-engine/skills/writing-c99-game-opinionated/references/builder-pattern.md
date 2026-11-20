# builder-pattern: Builder Pattern for Meshes

Name the mesh builder pair `{shape}_mesh{dim}_req` (returns exact vertex/index counts) / `{shape}_mesh{dim}_build` (fills a caller-provided buffer, returns a status).

```c
mesh_req_t req = sphere_mesh3d_req(1.0f, 32, 16);
vertex3f_t vertices[req.vertex_count];
uint32_t indices[req.index_count];
mesh3d_t mesh = {.vertices = vertices, .indices = indices,
                 .vertex_capacity = req.vertex_count,
                 .index_capacity = req.index_count};
status_t status = sphere_mesh3d_build(&mesh, 1.0f, 32, 16);
```
