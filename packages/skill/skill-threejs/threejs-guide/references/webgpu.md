# webgpu: WebGPU Renderer

Use WebGPU for compute shaders and lower CPU overhead; call `await renderer.init()` before rendering; use TSL for cross-renderer materials.

```javascript
import * as THREE from "three/webgpu";

const renderer = new THREE.WebGPURenderer({antialias: true});
renderer.setSize(w, h);
await renderer.init(); // Required async init

const material = new THREE.MeshStandardNodeMaterial({
  color: vec3(1.0, 0.0, 0.0),
});
const mesh = new THREE.Mesh(geometry, material);
scene.add(mesh);
renderer.render(scene, camera);
```

## Techniques

- Setup: `new THREE.WebGPURenderer()` + `await renderer.init()`; async required for device
- Fallback: Check `navigator.gpu` before creating; fall back to WebGL if unavailable
- Compute shaders: `Fn([args], returnType)` defines GPU-side computation; `renderer.compute()` executes
- Storage: `storageTexture()` and `storageBuffer()` for read/write GPU data
- TSL materials: Work with both WebGL (→GLSL) and WebGPU (→WGSL) automatically
- Debugging: `renderer.info`, `renderer.getDevice()` for device limits/features
- Limitations: Chrome 113+, Firefox Nightly, Safari TP; no polyfill; mobile support limited
- Migration: Change import from `'three'` to `'three/webgpu'`; add async init; use TSL for shaders
- Tips: Use compute for physics/particles; minimize CPU↔GPU transfers; batch similar operations
