# change-tracking-and-sync: Mirroring ECS State to External Systems

## Guideline

When a stateful external system (physics, renderer, audio) must mirror ECS data, default to re-pushing the matching entities every frame; add change-tracking machinery only where profiling shows the external call dominates. Never propagate changes through observer callbacks in parallel systems.

## Rationale

The ECS stores raw data optimized for batch iteration; it does not inherently know "what changed." The external system, however, needs creation, modification, and deletion told to it explicitly. Brute force (a normal filter-and-batch system that pushes every matching entity) is the simplest correct approach and is fast enough in most cases because the push itself is cache-friendly. Cleverer schemes trade simplicity for speed and each has a sharp edge — most dangerously, callbacks, which fire outside the declared read/write dependencies and therefore break the parallel scheduler (and can recurse).

## How to Apply

1. Start with **brute force**: a system that runs over all matching entities and pushes their state to the external system (`setGlobalPose(...)`).
2. If only a few of many change per frame, add per-component **dirty/version flags**, or a hierarchical tree of version flags for O(log N) skipping of unchanged subtrees.
3. Need polling without push? Maintain **change lists** per component/type that systems drain — this preserves automatic parallelization via declared read/write locks.
4. For lifecycle only (creation/teardown), callbacks are acceptable; for general propagation they are not.
5. Sequence dependent passes so the mirror runs _before_ anything that raycasts/queries the external system this frame.

## Five strategies, ranked by when to reach for them

- **Brute force (+ dirty flags)** — default. Simple, parallel-safe, cache-friendly.
- **Change lists** — when many sequential mirror passes benefit from polling instead of re-scanning.
- **"Changing" tag component** — add a temporary tag on mutation so a normal filtered system processes only changed entities; remove the tag after N idle frames. Costs entity-type proliferation and memory rearrangement; apply selectively (e.g. only Transform).
- **Orchestration** — a higher-level system explicitly notifies the external one. Only when special knowledge meaningfully simplifies the case; doesn't fit generic plugin engines.
- **Callbacks/observer** — lifecycle hooks only. Breaks parallelism and can recurse; avoid for propagation.

## Example

```c
// Good: brute-force mirror as an ordinary system; runs before dependent queries this frame
void mirror_kinematics_system(tm_kinematic_actor_t *kd, const tm_transform_t *td, uint32_t n) {
    while (n--) { kd->actor->setGlobalPose(to_physx(td)); ++kd, ++td; }
}
// Bad: on_transform_changed callback pushes to physics outside declared deps -> scheduler race + recursion
register_callback(transform_type, on_changed); // fires mid-frame, bypasses read/write locks
```

## Counter-Example

A turn-based or event-sparse game where almost nothing moves per frame may legitimately skip brute force in favor of change lists or orchestration from the start — there the per-frame scan, not the external call, is the waste.

## Gotcha

Batch mirroring opens a one-frame out-of-sync window. A 100 km/h car moves ~1 m per 30 Hz frame; a raycast fired between the ECS mutation and the mirror pass can miss it entirely. Order the passes, or query the ECS-side value where exactness matters.

## Related

[systems-and-iteration.md](./systems-and-iteration.md), [rendering-integration.md](./rendering-integration.md)
