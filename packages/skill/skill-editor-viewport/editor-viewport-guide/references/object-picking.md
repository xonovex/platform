# object-picking: GPU Id-Buffer Picking and Read-Back

## Guideline

To find what the user clicked, render each object's stable id into a GPU buffer during a picking pass and read back the single pixel under the cursor asynchronously — instead of CPU ray-casting against physics or acceleration-structure proxies — so picking is pixel-perfect, matches exactly what is drawn (alpha masks, skinning, deformation, voxels), and needs no separate spatial structure.

## Rationale

CPU ray picking needs a proxy for every pickable thing: a physics shape, a bounding volume, or a BVH that must be rebuilt for deformable and procedurally generated geometry. It silently fails for anything without a proxy (alpha-cut foliage, GPU-skinned characters, voxel terrain) and couples picking to physics. Rendering ids is the opposite: whatever the GPU rasterizes is exactly what can be picked, at the resolution the user sees, for "a couple of lines" of shared shader code. The cost is a read-back from GPU to CPU, which must be asynchronous — stalling to read a pixel the same frame flushes the pipeline. Queuing the read and consuming it a frame or two later hides the latency completely from the user. Because the picking shader runs over the same rasterization as the main scene, depth ordering is free: with a per-pixel closest-depth test the picked id is always the front-most surface under the cursor.

## How to Apply

1. Give every rendered object a stable integer id (entity id / handle) and pass it as a shader constant so it is already available where pixels are shaded.
2. Add a small shared shader feature that, when enabled, writes `{depth, id}` for the front-most surface — usually into a tiny structured buffer (one record), not a full-screen target.
3. In that feature: reject pixels below an opacity threshold (so you can click through thin/transparent surfaces), reject pixels not under the cursor, then keep the entry only if it is closer than the stored depth.
4. Serialize concurrent writes: `InterlockedMin` the depth, then take a tiny spinlock (e.g. a sign bit) with `InterlockedCompareExchange` before storing the id, so the stored id always belongs to the winning depth.
5. Per frame: `update_cpu` checks for a completed read-back and, on a click, queues a request with the cursor position and opacity threshold; `update_gpu` clears the buffer, updates constants, and queues the async read-back after the scene pass.
6. Consume the result when it arrives (a frame or two later) and turn the id into a selection.

## Example

```hlsl
// One shared shader feature, enabled only for the picking pass.
struct pick_t { float depth; uint64_t id; };          // 12 bytes, one record
RWStructuredBuffer<pick_t> pick;

void update_picking_buffer(float opacity, float depth, uint2 px, uint64_t id) {
  if (opacity < 0.5) return;                            // click through transparency
  if (any(px != g_cursor_px)) return;                   // only the pixel under cursor
  // Reverse-Z: closer == larger depth -> use InterlockedMax instead of Min.
  uint orig;
  InterlockedMin(pick[0].depth_bits, asuint(depth), orig);
  if (asfloat(orig) <= depth) return;                   // someone closer already won
  // Tiny spinlock so id stays consistent with the winning depth:
  // CAS a lock bit, write id, release. (Spin until acquired.)
  store_id_under_lock(id);
}
```

```c
// CPU side: async read-back, consumed later — never block the frame on it.
void picking_update_cpu(picking_o *p) {
  if (readback_ready(p)) p->picked_id = read_pick_record(p).id; // 1-2 frames late
  if (clicked) queue_request(p, cursor_px, /*opacity*/ 0.5f);
}
```

## Gotchas

- A synchronous read of the picked pixel stalls the GPU/CPU pipeline; always queue an async read-back and accept the small delay (it is not user-visible).
- The naive "depth-test then store" races: two threads can pass the depth test and the last writer wins, leaving an id that does not match the stored depth — gate the id write with an atomic-min plus a lock.
- Reverse-Z inverts "closest": use `InterlockedMax` / `>=` where a forward-Z buffer would use min/`<=`.
- Skipping the opacity threshold makes you pick transparent decals and foliage cards instead of what's behind them; reject sub-threshold pixels.
- Picking against physics/bounding proxies misses anything without a proxy (skinned, deformed, voxel, alpha-cut) — render ids so picking matches what is actually drawn.
- Extending the record to also store cursor-pixel distance enables fuzzy/nearest selection for gizmo handles, but keep that out of the hot path unless needed.

## Related

[references/selection-highlighting.md](./selection-highlighting.md), [references/render-editor-integration.md](./render-editor-integration.md), **gpu-rendering-guide**
