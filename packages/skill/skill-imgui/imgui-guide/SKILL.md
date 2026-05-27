---
name: imgui-guide
description: "Use when designing or implementing an immediate-mode GUI (IMGUI): batching the whole UI into one draw call with compact primitive buffers, keying controls by stable IDs, resolving ordering with frame-delayed state, keyboard focus / responder chains / event trickling, drag-and-drop, per-monitor DPI scaling, string localization, and screen-reader accessibility. Triggers on immediate-mode widgets, retained-vs-immediate UI, hover/active/focus IDs, UI draw batching, ImGui-style code, localizing UI strings, accessibility/screen-reader support for an IMGUI, even when the user doesn't say 'IMGUI'. Skip retained-mode / DOM / web component frameworks (use react-guide or frontend-design), GPU command-submission internals (use gpu-rendering-guide), and game-entity design (use ecs-guide)."
---

# Immediate-Mode GUI Guidelines

Architecture for an immediate-mode GUI: the UI is re-issued every frame from application data, controls are identified by id rather than retained objects, and the few things that must persist (hover, focus, drag, DPI) live in a small, deliberate state store. For how the resulting vertex/index data reaches the GPU, see **gpu-rendering-guide**.

## Requirements

- The UI is rebuilt every frame; widgets return their result inline (`if (button(id, ...))`), they are not retained objects.
- A small per-frame and cross-frame state store holds only what immediate mode cannot recompute: hover, active, focus, drag, layout.

## Essentials

- **One draw call** - Pack all widgets into compact primitive buffers and synthesize vertices in the shader; encode clip and texture so no state switches are needed, see [references/draw-batching.md](references/draw-batching.md)
- **Stable unique IDs** - Key hover, active, focus, drag, and tab order by per-control id — the substitute for retained widget objects, see [references/ids-and-state.md](references/ids-and-state.md)
- **Frame delay resolves ordering** - Immediate mode can't see later controls; defer hover/focus decisions one frame so the last-drawn control wins, see [references/frame-delay.md](references/frame-delay.md)
- **Process events in end\_\*(), consume by clearing** - Trickle keyboard events through a responder chain; one event per frame, see [references/events-and-focus.md](references/events-and-focus.md)
- **Drag-and-drop is a data operation** - Hold a single global drag id; the drop is a data-model mutation, not UI plumbing, see [references/drag-and-drop.md](references/drag-and-drop.md)
- **Own DPI** - Work in virtual coordinates, scale at the edges (vertex shader, rect conversion), key the font atlas by DPI, see [references/dpi-scaling.md](references/dpi-scaling.md)

## Localization & Accessibility

- **Hash the source string as the key** - Wrap visible text in `LOCALIZE(...)`, resolve per frame through a swappable localizer that falls back to the source string, and audit coverage with a pseudo-localization mode, see [references/localization.md](references/localization.md)
- **Build a side semantic tree while drawing** - As each control is drawn, register its role/label/rect/state into a per-frame list, then expose that list to platform accessibility APIs and automation — immediate mode does not preclude a screen-reader-visible tree, see [references/accessibility.md](references/accessibility.md)

## Gotchas

- Immediate mode cannot know whether a control drawn _later_ this frame will occlude the current one — resolve hover with a one-frame delay (compute `next_hover`, promote at frame end) so the topmost/last-drawn control wins.
- An input event can create a new control mid-frame that then reacts to the very event that created it; a one-frame delay is sometimes the only clean fix.
- Two controls sharing an id silently merge their hover/active/focus state — derive ids from a stable source (data object id, loop index folded into a scope), never from screen position.
- `SetProcessDpiAwareness(PROCESS_PER_MONITOR_DPI_AWARE)` must run _before_ any window is created, and DPI must be part of the font-atlas cache key or text renders blurry.
- Drop targets must consume the mouse-release _before_ the end-of-frame step that cancels a drag, or valid drops are lost.
- Hashing the source string as a localization key means editing the English copy silently orphans the existing translation — treat string edits as re-translation work, and run pseudo-localization to catch unmarked strings.
- A control drawn but not registered into the accessibility list is invisible to assistive tech; one registered but not drawn is a ghost — build the side list in lockstep with drawing, and actually feed it to the platform accessibility API.

## Example

```c
// A button: stateless to issue, identified by a stable id; state lives in the ui context.
bool button(tm_ui_o *ui, uint64_t id, rect_t r, const char *label) {
    if (rect_contains(r, ui->mouse))
        ui->next_hover = id;                 // resolved at frame end (last-drawn wins)
    const bool clicked = ui->hover == id && ui->left_mouse_released;
    draw_rect(ui->buffer, r, ui->hover == id ? COLOR_HOT : COLOR_NORMAL); // into shared primitive buffer
    draw_text(ui->buffer, r, label);
    return clicked;
}
```

## Progressive Disclosure

- Read [references/draw-batching.md](references/draw-batching.md) - Load when designing how UI geometry is buffered and submitted: primitive buffers, index encoding, single draw call
- Read [references/ids-and-state.md](references/ids-and-state.md) - Load when deciding what state persists across frames and how controls are identified
- Read [references/frame-delay.md](references/frame-delay.md) - Load when fixing hover/focus/occlusion ordering bugs that immediate mode can't resolve within one frame
- Read [references/events-and-focus.md](references/events-and-focus.md) - Load when adding keyboard focus, responder chains, tab order, or event trickling/consumption
- Read [references/drag-and-drop.md](references/drag-and-drop.md) - Load when implementing drag-and-drop, especially across windows
- Read [references/dpi-scaling.md](references/dpi-scaling.md) - Load when handling high-DPI / per-monitor scaling, font atlases, or virtual-vs-pixel coordinates
- Read [references/localization.md](references/localization.md) - Load when localizing UI strings: extraction, source-string-as-key hashing, swappable localizer, pseudo-localization
- Read [references/accessibility.md](references/accessibility.md) - Load when adding screen-reader support, exposing UI semantics, automation, zoom, or high-contrast themes to an immediate-mode UI
