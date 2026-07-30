---
name: c99-game-opinionated-guide
description: "Use when editing C99 game-engine or runtime code in projects that follow the opinionated caller-owns-memory, SoA, builder-pattern style. A focused overlay that covers only game/engine house-style decisions, not generic C99 idioms. Triggers on `.c`/`.h` files in game/engine directories and on prompts about vectors, matrices, quaternions, physics, meshes, spatial structures, tagged unions, inverse mass, builder patterns, even when the user doesn't say 'opinionated'."
---

# C99 Game Engine Opinionated Guidelines

## Requirements

- **Overlay on c99-guide** - This guide carries only the game/engine opinionated decisions; for generic C99 idioms: `const`-correctness, designated initializers (ZII), fixed-width types, value-oriented APIs, and baseline error/return patterns, follow **c99-guide**

## Math

- **Types** - 16-byte aligned vectors/matrices/quaternions, see [references/math-types.md](references/math-types.md)
- **Access** - Use accessor functions for matrices, direct fields for vectors, see [references/math-types.md](references/math-types.md)
- **Coordinates** - Right-handed Y-up, CCW winding, configurable clip depth, see [references/coordinate-system.md](references/coordinate-system.md)
- **Suffixes** - `_aos/_soa` layouts, `_simde` SIMD, `2d/3d` dimensions

## Geometry Pipeline

- **Analytic → Discrete → Packing** - Separate logic/rendering/GPU layers, see [references/geometry-pipeline.md](references/geometry-pipeline.md)
- **Builder pattern** - `*_req()` query size, `*_build()` write to caller buffer, see [references/builder-pattern.md](references/builder-pattern.md)

## Patterns

- **Inverse mass** - Store `1/mass`, static objects use `0.0`, see [references/physics-patterns.md](references/physics-patterns.md)
- **SoA optimization** - AoS for single objects, SoA for batches, see [references/mesh-types.md](references/mesh-types.md)
- **Validation** - Check capacity, bounds, NULL, overflow before operations

## Gotchas

- Vertex packing order matters for GPU upload: pack tightly and match the shader's attribute layout, not the C struct's natural padding
- Quaternion math is sensitive to normalization drift: re-normalize after long chains of multiplications
- Tagged unions with a sentinel `TYPE_INVALID = 0` save initialization bugs; designated initializers default fields to zero
- Builder patterns in C99 work via opaque structs + functions; never expose mutable struct fields across the public boundary

## Progressive disclosure

- Read [references/math-types.md](references/math-types.md) - Load when working with vectors, matrices, or quaternions
- Read [references/coordinate-system.md](references/coordinate-system.md) - Load when setting up camera, projection, or mesh normals
- Read [references/physics-patterns.md](references/physics-patterns.md) - Load when implementing rigid bodies, collision, or constraints
- Read [references/geometry-pipeline.md](references/geometry-pipeline.md) - Load when converting analytic shapes to renderable meshes
- Read [references/builder-pattern.md](references/builder-pattern.md) - Load when generating meshes or other variable-size data
- Read [references/vertex-packing.md](references/vertex-packing.md) - Load when preparing vertex data for GPU upload
- Read [references/mesh-types.md](references/mesh-types.md) - Load when working with 2D or 3D mesh structures
- Read [references/spatial-structures.md](references/spatial-structures.md) - Load when implementing broad-phase collision or spatial queries
