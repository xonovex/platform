# animation: Animation System

**Guideline:** Use AnimationMixer for GLTF keyframe animations; use clock.getDelta() for frame-rate independent procedural animation.

**Rationale:** Mixer efficiently plays complex GLTF animations; delta time ensures consistent speed across frame rates.

**Example:**
```javascript
const mixer = new THREE.AnimationMixer(gltf.scene);
const clip = THREE.AnimationClip.findByName(gltf.animations, 'Walk');
const action = mixer.clipAction(clip);
action.play();
const clock = new THREE.Clock();
renderer.setAnimationLoop(() => {
  mixer.update(clock.getDelta());
  renderer.render(scene, camera);
});
```

**Techniques:**
- AnimationMixer: Create from scene root; plays multiple clips simultaneously
- AnimationAction: play(), stop(), reset(); loop modes (Repeat, Once, PingPong); timeScale for speed
- Blending: crossFadeTo(nextAction, duration) for smooth transitions; weight blending for layered animations
- Procedural: Use elapsed time with Math.sin/cos for smooth curves; delta time for frame-rate independence
- Keyframe creation: VectorKeyframeTrack (position), QuaternionKeyframeTrack (rotation), AnimationClip (combines tracks)
- Performance: Single mixer for multiple objects; stopAllAction() and uncacheRoot() for cleanup
