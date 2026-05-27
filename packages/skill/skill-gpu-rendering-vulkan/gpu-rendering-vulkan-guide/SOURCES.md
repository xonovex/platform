# Sources

## Vulkan Specification

- **URL:** https://registry.khronos.org/vulkan/specs/1.3/html/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → all sections
  - Authoritative semantics for every `Vk*` type and `vk*` command used
- **Aspects extracted:**
  - Instance/physical-device/`VkDevice` creation, queue families, feature/extension enablement → `references/device-and-queues.md`
  - `VkPhysicalDeviceMemoryProperties`, memory types/heaps, `vkAllocateMemory`, `VkMemoryRequirements`, `bufferImageGranularity` → `references/device-memory.md`
  - `VkImage`/`VkBuffer`, layouts, `VkImageMemoryBarrier2`/`VkBufferMemoryBarrier2`, `vkCmdPipelineBarrier2`, queue-family transfer → `references/resources-and-barriers.md`
  - `VkDescriptorSetLayout`/`Pool`/`Set`, descriptor indexing / `UPDATE_AFTER_BIND`, push constants → `references/descriptors.md`
  - `VkPipeline`, `VkPipelineCache`, dynamic rendering, dynamic state → `references/pipelines.md`
  - `VkSemaphore` (binary + timeline), `VkFence`, `vkQueueSubmit2`, stage/access masks → `references/synchronization.md`
  - `VkCommandPool`/`VkCommandBuffer`, `VkSwapchainKHR`, acquire/present → `references/commands-and-swapchain.md`

## Khronos Vulkan-Guide and Vulkan-Samples

- **URLs:**
  - Khronos Vulkan-Guide — https://docs.vulkan.org/guide/latest/
  - Khronos Vulkan-Samples — https://github.com/KhronosGroup/Vulkan-Samples
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Setup, Pipelines & binding, Commands & present
  - Recommended modern practice (synchronization2, dynamic rendering, timeline semaphores, descriptor indexing)
- **Aspects extracted:**
  - Queue-family selection, swapchain creation/recreation, frames-in-flight wiring → `references/device-and-queues.md`, `references/commands-and-swapchain.md`
  - Dynamic rendering vs render passes, pipeline cache persistence → `references/pipelines.md`
  - Bindless via descriptor indexing, set frequency, push constants → `references/descriptors.md`

## Synchronization examples

- **URL:** https://github.com/KhronosGroup/Vulkan-Docs/wiki/Synchronization-Examples
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Resources & sync
  - Concrete stage/access-mask pairings for common hazards (write→read, transfer→sample, present)
- **Aspects extracted:**
  - `srcStageMask`/`srcAccessMask` → `dstStageMask`/`dstAccessMask` pairings and layout transitions → `references/resources-and-barriers.md`, `references/synchronization.md`

## GPU memory allocator documentation

- **URLs:**
  - Vulkan memory allocator documentation — https://gpuopen.com/learn/vulkan-memory-allocator/
  - https://gpuopen-librariesandsdks.github.io/VulkanMemoryAllocator/html/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Setup (device memory)
  - Sub-allocation from large blocks, memory-type selection, staging, persistent mapping, defragmentation
- **Aspects extracted:**
  - Few large `vkAllocateMemory` blocks + sub-allocation, allocation-count cap, alignment, staging upload → `references/device-memory.md`

## Game-engine development blog (archive)

- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Pipelines & binding, Commands & present, Setup (device memory)
  - Shipped Vulkan practice for descriptor management, command-buffer lifecycle, pipeline caching, and device-memory allocation
- **Aspects extracted:**
  - "Moving the engine to Bindless" — one global `VkDescriptorSet` with a large array per resource class, array index embedded in the resource handle, fence-deferred slot release, device-limit clamping, fallback "null" resources → `references/descriptors.md`
  - "Vulkan: Descriptor Sets Management" — shared sets as read-only blueprints, per-job `VkDescriptorPool` + lazy `VkCopyDescriptorSet`, whole-pool recycling to avoid fragmentation → `references/descriptors.md`
  - "Vulkan: Command Buffer Management" — pool per worker thread, fence-gated deferred resource deletion, recycling whole reset pools into a free pool-of-pools, primary/secondary split → `references/commands-and-swapchain.md`
  - "Vulkan: Pipelines and Render States" — deferred pipeline creation hashed on (formats, shader, state overrides), worker-thread-local staging merged post-frame, stacked render-state override blocks (last-wins, dynamic vs static) → `references/pipelines.md`
  - "Device Memory Management" — buddy 256 MB blocks ≤ block size vs dedicated above, linear staging allocator recycled by fence, tagged allocations + early visual debugging → `references/device-memory.md`

## Refresh Workflow

1. Re-read the upstream source(s) above (spec sections, Vulkan-Guide pages, sync examples, allocator docs)
2. Diff against the prior pull (or scan for newly added sections / API revisions / extension promotions)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
