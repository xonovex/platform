# filename: math

**Guideline:** Use Three.js math classes (Vector3, Matrix4, Quaternion) and reuse instances to avoid garbage collection in animation loops.

**Rationale:** Three.js math classes are GPU-optimized; allocating new instances every frame causes GC stalls and framerate drops.

**Example:**
```javascript
// Reuse pattern: Define once, mutate in loop
const v3_1 = new THREE.Vector3();
for (let i = 0; i < objects.length; i++) {
  v3_1.copy(objects[i].position).normalize();
  // Use v3_1...
}

// Vector operations
v.add(other); v.multiplyScalar(2); v.lerp(target, 0.5);
// Quaternion: q.slerp(target, t); Matrix4: m.compose(pos, quat, scale);
```

**Techniques:**
- Vector3: `.set()`, `.copy()`, `.add()`, `.multiplyScalar()`, `.normalize()`, `.dot()`, `.cross()`, `.lerp()`
- Quaternion: `.setFromAxisAngle()`, `.slerp()` for smooth rotation interpolation avoiding gimbal lock
- Matrix4: `.compose()` combines position/quaternion/scale; use `.invert()` for transforms
- Color: `.setHSL()`, `.lerp()` for smooth color transitions
- Bounding volumes: Box3/Sphere for collision detection with `.intersectsBox()`, `.containsPoint()`
- MathUtils: `degToRad()`, `lerp()`, `clamp()`, `smoothstep()` for common operations
