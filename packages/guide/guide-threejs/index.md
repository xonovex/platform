---
name: threejs-guidelines
description: Trigger on `.js/.ts` files with Three.js imports, `three` package usage, or 3D scene code. Use when building vanilla Three.js applications. Apply for scene setup, meshes, materials, lighting, loaders, animation, interaction, shaders, post-processing, performance optimization. Keywords: Three.js, WebGL, WebGPU, Scene, Renderer, Mesh, BufferGeometry, MeshStandardMaterial, GLTFLoader, Raycaster, ShaderMaterial, EffectComposer, InstancedMesh.
---

# Three.js Coding Guidelines

Vanilla Three.js for scene setup, rendering, meshes, materials, lights, animation, interaction, and effects.

## Essentials

- **Scene structure** - Organize Object3D hierarchy; use Groups for logical organization, see [reference/scene-fundamentals.md](reference/scene-fundamentals.md)
- **Render loop** - Frame-rate independence with `clock.getDelta()`, see [reference/scene-fundamentals.md](reference/scene-fundamentals.md)
- **Disposal** - Dispose geometries, materials, textures to prevent memory leaks, see [reference/scene-fundamentals.md](reference/scene-fundamentals.md)
- **Pixel ratio** - Cap at 2 with `Math.min(devicePixelRatio, 2)`, see [reference/scene-fundamentals.md](reference/scene-fundamentals.md)
- **Color space** - Use `SRGBColorSpace` for colors, `NoColorSpace` for data, see [reference/textures.md](reference/textures.md)
- **Shadows** - Enable on renderer, light, mesh; keep frustum tight, see [reference/lighting-shadows.md](reference/lighting-shadows.md)

## Core Topics

- Read [reference/scene-fundamentals.md](reference/scene-fundamentals.md) - Scene, Renderer, Object3D, cleanup, render loop
- Read [reference/geometry.md](reference/geometry.md) - Shapes, BufferGeometry, custom geometry, instancing
- Read [reference/cameras-controls.md](reference/cameras-controls.md) - Cameras, OrbitControls, FlyControls, viewport
- Read [reference/materials.md](reference/materials.md) - PBR materials, ShaderMaterial, properties
- Read [reference/textures.md](reference/textures.md) - Loading, UV mapping, render targets, environment
- Read [reference/lighting-shadows.md](reference/lighting-shadows.md) - Lights, shadows, IBL, light probes
- Read [reference/animation.md](reference/animation.md) - Keyframes, skeletal, morph targets, AnimationMixer
- Read [reference/interaction.md](reference/interaction.md) - Raycasting, selection, drag, coordinate conversion
- Read [reference/loaders.md](reference/loaders.md) - GLTF, FBX, textures, HDR, compression
- Read [reference/shaders.md](reference/shaders.md) - GLSL shaders, uniforms, varyings
- Read [reference/postprocessing.md](reference/postprocessing.md) - Bloom, DOF, SSAO, custom effects
- Read [reference/performance.md](reference/performance.md) - InstancedMesh, LOD, culling, batching, profiling
- Read [reference/patterns.md](reference/patterns.md) - Architecture patterns, asset management, state machines
- Read [reference/math.md](reference/math.md) - Vector3, Matrix4, Quaternion, Box3, curves, MathUtils
- Read [reference/node-materials.md](reference/node-materials.md) - TSL (Three Shading Language), node-based materials
- Read [reference/physics-vr.md](reference/physics-vr.md) - Physics engines (Rapier, Cannon), WebXR (VR/AR)
- Read [reference/webgpu.md](reference/webgpu.md) - WebGPU renderer, compute shaders, modern GPU
