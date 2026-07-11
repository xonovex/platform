# single-vs-multiple-components: One Component Instance Per Type

## Guideline

Allow at most one instance of a given component type per entity. Represent "many of a thing" with child entities or a single list-holding component — not by attaching the same component type twice.

## Rationale

Multiple instances force addressing by an (entity, type, instance-id) triplet instead of (entity, type): per-instance ids, hot-loop lookup indirection, and ambiguity (two `mass` and three `position` components — which pairs with which?). You also lose "_the_ position of an entity." Single-instance typing keeps matched arrays co-located and lookup-free.

## How to Apply

1. Default to one component per type per entity; let the entity type be a plain component bitmask.
2. Need several of something (multiple lights, multiple colliders)? Model each as a **child entity** holding one component, or store a **list inside a single component**.
3. Need a custom internal layout for that collection? Put an **index** in the public component that points into a structure the owning system controls privately.
4. Reserve genuine multi-instance support for cases with a strong, profiled requirement that the alternatives can't meet.

## Example

```c
// Bad: two light components on one entity — which one is "the" light? needs instance ids
add_component(e, light_component);
add_component(e, light_component); // ambiguous, forces (entity,type,id) addressing

// Good: a parent entity with one light-list component, or child entities each with one light
struct lights_component_t { light_t *lights; uint32_t count; }; // list-component
// or: spawn child entities, each carrying a single light_component
```

## Counter-Example

Frameworks that center on per-instance scripting (e.g. several independent "behavior" scripts on one object) sometimes accept multiple instances deliberately — but they pay the id/lookup/coupling cost described above. Know you are buying that cost.

## Related

[storage-and-archetypes.md](./storage-and-archetypes.md)
