---
name: gpu-rendering-guide
description: "Use when designing the architecture of a low-level GPU renderer on any explicit API (Vulkan/D3D12/Metal/WebGPU): render/frame graphs, shader/permutation systems, the descriptor/binding model, explicit GPU↔CPU synchronization, command recording and frames-in-flight, and GPU memory strategy. Triggers on render passes, resource barriers/transitions, transient/aliased targets, shader variants, bind groups/descriptor frequency, command buffers, double/triple buffering, even when the user doesn't say 'renderer' or name a specific API. Skip API-specific how-to (use gpu-rendering-vulkan-guide), high-level web 3D (use threejs-guide), and CPU-only data layout (use data-oriented-design-guide)."
---

# GPU Rendering Guidelines (Explicit-API Architecture)

API-agnostic architecture for a low-level GPU renderer. The concepts hold across explicit APIs (Vulkan, D3D12, Metal, WebGPU); names differ but the model is the same. For one API's concrete how-to use the matching API skill (e.g. gpu-rendering-vulkan-guide); for the general allocator principle behind GPU memory, see memory-management-guide.

## Requirements

- Target an explicit, low-level API where the app owns memory, synchronization, and pipeline state — not a driver that hides them.
- Assume validation/debug layers and a frame debugger are available during development.

## Essentials

- **Declare, don't sequence** - Passes declare resource reads/writes; the graph derives order, barriers, and layout transitions, see [references/render-graph.md](references/render-graph.md)
- **Shaders are build artifacts** - Author → compile to a binary intermediate offline → reflect for layouts, see [references/shader-system.md](references/shader-system.md)
- **You own GPU memory** - Sub-allocate from a few large blocks; never one allocation per resource, see [references/gpu-memory-strategy.md](references/gpu-memory-strategy.md)
- **You own synchronization** - Nothing is implicit; barriers, queue waits, and fences are explicit, see [references/synchronization.md](references/synchronization.md)

## Architecture

- **Render graph** - Per-frame DAG that prunes dead passes, orders work, and aliases transient targets, see [references/render-graph.md](references/render-graph.md)
- **Shader system** - Permutations vs ubershaders, pipeline keyed by shader+state, hot-reload, caching, see [references/shader-system.md](references/shader-system.md)
- **Command recording** - One recording context per thread per frame, multi-threaded record, see [references/command-recording-and-frames.md](references/command-recording-and-frames.md)
- **Frames in flight** - Double/triple-buffer per-frame resource sets behind a fence so the CPU can't outrun the GPU, see [references/command-recording-and-frames.md](references/command-recording-and-frames.md)
- **Sort keys** - A 64-bit key per command decouples GPU submission order from CPU record order; merge worker streams and sort before submit, see [references/command-recording-and-frames.md](references/command-recording-and-frames.md)
- **Programmable vertex fetch** - Pull vertices from storage buffers behind a loader interface instead of fixed-function input; skin in-shader from an indirected influence list, see [references/vertex-assembly-skinning.md](references/vertex-assembly-skinning.md)
- **GPU-resident simulation** - Keep large element state in GPU buffers, advance with compute, and drive draws from a GPU-tracked count via indirect args — no CPU enumeration or readback, see [references/gpu-compute-simulation.md](references/gpu-compute-simulation.md)

## Resources

- **GPU memory** - Device-local vs host-visible tiers, placement alignment, sub-allocation, see [references/gpu-memory-strategy.md](references/gpu-memory-strategy.md)
- **Uploads** - Staging buffer → device-local copy; persistent-mapped ring buffers for per-frame data, see [references/gpu-memory-strategy.md](references/gpu-memory-strategy.md)
- **Pipeline state** - Precompile and cache pipeline objects; avoid first-use compile stalls, see [references/binding-model.md](references/binding-model.md)
- **Binding model** - Group bindings by update frequency, bindless arrays + handles, inline constants for tiny data, see [references/binding-model.md](references/binding-model.md)

## Output

- **HDR output** - Wide-gamut PQ/scRGB swapchain at ≥10-bit/FP16, linear pipeline, own the final color-space + transfer encode scaled to the display's real peak nits, see [references/hdr-output.md](references/hdr-output.md)

## Synchronization

- **Resource barriers** - Producer/consumer stage+access scopes plus image-layout transitions, see [references/synchronization.md](references/synchronization.md)
- **Queue↔queue** - Queue-side waits order submissions; a monotonic timeline value generalizes them, see [references/synchronization.md](references/synchronization.md)
- **GPU→CPU** - Fences gate frame resource reuse and readback, see [references/synchronization.md](references/synchronization.md)

## Gotchas

- A barrier scoped too broadly (everything → everything) is correct but serializes the GPU; scope stage/access to what actually waits.
- Forgetting an image-layout transition is undefined behavior even when the data is "obviously" ready — layout is part of the contract, not a hint.
- One allocation per resource exhausts the (small, capped) device allocation limit and is slow; sub-allocate from large blocks.
- Host-coherent memory skips explicit flush/invalidate but is not free; large dynamic data still wants a device-local copy via staging.
- Recording into a per-frame context whose previous submission's fence has not signaled corrupts in-flight GPU work — gate reuse on the fence.
- Re-using a transient target the graph aliased to another lifetime, then reading it later, returns garbage; lifetimes must not overlap.
- Applying a display transfer function twice (an automatic sRGB backbuffer plus your own HDR encode) double-darkens and loses precision in the shadows — encode exactly once.
- Reading GPU-simulation results back to the CPU each frame reintroduces the full pipeline stall you went to the GPU to avoid; consume them on the GPU via indirect draw.

## Progressive Disclosure

- Read [references/render-graph.md](references/render-graph.md) - Load when ordering passes, automating barriers/transitions, or aliasing transient targets
- Read [references/shader-system.md](references/shader-system.md) - Load when compiling shaders, handling variants/permutations, reflection, or hot-reload
- Read [references/binding-model.md](references/binding-model.md) - Load when designing pipeline state and the descriptor/binding model, or going bindless
- Read [references/synchronization.md](references/synchronization.md) - Load when placing barriers, queue waits, fences, or reasoning about queue timelines
- Read [references/command-recording-and-frames.md](references/command-recording-and-frames.md) - Load when recording commands, threading recording, or sizing frames-in-flight
- Read [references/gpu-memory-strategy.md](references/gpu-memory-strategy.md) - Load when planning GPU memory tiers, sub-allocation, or uploads
- Read [references/vertex-assembly-skinning.md](references/vertex-assembly-skinning.md) - Load when designing vertex fetch, vertex packing, morph targets, or GPU skinning
- Read [references/gpu-compute-simulation.md](references/gpu-compute-simulation.md) - Load when simulating many elements (particles/agents) on the GPU with compute and indirect dispatch
- Read [references/hdr-output.md](references/hdr-output.md) - Load when adding HDR output, picking a swapchain color space/format, or encoding for the display
