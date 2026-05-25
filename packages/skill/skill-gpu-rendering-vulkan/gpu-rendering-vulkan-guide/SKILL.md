---
name: gpu-rendering-vulkan-guide
description: "Use when implementing a Vulkan renderer: the concrete Vulkan API for device/queues, device memory + memory types + staging, images/buffers + pipeline barriers and layout transitions, descriptor sets/layouts + bindless, pipelines + pipeline cache + dynamic rendering, timeline semaphores + fences, command pools/buffers, and the swapchain. Triggers on Vk* types and vkCmd*/vkCreate* calls, VkImageMemoryBarrier2, VkDescriptorSet, VkPipeline, VkSemaphore, VkCommandPool, swapchain acquire/present. Skip API-agnostic rendering architecture (use gpu-rendering-guide), other GPU APIs (D3D12/Metal), and high-level web 3D (use threejs-guide)."
---

# Vulkan Rendering Guidelines

The concrete Vulkan API for a low-level renderer. This skill is the Vulkan _how_; for the _why_ and the API-agnostic architecture (render graphs, binding model, sync model, frames-in-flight, memory strategy) see gpu-rendering-guide. For the general allocator principle behind device memory see memory-management-guide.

## Requirements

- Vulkan 1.3 (or 1.2 + the relevant extensions: `VK_KHR_synchronization2`, `VK_KHR_dynamic_rendering`, `VK_EXT_descriptor_indexing`, `VK_KHR_timeline_semaphore`).
- Validation layers + synchronization validation on in development; a frame debugger (RenderDoc-class) for capture.

## Setup

- **Device & queues** - `VkInstance` → physical device → `VkDevice`; select graphics/compute/transfer queue families, see [references/device-and-queues.md](references/device-and-queues.md)
- **Device memory** - `VkPhysicalDeviceMemoryProperties` types/heaps; few `vkAllocateMemory` + sub-allocation; staging, see [references/device-memory.md](references/device-memory.md)

## Resources & sync

- **Images/buffers & barriers** - `VkImage`/`VkBuffer` + views, `VkImageMemoryBarrier2` stage/access masks, layout transitions, see [references/resources-and-barriers.md](references/resources-and-barriers.md)
- **Synchronization** - `VkSemaphore` (binary + timeline), `VkFence`, pipeline stage/access masks, submit-time waits, see [references/synchronization.md](references/synchronization.md)

## Pipelines & binding

- **Descriptors** - `VkDescriptorSetLayout`/`Pool`/`Set`, update-after-bind for bindless, push constants, set frequency, see [references/descriptors.md](references/descriptors.md)
- **Pipelines** - `VkPipeline` (graphics/compute), `VkPipelineCache`, dynamic rendering vs render passes, dynamic state, see [references/pipelines.md](references/pipelines.md)

## Commands & present

- **Commands & swapchain** - `VkCommandPool` per thread per frame, primary/secondary buffers, `VkSwapchainKHR` acquire/present, fence per frame, see [references/commands-and-swapchain.md](references/commands-and-swapchain.md)

## Gotchas

- `vkAllocateMemory` is hard-capped (`maxMemoryAllocationCount`, often ~4096) and slow; sub-allocate from a few large blocks. The architecture rationale is in gpu-rendering-guide.
- A `VkImageMemoryBarrier2` that omits the layout transition, or uses `oldLayout` that does not match the image's current layout, is undefined behavior — track current layout per image/subresource.
- A `VkDescriptorSet` updated while the GPU may still read it (without `UPDATE_AFTER_BIND`) is a data race — gate on the frame `VkFence`.
- Binding a `VkPipeline` whose dynamic state you forgot to set (e.g. `vkCmdSetViewport`) draws nothing or validation-errors; declare every dynamic state you rely on.
- Ignoring `VK_ERROR_OUT_OF_DATE_KHR`/`VK_SUBOPTIMAL_KHR` from acquire/present leaves a stale swapchain after resize — recreate it.

## Progressive Disclosure

- Read [references/device-and-queues.md](references/device-and-queues.md) - Load when creating the instance/device or selecting queue families
- Read [references/device-memory.md](references/device-memory.md) - Load when allocating device memory, choosing memory types, or staging uploads
- Read [references/resources-and-barriers.md](references/resources-and-barriers.md) - Load when creating images/buffers or placing `VkImageMemoryBarrier2`/layout transitions
- Read [references/descriptors.md](references/descriptors.md) - Load when building descriptor set layouts/pools/sets, going bindless, or using push constants
- Read [references/pipelines.md](references/pipelines.md) - Load when building `VkPipeline`s, the pipeline cache, or dynamic rendering
- Read [references/synchronization.md](references/synchronization.md) - Load when using `VkSemaphore`/`VkFence`, stage/access masks, or submit-time waits
- Read [references/commands-and-swapchain.md](references/commands-and-swapchain.md) - Load when managing command pools/buffers, the swapchain, or frames-in-flight
