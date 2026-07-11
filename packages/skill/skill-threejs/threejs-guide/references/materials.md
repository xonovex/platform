# materials: Materials

Choose material type by performance need; reuse material instances to cut draw calls; handle transparency with proper depth writes and render order.

```javascript
const pbr = new THREE.MeshStandardMaterial({
  color: 0xffffff,
  roughness: 0.5,
  metalness: 0.0,
  map: texture,
  normalMap: normalTex,
  envMap: envTex,
});
// Reuse: mesh1.material = pbr; mesh2.material = pbr;
```

## Techniques

- Material types: MeshBasicMaterial (fastest, unlit), Lambert (matte), Phong (shiny), Standard/Physical (PBR realistic)
- PBR properties: roughness 0=mirror/1=diffuse; metalness 0=dielectric/1=metal; use maps for per-pixel variation
- Transparency: `.alphaTest` for hard edges (fast); `.transparent=true` + `.depthWrite=false` + `.renderOrder` for smooth (slow)
- Blending: `AdditiveBlending` for glow/fire; `MultiplyBlending` for shadows
- Performance: Material pooling with Map; reuse instances; use `onBeforeCompile()` for shader hacks; dispose() to prevent leaks
