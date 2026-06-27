# command-recording-and-frames: Command Recording and Frames in Flight

## Guideline

Record GPU work into command streams allocated from a context owned by one thread for one frame and reset wholesale; record independent passes in parallel into separate streams executed by a primary one; double- or triple-buffer all per-frame mutable resources behind one fence per frame so the CPU prepares frame N+1 while the GPU consumes frame N, and never overwrites a resource the GPU still reads.

## Rationale

Command-recording memory is not thread-safe and cannot be freed per-stream cheaply; the efficient model is one recording context per thread per frame slot, recording into it, submitting, and later resetting the whole context in O(1) once the GPU is done with it. That "once the GPU is done" is enforced by a fence: each frame slot has a fence the submission signals on completion, and before the CPU reuses that slot's context, command memory, bindings, and ring-buffer ranges it waits on that slot's fence. To keep both processors busy you keep a small number of frame slots in flight (typically 2 or 3): the CPU records and submits slot i while the GPU executes slot i-1, so neither stalls waiting for the other. Recording is also the parallelizable part of a frame — split passes across threads, each writing a secondary command stream with its own context, then have one primary stream execute them in order. The swapchain ties it together: acquire an image (waiting on a queue-side signal), render into it, present (waiting on the render-done signal), and gate the next reuse of the slot on its fence.

## Techniques

- **Context per thread per frame** - One recording context per (recording thread × frame slot); reset the whole context when the slot's fence signals; never free individual streams.
- **Primary vs secondary** - Primary streams are submitted to a queue; secondary streams are recorded by worker threads and invoked from a primary.
- **Multi-threaded recording** - Partition draws/passes across threads, each recording its own secondary stream into its own context; join, then a primary executes them. Recording scales; submission stays single-threaded.
- **Sort keys** - Tag each recorded command with a 64-bit sort key; merge all threads' streams and sort by key before submission. GPU execution order is then decoupled from the order (and thread) commands were recorded in, so workers record independently and ordering is resolved once, afterward.
- **Frames in flight** - N frame slots (2 = double, 3 = triple buffer), each with its own recording contexts, per-frame bindings, ring-buffer ranges, and one fence.
- **Per-frame resource sets** - Anything the GPU reads that the CPU also rewrites (uniforms, instance buffers, dynamic bindings) is duplicated per slot so writing slot i can't race the GPU reading slot i-1.
- **Acquire/submit/present loop** - acquire image (signal image-available) → record → submit (wait image-available, signal render-done, signal fence) → present (wait render-done), see [references/synchronization.md](./synchronization.md).
- **Fence-gated reuse** - At frame start, wait the fence for the slot about to be reused — not the one just submitted — so the CPU runs ahead by N-1 frames.

## Example

```c
// Neutral pseudocode; concrete pool/buffer/fence objects live in the per-API skill.
enum { FRAMES_IN_FLIGHT = 2 };
typedef struct {
    cmd_context ctx;        // reset wholesale, not per-stream
    cmd_stream  cmd;
    gpu_fence   in_flight;  // signaled when this slot's GPU work completes
    gpu_ring    per_frame_ubo; // duplicated so writes don't race the GPU
    binding_group group;
} frame_slot;
frame_slot slots[FRAMES_IN_FLIGHT];

void draw_frame(uint64_t frame) {
    frame_slot *s = &slots[frame % FRAMES_IN_FLIGHT];
    fence_wait(&s->in_flight);        // wait the slot we reuse, not the one just submitted
    fence_reset(&s->in_flight);
    cmd_context_reset(s->ctx);        // O(1) reclaim; the fence guarantees it is safe

    uint32_t img; acquire_next_image(swapchain, image_available[frame % 2], &img);
    cmd_begin(s->cmd);
    // multi-threaded: workers fill secondaries[]; the primary executes them in order.
    cmd_execute_secondaries(s->cmd, secondaries, n_workers);
    cmd_end(s->cmd);

    queue_submit(gfx, s->cmd, /*wait*/ image_available[frame % 2],
                 /*signal*/ render_done[frame % 2], /*fence*/ s->in_flight);
    queue_present(present, swapchain, img, /*wait*/ render_done[frame % 2]);
}
```

## Gotchas

- Recording into (or resetting) a context whose previous submission's fence has not signaled corrupts in-flight GPU work — always gate reuse on the slot fence.
- Sharing one recording context across threads is a data race; contexts are externally synchronized — one context per thread.
- Waiting on the fence of the frame you just submitted (instead of the slot you are about to reuse) collapses frames-in-flight back to a full CPU/GPU stall.
- Per-frame resources must be genuinely duplicated; a single shared dynamic buffer written each frame races the GPU reading last frame's value.
- The swapchain can return out-of-date/suboptimal on resize; recreate the swapchain and dependent targets, do not present into a stale image.
- More frames in flight adds input latency and multiplies per-frame memory; 2–3 is the usual sweet spot, not "as many as possible".
- Without a sort key, GPU order _is_ the order commands were recorded, so multi-threaded recording forces threads to coordinate their relative ordering; a per-command key lets each thread record obliviously and the order is sorted out after the join.

## Related

[references/synchronization.md](./synchronization.md), [references/gpu-memory-strategy.md](./gpu-memory-strategy.md), [references/render-graph.md](./render-graph.md), [references/binding-model.md](./binding-model.md)
