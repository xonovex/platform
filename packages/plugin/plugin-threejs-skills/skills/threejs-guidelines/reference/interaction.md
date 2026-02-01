# interaction: Raycasting and User Interaction

**Guideline:** Convert screen to normalized device coordinates, then raycast; cache raycaster instance and track state for hover/drag.

**Rationale:** Raycasting picks objects; state tracking enables hover highlights and drag interactions; coordinate conversion is essential.

**Example:**

```javascript
const raycaster = new THREE.Raycaster();
const pointer = new THREE.Vector2();
window.addEventListener("pointermove", (e) => {
  pointer.x = (e.clientX / innerWidth) * 2 - 1;
  pointer.y = -(e.clientY / innerHeight) * 2 + 1;
  raycaster.setFromCamera(pointer, camera);
  const hits = raycaster.intersectObjects(scene.children, true);
  if (hits.length > 0) console.log(hits[0].object, hits[0].point);
});
```

**Techniques:**

- NDC conversion: `pointer.x = (clientX / width) * 2 - 1`, `pointer.y = -(clientY / height) * 2 + 1`
- Intersection properties: object, point (world), distance, uv, face, instanceId
- Hover effects: Track state, restore original color on exit
- Drag implementation: pointerdown to start, pointermove to update, pointerup to release
- Performance: Filter interactive objects before raycast; don't raycast every frame on heavy scenes
