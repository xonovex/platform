# Sources

## Game-engine development blog (archive)

- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Integration, Picking and selection, Gizmos and math, Gotchas
  - The whole worked instance of the renderer↔editor viewport seam: the thin POD API, GPU id-buffer picking with async read-back, id-edge-detect selection outlines, a queried generic transform interface for gizmos, and the screen-space projection math that repairs axis dragging.
- **Aspects extracted:**
  - "The borderland between rendering and editor — Part 1" — features living between renderer and editor; a thin POD descriptor + single `render` entry point; per-viewport multi-grid support; compositing overlays after tone-mapping in sRGB to match the UI; CPU/GPU update split → `references/render-editor-integration.md`
  - "The borderland between rendering and editor — Part 2: Picking" — GPU id-buffer picking vs physics/CPU-ray picking; `{depth, id}` record; opacity threshold + cursor-pixel test + closest-depth test; `InterlockedMin` plus a sign-bit spinlock (`InterlockedCompareExchange`) so id matches the winning depth; per-viewport `create`/`destroy`, `update_cpu`/`update_gpu`, async read-back consumed a frame or two later; passing entity id as a constant; fuzzy distance-based selection extension → `references/object-picking.md`
  - "The borderland between rendering and editor — Part 3: Selection highlighting" — single-channel selection target written via a shared `selection` shader feature (low 8 bits of entity id); fullscreen `GatherRed` 3x3 id edge detection (counter-clockwise XYZW tap order) accumulating alpha from differing neighbors; reverse-Z depth, neighborhood `max` of selection depth to survive TAA jitter, linearize-and-compare to scene depth, dim occluded outline to 30% rather than discard; zero-alpha early-out; per-color future extension → `references/selection-highlighting.md`
  - "Making the move/rotate/scale gizmos work with any component" — queried function-pointer interface (not inheritance) registered by name; `gizmo_get_transform` returning both world and local; `gizmo_set_transform` with an `undo_scope` (0 = in-progress, non-zero = commit); walking the ownership chain on selection; `gizmo_priority` to disambiguate multiple transformable components; spline control points / wire endpoints as use cases; keep gizmo rendering/interaction central → `references/manipulation-gizmos.md`
  - "Linear algebra shenanigans — gizmo repair" — the backward-drag bug from a fake 3D mouse ray (skew-line distance, not projection); the fix: pull the axis into screen space, 2D point-on-line projection, lift back to world, line-line intersect; apply the delta from a stored start parameter; inverse-parent-rotation for local mode; `point_on_line_projection_2d` (`uu < 0.00001f` guard) and `line_line_intersection` (`uv*uv > 1 - 1e-5f` parallel guard, `1 - uv^2` denominator); perspective divide; constant screen-size handles → `references/gizmo-math.md`

## Real-time editor viewport interaction (general prior art)

- **URLs:**
  - Jump flood / silhouette outline techniques — https://bgolus.medium.com/the-quest-for-very-wide-outlines-ba82ed442cd9
  - GPU object picking via id buffers — https://learnopengl.com/Guest-Articles/2021/Scene/Mouse-Picting
  - Ray–plane / ray–line intersection and screen-to-world rays — https://www.scratchapixel.com/lessons/3d-basic-rendering/minimal-ray-tracer-rendering-simple-shapes/ray-plane-and-ray-disk-intersection.html
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Picking and selection, Gizmos and math
  - Cross-confirmation that id-buffer picking, edge-detected/silhouette outlines, and screen-space ray projection are general editor-viewport techniques, not specific to one engine.
- **Aspects extracted:**
  - Id/color buffer picking and single-pixel read-back as a general technique → `references/object-picking.md`
  - Outline-by-edge-detection and silhouette compositing as general techniques → `references/selection-highlighting.md`
  - Screen-to-world ray construction and ray/line intersection fundamentals → `references/gizmo-math.md`

## Refresh Workflow

1. Re-read the upstream source(s) above (the five borderland/gizmo posts and the general prior-art links)
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
