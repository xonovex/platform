# scene-fundamentals: Scene, Renderer, and Core Concepts

**Guideline:** Configure Scene/Renderer; use Object3D hierarchy with proper dispose patterns; animate with setAnimationLoop and delta time.

**Rationale:** Proper setup ensures memory safety, correct rendering, and frame-rate independence; improper disposal causes memory leaks.

**Example:**
```javascript
const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
renderer.outputColorSpace = THREE.SRGBColorSpace;
renderer.shadowMap.enabled = true;
const clock = new THREE.Clock();
renderer.setAnimationLoop(() => {
  const delta = clock.getDelta();
  mesh.rotation.y += delta * 0.5;
  renderer.render(scene, camera);
});
```

**Techniques:**
- Renderer config: antialias, alpha (transparency), outputColorSpace (sRGB), toneMaps, shadowMap
- Scene setup: background (color/texture), environment map (PBR), fog, traverse()
- Object3D: position/rotation/scale, add/remove, visible (hide without disposal), layers, userData
- Dispose order: geometry → texture → material → renderer (prevents memory leaks)
- Animation: Use setAnimationLoop() not requestAnimationFrame (WebXR); use delta for frame-rate independence
- Coordinate system: Right-handed (+X right, +Y up, +Z toward viewer); helpers for debugging
