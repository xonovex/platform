# rendering-integration: Feeding the Renderer from ECS Data

**Guideline:** Let rendering-relevant components implement renderer-facing interfaces, cull once per viewer into a visibility bitmask, and extend the render graph by injecting modules — rather than walking a scene graph or pushing draw calls per object.

**Rationale:** A frame may touch hundreds of thousands of renderable components across several viewers (main camera, shadow cascades, reflections). Iterating that data once, in parallel, per component type — and expressing "which viewers see this object" as bits — keeps the bridge cache-friendly and avoids visiting any renderable more than once per frame (which is also what makes per-camera state mutation safe). Plugin interfaces keep the core renderer ignorant of specific component types, so new renderable kinds need no core changes.

**How to Apply:**

1. Give renderable components a **render interface** (culling callbacks + a render function) and auxiliary data (lights, decals, probes, post volumes) a **shader interface** (`update()` to produce view-dependent data, plus a hook to inject render-graph modules).
2. Run frustum culling **per viewer**; record results as a per-renderable **bitmask** of which viewers see it. With multiple viewers, fill all the bits in a single pass over the component data.
3. Bundle each viewer's sort key, view-dependent data, visibility mask, and camera settings into a "viewer" record passed to render functions.
4. Inject GPU work by adding render-graph **modules** from the shader interface; let the graph derive ordering, barriers, and transitions (the render graph is owned by **gpu-rendering-guide**).
5. Generate extra viewers (shadows, reflections) **during** graph execution; they fold into the viewer array. Do not try to register them before execution.

**Example:**

```c
// Good: cull once per viewer into bits, then render each renderable once for all its viewers
for (uint32_t v = 0; v < num_viewers; ++v)
    cull_into_bits(renderables, n, viewers[v], /*bit=*/v); // single object may set several bits
for (uint32_t i = 0; i < n; ++i)
    renderables[i].render(viewers_seeing(visibility[i]), sort_keys); // visited once; per-viewer mutation safe

// Bad: re-traverse the whole scene once per camera, mutating shared object state each pass -> races + waste
for_each_camera(c) for_each_object(o) { o->lod = pick_lod(o, c); draw(o, c); }
```

**Counter-Example:** A single-camera, low-object-count tool view doesn't need bitmask visibility or multi-viewer machinery — a straight cull-and-draw is clearer. The bitmask model earns its keep with multiple simultaneous viewers.

**Related:** [systems-and-iteration.md](./systems-and-iteration.md), [change-tracking-and-sync.md](./change-tracking-and-sync.md)
