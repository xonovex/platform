# filename: patterns

**Guideline:** Organize scenes in Groups; implement proper cleanup to prevent memory leaks; use delta time for frame-rate independent updates.

**Rationale:** Hierarchical organization enables efficient transforms; cleanup prevents GPU memory exhaustion; delta time ensures consistent movement across devices.

**Example:**
```javascript
const scene = new THREE.Scene();
const clock = new THREE.Clock();
const world = new THREE.Group();
world.add(player, enemies, effects);
scene.add(world);

function animate() {
  const delta = clock.getDelta();
  player.position.x += moveSpeed * delta;
  mesh.rotation.y += angularVelocity * delta;
  renderer.render(scene, camera);
}
```

**Techniques:**
- Scene hierarchy: Group objects (world → player/enemies/effects) for batch transforms
- Frame-rate independence: Multiply movement by `clock.getDelta()`; use `getElapsedTime()` for animations
- Cleanup pattern: Dispose geometry/materials; `mixer.stopAllAction()`, remove listeners, call on `beforeunload`
- Object pooling: Reuse bullets/particles with `.visible` flag instead of create/destroy for performance
- Component pattern: Encapsulate entities (Player class with update/takeDamage methods)
- Event bus: Centralize game events for decoupled communication between systems
