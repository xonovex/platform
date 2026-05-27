# binding-model: Pipeline State and the Descriptor/Binding Model

**Guideline:** Bake render state into immutable pipeline-state objects that are precompiled and cached so no compile happens on the hot path; describe shader resource access with binding groups organized by update frequency; prefer large bindless arrays indexed by handle for material/texture access; and route small, frequently-changing per-draw data through inline constants instead of binding updates.

**Rationale:** On an explicit API the full graphics state — shaders, blend, depth/stencil, rasterizer, vertex input, attachment formats — is frozen into one pipeline object at creation. This lets the driver compile a fully specialized GPU program ahead of time, but it means the first use of an uncompiled pipeline stalls the frame, so pipelines must be built and cached during load, not on demand. Bindings are how shaders reach buffers and images; grouping them by how often they change (per-frame, per-material, per-draw) means a frame rebinds only the groups that actually changed, minimizing churn. (Each API names this differently — descriptor sets, root parameters/descriptor tables, argument buffers, bind groups — but the frequency model is the same.) The traditional bind-a-group-per-draw model becomes a bottleneck with many materials; a bindless model declares one huge resource array, uploads texture/buffer handles once, and lets the shader index it by an integer passed per draw — collapsing thousands of binds into one. Inline constants carry a tiny amount of data (a transform index, a few floats) directly in the command stream with no binding update at all, which is the cheapest path for data that changes every single draw.

**Techniques:**

- **Precompiled pipeline** - Build pipelines at load from the (shader + state) key; persist the driver's pipeline-cache blob to disk so cold start is fast, see [references/shader-system.md](./shader-system.md).
- **Pipeline cache** - Hash the full state into a key; dedupe identical pipelines; warm likely combinations before they are first drawn.
- **Binding frequency** - Group 0 = per-frame (camera, lights), group 1 = per-material (textures, params), group 2 = per-draw (object data). Rebind only what changed.
- **Bindless** - One large resource array of all textures/buffers; the shader indexes by handle; add resources without rebinding. Requires the API's bindless/descriptor-indexing capability.
- **Inline constants** - A small block (tens to ~128 bytes, API-dependent) for per-draw scalars/indices; updated inline in the command stream, no binding write.
- **Dynamic offsets** - One binding for a ring buffer; supply a per-draw byte offset to address the right sub-range without a new binding.
- **Resource binders carry only handles** - Bundle a draw's resources into a binder object (≈ one binding group) and serialize only its small integer handle onto the command stream, not the resource contents. Commands stay self-contained, so streams recorded on worker threads can be translated to the backend in parallel without backtracking to reconstruct binding state.

**Example:**

```c
// Precompile and cache the pipeline at load; never compile on the draw path.
// (Neutral pseudocode; concrete API objects live in the per-API skill.)
pipeline_key key = { .vs = vs, .fs = fs, .blend = BLEND_OPAQUE, .depth = DEPTH_LESS,
                     .color_fmt = RGBA16F, .depth_fmt = DEPTH32F };
pipeline pso = pipeline_cache_get_or_build(cache, &key);   // first build at load; hits thereafter

// Binding groups organized by update frequency: bind heavy groups rarely.
cmd_bind_group(cmd, 0, group_per_frame);                   // once per frame
for (material m : materials) {
    cmd_bind_group(cmd, 1, m.group);                       // once per material
    for (draw d : m.draws) {
        // Per-draw data via inline constant -> no binding update, cheapest path.
        cmd_push_constants(cmd, STAGE_VERTEX, &d.transform_index, sizeof(uint32_t));
        cmd_draw_indexed(cmd, d.index_count, d.first_index);
    }
}

// Bindless: bind one giant array once; the shader indexes it by handle.
// e.g. sampler2D textures[];   sample(textures[pc.tex_id], uv)
```

**Gotchas:**

- First use of an uncompiled pipeline compiles synchronously and hitches; build and warm the cache at load, and persist the driver cache blob across runs.
- Changing any baked state (a blend mode, an attachment format) needs a different pipeline — there is no partial state change; plan the permutation set, see [references/shader-system.md](./shader-system.md).
- Bindless needs the API's descriptor-indexing/bindless capability and non-uniform indexing qualifiers in the shader; a divergent index without the qualifier is undefined.
- Inline-constant space is small and shared across stages; overflowing the guaranteed minimum is not portable — keep it to indices and a few scalars.
- Binding-group pool sizing is fixed at pool creation in some APIs; underestimating exhausts the pool mid-frame — size for the worst case or use a growable strategy.
- A binding written while the GPU may still read it (without update-after-bind) is a data race — gate updates on the frame fence, see [references/synchronization.md](./synchronization.md).

**Related:** [references/shader-system.md](./shader-system.md), [references/synchronization.md](./synchronization.md), [references/command-recording-and-frames.md](./command-recording-and-frames.md), [references/gpu-memory-strategy.md](./gpu-memory-strategy.md)
