# storage-and-archetypes: Group Entities by Component Set

## Guideline

Represent an entity's type as the bitmask of components it has, and store all entities sharing a type contiguously so a system can walk the matching component arrays linearly with no per-entity lookup.

## Rationale

Systems process components in bulk. If the components a system needs are co-located in memory, iteration is a tight pointer walk with predictable prefetch and no indirection. If they are scattered (one allocation per entity, or a map keyed by id), every step is a cache miss and the transform is dominated by data movement, not arithmetic. Bitmask typing also makes "which entities does this system touch?" a cheap mask test instead of a per-entity branch.

## How to Apply

1. Assign each component type a bit. An entity's _type_ is the OR of its component bits.
2. Bucket entities by type; within a bucket, store each component in its own contiguous array (SoA), parallel-indexed.
3. A system declares a required mask; it runs over every bucket whose type is a superset of that mask, walking the co-located arrays.
4. Adding/removing a component changes the entity's type — move its data to the new bucket. Treat this as an occasional structural edit, not a per-frame operation.
5. Keep components plain data; put behavior in systems so the same layout serves any number of systems.

## Example

```c
// Bad: entity owns scattered, individually-allocated components; systems chase pointers
struct entity { transform_t *t; velocity_t *v; mesh_t *m; /* each malloc'd */ };
for (uint32_t i = 0; i < n; ++i)
    entities[i]->t->pos = add(entities[i]->t->pos, scale(entities[i]->v->vel, dt)); // miss per deref

// Good: per-type contiguous arrays, walked in lockstep
typedef struct { transform_t *transform; velocity_t *velocity; uint32_t count; } movers_t;
for (uint32_t i = 0; i < m.count; ++i)
    m.transform[i].pos = add(m.transform[i].pos, scale(m.velocity[i].vel, dt));
```

## Counter-Example

A tiny world (a handful of entities, e.g. UI singletons or top-level managers) gains nothing from archetype bucketing — a struct of fields is clearer. Archetype storage pays off when "where there is one, there are many."

## Related

[systems-and-iteration.md](./systems-and-iteration.md), [single-vs-multiple-components.md](./single-vs-multiple-components.md)
