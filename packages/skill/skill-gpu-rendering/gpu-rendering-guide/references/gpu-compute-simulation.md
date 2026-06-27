# gpu-compute-simulation: GPU-Resident Simulation with Compute

## Guideline

For large element counts (particles, agents, cloth), keep the simulation state resident in GPU storage buffers, advance it with compute dispatches, and let the GPU drive its own draw/dispatch counts via indirect arguments — so the CPU never enumerates elements or stalls on readback.

## Rationale

A simulation that lives on the CPU and uploads results every frame is bounded by upload bandwidth and by the CPU loop, and reading per-element results back to the CPU forces a pipeline stall (you wait for the GPU to finish, killing parallelism). Moving the state into persistent GPU buffers eliminates both: the data never crosses the bus after init, the update is a single dispatch over all active elements, and the count of live elements is tracked on the GPU and fed straight into an indirect draw/dispatch so the CPU issues a fixed, count-agnostic command. This is what lets element counts scale into the hundreds of thousands without the CPU touching any element.

## How to Apply

1. Lay out element state as typed channels in storage buffers, SoA per channel, sized `capacity × stride`; treat the buffer as a ring so a spawn overwrites the oldest element instead of failing.
2. Split the work into event-triggered stages: `init` (once), `spawn` (on demand), `update` (per frame) — each a compute dispatch over the active range.
3. Advance per-element lifetime by the frame delta inside `update`; cull dead elements by compaction or by tracking a live count on the GPU.
4. Keep the live-element count in a GPU buffer and use it as the argument to an indirect dispatch/draw; never read it back to size CPU-side dispatches.
5. Double-buffer state that needs last-frame values (e.g. previous position for velocity), ping-ponging read/write buffers across frames.
6. Use atomics or append/consume buffers for spawn slots and compaction so concurrent threads don't collide.

## Example

```hlsl
// Update stage: one thread per element, state stays in GPU buffers across frames.
RWStructuredBuffer<float3> pos_curr;   // ping-pong with pos_prev each frame
RWStructuredBuffer<float3> pos_prev;
RWStructuredBuffer<float>  life;
RWByteAddressBuffer        live_count; // also feeds the indirect draw args

[numthreads(64, 1, 1)]
void update(uint id : SV_DispatchThreadID) {
    if (id >= capacity) return;
    float t = life[id] - delta_time;
    if (t <= 0.0) return;                 // dead: skip (compaction handles slot reuse)
    life[id]      = t;
    pos_prev[id]  = pos_curr[id];         // keep previous for motion vectors
    pos_curr[id] += velocity(id) * delta_time;
    live_count.InterlockedAdd(0, 1);      // GPU counts its own work
}
// CPU side: DispatchIndirect / DrawIndirect read their counts from live_count — fixed CPU command.
```

## Gotchas

- Absolute indexing into the ring while another stage writes the same slots is a read/write hazard — separate stages with a barrier, or index relative to the active range.
- Reading any per-element result back to the CPU reintroduces the stall you moved to the GPU to avoid; consume results on the GPU (indirect draw) instead.
- A draw command that hardcodes a vertex count can't scale procedurally; source counts from the GPU-side live count.
- Ring overwrite means a spawn never fails but silently evicts the oldest element when `count > capacity` — size capacity for the worst case you care about.
- CPU-side rate spawning defeats the model for high rates; push spawn events into a GPU buffer the spawn dispatch consumes.

## Related

[references/gpu-memory-strategy.md](./gpu-memory-strategy.md), [references/synchronization.md](./synchronization.md), [references/command-recording-and-frames.md](./command-recording-and-frames.md), **data-oriented-design-guide**
