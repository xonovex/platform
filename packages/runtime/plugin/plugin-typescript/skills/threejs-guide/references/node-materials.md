# node-materials: TSL Node Materials

Use TSL (Three Shading Language) for materials that run on both WebGL (→GLSL) and WebGPU (→WGSL); compose shader nodes instead of writing raw GLSL.

```javascript
import {float, mix, sin, uniform, vec3} from "three/tsl";

const timeUniform = uniform(float, "time");
const material = new THREE.MeshStandardNodeMaterial({
  color: mix(
    vec3(1.0, 0.0, 0.0),
    vec3(0.0, 0.0, 1.0),
    sin(timeUniform).mul(0.5).add(0.5),
  ),
});
timeUniform.value = clock.getElapsedTime();
```

## Techniques

- TSL nodes: `float()`, `vec2/3/4()`, `sin()`, `cos()`, `mix()`, `step()`, `normalize()`
- Uniforms: `uniform(type, name)` for runtime control; update `.value` each frame
- Varyings: `varying(type)` for vertex→fragment data passing
- Functions: `Fn([args], returnType, ({params}) => {...})` for reusable shader logic
- Textures: `texture(textureObject)` for sampling; `storageTexture()` for compute read/write
- Positions: `positionLocal`, `positionWorld`, `normalWorld` for vertex data
- Material types: MeshBasicNodeMaterial, MeshStandardNodeMaterial, MeshPhysicalNodeMaterial
- Lighting: `lightingContext()` for advanced PBR lighting effects
- Composition: Operators like `.mul()`, `.add()`, `.dot()`, `.cross()` chain naturally
