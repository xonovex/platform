---
name: editor-viewport-guide
description: "Use when building the interactive 3D viewport that bridges a real-time renderer and an editor: GPU id-buffer object picking, selection outline/highlight rendering, and move/rotate/scale gizmos that work on any object plus the linear algebra behind them. Triggers on click-to-select in a 3D scene, reading back the picked object under the cursor, drawing a crisp selection outline, transform handles/gizmos, dragging an axis at a steep camera angle, constant screen-size handles, ray/plane projection, even when the user doesn't say 'viewport' or 'gizmo'."
---

# Editor viewport Guidelines

The interactive seam where a real-time renderer meets an editor: picking what the user clicks, highlighting what is selected, and manipulating it with transform gizmos. These features straddle two subsystems, so build a thin, render-pipeline-agnostic API the editor drives with plain data. Renderer plumbing (passes, id targets, readback primitives) belongs to gpu-rendering-guide; the generic per-object transform interface is owned by ecs-guide / data-model-guide; vector/matrix/quaternion types come from c99-game-opinionated-guide.

## Requirements

- A renderer that can run extra passes and expose a render-target read-back path (see gpu-rendering-guide).
- A way to identify each rendered object by a stable integer/handle (entity id or equivalent) that shaders can emit.
- Vector/matrix/quaternion math and screen↔world camera transforms (see c99-game-opinionated-guide).

## Essentials

- **A thin data seam** - The editor passes plain POD (transforms, colors, ids); the viewport layer owns all pipeline details, see [references/render-editor-integration.md](references/render-editor-integration.md)
- **Pick on the GPU** - Render object ids to a buffer and read back the pixel under the cursor instead of CPU ray-casting proxies, see [references/object-picking.md](references/object-picking.md)
- **Outline by id, not depth** - Write a selection id to a separate target and edge-detect it in a fullscreen pass, see [references/selection-highlighting.md](references/selection-highlighting.md)
- **Gizmos are generic** - Decouple handles from object type via a get/set-transform interface any component implements, see [references/manipulation-gizmos.md](references/manipulation-gizmos.md)
- **Do the geometry in screen space** - Project the cursor onto a screen-space axis, then lift back to world; never invent a 3D mouse ray and intersect, see [references/gizmo-math.md](references/gizmo-math.md)

## Integration

- **Per-viewport instances** - Picking and highlight state is created/destroyed per viewport; multiple viewports run independently, see [references/render-editor-integration.md](references/render-editor-integration.md)
- **Composite after post** - Editor overlays (grid, outline) draw after tone-mapping in sRGB so colors match the UI, see [references/render-editor-integration.md](references/render-editor-integration.md)
- **Opt-in per shader** - Any object's shader joins picking/selection by enabling a small shared shader feature, not by bespoke code, see [references/render-editor-integration.md](references/render-editor-integration.md)

## Picking and selection

- **Async readback** - Queue the id read after the frame; consume it a frame or two later — the latency is invisible, see [references/object-picking.md](references/object-picking.md)
- **Closest-wins, locked** - Use an atomic min on depth plus a tiny spinlock to keep the id consistent with the winning depth, see [references/object-picking.md](references/object-picking.md)
- **Gather-based edge detect** - Sample a neighborhood of the id target with `Gather`; alpha rises where neighbor ids differ, see [references/selection-highlighting.md](references/selection-highlighting.md)
- **Dim, don't drop, when occluded** - Compare selection depth (max over a few taps, to survive jitter) against scene depth and dim hidden outline, see [references/selection-highlighting.md](references/selection-highlighting.md)

## Gizmos and math

- **One gizmo, many components** - Resolve selection up the ownership chain to the highest-priority component that exposes a transform, see [references/manipulation-gizmos.md](references/manipulation-gizmos.md)
- **World vs local on a modifier** - `get_transform` returns both; apply the delta in world or local space depending on the user's mode, see [references/manipulation-gizmos.md](references/manipulation-gizmos.md)
- **Constant screen-size handles** - Scale handle geometry by view distance so it stays a fixed pixel size, see [references/gizmo-math.md](references/gizmo-math.md)
- **Robust deltas** - Store the drag-start parameter and apply only the difference; guard parallel/degenerate cases, see [references/gizmo-math.md](references/gizmo-math.md)

## Gotchas

- A 3D "mouse ray" built by assigning arbitrary z to the cursor and intersecting it with the axis is a skew-line distance problem, not a projection — it sends objects backward at steep camera angles; project in 2D screen space instead.
- Reading the id pixel synchronously stalls the GPU pipeline; queue an async read-back and accept the one-to-two-frame delay.
- Comparing the selection depth directly against a jittered (TAA) scene depth makes the outline shimmer; take the closest depth over a small neighborhood before comparing.
- With reverse-Z, "closest" is the maximum depth, not the minimum — flip your `min`/`max` and atomic comparisons accordingly.
- A naive id write (depth test then store) races between threads; without an atomic-min plus lock the stored id can belong to a farther surface than the stored depth.
- Letting each component draw and hit-test its own gizmo duplicates code and drifts behavior; keep gizmo rendering/interaction central and let components supply only transform get/set.
- Applying the absolute projected parameter as the position (instead of the delta from drag-start) snaps the object to the cursor on the first frame; always subtract the start value.
- Dividing by a near-zero denominator when the drag axis is nearly parallel to the view (or the line direction is near-zero) explodes the delta; clamp with an epsilon and fall back to no movement.

## Progressive Disclosure

- Read [references/render-editor-integration.md](references/render-editor-integration.md) - Load when designing the API seam between renderer and editor, per-viewport overlay/state, or where editor draws composite
- Read [references/object-picking.md](references/object-picking.md) - Load when implementing click-to-select, an id/color buffer, GPU read-back, or comparing against CPU ray picking
- Read [references/selection-highlighting.md](references/selection-highlighting.md) - Load when drawing a selection outline/silhouette, edge-detecting an id target, or handling occluded/jittered outlines
- Read [references/manipulation-gizmos.md](references/manipulation-gizmos.md) - Load when making move/rotate/scale gizmos work on arbitrary components via a generic transform interface
- Read [references/gizmo-math.md](references/gizmo-math.md) - Load when fixing gizmo drag math, screen-space projection, constant-size handles, ray/plane intersection, or degenerate cases
