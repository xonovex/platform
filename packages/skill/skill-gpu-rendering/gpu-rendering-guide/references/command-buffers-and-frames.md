# command-buffers-and-frames: Command Buffer Recording and Frames in Flight

**Guideline:** Record GPU work into command buffers allocated from a pool that is owned by one thread for one frame and reset wholesale by the pool; record independent passes in parallel into secondary buffers executed by a primary; double- or triple-buffer all per-frame mutable resources behind one fence per frame so the CPU prepares frame N+1 while the GPU consumes frame N, and never overwrites a resource the GPU still reads.

**Rationale:** Command-pool memory is not thread-safe and cannot be freed per-buffer cheaply; the efficient model is one pool per recording thread per frame slot, recording into it, submitting, and later resetting the whole pool in O(1) once the GPU is done with it. That "once the GPU is done" is enforced by a fence: each frame slot has a fence the submission signals on completion, and before the CPU reuses that slot's pool, command buffers, descriptors, and ring-buffer ranges it waits on that slot's fence. To keep both processors busy you keep a small number of frame slots in flight (typically 2 or 3): the CPU records and submits slot i while the GPU executes slot i-1, so neither stalls waiting for the other. Recording is also the parallelizable part of a frame — split passes across threads, each writing a secondary command buffer with its own pool, then have one primary buffer execute them in order. The swapchain ties it together: acquire an image (waiting on a semaphore), render into it, present (waiting on the render-done semaphore), and gate the next reuse of the slot on its fence.

**Techniques:**

- **Pool per thread per frame** - One command pool per (recording thread × frame slot); reset the whole pool when the slot's fence signals; never free individual buffers.
- **Primary vs secondary** - Primary buffers are submitted to a queue; secondary buffers are recorded by worker threads and invoked from a primary via execute-commands.
- **Multi-threaded recording** - Partition draws/passes across threads, each recording its own secondary buffer into its own pool; join, then a primary executes them. Recording scales; submission stays single-threaded.
- **Frames in flight** - N frame slots (2 = double, 3 = triple buffer), each with its own command pools, per-frame descriptor sets, ring-buffer ranges, and one fence.
- **Per-frame resource sets** - Anything the GPU reads that the CPU also rewrites (uniforms, instance buffers, dynamic descriptors) is duplicated per slot so writing slot i can't race the GPU reading slot i-1.
- **Acquire/submit/present loop** - acquire image (signal image-available) → record → submit (wait image-available, signal render-done, signal fence) → present (wait render-done), see [references/synchronization.md](./synchronization.md).
- **Fence-gated reuse** - At frame start, wait the fence for the slot about to be reused — not the one just submitted — so the CPU runs ahead by N-1 frames.

**Example:**

```c
enum { FRAMES_IN_FLIGHT = 2 };
typedef struct {
    VkCommandPool pool;          // reset wholesale, not per-buffer
    VkCommandBuffer cmd;
    VkFence in_flight;           // signaled when this slot's GPU work completes
    gpu_ring per_frame_ubo;      // duplicated so writes don't race the GPU
    VkDescriptorSet set;
} frame_slot;
frame_slot slots[FRAMES_IN_FLIGHT];

void draw_frame(uint64_t frame) {
    frame_slot *s = &slots[frame % FRAMES_IN_FLIGHT];
    vkWaitForFences(dev, 1, &s->in_flight, VK_TRUE, UINT64_MAX); // wait the slot we reuse
    vkResetFences(dev, 1, &s->in_flight);
    vkResetCommandPool(dev, s->pool, 0);                        // O(1) reclaim, fence guarantees safe

    uint32_t img; vkAcquireNextImageKHR(dev, swapchain, ~0ull, image_available[frame%2], NULL, &img);
    vkBeginCommandBuffer(s->cmd, &begin);
    // multi-threaded: workers fill secondaries[]; primary executes them in order.
    vkCmdExecuteCommands(s->cmd, n_workers, secondaries);
    vkEndCommandBuffer(s->cmd);

    VkSubmitInfo si = { .pWaitSemaphores = &image_available[frame%2],
                        .pSignalSemaphores = &render_done[frame%2], ... };
    vkQueueSubmit(gfx, 1, &si, s->in_flight);                   // fence signals when GPU done
    vkQueuePresentKHR(present, &(VkPresentInfoKHR){ .pWaitSemaphores = &render_done[frame%2], ... });
}
```

**Gotchas:**

- Recording into (or resetting) a command buffer whose previous submission's fence has not signaled corrupts in-flight GPU work — always gate reuse on the slot fence.
- Sharing one command pool across threads is a data race; pools are externally synchronized — one pool per thread.
- Waiting on the fence of the frame you just submitted (instead of the slot you are about to reuse) collapses frames-in-flight back to a full CPU/GPU stall.
- Per-frame resources must be genuinely duplicated; a single shared dynamic buffer written each frame races the GPU reading last frame's value.
- The swapchain can return out-of-date/suboptimal on resize; recreate the swapchain and dependent targets, do not present into a stale image.
- More frames in flight adds input latency and multiplies per-frame memory; 2–3 is the usual sweet spot, not "as many as possible".

**Related:** [references/synchronization.md](./synchronization.md), [references/device-memory.md](./device-memory.md), [references/render-graph.md](./render-graph.md), [references/pipelines-and-descriptors.md](./pipelines-and-descriptors.md)
