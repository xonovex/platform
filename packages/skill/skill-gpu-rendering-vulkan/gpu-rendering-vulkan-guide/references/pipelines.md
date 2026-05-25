# pipelines: VkPipeline, Pipeline Cache, and Dynamic Rendering

**Guideline:** Build immutable `VkPipeline` objects (graphics/compute) at load, backed by a `VkPipelineCache` persisted to disk so no compile happens on the hot path; prefer dynamic rendering (`VK_KHR_dynamic_rendering`) over `VkRenderPass`/`VkFramebuffer` objects; and declare frequently-changing state as dynamic (`VkPipelineDynamicStateCreateInfo`) so one pipeline covers many viewport/scissor/etc. values.

**Rationale:** A graphics `VkPipeline` bakes the entire state — `VkShaderModule`s, vertex input, input assembly, rasterization, blend, depth/stencil, and (with dynamic rendering) attachment formats — into one immutable object the driver fully specializes ahead of time. First use of an uncompiled pipeline compiles synchronously and hitches the frame, so pipelines are built at load and looked up by a state key; a `VkPipelineCache` deduplicates compiles within a run and, serialized to disk, skips them across runs. Dynamic rendering removes the `VkRenderPass`/`VkFramebuffer`/subpass boilerplate: a pipeline declares attachment formats via `VkPipelineRenderingCreateInfo` and rendering begins with `vkCmdBeginRendering`. Declaring viewport/scissor (and more) as dynamic state lets one pipeline serve many values without a recompile. The agnostic rationale (precompile + cache, no partial state change) is in gpu-rendering-guide (binding-model / shader-system).

**Techniques:**

- **Pipeline cache** - Create one `VkPipelineCache`, pass it to every `vkCreateGraphicsPipelines`/`vkCreateComputePipelines`; `vkGetPipelineCacheData` to a file on shutdown, reload it at startup.
- **Build at load** - Enumerate the (shader + state) keys your materials need and compile them before first draw; never compile on the draw path.
- **Dynamic rendering** - Set `VkPipelineRenderingCreateInfo{ colorAttachmentCount, pColorAttachmentFormats, depthAttachmentFormat }` in `pNext`; render with `vkCmdBeginRendering`/`vkCmdEndRendering` and `VkRenderingAttachmentInfo` — no `VkRenderPass`.
- **Dynamic state** - List `VK_DYNAMIC_STATE_VIEWPORT`/`SCISSOR` (and extended dynamic state where enabled) so the pipeline omits those from its key; set them per command buffer with `vkCmdSetViewport`/`vkCmdSetScissor`.
- **Compute pipeline** - `vkCreateComputePipelines` from a single compute `VkShaderModule` + layout; dispatch with `vkCmdDispatch`.

**Example:**

```c
VkPipelineRenderingCreateInfo rci = {  // dynamic rendering: declare formats, no VkRenderPass
    .sType = VK_STRUCTURE_TYPE_PIPELINE_RENDERING_CREATE_INFO,
    .colorAttachmentCount = 1, .pColorAttachmentFormats = (VkFormat[]){VK_FORMAT_R16G16B16A16_SFLOAT},
    .depthAttachmentFormat = VK_FORMAT_D32_SFLOAT,
};
VkDynamicState dyn[] = {VK_DYNAMIC_STATE_VIEWPORT, VK_DYNAMIC_STATE_SCISSOR};
VkPipelineDynamicStateCreateInfo dsi = {.sType = VK_STRUCTURE_TYPE_PIPELINE_DYNAMIC_STATE_CREATE_INFO,
    .dynamicStateCount = 2, .pDynamicStates = dyn};
VkGraphicsPipelineCreateInfo gpci = {.sType = VK_STRUCTURE_TYPE_GRAPHICS_PIPELINE_CREATE_INFO,
    .pNext = &rci, .pDynamicState = &dsi, .layout = layout, /* stages, vertex input, blend, depth */ };
VkPipeline pso; vkCreateGraphicsPipelines(dev, pipeline_cache, 1, &gpci, NULL, &pso); // cache-backed

// At draw time: bind, set dynamic state, render between begin/end.
VkRenderingAttachmentInfo color = {.sType = VK_STRUCTURE_TYPE_RENDERING_ATTACHMENT_INFO,
    .imageView = hdr_view, .imageLayout = VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL,
    .loadOp = VK_ATTACHMENT_LOAD_OP_CLEAR, .storeOp = VK_ATTACHMENT_STORE_OP_STORE};
VkRenderingInfo ri = {.sType = VK_STRUCTURE_TYPE_RENDERING_INFO, .renderArea = area,
    .layerCount = 1, .colorAttachmentCount = 1, .pColorAttachments = &color};
vkCmdBeginRendering(cmd, &ri);
vkCmdBindPipeline(cmd, VK_PIPELINE_BIND_POINT_GRAPHICS, pso);
vkCmdSetViewport(cmd, 0, 1, &viewport); vkCmdSetScissor(cmd, 0, 1, &scissor);
// ... draws ...
vkCmdEndRendering(cmd);
```

**Gotchas:**

- First use of an uncompiled `VkPipeline` compiles synchronously and stalls; build at load with a warm `VkPipelineCache`, and persist its blob across runs.
- Changing any baked state (a blend mode, an attachment format) needs a different `VkPipeline` — there is no partial change; plan the permutation set, see [references/descriptors.md](./descriptors.md).
- A pipeline's attachment formats (render pass or `VkPipelineRenderingCreateInfo`) must match the actual attachments at draw time, or rendering is invalid.
- Forgetting to `vkCmdSetViewport`/`vkCmdSetScissor` for a dynamic-state pipeline leaves them undefined — set every dynamic state you declared.
- The serialized pipeline-cache blob is driver/device-specific; validate its header (`VkPipelineCacheHeaderVersionOne`) before trusting it, and rebuild on mismatch.

**Related:** [references/descriptors.md](./descriptors.md), [references/resources-and-barriers.md](./resources-and-barriers.md), [references/commands-and-swapchain.md](./commands-and-swapchain.md)
