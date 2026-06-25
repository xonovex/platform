---
name: c99-game-opinionated-guide
description: "Use when editing C99 game-engine or runtime code in projects that follow the opinionated caller-owns-memory, SoA, builder-pattern style. Triggers on `.c`/`.h` files in game/engine directories and on prompts about vectors, matrices, quaternions, physics, meshes, spatial structures, tagged unions, inverse mass, builder patterns, even when the user doesn't say 'opinionated'."
---

# C99 Game Engine Opinionated Guidelines

## Requirements

- C99; extends `c99-guide` skill.

## Math

- **Types** - 16-byte aligned vectors/matrices/quaternions, see [references/math-types.md](references/math-types.md)
- **Access** - Use accessor functions for matrices, direct fields for vectors, see [references/math-types.md](references/math-types.md)
- **Coordinates** - Right-handed Y-up, CCW winding, configurable clip depth, see [references/coordinate-system.md](references/coordinate-system.md)
- **Suffixes** - `_aos/_soa` layouts, `_simde` SIMD, `2d/3d` dimensions

## Geometry Pipeline

- **Analytic → Discrete → Packing** - Separate logic/rendering/GPU layers, see [references/geometry-pipeline.md](references/geometry-pipeline.md)
- **Builder pattern** - `*_req()` query size, `*_build()` write to caller buffer, see [references/builder-pattern.md](references/builder-pattern.md)

## Patterns

- **Caller-owns-memory** - Libraries never allocate, app provides arrays
- **Tagged unions** - Type enum + union for polymorphism, see [references/tagged-unions.md](references/tagged-unions.md)
- **Inverse mass** - Store `1/mass`, static objects use `0.0`, see [references/physics-patterns.md](references/physics-patterns.md)
- **SoA optimization** - AoS for single objects, SoA for batches, see [references/mesh-types.md](references/mesh-types.md)
- **Validation** - Check capacity, bounds, NULL, overflow before operations

## Gotchas

- Vertex packing order matters for GPU upload — pack tightly and match the shader's attribute layout, not the C struct's natural padding
- Quaternion math is sensitive to normalization drift — re-normalize after long chains of multiplications
- Tagged unions with a sentinel `TYPE_INVALID = 0` save initialization bugs; designated initializers default fields to zero
- Builder patterns in C99 work via opaque structs + functions; never expose mutable struct fields across the public boundary

## Progressive disclosure

- Read [references/math-types.md](references/math-types.md) - Load when working with vectors, matrices, or quaternions
- Read [references/coordinate-system.md](references/coordinate-system.md) - Load when setting up camera, projection, or mesh normals
- Read [references/tagged-unions.md](references/tagged-unions.md) - Load when implementing shape or entity polymorphism
- Read [references/physics-patterns.md](references/physics-patterns.md) - Load when implementing rigid bodies, collision, or constraints
- Read [references/geometry-pipeline.md](references/geometry-pipeline.md) - Load when converting analytic shapes to renderable meshes
- Read [references/builder-pattern.md](references/builder-pattern.md) - Load when generating meshes or other variable-size data
- Read [references/vertex-packing.md](references/vertex-packing.md) - Load when preparing vertex data for GPU upload
- Read [references/mesh-types.md](references/mesh-types.md) - Load when working with 2D or 3D mesh structures
- Read [references/spatial-structures.md](references/spatial-structures.md) - Load when implementing broad-phase collision or spatial queries
