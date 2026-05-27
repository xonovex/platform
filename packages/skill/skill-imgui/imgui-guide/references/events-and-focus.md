# events-and-focus: Keyboard Focus, Responder Chains, Event Trickling

**Guideline:** Track keyboard focus as a responder chain (root → focused control) built from an explicit scope stack, trickle events through it by processing them in the `end_*()` functions, and consume an event by clearing its flag. Feed one input event per frame.

**Rationale:** Immediate mode has no widget tree to walk for focus, so the retained-mode concepts (first responder, responder chain, event bubbling) have to be reconstructed from ids and explicit scope tracking. Processing events in `end_*()` runs them in reverse of issue order, giving outer/earlier controls first refusal; clearing the flag marks the event consumed so nothing else reacts. Handling one event per frame avoids juggling several events' worth of state mutation within a single frame.

**How to Apply:**

1. Maintain a fixed-size **responder chain** array (root → first responder); the last entry is the finest-grained focused control.
2. Build it from a **scope stack** via `begin_responder_scope(id)` / `end_responder_scope(id)`. When a control gains focus, capture its chain from the current scope stack.
3. **Trickle**: handle keyboard events in `end_*()` (reverse of call order). A control acts only if `in_responder_chain(my_id)`, then **consumes** by clearing the key flag (`key_pressed[K] = false`).
4. **Tab order** is implicit from render order, resolved with a one-frame delay (Tab → `focus_on_next`; Shift-Tab → `tab_focus_on_id = last_id`). See frame-delay.
5. Test conditions _before_ the mutations they depend on, or a control will focus itself the same frame.

**Example:**

```c
// Trickle + consume in end_*(): only the focused chain reacts, and it eats the event
void end_pane(tm_ui_o *ui, uint64_t my_id) {
    if (ui->key_pressed[KEY_PGDN] && in_responder_chain(ui, my_id)) {
        scroll_page_down(my_id);
        ui->key_pressed[KEY_PGDN] = false; // consumed; outer controls won't also scroll
    }
    end_responder_scope(ui, my_id);
}
```

**Counter-Example:** A UI with no keyboard interaction (pure mouse tool palette) needs neither responder chains nor trickling — hover/active ids suffice.

**Gotcha:** Overlapping sibling controls break a geometric/rect-based chain; you need the explicit scope stack, not "whatever is under the mouse," to rebuild the chain on focus change.

**Related:** [frame-delay.md](./frame-delay.md), [ids-and-state.md](./ids-and-state.md)
