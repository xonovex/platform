# patterns: Architecture Patterns

Organize scenes in Groups; multiply movement by delta time for frame-rate independence; pool objects and clean up to avoid GPU memory exhaustion.

```javascript
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

## Techniques

- Scene hierarchy: Group objects (world → player/enemies/effects) for batch transforms
- Frame-rate independence: Multiply movement by `clock.getDelta()`; use `getElapsedTime()` for animations
- Cleanup pattern: Dispose geometry/materials; `mixer.stopAllAction()`, remove listeners, call on `beforeunload`
- Object pooling: Reuse bullets/particles with `.visible` flag instead of create/destroy
- Component pattern: Encapsulate entities (Player class with update/takeDamage methods)
- Event bus: Centralize game events for decoupled communication between systems
