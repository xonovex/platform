# Sources

## Game-engine development blog (archive)

- **URL:** https://archive-host.github.io/blog_archive/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → all sections
  - The immediate-mode rendering, identity, frame-delay, event, drag, and DPI model
- **Aspects extracted:**
  - "One Draw Call UI" — stateless drawing layer, single draw call, overlay buffer appended with index offset, three layers (drawing / UI / retained docking) → `references/draw-batching.md`
  - "UI rendering using Primitive Buffers" — compact tightly-packed primitives, 32-bit index encoding (type / corner / offset), shader-side vertex synthesis, clip offset stored in primitive → `references/draw-batching.md`
  - "Keyboard Focus and Event Trickling in Immediate Mode GUIs" — responder chain array, scope begin/end stack, trickling in `end_*()`, consume-by-clearing, tab order via frame delay, one-event-per-frame → `references/events-and-focus.md`, `references/frame-delay.md`
  - "Implementing drag-and-drop in an IMGUI" — data-representation-first, global drag id, prepare-drag latch, drop type-check, end-of-frame cancel ordering → `references/drag-and-drop.md`
  - "DPI-aware IMGUI" — per-monitor awareness, virtual vs pixel coordinates, DPI-keyed font atlas (stb_truetype), scale in vertex shader, viewport virtual-resolution default → `references/dpi-scaling.md`
  - Last-drawn-wins hover via `next_hover` promoted at frame end → `references/frame-delay.md`, `references/ids-and-state.md`

## Immediate-mode GUI prior art

- **URLs:**
  - Casey Muratori, "Immediate-Mode Graphical User Interfaces" (2005) — https://caseymuratori.com/blog_0001
  - Dear ImGui — https://github.com/ocornut/imgui
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Requirements, Essentials
  - The foundational immediate-mode model and id-based control identity
- **Aspects extracted:**
  - Widgets re-issued each frame, return result inline; id-keyed hot/active state → `references/ids-and-state.md`
  - Id-stack scoping to avoid collisions → `references/ids-and-state.md`

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
