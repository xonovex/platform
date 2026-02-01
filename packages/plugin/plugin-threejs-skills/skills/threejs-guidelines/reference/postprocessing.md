# postprocessing: Post-Processing Effects

**Guideline:** Use EffectComposer to chain passes (RenderPass→effects→screen); render to intermediate targets for complex pipelines.

**Rationale:** Post-processing applies cinematic effects (bloom, grain, color grading) after rendering without modifying materials.

**Example:**

```javascript
const composer = new EffectComposer(renderer);
composer.addPass(new RenderPass(scene, camera));
composer.addPass(
  new UnrealBloomPass(
    new THREE.Vector2(w, h),
    1.5, // Strength
    0.4, // Radius
    0.85, // Threshold
  ),
);
function animate() {
  composer.render();
}
```

**Techniques:**

- UnrealBloomPass: Glow on emissive materials; adjust threshold for control
- FilmPass: Grain/noise effect with intensity and grayscale toggle
- ShaderPass: Custom fragment shader effects with uniforms for animation
- WebGLRenderTarget: Intermediate render passes for multi-stage pipelines
- Color grading: ShaderPass with saturation/brightness uniforms
- Performance: Bloom expensive; SSAO slow; half-res passes for efficiency; only final pass sets renderToScreen
