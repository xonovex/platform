# gpu-memory-strategy: GPU Memory Strategy and Uploads

**Guideline:** Allocate a small number of large GPU memory blocks and sub-allocate resources out of them with correct placement alignment; choose the memory tier by access pattern (device-local for GPU-resident data, host-visible for CPU-written data); upload static data through a staging buffer into device-local memory; keep dynamic per-frame data in persistently-mapped ring buffers — never one allocation per resource.

**Rationale:** On an explicit API the driver does not allocate for you; you pick the tier and place the resource. The total number of device allocations is hard-capped (often a few thousand) and each allocation is slow, so one allocation per buffer/image exhausts the limit and stalls. A sub-allocator carves resources out of a handful of big blocks: cheap, cap-friendly, and contiguous — the same arena/pool principle as CPU allocators, see memory-management-guide for the underlying idea. Memory tiers differ physically: device-local is fast for the GPU but usually not CPU-mappable, while host-visible is CPU-writable but slower for the GPU to read. Static assets therefore want a staging copy: write to a host-visible staging buffer, then copy into a device-local resource the shaders actually sample. Dynamic per-frame data (uniforms, instance arrays) changes every frame, so it lives in a persistently-mapped host-visible ring buffer sized for frames-in-flight: map once, write a fresh sub-range each frame, never re-map. Long-lived sub-allocation fragments over time, so a defragmentation pass that relocates live resources keeps blocks compact. (Per-API specifics — exact memory-type flags, alignment queries, the staging copy command — belong in the per-API skill; this is the strategy.)

**Techniques:**

- **Few big blocks** - Allocate large blocks (e.g. 64–256 MB); sub-allocate with a block/buddy/free-list allocator. Keeps allocation count tiny. (Same arena/pool idea as memory-management-guide, applied to GPU memory.)
- **Tier by use** - Device-local for GPU-resident (textures, vertex/index, render targets); host-visible for CPU-written; a device-local + host-visible tier (resizable BAR) when present, for direct GPU-visible CPU writes.
- **Placement alignment** - Honor each resource's reported alignment when placing it in a block; images and buffers report different requirements.
- **Staging upload** - CPU → host-visible staging buffer → copy command → device-local resource; the staging buffer is reusable/recyclable.
- **Persistent mapping** - Map a host-visible block once at creation and keep the pointer; avoids per-frame map/unmap cost.
- **Ring buffer** - One mapped buffer holding N per-frame sub-ranges; frame i writes range i; a fence guards reuse, see [references/command-recording-and-frames.md](./command-recording-and-frames.md).
- **Defragmentation** - Periodically relocate live allocations to compact blocks, then fix up the handles/views that referenced them.
- **General-purpose allocator** - A reusable allocator over these techniques (block management, tier selection, defrag) so call sites request `(size, usage)` and never touch raw device memory.
- **Visual allocator debugging** - Give every allocation a debug tag and build a simple visualization of block occupancy early; GPU memory waste (e.g. power-of-two rounding in a buddy allocator, or blocks hoarded and never released) is otherwise invisible and easy to ship.

**Example:**

```c
// Bad: one device allocation per resource -> hits the allocation cap, slow.
for (int i = 0; i < n; i++) gpu_allocate_memory(dev, &info, &mem[i]);

// Good: sub-allocate from a few large blocks; placement honors alignment.
// (Neutral pseudocode; concrete bind/copy calls live in the per-API skill.)
gpu_alloc a = gpu_malloc(allocator, req.size, req.alignment, MEM_DEVICE_LOCAL);
gpu_bind_image_memory(image, a.block, a.offset);

// Static upload: host-visible staging -> device-local copy.
gpu_alloc stage = gpu_malloc(allocator, size, 1, MEM_HOST_VISIBLE);
memcpy(stage.mapped, pixels, size);                  // persistently mapped, no map/unmap
cmd_copy_buffer_to_image(cmd, stage.buffer, image, region);
// barrier: transfer-write -> shader-read, layout transfer-dst -> shader-read-only

// Dynamic per-frame data: persistently-mapped ring, sub-range per frame-in-flight.
uint8_t *slot = ring.mapped + frame_index * ring.stride;  // fence on frame_index gates reuse
memcpy(slot, &per_frame_ubo, sizeof per_frame_ubo);
```

**Gotchas:**

- The device allocation count is capped; per-resource allocation works on small scenes then fails at scale — sub-allocate from the start.
- Device-local memory is usually not CPU-mappable; writing it directly is invalid — go through staging.
- Without host-coherent memory you must flush after writing and invalidate before reading; coherent memory skips this but can be slower for the GPU.
- Buffer and image alignment requirements differ and some implementations require buffer/image granularity separation within a block — query and respect both.
- Writing a ring-buffer sub-range still in use by an in-flight frame races the GPU; size the ring to frames-in-flight and gate on the per-frame fence.
- Defragmentation invalidates offsets/views that referenced the moved resource; update every binding and view, or you sample stale memory.
- A buddy/power-of-two block allocator rounds each request up to the next power of two; a 9 MB texture can quietly consume a 16 MB slot. Measure occupancy before assuming the allocator is tight.

**Related:** [references/command-recording-and-frames.md](./command-recording-and-frames.md), [references/synchronization.md](./synchronization.md), [references/render-graph.md](./render-graph.md), [references/binding-model.md](./binding-model.md)
