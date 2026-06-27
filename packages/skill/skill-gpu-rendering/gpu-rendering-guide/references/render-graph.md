# render-graph: Render Graphs and Transient Resource Aliasing

## Guideline

Express a frame as a graph of passes that each declare which resources they read and write, and let the graph derive execution order, insert barriers and image-layout transitions, prune unused passes, and alias transient render targets whose lifetimes do not overlap — instead of hand-sequencing barriers and managing target memory by hand.

## Rationale

Hand-managed barriers are the single largest source of correctness bugs and silent over-synchronization in an explicit renderer: every pass must know the prior and next usage of every resource it touches, which couples passes together and breaks the moment one is reordered. A render graph (or frame graph) inverts this. Passes are declarative and order-independent; the graph topologically sorts them from the read/write dependencies, computes the minimal barrier (correct stage/access scopes and layout) at each producer→consumer edge automatically, and drops any pass whose outputs nothing consumes. Because the graph knows every resource's first and last use, it can place two transient targets in the same physical memory when their lifetimes are disjoint, cutting VRAM for the framebuffer-heavy passes (G-buffer, bloom chains, SSAO) that dominate a modern frame. Rebuilding the graph per frame is cheap relative to the GPU work it schedules.

## How to Apply

1. Each frame, register passes; each pass declares its inputs (reads) and outputs (writes) as virtual resource handles, not physical allocations.
2. Build a DAG from the declared dependencies; topologically sort to get execution order.
3. Cull: walk back from the final presented/consumed resources and drop any pass nothing reaches.
4. Compute resource lifetimes (first write → last read); assign physical memory, aliasing transient targets with disjoint lifetimes onto the same allocation.
5. Walk each producer→consumer edge and emit the barrier: source/destination stage+access scopes and the image-layout transition implied by the usages.
6. Execute passes in sorted order, invoking each pass's record callback with its resolved physical resources.

## Example

```c
// Declarative pass setup — the pass states intent, not barriers or ordering.
// (Neutral pseudocode; concrete API calls live in the per-API skill.)
rg_handle depth = rg_create_image(g, &(rg_image_desc){.format = DEPTH32F, .transient = true});
rg_handle gbuf  = rg_create_image(g, &(rg_image_desc){.format = RGBA8,    .transient = true});
rg_handle lit   = rg_create_image(g, &(rg_image_desc){.format = RGBA16F,  .transient = true});

rg_add_pass(g, "gbuffer", (rg_pass){
    .writes  = {depth, gbuf},                // graph infers color/depth-attachment layouts
    .execute = record_gbuffer,
});
rg_add_pass(g, "lighting", (rg_pass){
    .reads   = {depth, gbuf},                // graph inserts the read-after-write barrier here:
    .writes  = {lit},                        //   gbuf: color-attachment -> shader-read layout
    .execute = record_lighting,              //   src = color-attachment-output, dst = fragment-shader
});
// `depth` and `gbuf` are never read after lighting -> the graph may alias their memory
// to a later transient (e.g. a bloom mip). A debug-only pass writing an unread target is culled.

rg_compile(g);   // sort, cull, assign+alias memory, plan barriers/transitions
rg_execute(g, cmd);
```

## Gotchas

- The graph can only insert a correct barrier for usage it was told about; a pass that touches a resource it did not declare gets no barrier and corrupts silently — declare every read and write.
- Aliasing requires non-overlapping lifetimes; if a "transient" target is read after the graph reused its memory, you get garbage — mark anything that must persist as non-transient.
- An aliased resource's contents are undefined on first use in its new lifetime; always fully overwrite (or clear) before reading.
- Per-frame graph rebuild must be cheap (arena-allocated, no global heap churn) or it eats the frame it is meant to schedule.
- Async-compute and multi-queue edges need queue-ownership transfers and cross-queue waits, not just intra-queue barriers — the graph must model the queue, see [references/synchronization.md](./synchronization.md).

## Related

[references/synchronization.md](./synchronization.md), [references/gpu-memory-strategy.md](./gpu-memory-strategy.md), [references/command-recording-and-frames.md](./command-recording-and-frames.md), [references/binding-model.md](./binding-model.md)
