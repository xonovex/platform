# Sources

## Vulkan Specification and Vulkan-Guide

- **URLs:**
  - Vulkan specification — https://registry.khronos.org/vulkan/specs/1.3/html/
  - Khronos Vulkan-Guide — https://docs.vulkan.org/guide/latest/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Requirements, Synchronization, Resources
  - Explicit ownership of memory, synchronization, and pipeline state; primitive semantics
- **Aspects extracted:**
  - Pipeline barriers (stage/access masks), image-layout transitions, semaphores, timeline semaphores, fences, events, queue families → `references/synchronization.md`
  - Command pools/buffers, primary vs secondary, pool reset → `references/command-buffers-and-frames.md`
  - Memory types/heaps (DEVICE_LOCAL vs HOST_VISIBLE|HOST_COHERENT), allocation limits, alignment → `references/device-memory.md`
  - Descriptor set layouts, descriptor indexing / update-after-bind, push constants, pipeline cache → `references/pipelines-and-descriptors.md`

## FrameGraph: Extensible Rendering Architecture in Frostbite (GDC 2017)

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
  - Async-compute/multi-queue edges needing semaphores + ownership transfers → `references/render-graph.md`, `references/synchronization.md`

## GPU memory allocator documentation (AMD VMA)

- **URL:** https://gpuopen.com/learn/vulkan-memory-allocator/ and https://gpuopen-librariesandsdks.github.io/VulkanMemoryAllocator/html/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Resources, Essentials
  - Sub-allocation from large blocks, staging, persistent mapping, defragmentation
- **Aspects extracted:**
  - Few large allocations + sub-allocation, allocation-count cap, placement alignment → `references/device-memory.md`
  - Staging upload to device-local, persistent mapping, ring buffers, defragmentation → `references/device-memory.md`

## Shader compilation and reflection tooling

- **URLs:**
  - glslang — https://github.com/KhronosGroup/glslang
  - SPIRV-Reflect — https://github.com/KhronosGroup/SPIRV-Reflect
  - SPIR-V specification — https://registry.khronos.org/SPIR-V/specs/unified1/SPIRV.html
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Architecture
  - Offline compile, reflection-derived layouts, variants, caching
- **Aspects extracted:**
  - Source → SPIR-V offline, reflection for descriptor/layout info → `references/shader-system.md`
  - Permutations vs ubershader + specialization constants, PSO keying, hot-reload, disk cache → `references/shader-system.md`, `references/pipelines-and-descriptors.md`

## Explicit-API rendering architecture references

- **URLs:**
  - Microsoft Direct3D 12 programming guide — https://learn.microsoft.com/en-us/windows/win32/direct3d12/directx-12-programming-guide
  - Apple Metal documentation — https://developer.apple.com/documentation/metal
  - "Writing an efficient Vulkan renderer" — https://zeux.io/2020/02/27/writing-an-efficient-vulkan-renderer/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Architecture, Resources, Synchronization
  - Cross-API confirmation that PSOs, explicit barriers, descriptor binding, and frames-in-flight are general explicit-API concepts, not Vulkan-only
- **Aspects extracted:**
  - PSO precompilation/caching, descriptor frequency, bindless, push constants → `references/pipelines-and-descriptors.md`
  - Frames-in-flight, multi-threaded command recording, swapchain loop → `references/command-buffers-and-frames.md`

## Refresh Workflow

1. Re-read the upstream source(s) above (spec sections, talk, allocator/tooling docs)
2. Diff against the prior pull (or scan for newly added sections / API revisions)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
