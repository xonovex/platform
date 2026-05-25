# Sources

## Explicit-API rendering architecture (cross-API)

- **URLs:**
  - Microsoft Direct3D 12 programming guide — https://learn.microsoft.com/en-us/windows/win32/direct3d12/directx-12-programming-guide
  - Apple Metal documentation — https://developer.apple.com/documentation/metal
  - WebGPU specification — https://www.w3.org/TR/webgpu/
  - Vulkan specification — https://registry.khronos.org/vulkan/specs/1.3/html/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Requirements, Architecture, Resources, Synchronization
  - Cross-API confirmation that pipeline-state objects, explicit barriers, frequency-grouped binding, bindless, frames-in-flight, and queue-side waits are general explicit-API concepts, not specific to one API
- **Aspects extracted:**
  - Pipeline objects + cache, binding frequency, bindless, inline constants → `references/binding-model.md`
  - Frames-in-flight, multi-threaded command recording, swapchain loop → `references/command-recording-and-frames.md`
  - Resource barriers (stage/access scopes), layout transitions, cross-queue waits, timeline values, fences → `references/synchronization.md`
  - Memory tiers (device-local vs host-visible), allocation limits, alignment, staging → `references/gpu-memory-strategy.md`

## FrameGraph: Extensible Rendering Architecture (GDC 2017)

- **URL:** https://www.gdcvault.com/play/1024612/FrameGraph-Extensible-Rendering-Architecture-in
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Architecture
  - The declarative-pass / derived-execution framing and transient resource aliasing
- **Aspects extracted:**
  - Passes declare reads/writes; graph derives order, barriers, and layout transitions; culls dead passes → `references/render-graph.md`
  - Transient render-target memory aliasing by disjoint lifetime → `references/render-graph.md`

## Render-graph / frame-graph technique writeups

- **URLs:**
  - "Render graphs and Vulkan — a deep dive" — https://themaister.net/blog/2017/08/15/render-graphs-and-vulkan-a-deep-dive/
  - "Organizing GPU Work with Directed Acyclic Graphs" — https://levelup.gitconnected.com/organizing-gpu-work-with-directed-acyclic-graphs-f3fd5f2c2af3
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Architecture, Gotchas
  - Practical per-frame graph compilation, barrier derivation, and aliasing edge cases
- **Aspects extracted:**
  - Topological sort, lifetime computation, automatic barrier/transition placement → `references/render-graph.md`
  - Async-compute/multi-queue edges needing cross-queue waits + ownership transfers → `references/render-graph.md`, `references/synchronization.md`

## GPU memory strategy

- **URLs:**
  - GPU memory allocator documentation — https://gpuopen.com/learn/vulkan-memory-allocator/
  - "Writing an efficient Vulkan renderer" — https://zeux.io/2020/02/27/writing-an-efficient-vulkan-renderer/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Resources, Essentials
  - Sub-allocation from large blocks, staging, persistent mapping, defragmentation (the strategy, not one API's calls)
- **Aspects extracted:**
  - Few large allocations + sub-allocation, allocation-count cap, placement alignment → `references/gpu-memory-strategy.md`
  - Staging upload to device-local, persistent mapping, ring buffers, defragmentation → `references/gpu-memory-strategy.md`
  - General allocator principle deferred to memory-management-guide → `references/gpu-memory-strategy.md`

## Shader compilation and reflection tooling

- **URLs:**
  - glslang — https://github.com/KhronosGroup/glslang
  - SPIRV-Reflect — https://github.com/KhronosGroup/SPIRV-Reflect
  - DirectX Shader Compiler (DXC) — https://github.com/microsoft/DirectXShaderCompiler
  - SPIR-V specification — https://registry.khronos.org/SPIR-V/specs/unified1/SPIRV.html
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Architecture
  - Offline compile to a binary intermediate, reflection-derived layouts, variants, caching
- **Aspects extracted:**
  - Source → binary intermediate offline, reflection for binding/layout info → `references/shader-system.md`
  - Permutations vs ubershader + specialization constants, pipeline keying, hot-reload, disk cache → `references/shader-system.md`, `references/binding-model.md`

## Refresh Workflow

1. Re-read the upstream source(s) above (spec sections, talk, allocator/tooling docs)
2. Diff against the prior pull (or scan for newly added sections / API revisions)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
