---
name: ecs-guide
description: "Use when designing or implementing a data-oriented Entity-Component-System: archetype/bitmask storage, contiguous per-type component arrays, filter-and-batch systems, change tracking, and syncing ECS state into stateful external systems (renderer, physics). Triggers on entities, components, archetypes, systems, world iteration, component bitmasks, mirroring transforms to physics, even when the user doesn't say 'ECS'."
---

# Entity-Component-System Guidelines (Data-Oriented)

Architecture for a data-oriented ECS: how entities and components are stored, how systems iterate them, and how to bridge the ECS to stateful external systems (renderer, physics) without losing parallelism or cache locality. For the underlying cache/layout reasoning see **data-oriented-design-guide**; for the renderer the ECS feeds, see **gpu-rendering-guide**; for the allocators behind component storage see **memory-management-guide**.

## Requirements

- A component is plain data (POD-ish struct); behavior lives in systems, not components.
- Storage groups entities by their set of components, so a system can walk matching components linearly.

## Essentials

- **Entity type = component bitmask** - Group entities of identical component sets contiguously; iterate by type, see [references/storage-and-archetypes.md](references/storage-and-archetypes.md)
- **One component instance per type** - Model "many of a thing" with child entities or a list-component, not duplicate components, see [references/single-vs-multiple-components.md](references/single-vs-multiple-components.md)
- **Systems are filter-and-batch loops** - Select types matching a bitmask, walk co-located arrays in a tight loop, go wide across cores, see [references/systems-and-iteration.md](references/systems-and-iteration.md)
- **Default to brute-force sync** - Re-push matching entities each frame; add dirty flags / change lists only when profiling proves it, see [references/change-tracking-and-sync.md](references/change-tracking-and-sync.md)
- **Never propagate via callbacks** - Observer callbacks bypass declared read/write deps and break the parallel scheduler; reserve them for lifecycle, see [references/change-tracking-and-sync.md](references/change-tracking-and-sync.md)
- **Feed renderers through interfaces + visibility bitmasks** - Components expose render/shader interfaces; cull once per viewer into a bitmask, see [references/rendering-integration.md](references/rendering-integration.md)

## Gotchas

- Mirroring ECS state to an external system in a batch pass leaves a one-frame out-of-sync window: a 100 km/h car moves a full meter at 30 Hz, so a raycast fired before the mirror pass misses it. Order dependent passes accordingly.
- Adding or removing a component moves the entity's data to a different type bucket — cheap per entity, but doing it every frame (e.g. a "Changing" tag component) churns memory and shrinks batches.
- Cross-component lookups by id in a hot loop defeat the whole point; co-locate the data instead so the match needs no indirection.
- The render graph can only be extended _before_ execution; extra viewers (shadows, reflections) are generated _during_ execution and fold into the viewer array — don't try to register them up front.

## Example

```c
// A system selects entity types whose bitmask includes (transform, velocity),
// then walks the co-located component arrays with no per-entity lookup.
void velocity_system(tm_transform_t *td, const tm_velocity_t *vd, uint32_t n, float dt) {
    while (n--) {
        td->pos = vec3_mul_add(td->pos, vd->vel, dt);
        ++td, ++vd;
    }
}
```

## Progressive Disclosure

- Read [references/storage-and-archetypes.md](references/storage-and-archetypes.md) - Load when deciding how to lay out entities/components in memory or how to group them for iteration
- Read [references/single-vs-multiple-components.md](references/single-vs-multiple-components.md) - Load when tempted to attach two of the same component, or representing collections (multiple lights, etc.)
- Read [references/systems-and-iteration.md](references/systems-and-iteration.md) - Load when writing the system loop, filtering by components, or parallelizing across cores
- Read [references/change-tracking-and-sync.md](references/change-tracking-and-sync.md) - Load when an external system (physics, renderer, audio) must mirror ECS state and you need to know what changed
- Read [references/rendering-integration.md](references/rendering-integration.md) - Load when feeding the renderer from ECS data: culling, viewers, visibility bitmasks, render-graph injection
