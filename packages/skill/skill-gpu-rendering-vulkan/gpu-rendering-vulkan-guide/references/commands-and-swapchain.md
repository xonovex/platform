# commands-and-swapchain: VkCommandPool, Command Buffers, and the Swapchain

**Guideline:** Allocate command buffers from one `VkCommandPool` per recording thread per frame slot and reset the whole pool with `vkResetCommandPool` once the slot's `VkFence` signals; record independent passes into secondary `VkCommandBuffer`s executed by a primary; drive presentation with `vkAcquireNextImageKHR`/`vkQueuePresentKHR` over a `VkSwapchainKHR`, keeping N frame slots in flight each with its own pools, per-frame resources, and a fence.

**Rationale:** A `VkCommandPool` is externally synchronized and frees buffers cheaply only as a whole (`vkResetCommandPool` / `vkResetCommandBuffer`), so the efficient model is one pool per (thread × frame slot): record, submit, and later reset the whole pool in O(1) once the GPU is done — proven done by the slot's `VkFence`. Keeping 2–3 frame slots in flight lets the CPU record slot i while the GPU runs slot i-1. Recording parallelizes: worker threads each fill a secondary `VkCommandBuffer` from their own pool, and a primary buffer runs them with `vkCmdExecuteCommands`. The swapchain ties it together — `vkAcquireNextImageKHR` returns an image index and signals an image-available semaphore, you render into that image, then `vkQueuePresentKHR` waits the render-finished semaphore. The agnostic frames-in-flight model is in gpu-rendering-guide (command-recording-and-frames).

**Techniques:**

- **Pool per thread per frame** - `vkCreateCommandPool` per (recording thread × frame slot); `vkResetCommandPool` when the slot fence signals; never `vkFreeCommandBuffers` per buffer on the hot path.
- **Primary/secondary** - Primary buffers go to `vkQueueSubmit2`; secondaries are recorded with `VkCommandBufferInheritanceInfo` and invoked via `vkCmdExecuteCommands`.
- **Frames in flight** - N slots (2–3), each owning command pools, per-frame descriptor sets, ring-buffer ranges, image-available/render-finished semaphores, and one `VkFence`.
- **Swapchain creation** - `vkCreateSwapchainKHR` from surface caps/formats/present modes; retrieve images with `vkGetSwapchainImagesKHR`; create a `VkImageView` per image.
- **Acquire/submit/present** - `vkAcquireNextImageKHR` (signal image-available) → record → `vkQueueSubmit2` (wait image-available, signal render-finished, signal fence) → `vkQueuePresentKHR` (wait render-finished), see [references/synchronization.md](./synchronization.md).
- **Recreation** - On `VK_ERROR_OUT_OF_DATE_KHR`/`VK_SUBOPTIMAL_KHR` or resize, `vkDeviceWaitIdle`, destroy, and recreate the swapchain and its views/targets.

**Example:**

```c
enum { FRAMES_IN_FLIGHT = 2 };
typedef struct {
    VkCommandPool   pool;        // reset wholesale, not per-buffer
    VkCommandBuffer cmd;
    VkFence         in_flight;   // signaled when this slot's GPU work completes
    VkSemaphore     image_available, render_finished;
} frame_slot;
frame_slot slots[FRAMES_IN_FLIGHT];

void draw_frame(uint64_t frame) {
    frame_slot *s = &slots[frame % FRAMES_IN_FLIGHT];
    vkWaitForFences(dev, 1, &s->in_flight, VK_TRUE, UINT64_MAX); // wait the slot we reuse
    vkResetFences(dev, 1, &s->in_flight);

    uint32_t img;
    VkResult r = vkAcquireNextImageKHR(dev, swapchain, UINT64_MAX, s->image_available, VK_NULL_HANDLE, &img);
    if (r == VK_ERROR_OUT_OF_DATE_KHR) { recreate_swapchain(); return; }

    vkResetCommandPool(dev, s->pool, 0);                        // O(1); fence guarantees safe
    vkBeginCommandBuffer(s->cmd, &(VkCommandBufferBeginInfo){
        .sType = VK_STRUCTURE_TYPE_COMMAND_BUFFER_BEGIN_INFO,
        .flags = VK_COMMAND_BUFFER_USAGE_ONE_TIME_SUBMIT_BIT});
    vkCmdExecuteCommands(s->cmd, n_workers, secondaries);       // workers filled these
    vkEndCommandBuffer(s->cmd);

    VkSemaphoreSubmitInfo wait = {.sType = VK_STRUCTURE_TYPE_SEMAPHORE_SUBMIT_INFO,
        .semaphore = s->image_available, .stageMask = VK_PIPELINE_STAGE_2_COLOR_ATTACHMENT_OUTPUT_BIT};
    VkSemaphoreSubmitInfo sig = {.sType = VK_STRUCTURE_TYPE_SEMAPHORE_SUBMIT_INFO,
        .semaphore = s->render_finished, .stageMask = VK_PIPELINE_STAGE_2_ALL_GRAPHICS_BIT};
    VkCommandBufferSubmitInfo cbi = {.sType = VK_STRUCTURE_TYPE_COMMAND_BUFFER_SUBMIT_INFO, .commandBuffer = s->cmd};
    VkSubmitInfo2 si = {.sType = VK_STRUCTURE_TYPE_SUBMIT_INFO_2,
        .waitSemaphoreInfoCount = 1, .pWaitSemaphoreInfos = &wait,
        .commandBufferInfoCount = 1, .pCommandBufferInfos = &cbi,
        .signalSemaphoreInfoCount = 1, .pSignalSemaphoreInfos = &sig};
    vkQueueSubmit2(gfx_q, 1, &si, s->in_flight);                // fence signals on completion
    vkQueuePresentKHR(present_q, &(VkPresentInfoKHR){.sType = VK_STRUCTURE_TYPE_PRESENT_INFO_KHR,
        .waitSemaphoreCount = 1, .pWaitSemaphores = &s->render_finished,
        .swapchainCount = 1, .pSwapchains = &swapchain, .pImageIndices = &img});
}
```

**Gotchas:**

- Resetting a pool or recording into a buffer whose previous submission's fence has not signaled corrupts in-flight GPU work — always gate on the slot fence.
- A `VkCommandPool` is not thread-safe; two threads recording from one pool race — one pool per thread.
- Waiting on the fence you just submitted (not the slot about to be reused) collapses frames-in-flight into a stall.
- Ignoring `VK_SUBOPTIMAL_KHR`/`VK_ERROR_OUT_OF_DATE_KHR` from acquire or present leaves a stale swapchain after resize — recreate it (and never present into a stale image).
- The image-available semaphore must be one not currently pending on another acquire; using a per-frame-slot semaphore avoids reusing one mid-flight.

**Related:** [references/synchronization.md](./synchronization.md), [references/device-memory.md](./device-memory.md), [references/resources-and-barriers.md](./resources-and-barriers.md), [references/device-and-queues.md](./device-and-queues.md)
