# frame-delay: Resolve Ordering with a One-Frame Delay

## Guideline

When a decision depends on a control that hasn't been issued yet this frame (occlusion, focus transitions), record the intent into a "next" field and promote it at frame end, so the decision uses complete information one frame later.

## Rationale

Immediate mode issues controls in order and cannot look ahead: when deciding whether the mouse is over control A, it doesn't yet know that overlapping control B will be drawn later and should win. Computing the result during the frame and committing it at frame end (so the last-drawn / topmost control wins) sidesteps this without adding a retained hierarchy. The same pattern resolves tab focus and drag cancellation.

## How to Apply

1. While issuing controls, write the candidate into a `next_*` field (e.g. `next_hover = id` for any control under the cursor); the last write of the frame wins because it was drawn last/on top.
2. At frame end, promote `next_hover` → `hover`. Controls read `hover` (last frame's resolved value) this frame.
3. Track a layer for overlays (`next_hover_layer`) so a lower layer can't steal input from an overlay drawn above it.
4. Apply the same delay to focus transitions: Tab sets `focus_on_next`; the _next_ control to render claims focus. Shift-Tab records `tab_focus_on_id = last_id` and takes effect next frame.
5. Beware events that create a control which then reacts to the creating event — defer that reaction a frame too.

## Example

```c
// During the frame: every control under the cursor claims next_hover; last-drawn wins
if (rect_contains(r, ui->mouse)) ui->next_hover = id;
...
// End of frame: promote
ui->hover = ui->next_hover;
ui->next_hover = 0;
```

## Counter-Example

State that is fully determined the moment it's computed (a click on a control already known to be hovered) needs no delay — only look-ahead-dependent decisions do.

## Related

[ids-and-state.md](./ids-and-state.md), [events-and-focus.md](./events-and-focus.md)
