# filename: physics-vr

**Guideline:** Use Rapier (or Cannon-es) for physics; sync physics bodies to meshes each frame; use `renderer.setAnimationLoop()` for WebXR.

**Rationale:** Three.js has no built-in physics; external engines require frame-based sync. WebXR requires specific animation loop API.

**Example:**

```javascript
// Physics setup
const world = new RAPIER.World(new RAPIER.Vector3(0, -9.81, 0));
const bodyDesc = RAPIER.RigidBodyDesc.dynamic().setTranslation(0, 10, 0);
const body = world.createRigidBody(bodyDesc);
const shape = RAPIER.ColliderDesc.cuboid(0.5, 0.5, 0.5);
world.createCollider(shape, body);

// Sync loop
function animate() {
  world.step();
  const {x, y, z} = body.translation();
  const {x: qx, y: qy, z: qz, w: qw} = body.rotation();
  mesh.position.set(x, y, z);
  mesh.quaternion.set(qx, qy, qz, qw);
  renderer.render(scene, camera);
}
renderer.setAnimationLoop(animate);
```

**Techniques:**

- Physics engines: Rapier (recommended), Cannon-es, Oimo; create world with gravity
- Bodies: RigidBodyDesc dynamic/static/kinematic; add shapes (colliders)
- Forces: `.applyForce()`, `.applyImpulse()`, `.setLinvel()`, `.setAngvel()`
- Raycasting: `world.castRay()` for hit detection with physics bodies
- WebXR setup: `renderer.xr.enabled = true`; `renderer.setAnimationLoop(animate)` (not requestAnimationFrame)
- XR input: `renderer.xr.getController(0/1)` for controllers; listen to 'select'/'squeeze' events
- AR hit testing: `frame.getHitTestResults()` returns positions to place objects
- VR comfort: Minimize latency, avoid sudden motions; limit physics timestep to ≤16.67ms
