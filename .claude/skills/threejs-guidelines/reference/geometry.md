# geometry: Geometry and BufferGeometry

**Guideline:** Use built-in geometries for standard shapes; create efficient custom BufferGeometry with position/UV/indices/normals; use InstancedMesh.

**Rationale:** Built-ins are optimized; proper BufferGeometry avoids redundant data; InstancedMesh enables thousands of copies in one draw call.

**Example:**
```javascript
const geometry = new THREE.BufferGeometry();
geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
geometry.setAttribute('uv', new THREE.BufferAttribute(uvs, 2));
geometry.setIndex(new THREE.BufferAttribute(indices, 1));
geometry.computeVertexNormals();
```

**Techniques:**
- Built-ins: BoxGeometry, SphereGeometry, PlaneGeometry, CylinderGeometry, IcosahedronGeometry, etc.
- Path-based: LatheGeometry (vases), TubeGeometry (splines), ExtrudeGeometry (2D extrude), TextGeometry (fonts)
- Custom BufferGeometry: Position (required), UV, indices (reuse vertices), normals, bounds
- InstancedMesh: Per-instance transforms via setMatrixAt(); per-instance colors via setColorAt()
- Optimization: Merge geometries with BufferGeometryUtils; dispose with geometry.dispose()
- Morph targets: Blend multiple shape deformations with morphTargetInfluences
