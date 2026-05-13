# filename: lighting-shadows

**Guideline:** Use light types appropriate to scene; keep shadow frustums tight; leverage IBL (Environment maps) for realistic ambient lighting.

**Rationale:** Lighting creates atmosphere; tight shadow frustums improve quality; IBL is more efficient than many point lights for PBR.

**Example:**

```javascript
renderer.shadowMap.enabled = true;
renderer.shadowMap.type = THREE.PCFSoftShadowMap;

const dirLight = new THREE.DirectionalLight(0xffffff, 1);
dirLight.castShadow = true;
dirLight.shadow.camera.left = -10;
dirLight.shadow.camera.right = 10;
mesh.castShadow = true;
mesh.receiveShadow = true;

new RGBELoader().load("env.hdr", (tex) => {
  tex.mapping = THREE.EquirectangularReflectionMapping;
  scene.environment = tex;
});
```

**Techniques:**

- Light types: AmbientLight (fill), HemisphereLight (sky/ground outdoors), DirectionalLight (sun), PointLight (bulb), SpotLight (flashlight)
- Shadow types: PCFSoftShadowMap (best quality); tight frustum (camera.left/right/top/bottom) and mapSize (512-2048)
- Shadow bias: `.bias = -0.0001`, `.normalBias = 0.02` to fix acne/peter-panning
- IBL: Load HDR with RGBELoader; set `scene.environment` and `scene.background`
- Performance: Limit to 3-5 lights; use contact shadows or baked lighting for static scenes
- Helpers: DirectionalLightHelper, PointLightHelper, CameraHelper for debug
