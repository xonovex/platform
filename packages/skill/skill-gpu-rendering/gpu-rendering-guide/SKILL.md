---
name: gpu-rendering-guide
description: "Use when building or reviewing a low-level GPU renderer on an explicit API (Vulkan/D3D12/Metal-class): render graphs/frame graphs, shader systems & permutations, GPU device-memory allocation, pipelines/descriptors, and GPU↔CPU synchronization. Triggers on prompts about render passes, resource barriers/transitions, transient/aliased render targets, SPIR-V/shader variants, descriptor sets/bindless, command buffers, frames-in-flight, semaphores/fences/timeline, staging uploads, even when the user doesn't say 'Vulkan'. Skip high-level web 3D (use threejs-guide), pure shading-language math with no pipeline/resource concerns, and CPU-only data layout (use data-oriented-design-guide)."
---

# GPU Rendering Guidelines (Explicit API)

## Requirements

- Target an explicit, low-level API (Vulkan 1.2+, D3D12, or Metal-class): the app owns memory, synchronization, and pipeline state.
- Assume validation layers + a frame debugger (RenderDoc-class) are available during development.

## Essentials

- **Declare, don't sequence** - Passes declare resource reads/writes; the graph derives order, barriers, and layout transitions, see [references/render-graph.md](references/render-graph.md)
- **Shaders are build artifacts** - Author → compile to SPIR-V offline → reflect for layouts, see [references/shader-system.md](references/shader-system.md)
- **You own GPU memory** - Sub-allocate from a few large allocations; never one allocation per resource, see [references/device-memory.md](references/device-memory.md)
- **You own synchronization** - Nothing is implicit; barriers, semaphores, and fences are explicit, see [references/synchronization.md](references/synchronization.md)

## Architecture

- **Render graph** - Per-frame DAG that prunes dead passes, orders work, and aliases transient targets, see [references/render-graph.md](references/render-graph.md)
- **Shader system** - Permutations vs ubershaders, PSO keyed by shader+state, hot-reload, caching, see [references/shader-system.md](references/shader-system.md)
- **Command recording** - One pool per thread per frame, primary/secondary, multi-threaded record, see [references/command-buffers-and-frames.md](references/command-buffers-and-frames.md)
- **Frames in flight** - Double/triple-buffer per-frame resource sets behind a fence so the CPU can't outrun the GPU, see [references/command-buffers-and-frames.md](references/command-buffers-and-frames.md)

## Resources

- **Device memory** - DEVICE_LOCAL vs HOST_VISIBLE|HOST_COHERENT, placement alignment, defrag, see [references/device-memory.md](references/device-memory.md)
- **Uploads** - Staging buffer → device-local copy; persistent-mapped ring buffers for per-frame data, see [references/device-memory.md](references/device-memory.md)
- **Pipelines** - Precompile and cache PSOs; avoid first-use compile stalls, see [references/pipelines-and-descriptors.md](references/pipelines-and-descriptors.md)
- **Descriptors** - Layouts by update frequency, bindless arrays + handles, push constants for tiny data, see [references/pipelines-and-descriptors.md](references/pipelines-and-descriptors.md)

## Synchronization

- **Pipeline barriers** - src/dst stage+access masks plus image-layout transitions, see [references/synchronization.md](references/synchronization.md)
- **Queue↔queue** - Semaphores order submissions; timeline semaphores carry monotonic values, see [references/synchronization.md](references/synchronization.md)
- **GPU→CPU** - Fences gate frame resource reuse and readback, see [references/synchronization.md](references/synchronization.md)

## Gotchas

- A barrier that is too broad (e.g. ALL_COMMANDS → ALL_COMMANDS) is correct but serializes the GPU; scope stage/access masks to what actually waits.
- Forgetting an image-layout transition is undefined behavior even when the data is "obviously" ready — layout is part of the contract, not just a hint.
- One allocation per resource exhausts the (small, capped) device allocation limit and is slow; sub-allocate from large blocks.
- HOST_COHERENT skips explicit flush/invalidate but is not free; large dynamic data still wants a device-local copy via staging.
- Recording into a command buffer whose previous submission's fence has not signaled corrupts in-flight GPU work — gate reuse on the fence.
- Re-using a transient target the graph aliased to another lifetime, then reading it later, returns garbage; lifetimes must not overlap.

## Progressive Disclosure

- Read [references/render-graph.md](references/render-graph.md) - Load when ordering passes, automating barriers/transitions, or aliasing transient targets
- Read [references/shader-system.md](references/shader-system.md) - Load when compiling shaders, handling variants/permutations, reflection, or hot-reload
- Read [references/device-memory.md](references/device-memory.md) - Load when allocating GPU memory, choosing heaps, or uploading data
- Read [references/pipelines-and-descriptors.md](references/pipelines-and-descriptors.md) - Load when building pipeline objects, descriptor sets, or going bindless
- Read [references/synchronization.md](references/synchronization.md) - Load when placing barriers, semaphores, fences, or reasoning about queues
- Read [references/command-buffers-and-frames.md](references/command-buffers-and-frames.md) - Load when recording command buffers, threading recording, or sizing frames-in-flight
