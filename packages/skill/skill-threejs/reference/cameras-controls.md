# filename: cameras-controls

**Guideline:** Choose camera type (Perspective, Orthographic, Cube); select controls matching interaction (OrbitControls, FlyControls, PointerLockControls).

**Rationale:** Proper camera frustum prevents z-fighting; tight near/far prevents clipping; controls must match intended interaction pattern.

**Example:**

```javascript
const camera = new THREE.PerspectiveCamera(75, w / h, 0.1, 1000);
camera.position.set(5, 5, 5);

const controls = new OrbitControls(camera, renderer.domElement);
controls.enableDamping = true;
controls.dampingFactor = 0.05;
controls.target.set(0, 0, 0);
controls.update();
```

**Techniques:**

- Camera types: PerspectiveCamera (3D standard), OrthographicCamera (CAD/isometric), CubeCamera (reflections)
- Near/far planes: Large scenes use `near=1, far=100000`; small scenes `near=0.001, far=100` to avoid z-fighting
- OrbitControls: `enableDamping=true` for smooth motion; set `minDistance`/`maxDistance` and polar angle limits
- FlyControls: Pass `clock.getDelta()` to `.update()` for frame-rate independence
- PointerLockControls: Call `.lock()` on click; check `.isLocked`; use `.moveForward()`, `.moveRight()`
- Multi-camera: `.setViewport()`, `.setScissor()`, render twice for split-screen/PIP
- Coordinate conversion: `.project()` for world→screen; `.unproject()` for screen→world
