# pipelines-and-descriptors: Pipeline State Objects and Descriptor Binding

**Guideline:** Bake render state into immutable pipeline-state objects that are precompiled and cached so no compile happens on the hot path; describe shader resource access with descriptor-set layouts organized by update frequency; prefer large bindless descriptor arrays indexed by handle for material/texture access; and route small, frequently-changing per-draw data through push constants instead of descriptor updates.

**Rationale:** On an explicit API the full graphics state — shaders, blend, depth/stencil, rasterizer, vertex input, attachment formats — is frozen into one PSO at creation. This lets the driver compile a fully specialized GPU program ahead of time, but it means the first use of an uncompiled PSO stalls the frame, so PSOs must be built and cached during load, not on demand. Descriptors are how shaders reach buffers and images; grouping them into sets by how often they change (per-frame, per-material, per-draw) means a frame rebinds only the sets that actually changed, minimizing churn. The traditional bind-a-set-per-draw model becomes a bottleneck with many materials; a bindless model declares one huge descriptor array (update-after-bind), uploads texture/buffer handles once, and lets the shader index it by an integer passed per draw — collapsing thousands of binds into one. Push constants carry a tiny amount of data (a transform index, a few floats) directly in the command stream with no descriptor update at all, which is the cheapest path for the data that changes every single draw.

**Techniques:**

- **Precompiled PSO** - Build pipelines at load from the (shader + state) key; persist the driver's pipeline-cache blob to disk so cold start is fast, see [references/shader-system.md](./shader-system.md).
- **PSO cache** - Hash the full state into a key; dedupe identical pipelines; warm likely combinations before they are first drawn.
- **Set frequency** - Layout set 0 = per-frame (camera, lights), set 1 = per-material (textures, params), set 2 = per-draw (object data). Rebind only what changed.
- **Bindless / update-after-bind** - One large descriptor array of all textures/buffers; the shader indexes by handle; add resources without rebinding. Requires the descriptor-indexing feature.
- **Push constants** - A small block (often ≤128 bytes guaranteed) for per-draw scalars/indices; updated inline in the command buffer, no descriptor write.
- **Dynamic offsets** - One descriptor for a ring buffer; supply a per-draw byte offset to address the right sub-range without a new descriptor.

**Example:**

```c
// Precompile and cache the PSO at load; never compile on the draw path.
pso_key key = { .vs = vs, .fs = fs, .blend = BLEND_OPAQUE, .depth = DEPTH_LESS,
                .color_fmt = RGBA16F, .depth_fmt = D32 };
VkPipeline pso = pso_cache_get_or_build(cache, &key);   // first build at load; hits thereafter

// Descriptor sets organized by update frequency: bind heavy sets rarely.
vkCmdBindDescriptorSets(cmd, ..., 0, 1, &set_per_frame, 0, NULL);    // once per frame
for (material m : materials) {
    vkCmdBindDescriptorSets(cmd, ..., 1, 1, &m.set, 0, NULL);        // once per material
    for (draw d : m.draws) {
        // Per-draw data via push constant -> no descriptor update, cheapest path.
        uint32_t pc = (uint32_t)d.transform_index;
        vkCmdPushConstants(cmd, layout, STAGE_VERTEX, 0, sizeof pc, &pc);
        vkCmdDrawIndexed(cmd, d.index_count, 1, d.first_index, 0, 0);
    }
}

// Bindless: bind one giant array once; the shader indexes it by handle.
// layout(set=0,binding=0) uniform sampler2D textures[];  texture(textures[pc.tex_id], uv)
```

**Gotchas:**

- First use of an uncompiled PSO compiles synchronously and hitches; build and warm the cache at load, and persist the driver cache blob across runs.
- Changing any baked state (a blend mode, an attachment format) needs a different PSO — there is no partial state change; plan the PSO permutation set, see [references/shader-system.md](./shader-system.md).
- Bindless needs the descriptor-indexing/update-after-bind capability and non-uniform indexing qualifiers in the shader; a divergent index without the qualifier is undefined.
- Push-constant space is small and shared across stages; overflowing the guaranteed minimum is not portable — keep it to indices and a few scalars.
- Descriptor set/pool sizing is fixed at pool creation; underestimating exhausts the pool mid-frame — size for the worst case or use a growable pool strategy.
- A descriptor written while the GPU may still read it (without update-after-bind) is a data race — gate updates on the frame fence, see [references/synchronization.md](./synchronization.md).

**Related:** [references/shader-system.md](./shader-system.md), [references/synchronization.md](./synchronization.md), [references/command-buffers-and-frames.md](./command-buffers-and-frames.md), [references/device-memory.md](./device-memory.md)
