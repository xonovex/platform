# synchronization: Explicit GPU↔CPU and GPU↔GPU Synchronization

**Guideline:** Synchronize everything explicitly with the right primitive for the relationship: pipeline barriers for dependencies within a queue (scoped src/dst stage+access masks plus image-layout transitions), semaphores to order work between submissions and across queues, timeline semaphores for monotonic counter waits, and fences for GPU→CPU completion — scoping each to exactly the work that must wait, neither more nor less.

**Rationale:** An explicit API guarantees no ordering between commands unless you ask for it: a write in one pass is not visible to a read in the next, and an image is not in a usable layout, until a barrier says so. The barrier names a source scope (stages that must finish, the access that must be made available) and a destination scope (stages that must wait, the access made visible), plus any layout transition. Getting these masks tight is the whole game: too broad (ALL_COMMANDS → ALL_COMMANDS) is always correct but drains the pipeline and erases the parallelism the GPU exists for; too narrow is a hazard that produces garbage or a hang, often only on some hardware. Barriers only order work inside one queue; crossing queues (graphics→compute, or to the presentation engine) needs a semaphore, which is a GPU-side wait between submissions. Timeline semaphores generalize this to a monotonically increasing value so one object expresses many waits and supports wait-before-signal. Fences are the only GPU→CPU signal — the CPU blocks on a fence to know a frame's resources are free to reuse or that readback data has landed.

**Techniques:**

- **Pipeline barrier** - In-queue dependency: src stage/access (producer) → dst stage/access (consumer), optionally an image-layout transition or queue-family ownership transfer.
- **Layout transition** - Images must be in the layout matching their next use (COLOR_ATTACHMENT_OPTIMAL, SHADER_READ_ONLY_OPTIMAL, TRANSFER_DST_OPTIMAL, PRESENT_SRC); the transition rides in the barrier.
- **Binary semaphore** - Orders one submission after another on the GPU (queue→queue); signaled by one submit, waited by the next. Used for swapchain acquire/present and async-compute handoff.
- **Timeline semaphore** - A 64-bit counter; wait for value ≥ N, signal value N. One object replaces many binary semaphores and allows the CPU to wait too, often retiring fences.
- **Fence** - GPU→CPU signal; the CPU waits on it to reclaim per-frame command buffers, descriptors, and ring-buffer ranges, see [references/command-buffers-and-frames.md](./command-buffers-and-frames.md).
- **Event** - Fine-grained split barrier within a queue: set after producing, wait before consuming, letting unrelated work fill the gap.
- **Queue families** - A resource used on two queues needs an ownership transfer (release on src, acquire on dst) or CONCURRENT sharing mode.

**Example:**

```c
// In-queue barrier: render-target write -> sampled read, with the layout transition.
VkImageMemoryBarrier b = {
    .srcAccessMask = VK_ACCESS_COLOR_ATTACHMENT_WRITE_BIT,    // make the write available
    .dstAccessMask = VK_ACCESS_SHADER_READ_BIT,               // visible to the sampler
    .oldLayout     = VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL,
    .newLayout     = VK_IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL, // required for sampling
    .image = gbuffer,
};
vkCmdPipelineBarrier(cmd,
    VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT,            // src stage: scoped, not ALL_COMMANDS
    VK_PIPELINE_STAGE_FRAGMENT_SHADER_BIT,                    // dst stage: only the consumer waits
    0, 0, NULL, 0, NULL, 1, &b);

// Queue/present ordering uses semaphores, not barriers.
vkAcquireNextImageKHR(dev, swapchain, UINT64_MAX, image_available, NULL, &idx);
VkSubmitInfo s = { .waitSemaphoreCount = 1, .pWaitSemaphores = &image_available,
                   .signalSemaphoreCount = 1, .pSignalSemaphores = &render_done };
vkQueueSubmit(gfx, 1, &s, frame_fence);                      // fence: GPU->CPU, reclaim frame resources
vkQueuePresentKHR(present, &(VkPresentInfoKHR){ .pWaitSemaphores = &render_done, ... });
```

**Gotchas:**

- A correct-but-broad barrier (ALL_COMMANDS → ALL_COMMANDS, or a full memory barrier per draw) serializes the GPU and silently tanks throughput — scope masks to the real producer/consumer.
- Under-synchronizing (a missing barrier between write and read) is a hazard that may "work" on the hardware you tested and corrupt on another — rely on validation/sync-validation layers, not luck.
- A layout transition is mandatory even when the data is ready; using an image in the wrong layout is undefined behavior.
- Semaphores order GPU work but do not make memory available/visible the way a barrier does within a queue — you often still need both at a queue boundary.
- Forgetting a queue-family ownership transfer for a resource shared across queues corrupts it on at least one queue; use a transfer or CONCURRENT mode.
- Waiting on a fence every frame with a full stall defeats frames-in-flight; wait on the fence for the frame slot you are about to reuse, not the one just submitted, see [references/command-buffers-and-frames.md](./command-buffers-and-frames.md).

**Related:** [references/command-buffers-and-frames.md](./command-buffers-and-frames.md), [references/render-graph.md](./render-graph.md), [references/device-memory.md](./device-memory.md), [references/pipelines-and-descriptors.md](./pipelines-and-descriptors.md)
