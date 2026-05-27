# descriptors: Descriptor Sets, Layouts, Pools, Bindless, and Push Constants

**Guideline:** Describe shader resource access with `VkDescriptorSetLayout`s grouped by update frequency, allocate `VkDescriptorSet`s from a `VkDescriptorPool` sized for the worst case, bind heavy sets rarely; for material/texture access use a large update-after-bind descriptor array (descriptor indexing) indexed by handle; route tiny per-draw data through push constants.

**Rationale:** A `VkDescriptorSetLayout` declares the bindings (type, count, stage flags) a set provides; a `VkDescriptorSet` is concrete storage allocated from a `VkDescriptorPool`. Grouping bindings by how often they change — set 0 per-frame, set 1 per-material, set 2 per-draw — means `vkCmdBindDescriptorSets` rebinds only what changed. The bind-a-set-per-draw model bottlenecks with many materials; `VK_EXT_descriptor_indexing` with `UPDATE_AFTER_BIND` lets one giant descriptor array hold every texture, the shader indexing it by a handle passed per draw (`nonuniformEXT` when divergent) — collapsing thousands of binds into one. Push constants carry a small block (`maxPushConstantsSize`, guaranteed ≥128 bytes) inline in the command buffer with no descriptor write, the cheapest path for a transform index or a few scalars. The architecture rationale (frequency grouping, bindless, inline constants) is in gpu-rendering-guide (binding-model).

**Techniques:**

- **Layout by frequency** - One `VkDescriptorSetLayout` per update rate; bind set 0 (per-frame) once, set 1 (per-material) per material, set 2 (per-draw) or push constants per draw.
- **Pool sizing** - `VkDescriptorPoolCreateInfo.pPoolSizes` must cover the worst-case count per descriptor type per frame-in-flight; undersize and `vkAllocateDescriptorSets` returns `VK_ERROR_OUT_OF_POOL_MEMORY`.
- **Bindless** - A binding with a large `descriptorCount` and `VK_DESCRIPTOR_BINDING_UPDATE_AFTER_BIND_BIT | PARTIALLY_BOUND_BIT | VARIABLE_DESCRIPTOR_COUNT_BIT`; enable `descriptorIndexing`; index in GLSL with `nonuniformEXT(handle)`.
- **Push constants** - `VkPushConstantRange{ stageFlags, offset, size }` in the pipeline layout; `vkCmdPushConstants` per draw. Keep to indices/scalars within the guaranteed minimum.
- **Dynamic offsets** - A `*_DYNAMIC` descriptor for a ring buffer; pass a per-draw byte offset to `vkCmdBindDescriptorSets` to address the right sub-range.
- **One global bindless set** - Instead of many per-pipeline sets, keep a single renderer-wide `VkDescriptorSet` with one large array binding per resource class (storage buffer, sampled image, storage image, sampler, acceleration structure). Allocate an array index per resource at creation from a per-type free list (scan for a free range, else bump a `next` counter) and embed it in the resource handle; shaders index by that integer. Defer freeing a slot until a `VkFence` proves no in-flight command buffer still references it. Removing the per-set bind/lookup is what makes command-buffer building dramatically cheaper (reported in practice as roughly a 6× reduction in build cost).
- **Per-job blueprint copy** - When sets must be mutated mid-frame, keep shared sets as read-only blueprints and give each worker/job its own `VkDescriptorPool` plus a parallel set array; lazily copy a blueprint into the job-local set (via `VkCopyDescriptorSet` / `vkUpdateDescriptorSets`) only on first use, then reset and recycle the whole pool after the fence. Whole-pool recycling avoids the fragmentation that per-set versioning + GC causes.

**Example:**

```c
// Per-draw transform index via push constant -> no descriptor write.
VkPushConstantRange pcr = {.stageFlags = VK_SHADER_STAGE_VERTEX_BIT, .offset = 0, .size = 4};
// ... include pcr in VkPipelineLayoutCreateInfo ...

vkCmdBindDescriptorSets(cmd, VK_PIPELINE_BIND_POINT_GRAPHICS, layout,
                        0, 1, &set_per_frame, 0, NULL);     // set 0 once per frame
for (material *m = materials; m < end; m++) {
    vkCmdBindDescriptorSets(cmd, VK_PIPELINE_BIND_POINT_GRAPHICS, layout,
                            1, 1, &m->set, 0, NULL);         // set 1 per material
    for (draw *d = m->draws; d < d_end; d++) {
        vkCmdPushConstants(cmd, layout, VK_SHADER_STAGE_VERTEX_BIT, 0, 4, &d->transform_index);
        vkCmdDrawIndexed(cmd, d->index_count, 1, d->first_index, 0, 0);
    }
}
// Bindless: layout(set=0,binding=0) uniform sampler2D tex[];
//           texture(tex[nonuniformEXT(pc.tex_id)], uv)
```

**Gotchas:**

- A `VkDescriptorSet` updated with `vkUpdateDescriptorSets` while the GPU may still read it (without `UPDATE_AFTER_BIND`) is a data race — gate updates on the frame `VkFence`, see [references/commands-and-swapchain.md](./commands-and-swapchain.md).
- Pool sizes are fixed at creation; size for worst case × frames-in-flight or use one pool per frame slot reset with `vkResetDescriptorPool`.
- Bindless requires the descriptor-indexing feature enabled at device creation _and_ `nonuniformEXT` for divergent indices — a divergent index without it is undefined.
- Push-constant space is small and shared across stages; exceeding `maxPushConstantsSize` (or overlapping stage ranges) is invalid — keep it tiny.
- The pipeline layout's set layouts and push-constant ranges must match the shader's declared bindings exactly, or binding silently targets the wrong resource.
- A bindless array index freed and immediately reused while a prior frame's command buffer still references the slot samples the wrong resource — release bind points through fence-gated deferred deletion, and reserve a slot of fallback resources so a "null" handle reads something valid.
- The bindless descriptor-array sizes you request may exceed the device's `maxDescriptorSet*` / `maxPerStageDescriptor*` limits; query and clamp (or log) rather than assuming a fixed 512K.

**Related:** [references/pipelines.md](./pipelines.md), [references/resources-and-barriers.md](./resources-and-barriers.md), [references/commands-and-swapchain.md](./commands-and-swapchain.md), [references/device-memory.md](./device-memory.md)
