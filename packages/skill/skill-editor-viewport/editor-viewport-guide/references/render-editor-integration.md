# render-editor-integration: The Seam Between a Real-Time Renderer and an Interactive Editor

## Guideline

Bridge the renderer and the editor with a thin, render-pipeline-agnostic API: the editor describes what it wants with plain old data (transforms, colors, ids), and a viewport layer translates that into passes, draws, and read-backs — so editor code never touches command buffers, constant buffers, or pipeline state, and a single feature can serve many viewports at once.

## Rationale

Picking, selection highlighting, and overlay rendering (grids, gizmos) live in the borderland between two subsystems that change at different rates and are owned by different code. If editor code reaches directly into the renderer, every pipeline change ripples into tools and every tool couples to one rendering backend. A narrow data seam inverts that: the editor populates a small struct and calls one function; the viewport layer decides how to realize it. This keeps the editor portable across rendering backends, lets the renderer evolve freely, and makes each feature trivially multi-viewport because the per-viewport state lives behind the seam, not in the editor. Compositing editor overlays after tone-mapping and in the same sRGB color space as the rest of the UI keeps their colors consistent and predictable regardless of the scene's HDR pipeline.

## How to Apply

1. Define a per-feature POD descriptor the editor fills in (e.g. a grid's transform/cell-size/color, a selection set of ids) — no handles to GPU resources, no pipeline knobs.
2. Expose one `render`/`update` entry point per feature that takes the descriptor plus opaque renderer context handles and does all the buffer writes and draw-call issuing internally.
3. Create per-viewport instances of stateful features (picking, highlight) with explicit `create`/`destroy`; never share mutable viewport state through globals.
4. Split CPU-side bookkeeping (`update_cpu`: consume read-backs, queue requests) from GPU-side scheduling (`update_gpu`: clear, fill constant buffers, queue passes), called once per frame.
5. Composite editor overlays after post-processing/tone-mapping, in 8-bit sRGB, so their colors match the UI exactly.
6. Make objects opt into editor features by enabling a small shared shader feature (a couple of lines), not by writing bespoke per-object code.

## Example

```c
// The editor only ever fills in plain data; the viewport layer owns the GPU.
typedef struct viewport_grid_t {
  mat44_t  transform;   // world placement
  float    grid_size;   // extent
  float    cell_size;   // smallest cell
  color_srgb_t thin, thick; // sRGB, same space as the UI
} viewport_grid_t;

// One entry point. Editor passes data + opaque renderer context; nothing else.
void grid_render(const viewport_grid_t *g, render_context_o *ctx,
                 shader_o *grid_shader, res_cmd_buf_o *res, cmd_buf_o *cmd);

// Stateful features are per-viewport, with explicit lifetime + a cpu/gpu split.
picking_o *picking_create(viewport_o *vp);
void       picking_destroy(picking_o *p);
void       picking_update_cpu(picking_o *p);   // consume read-back, queue request
void       picking_update_gpu(picking_o *p, cmd_buf_o *cmd); // clear, fill cbuf, queue read
```

## Gotchas

- Reaching from editor code into command/constant buffers couples tools to one backend and breaks on every pipeline change — keep the seam to plain data + opaque handles.
- Overlays drawn before tone-mapping inherit HDR/exposure and look wrong; composite after post in sRGB so colors match the UI.
- Sharing picking/highlight state across viewports through a global produces cross-viewport bleed; instantiate per viewport with `create`/`destroy`.
- Doing all work in one per-frame function couples read-back consumption to GPU scheduling; split `update_cpu` from `update_gpu`.
- Per-object bespoke integration for picking/selection does not scale; route every object through one shared, opt-in shader feature.

## Related

[references/object-picking.md](./object-picking.md), [references/selection-highlighting.md](./selection-highlighting.md), [references/manipulation-gizmos.md](./manipulation-gizmos.md), **gpu-rendering-guide**
