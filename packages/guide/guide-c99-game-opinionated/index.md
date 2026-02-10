---
name: c99-game-opinionated-guidelines
description: Trigger on `.c/.h` files in game/engine directories. Opinionated game engine patterns: caller-owns-memory, inverse mass, SoA optimization, builder pattern. Keywords: vectors, matrices, quaternions, physics, meshes, spatial structures, tagged unions, inverse mass.
---

# C99 Game Engine Opinionated Guidelines

## Requirements

- C99; extends `c99-guidelines` skill.

## Math

- **Types** - 16-byte aligned vectors/matrices/quaternions, see [reference/math-types.md](reference/math-types.md)
- **Access** - Use accessor functions for matrices, direct fields for vectors, see [reference/math-types.md](reference/math-types.md)
- **Coordinates** - Right-handed Y-up, CCW winding, configurable clip depth, see [reference/coordinate-system.md](reference/coordinate-system.md)
- **Suffixes** - `_aos/_soa` layouts, `_simde` SIMD, `2d/3d` dimensions

## Geometry Pipeline

- **Analytic → Discrete → Packing** - Separate logic/rendering/GPU layers, see [reference/geometry-pipeline.md](reference/geometry-pipeline.md)
- **Builder pattern** - `*_req()` query size, `*_build()` write to caller buffer, see [reference/builder-pattern.md](reference/builder-pattern.md)

## Patterns

- **Caller-owns-memory** - Libraries never allocate, app provides arrays
- **Tagged unions** - Type enum + union for polymorphism, see [reference/tagged-unions.md](reference/tagged-unions.md)
- **Inverse mass** - Store `1/mass`, static objects use `0.0`, see [reference/physics-patterns.md](reference/physics-patterns.md)
- **SoA optimization** - AoS for single objects, SoA for batches, see [reference/mesh-types.md](reference/mesh-types.md)
- **Validation** - Check capacity, bounds, NULL, overflow before operations

## Progressive disclosure

- Read [reference/math-types.md](reference/math-types.md) - When working with vectors, matrices, or quaternions
- Read [reference/coordinate-system.md](reference/coordinate-system.md) - When setting up camera, projection, or mesh normals
- Read [reference/tagged-unions.md](reference/tagged-unions.md) - When implementing shape or entity polymorphism
- Read [reference/physics-patterns.md](reference/physics-patterns.md) - When implementing rigid bodies, collision, or constraints
- Read [reference/geometry-pipeline.md](reference/geometry-pipeline.md) - When converting analytic shapes to renderable meshes
- Read [reference/builder-pattern.md](reference/builder-pattern.md) - When generating meshes or other variable-size data
- Read [reference/vertex-packing.md](reference/vertex-packing.md) - When preparing vertex data for GPU upload
- Read [reference/mesh-types.md](reference/mesh-types.md) - When working with 2D or 3D mesh structures
- Read [reference/spatial-structures.md](reference/spatial-structures.md) - When implementing broad-phase collision or spatial queries
