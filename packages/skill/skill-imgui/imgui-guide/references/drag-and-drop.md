# drag-and-drop: Build Drag-and-Drop from the Data Model

**Guideline:** Start from the data representation — a single global id naming the dragged object(s) — not from UI mechanics. The drop is a data-model mutation; the UI just reflects it on the next re-render.

**Rationale:** Drag state must outlive any single window's UI context (you drag _between_ windows), so it can't live in per-control or per-window state. Reducing the whole gesture to "what global id is being dragged?" makes initiation, hover-highlighting, type-checking, and the drop itself simple data operations. Designing the data first keeps the UI code thin and the cross-window case free.

**How to Apply:**

1. Keep one global field, `dragged_objects` (a global id; `0` = nothing dragging), with `start_dragging(id)` / `stop_dragging()` / `get_dragged_objects()`.
2. **Initiate** with a prepare-drag latch: on mouse-press over an item, arm `prepare_drag = item_id`; start the actual drag only once the cursor _leaves_ the item while still held; abort if released first.
3. **Drop**: query the global id, type-check it against the target's accepted type, highlight on valid hover, and on mouse-release commit the change and `stop_dragging()`.
4. At end of frame, if the mouse released over no valid target, cancel the drag — but only _after_ all drop targets have had a chance to process the release.
5. Represent the payload as a single object holding references to the dragged items; the drop mutates the data model, and the UI updates next frame.

**Example:**

```c
// Initiate: arm on press, start when the cursor leaves the item while held
if (ui->hover == item_id && ui->left_mouse_pressed) ui->prepare_drag = item_id;
if (!ui->left_mouse_is_down)                          ui->prepare_drag = 0;
if (ui->prepare_drag == item_id && ui->hover != item_id) {
    drag->start_dragging(object_id);
    ui->prepare_drag = 0;
}
// Drop target: type-check, highlight, commit on release
uint64_t d = drag->dragged_objects();
if (d && ui->hover == target_id && truth->object_type(tt, d) == ACCEPTED_TYPE) {
    draw_highlight(target_id);
    if (ui->left_mouse_released) { apply_drop(d, target_id); drag->stop_dragging(); }
}
```

**Counter-Example:** Reordering items within one list, where no cross-window transfer happens, can use simpler local state — the global-id model pays off precisely when the drag crosses contexts.

**Gotcha:** The end-of-frame cancel must run _after_ every drop target's release handling, or a legitimate drop is silently cancelled. Order matters more than anywhere else in IMGUI.

**Related:** [ids-and-state.md](./ids-and-state.md), [frame-delay.md](./frame-delay.md)
