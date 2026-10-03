# spatial-structures: Spatial Data Structures

Dimension-suffixed structures (`_2d`/`_3d`) for broad-phase queries, chosen by distribution: uniform grid (dense/uniform, O(1) insert), quad/octree (sparse/varying, O(log n)), BVH (static geometry, ray casts), KD-tree (point clouds, nearest-neighbor). Provide `_soa` variants for batch same-type queries.

```c
typedef struct {
    uint32_t *cells, *objects;
    aabb2d_t bounds;
    uint32_t cells_x, cells_y;
} grid2d_t;

uint32_t cell_x = (uint32_t)((pos.x - g->bounds.min.x) / cell_width);
uint32_t cell_y = (uint32_t)((pos.y - g->bounds.min.y) / cell_height);
uint32_t cell_idx = cell_y * g->cells_x + cell_x;
```
