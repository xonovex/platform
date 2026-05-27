# single-vs-multiple-components: One Component Instance Per Type

**Guideline:** Allow at most one instance of a given component type per entity. Represent "many of a thing" with child entities or a single list-holding component — not by attaching the same component type twice.

**Rationale:** Permitting multiple instances of one type forces every component to be addressed by an (entity, type, instance-id) triplet instead of just (entity, type). That adds per-instance ids, lookup indirection in hot loops, and ambiguity: when two `mass` and three `position` components coexist, which pairs with which? You also lose semantic clarity — you can no longer speak of "_the_ position of an entity." Single-instance typing keeps matched component arrays co-located and reachable without a lookup, which is the property that makes iteration fast and systems loosely coupled.

**How to Apply:**

1. Default to one component per type per entity; let the entity type be a plain component bitmask.
2. Need several of something (multiple lights, multiple colliders)? Model each as a **child entity** holding one component, or store a **list inside a single component**.
3. Need a custom internal layout for that collection? Put an **index** in the public component that points into a structure the owning system controls privately.
4. Reserve genuine multi-instance support for cases with a strong, profiled requirement that the alternatives can't meet.

**Example:**

```c
// Bad: two light components on one entity — which one is "the" light? needs instance ids
add_component(e, light_component);
add_component(e, light_component); // ambiguous, forces (entity,type,id) addressing

// Good: a parent entity with one light-list component, or child entities each with one light
struct lights_component_t { light_t *lights; uint32_t count; }; // list-component
// or: spawn child entities, each carrying a single light_component
```

**Counter-Example:** Frameworks that center on per-instance scripting (e.g. several independent "behavior" scripts on one object) sometimes accept multiple instances deliberately — but they pay the id/lookup/coupling cost described above. Know you are buying that cost.

**Related:** [storage-and-archetypes.md](./storage-and-archetypes.md)
