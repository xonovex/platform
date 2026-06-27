# device-memory: VkDeviceMemory, Memory Types, and Staging

## Guideline

Query `VkPhysicalDeviceMemoryProperties`, allocate a few large `VkDeviceMemory` blocks of the right memory type, and sub-allocate every `VkImage`/`VkBuffer` out of them honoring `VkMemoryRequirements.alignment`; upload static data via a `HOST_VISIBLE` staging buffer copied with `vkCmdCopyBuffer`/`vkCmdCopyBufferToImage` into a `DEVICE_LOCAL` resource; persistently map host-visible blocks. Never call `vkAllocateMemory` per resource.

## Rationale

Vulkan exposes memory as types (each a heap + property-flag combination) you select explicitly; the driver does not place resources for you. `vkAllocateMemory` count is capped by `maxMemoryAllocationCount` (often ~4096) and each call is slow, so one allocation per resource fails at scale — sub-allocate from large blocks instead, the same arena/pool principle as memory-management-guide and the strategy described in gpu-rendering-guide (gpu-memory-strategy). `DEVICE_LOCAL` memory is the fast GPU path but usually lacks `HOST_VISIBLE`, so the CPU cannot write it directly; static data is therefore written into a `HOST_VISIBLE | HOST_COHERENT` staging buffer and copied on a queue into the device-local resource. Each `VkImage`/`VkBuffer` reports its own `VkMemoryRequirements` (size, alignment, allowed `memoryTypeBits`); placement must satisfy that alignment, and buffer↔image co-placement must respect `bufferImageGranularity`.

## Techniques

- **Memory type selection** - Scan `VkPhysicalDeviceMemoryProperties.memoryTypes`; require the resource's `memoryTypeBits` and the property flags you want (`DEVICE_LOCAL`; `HOST_VISIBLE | HOST_COHERENT`; `DEVICE_LOCAL | HOST_VISIBLE` for resizable-BAR direct writes).
- **Few large blocks** - `vkAllocateMemory` a 64–256 MB block per type; sub-allocate with a block/free-list allocator so call sites request `(size, usage)`.
- **Alignment** - Use `vkGetBufferMemoryRequirements2` / `vkGetImageMemoryRequirements2`; align the sub-allocation offset to `req.alignment` and keep buffers/images apart by `bufferImageGranularity`.
- **Staging upload** - Map a `HOST_VISIBLE` staging buffer, `memcpy`, `vkCmdCopyBufferToImage` / `vkCmdCopyBuffer` into the `DEVICE_LOCAL` target, then a barrier to the read state, see [references/resources-and-barriers.md](./resources-and-barriers.md).
- **Persistent mapping** - `vkMapMemory` a host-visible block once and keep the pointer; without `HOST_COHERENT`, `vkFlushMappedMemoryRanges` after writes and `vkInvalidateMappedMemoryRanges` before reads.
- **General-purpose allocator** - A reusable Vulkan allocator over these techniques (block management, type selection, defrag) so call sites never touch raw `VkDeviceMemory`.
- **Block strategy by size** - A workable split: device-local requests ≤ a block size (e.g. 256 MB) sub-allocate from buddy-managed 256 MB blocks; larger requests get a dedicated `vkAllocateMemory`; staging uses a linear allocator over `max(size, 256 MB)` blocks recycled once the transfer fence signals.
- **Tagged visual debugging** - Pass a debug tag with every allocation and build a simple occupancy visualization early; buddy power-of-two rounding waste and never-released empty blocks are invisible otherwise (a ~400-line occupancy visualizer built in about a day immediately exposes the waste).

## Example

```c
VkPhysicalDeviceMemoryProperties mp; vkGetPhysicalDeviceMemoryProperties(phys, &mp);

static uint32_t pick_type(const VkPhysicalDeviceMemoryProperties *mp,
                          uint32_t type_bits, VkMemoryPropertyFlags want) {
    for (uint32_t i = 0; i < mp->memoryTypeCount; i++)
        if ((type_bits & (1u << i)) &&
            (mp->memoryTypes[i].propertyFlags & want) == want) return i;
    return UINT32_MAX;
}

VkMemoryRequirements req; vkGetImageMemoryRequirements(dev, image, &req);
gpu_alloc a = block_alloc(allocator, req.size, req.alignment,        // sub-allocate, honor alignment
    pick_type(&mp, req.memoryTypeBits, VK_MEMORY_PROPERTY_DEVICE_LOCAL_BIT));
vkBindImageMemory(dev, image, a.memory, a.offset);

// Static upload: HOST_VISIBLE staging -> DEVICE_LOCAL via a copy on the queue.
void *dst; vkMapMemory(dev, stage.memory, stage.offset, size, 0, &dst);
memcpy(dst, pixels, size);
VkBufferImageCopy region = {.imageSubresource = {VK_IMAGE_ASPECT_COLOR_BIT, 0, 0, 1},
                            .imageExtent = {w, h, 1}};
vkCmdCopyBufferToImage(cmd, stage.buffer, image,
                       VK_IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL, 1, &region);
```

## Gotchas

- `maxMemoryAllocationCount` is small; per-resource `vkAllocateMemory` works in a demo then fails in a real scene — sub-allocate from the start.
- `DEVICE_LOCAL`-only memory has no host pointer; `vkMapMemory` on it fails — write through staging.
- Without `HOST_COHERENT`, a written-but-unflushed range is not visible to the GPU; coherent memory skips flush/invalidate but can be slower for the GPU to read.
- Ignoring `req.alignment` or `bufferImageGranularity` produces a valid bind that corrupts neighboring resources on some hardware.
- A staging buffer reused before its copy's fence signals overwrites in-flight upload data — recycle staging by frame fence, see [references/commands-and-swapchain.md](./commands-and-swapchain.md).

## Related

[references/resources-and-barriers.md](./resources-and-barriers.md), [references/commands-and-swapchain.md](./commands-and-swapchain.md), [references/synchronization.md](./synchronization.md)
