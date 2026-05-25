# device-memory: Explicit GPU Device-Memory Allocation

**Guideline:** Allocate a small number of large device-memory blocks and sub-allocate resources out of them with correct placement alignment; choose the memory type by access pattern (DEVICE_LOCAL for GPU-resident data, HOST_VISIBLE|HOST_COHERENT for CPU-written data); upload static data through a staging buffer into device-local memory; keep dynamic per-frame data in persistently-mapped ring buffers — never one allocation per resource.

**Rationale:** On an explicit API the driver does not allocate for you; you pick the heap and place the resource. The total number of device allocations is hard-capped (often a few thousand) and each allocation is slow, so one `vkAllocateMemory` per buffer/image exhausts the limit and stalls. A sub-allocator carves resources out of a handful of big blocks: cheap, cap-friendly, and contiguous. Memory types differ physically — DEVICE_LOCAL is fast for the GPU but usually not CPU-mappable, while HOST_VISIBLE is CPU-writable but slower for the GPU to read. Static assets therefore want a staging copy: write to a HOST_VISIBLE staging buffer, then `vkCmdCopyBuffer`/`CopyBufferToImage` into a DEVICE_LOCAL resource the shaders actually sample. Dynamic per-frame data (uniforms, instance arrays) changes every frame, so it lives in a persistently-mapped HOST_VISIBLE ring buffer sized for frames-in-flight: map once, write a fresh sub-range each frame, never re-map. Long-lived sub-allocation fragments over time, so a defragmentation pass that relocates live resources keeps blocks compact.

**Techniques:**

- **Few big blocks** - Allocate large blocks (e.g. 64–256 MB); sub-allocate with a block/buddy/free-list allocator. Keeps allocation count tiny.
- **Memory type by use** - DEVICE_LOCAL for GPU-resident (textures, vertex/index, render targets); HOST_VISIBLE|HOST_COHERENT for CPU-written; DEVICE_LOCAL|HOST_VISIBLE (resizable BAR) when present, for direct GPU-visible CPU writes.
- **Placement alignment** - Honor the resource's reported alignment when placing it in a block; images and buffers report different requirements.
- **Staging upload** - CPU → HOST_VISIBLE staging buffer → copy command → DEVICE_LOCAL resource; the staging buffer is reusable/recyclable.
- **Persistent mapping** - Map a HOST_VISIBLE block once at creation and keep the pointer; avoids per-frame map/unmap cost.
- **Ring buffer** - One mapped buffer holding N per-frame sub-ranges; frame i writes range i; a fence guards reuse, see [references/command-buffers-and-frames.md](./command-buffers-and-frames.md).
- **Defragmentation** - Periodically relocate live allocations to compact blocks, then fix up the handles/descriptors that referenced them.
- **General-purpose allocator** - A reusable allocator over these techniques (block management, memory-type selection, defrag) so call sites request `(size, usage)` and never touch raw device memory.

**Example:**

```c
// Bad: one device allocation per resource -> hits the allocation cap, slow.
for (int i = 0; i < n; i++) vkAllocateMemory(dev, &info, NULL, &mem[i]);

// Good: sub-allocate from a few large blocks; placement honors alignment.
gpu_alloc a = gpu_malloc(allocator, req.size, req.alignment, MEM_DEVICE_LOCAL);
vkBindImageMemory(dev, image, a.block, a.offset);

// Static upload: HOST_VISIBLE staging -> device-local copy.
gpu_alloc stage = gpu_malloc(allocator, size, 1, MEM_HOST_VISIBLE | MEM_HOST_COHERENT);
memcpy(stage.mapped, pixels, size);                       // persistently mapped, no map/unmap
vkCmdCopyBufferToImage(cmd, stage.buffer, image, LAYOUT_TRANSFER_DST_OPTIMAL, 1, &region);
// barrier: TRANSFER_WRITE -> SHADER_READ, layout TRANSFER_DST -> SHADER_READ_ONLY

// Dynamic per-frame data: persistently-mapped ring, sub-range per frame-in-flight.
uint8_t *slot = ring.mapped + frame_index * ring.stride;  // fence on frame_index gates reuse
memcpy(slot, &per_frame_ubo, sizeof per_frame_ubo);
```

**Gotchas:**

- The device allocation count is capped; per-resource allocation works on small scenes then fails at scale — sub-allocate from the start.
- DEVICE_LOCAL memory is usually not CPU-mappable; writing it directly is invalid — go through staging.
- Without HOST_COHERENT you must `vkFlushMappedMemoryRanges` after writing and invalidate before reading; coherent memory skips this but can be slower for the GPU.
- Buffer and image alignment requirements differ and some implementations require buffer/image granularity separation within a block — query and respect both.
- Writing a ring-buffer sub-range still in use by an in-flight frame races the GPU; size the ring to frames-in-flight and gate on the per-frame fence.
- Defragmentation invalidates offsets/descriptors that referenced the moved resource; update every descriptor and view, or you sample stale memory.

**Related:** [references/command-buffers-and-frames.md](./command-buffers-and-frames.md), [references/synchronization.md](./synchronization.md), [references/render-graph.md](./render-graph.md), [references/pipelines-and-descriptors.md](./pipelines-and-descriptors.md)
