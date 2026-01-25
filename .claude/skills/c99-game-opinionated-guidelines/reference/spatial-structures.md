# spatial-structures: Spatial Data Structures

**Guideline:** Use dimension-suffixed structures (grid, tree, bvh) for broad-phase queries. Choose by density/pattern.

**Rationale:** Different structures suit different spatial distributions; dimension suffixes (_2d, _3d) enable consistent APIs across dimensions.

**Example:**

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

**Techniques:**
- Uniform grids: Use for dense/uniform object distributions with O(1) insertion
- Quadtrees/Octrees: Use for sparse/varying density with O(log n) operations
- BVH structures: Use for static geometry and ray casting queries
- KD-trees: Use for point clouds and nearest-neighbor searches
- SoA variants: Provide `_soa` versions for batch processing same-type queries
