# systems-and-iteration: Filter-and-Batch System Loops

## Guideline

A system is a filter (a required component mask) plus a batch loop over the matched, co-located component arrays. Keep the loop branch-light and free of per-element lookups, and split the work across worker cores.

## Rationale

The whole archetype layout exists so this loop can be a tight, prefetch-friendly walk. Per-element id lookups, virtual dispatch, or branching on rare conditions inside the loop reintroduce the misses the layout was meant to avoid. Because each bucket is independent and components are plain data, the loop parallelizes by simply partitioning the array range across jobs.

## How to Apply

1. Express the system's input as a required component mask; iterate every bucket whose type is a superset.
2. Pass the matched arrays plus a count into the loop; advance raw pointers, don't index through entity handles.
3. Hoist invariants (timestep, constants) out of the loop; avoid calling into other systems' data mid-loop.
4. Parallelize by splitting the count into ranges, one job per range; keep systems free of shared mutable global state so jobs don't serialize.
5. Declare the components each system reads and writes so a scheduler can run non-conflicting systems concurrently.

## Example

```c
// Good: pointer-walk over co-located arrays, trivially splittable into jobs
void velocity_system(tm_transform_t *td, const tm_velocity_t *vd, uint32_t n, float dt) {
    while (n--) {
        td->pos = vec3_mul_add(td->pos, vd->vel, dt);
        ++td, ++vd;
    }
}
// Bad: per-element lookup + cross-system call inside the hot loop
for (uint32_t i = 0; i < n; ++i) {
    transform_t *t = world_lookup_transform(world, ids[i]); // map miss per element
    physics_apply(world, ids[i]); // serializes against the physics system
}
```

## Counter-Example

A system that genuinely needs random access across the whole world (a global solver, a spatial query) won't fit the linear-walk shape — give it its own acceleration structure rather than forcing it through the component loop.

## Related

[storage-and-archetypes.md](./storage-and-archetypes.md), [change-tracking-and-sync.md](./change-tracking-and-sync.md)
