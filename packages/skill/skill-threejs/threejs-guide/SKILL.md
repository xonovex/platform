---
name: threejs-guide
description: "Use when building or editing 3D scenes in vanilla Three.js for WebGL/WebGPU. Triggers on `.js`/`.ts` files with `three` or `three/*` imports, and on prompts about Scene, Renderer, Mesh, BufferGeometry, materials, lighting, GLTF loaders, raycasting, shaders, post-processing, instancing, or perf tuning — even when the user doesn't say 'Three.js'. Skip React Three Fiber, Babylon.js, and other 3D engines."
---

# Three.js Coding Guidelines

Vanilla Three.js for scene setup, rendering, meshes, materials, lights, animation, interaction, and effects.

## Essentials

- **Scene structure** - Organize Object3D hierarchy; use Groups for logical organization, see [references/scene-fundamentals.md](references/scene-fundamentals.md)
- **Render loop** - Frame-rate independence with `clock.getDelta()`, see [references/scene-fundamentals.md](references/scene-fundamentals.md)
- **Disposal** - Dispose geometries, materials, textures to prevent memory leaks, see [references/scene-fundamentals.md](references/scene-fundamentals.md)
- **Pixel ratio** - Cap at 2 with `Math.min(devicePixelRatio, 2)`, see [references/scene-fundamentals.md](references/scene-fundamentals.md)
- **Color space** - Use `SRGBColorSpace` for colors, `NoColorSpace` for data, see [references/textures.md](references/textures.md)
- **Shadows** - Enable on renderer, light, mesh; keep frustum tight, see [references/lighting-shadows.md](references/lighting-shadows.md)

## Core Topics

- Read [references/scene-fundamentals.md](references/scene-fundamentals.md) - Scene, Renderer, Object3D, cleanup, render loop
- Read [references/geometry.md](references/geometry.md) - Shapes, BufferGeometry, custom geometry, instancing
- Read [references/cameras-controls.md](references/cameras-controls.md) - Cameras, OrbitControls, FlyControls, viewport
- Read [references/materials.md](references/materials.md) - PBR materials, ShaderMaterial, properties
- Read [references/textures.md](references/textures.md) - Loading, UV mapping, render targets, environment
- Read [references/lighting-shadows.md](references/lighting-shadows.md) - Lights, shadows, IBL, light probes
- Read [references/animation.md](references/animation.md) - Keyframes, skeletal, morph targets, AnimationMixer
- Read [references/interaction.md](references/interaction.md) - Raycasting, selection, drag, coordinate conversion
- Read [references/loaders.md](references/loaders.md) - GLTF, FBX, textures, HDR, compression
- Read [references/shaders.md](references/shaders.md) - GLSL shaders, uniforms, varyings
- Read [references/postprocessing.md](references/postprocessing.md) - Bloom, DOF, SSAO, custom effects
- Read [references/performance.md](references/performance.md) - InstancedMesh, LOD, culling, batching, profiling
- Read [references/patterns.md](references/patterns.md) - Architecture patterns, asset management, state machines
- Read [references/math.md](references/math.md) - Vector3, Matrix4, Quaternion, Box3, curves, MathUtils
- Read [references/node-materials.md](references/node-materials.md) - TSL (Three Shading Language), node-based materials
- Read [references/physics-vr.md](references/physics-vr.md) - Physics engines (Rapier, Cannon), WebXR (VR/AR)
- Read [references/webgpu.md](references/webgpu.md) - WebGPU renderer, compute shaders, modern GPU

## Gotchas

- Geometries, materials, and textures must be explicitly `.dispose()`d — JS GC doesn't free GPU memory; long sessions leak VRAM
- `BufferGeometry` replaced legacy `Geometry` years ago — old tutorials using `Geometry` silently fail on current versions
- `renderer.setPixelRatio(window.devicePixelRatio)` is critical for retina — without it, scenes render at 1× and look blurry
- Frustum culling is automatic for `Mesh`, but `LineSegments`/`Points` need explicit `geometry.computeBoundingSphere()` first
- Loading a `.glb`/`.gltf` is async — accessing `scene.children` immediately after `loader.load()` returns an empty array
