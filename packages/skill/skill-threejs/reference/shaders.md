# filename: shaders

**Guideline:** Use ShaderMaterial with uniforms for parameter control; modify built-in shaders via `onBeforeCompile`; move expensive calculations to vertex shader.

**Rationale:** Custom shaders enable advanced effects; uniform updates avoid recompilation; vertex-side calculations reduce fragment shader load.

**Example:**

```javascript
const material = new THREE.ShaderMaterial({
  uniforms: {time: {value: 0}},
  vertexShader: `varying vec2 vUv;
    void main() {
      vUv = uv;
      gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
    }`,
  fragmentShader: `varying vec2 vUv;
    void main() { gl_FragColor = vec4(vUv, 1.0, 1.0); }`,
});
material.uniforms.time.value = clock.getElapsedTime();
```

**Techniques:**

- Uniform types: float, vec2/vec3, Color, sampler2D (texture), Matrix4, arrays
- Varyings: Pass data vertex→fragment (vUv, vNormal, vWorldPos)
- Built-in matrices: modelMatrix, modelViewMatrix, projectionMatrix, normalMatrix auto-provided
- onBeforeCompile: Inject into Three.js shaders using `#include <begin_vertex>` injection points
- GLSL math: `mix()`, `step()`, `smoothstep()`, `dot()`, `normalize()`, `cross()`
- Performance: Use `lowp/mediump` precision; minimize branching; move expensive ops to vertex shader
