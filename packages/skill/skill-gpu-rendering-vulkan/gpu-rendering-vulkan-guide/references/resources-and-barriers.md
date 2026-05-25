# resources-and-barriers: Images, Buffers, Pipeline Barriers, and Layout Transitions

**Guideline:** Create `VkImage`/`VkBuffer` with the right usage flags and views (`VkImageView`/`VkBufferView`), track each image's current `VkImageLayout`, and synchronize every write→read and read→write with `VkImageMemoryBarrier2`/`VkBufferMemoryBarrier2` carrying scoped `srcStageMask`/`srcAccessMask` and `dstStageMask`/`dstAccessMask`, the layout transition, and any queue-family ownership transfer.

**Rationale:** Vulkan images have a _layout_ (an opaque, hardware-specific memory arrangement) that must match the next use — sampling needs `SHADER_READ_ONLY_OPTIMAL`, rendering needs `COLOR_ATTACHMENT_OPTIMAL`, a copy destination needs `TRANSFER_DST_OPTIMAL`, presenting needs `PRESENT_SRC_KHR` — and the only way to change it (and to make a prior write visible to a later read) is a pipeline barrier. The barrier's masks define the dependency: `src` is the producer stages/access that must complete and be made available; `dst` is the consumer stages/access that wait and to which it is made visible. Synchronization2 (`VkImageMemoryBarrier2`) folds the stage masks into the barrier struct and supersedes the legacy `vkCmdPipelineBarrier`. The agnostic model (why scopes must be tight, the producer/consumer framing) is in gpu-rendering-guide (synchronization).

**Techniques:**

- **Usage flags** - Set `VkImageUsageFlags`/`VkBufferUsageFlags` for every use (e.g. `COLOR_ATTACHMENT_BIT | SAMPLED_BIT`, `TRANSFER_DST_BIT`); a use without its flag is invalid.
- **Views** - Create a `VkImageView` per (format, aspect, mip/layer range) the shader/attachment needs; buffers use `VkBufferView` only for texel buffers.
- **Image barrier** - `VkImageMemoryBarrier2{ srcStageMask, srcAccessMask, dstStageMask, dstAccessMask, oldLayout, newLayout, subresourceRange, image }`, submitted via `vkCmdPipelineBarrier2(cmd, &VkDependencyInfo{...})`.
- **Buffer barrier** - `VkBufferMemoryBarrier2` for buffer write→read hazards (no layout); often a global `VkMemoryBarrier2` is simpler for broad buffer dependencies.
- **Layout tracking** - Keep the current layout per image (per subresource for partial transitions); `oldLayout` must match it, or pass `UNDEFINED` to discard contents.
- **Queue-family transfer** - For a resource moving between queue families, emit a release barrier on the source (set `srcQueueFamilyIndex`/`dstQueueFamilyIndex`) and a matching acquire barrier on the destination, see [references/device-and-queues.md](./device-and-queues.md).

**Example:**

```c
// Color-attachment write -> sampled read, synchronization2 style.
VkImageMemoryBarrier2 b = {
    .sType         = VK_STRUCTURE_TYPE_IMAGE_MEMORY_BARRIER_2,
    .srcStageMask  = VK_PIPELINE_STAGE_2_COLOR_ATTACHMENT_OUTPUT_BIT,
    .srcAccessMask = VK_ACCESS_2_COLOR_ATTACHMENT_WRITE_BIT,   // make the write available
    .dstStageMask  = VK_PIPELINE_STAGE_2_FRAGMENT_SHADER_BIT,
    .dstAccessMask = VK_ACCESS_2_SHADER_SAMPLED_READ_BIT,      // visible to the sampler
    .oldLayout     = VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL,
    .newLayout     = VK_IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL, // required for sampling
    .srcQueueFamilyIndex = VK_QUEUE_FAMILY_IGNORED,
    .dstQueueFamilyIndex = VK_QUEUE_FAMILY_IGNORED,
    .image = gbuffer,
    .subresourceRange = {VK_IMAGE_ASPECT_COLOR_BIT, 0, 1, 0, 1},
};
VkDependencyInfo dep = {.sType = VK_STRUCTURE_TYPE_DEPENDENCY_INFO,
                        .imageMemoryBarrierCount = 1, .pImageMemoryBarriers = &b};
vkCmdPipelineBarrier2(cmd, &dep);
```

**Gotchas:**

- `oldLayout` must equal the image's actual current layout (or be `UNDEFINED` to discard); a mismatch is undefined behavior the validation layers usually catch.
- `ALL_COMMANDS`/`ALL_COMMANDS` with full access masks is correct but serializes the GPU — scope masks to the real producer/consumer.
- A missing barrier between a write and a dependent read is a hazard that may pass on your GPU and corrupt elsewhere — run synchronization validation.
- `VK_ACCESS_2_SHADER_READ_BIT` is broad; prefer the specific `SHADER_SAMPLED_READ`/`SHADER_STORAGE_READ` bits where available.
- A queue-family release without a matching acquire (or vice versa) corrupts the resource on one queue — the pair must match exactly.

**Related:** [references/device-memory.md](./device-memory.md), [references/synchronization.md](./synchronization.md), [references/device-and-queues.md](./device-and-queues.md), [references/pipelines.md](./pipelines.md)
