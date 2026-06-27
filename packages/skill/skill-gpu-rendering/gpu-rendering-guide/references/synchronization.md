# synchronization: Explicit GPU↔CPU and GPU↔GPU Synchronization

## Guideline

Synchronize everything explicitly with the right primitive for the relationship: resource barriers for dependencies within a queue (scoped source/destination stage+access plus image-layout transitions), cross-queue waits to order work between submissions and across queues, a monotonic timeline value for counter-style waits, and fences for GPU→CPU completion — scoping each to exactly the work that must wait, neither more nor less.

## Rationale

An explicit API guarantees no ordering between commands unless you ask for it: a write in one pass is not visible to a read in the next, and an image is not in a usable layout, until a barrier says so. The barrier names a source scope (stages that must finish, the access that must be made available) and a destination scope (stages that must wait, the access made visible), plus any layout transition. Getting these scopes tight is the whole game: too broad (everything → everything) is always correct but drains the pipeline and erases the parallelism the GPU exists for; too narrow is a hazard that produces garbage or a hang, often only on some hardware. Barriers only order work inside one queue; crossing queues (graphics→compute, or to the presentation engine) needs a queue-side wait — a GPU-side ordering between submissions (binary semaphore in Vulkan/WebGPU terms, a fence value in D3D12). A monotonic timeline value generalizes this to a counter so one object expresses many waits and supports wait-before-signal. Fences are the only GPU→CPU signal — the CPU blocks on a fence to know a frame's resources are free to reuse or that readback data has landed.

## Techniques

- **Resource barrier** - In-queue dependency: source stage/access (producer) → destination stage/access (consumer), optionally an image-layout transition or queue-ownership transfer.
- **Layout transition** - Images must be in the layout matching their next use (color-attachment, shader-read, transfer-destination, present); the transition rides in the barrier.
- **Cross-queue wait** - Orders one submission after another on the GPU (queue→queue); signaled by one submit, waited by the next. Used for swapchain acquire/present and async-compute handoff.
- **Timeline value** - A 64-bit counter; wait for value ≥ N, signal value N. One object replaces many binary waits and lets the CPU wait too, often retiring fences.
- **Fence** - GPU→CPU signal; the CPU waits on it to reclaim per-frame command memory, bindings, and ring-buffer ranges, see [references/command-recording-and-frames.md](./command-recording-and-frames.md).
- **Split barrier / event** - Fine-grained barrier within a queue: signal after producing, wait before consuming, letting unrelated work fill the gap.
- **Queue ownership** - A resource used on two queues needs an ownership transfer (release on source, acquire on destination) or a concurrent/shared mode.

## Example

```c
// In-queue barrier: render-target write -> sampled read, with the layout transition.
// (Neutral pseudocode; concrete barrier structs live in the per-API skill.)
gpu_barrier b = {
    .src_access = ACCESS_COLOR_ATTACHMENT_WRITE,   // make the write available
    .dst_access = ACCESS_SHADER_READ,              // visible to the sampler
    .old_layout = LAYOUT_COLOR_ATTACHMENT,
    .new_layout = LAYOUT_SHADER_READ_ONLY,         // required for sampling
    .image      = gbuffer,
};
cmd_pipeline_barrier(cmd,
    STAGE_COLOR_ATTACHMENT_OUTPUT,                 // src stage: scoped, not "all commands"
    STAGE_FRAGMENT_SHADER,                         // dst stage: only the consumer waits
    &b, 1);

// Queue/present ordering uses cross-queue waits, not barriers.
acquire_next_image(swapchain, image_available, &idx);
queue_submit(gfx, cmd, /*wait*/ image_available, /*signal*/ render_done, /*fence*/ frame_fence);
queue_present(present, swapchain, idx, /*wait*/ render_done);
```

## Gotchas

- A correct-but-broad barrier (everything → everything, or a full memory barrier per draw) serializes the GPU and silently tanks throughput — scope to the real producer/consumer.
- Under-synchronizing (a missing barrier between write and read) is a hazard that may "work" on the hardware you tested and corrupt on another — rely on validation/sync-validation layers, not luck.
- A layout transition is mandatory even when the data is ready; using an image in the wrong layout is undefined behavior.
- Cross-queue waits order GPU work but do not make memory available/visible the way a barrier does within a queue — you often still need both at a queue boundary.
- Forgetting a queue-ownership transfer for a resource shared across queues corrupts it on at least one queue; use a transfer or a concurrent/shared mode.
- Waiting on a fence every frame with a full stall defeats frames-in-flight; wait on the fence for the frame slot you are about to reuse, not the one just submitted, see [references/command-recording-and-frames.md](./command-recording-and-frames.md).

## Related

[references/command-recording-and-frames.md](./command-recording-and-frames.md), [references/render-graph.md](./render-graph.md), [references/gpu-memory-strategy.md](./gpu-memory-strategy.md), [references/binding-model.md](./binding-model.md)
