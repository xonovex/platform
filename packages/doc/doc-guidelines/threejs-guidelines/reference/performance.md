# performance: Performance Optimization

**Guideline:** Minimize draw calls with InstancedMesh; use LOD for distance; dispose unused; reuse geometry/materials; profile before optimizing.

**Rationale:** GPU efficiency depends on draw calls and memory; InstancedMesh renders 1000+ copies in one call vs. 1000 separate meshes.

**Example:**

```javascript
const mesh = new THREE.InstancedMesh(geometry, material, 1000);
for (let i = 0; i < 1000; i++) {
  dummy.position.set(...);
  dummy.updateMatrix();
  mesh.setMatrixAt(i, dummy.matrix);
}
mesh.instanceMatrix.needsUpdate = true;
```

**Techniques:**

- InstancedMesh: Render thousands of copies in one draw call (vs. 1000 draw calls)
- LOD (Level of Detail): High detail near (0-50), medium mid-range (50-200), low far (200+)
- Merge geometries: Combine static meshes with BufferGeometryUtils.mergeGeometries()
- Memory: Dispose geometry/material/texture; reuse between meshes; unload offscreen
- Rendering: Shadow map 1024×1024; limit lights to 3-5; prefer alphaTest over transparency
- Profiling: `renderer.info.render.calls` (target <1000), `renderer.info.memory`
