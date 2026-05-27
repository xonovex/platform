# ids-and-state: Identify Controls by Stable ID

**Guideline:** Identify every interactive control by a stable unique id and key all persistent interaction state (hover, active, focus, drag, tab order) by that id. Keep the persistent store minimal — only what cannot be recomputed each frame.

**Rationale:** In immediate mode there are no retained widget objects to hang state on, yet hover, the currently-pressed control, keyboard focus, and an in-progress drag must survive between frames. An id is the substitute identity. If ids are unstable (derived from screen position) or collide (reused in a loop without scoping), two controls share one state slot and interaction jumps between them.

**How to Apply:**

1. Derive ids from a stable source: the underlying data object's id, or a scope combined with a loop index — never from layout/screen position.
2. Use a scope/id stack so repeated widgets (list rows, tree nodes) get distinct ids under a parent scope.
3. Store only the irreducible state: `hover`, `active`/pressed, focus (responder chain), drag id, and layout the caller can't recompute. Everything else is rebuilt from app data each frame.
4. Represent "nothing" as the zero/null id (no hover, no drag, no focus) so a cleared field is a valid default.

**Example:**

```c
// Good: id from the data object; state keyed by id in the ui context
for (uint32_t i = 0; i < n; ++i)
    if (button(ui, items[i].object_id, row_rect(i), items[i].name))
        select(items[i].object_id);

// Bad: id from screen position — scrolling or relayout makes "the same button" a different id
uint64_t id = hash(rect.x, rect.y); // hover/focus jump around on layout change
```

**Counter-Example:** Purely decorative, non-interactive primitives (labels, separators) need no id — only controls that can be hovered, focused, dragged, or clicked require one.

**Related:** [frame-delay.md](./frame-delay.md), [events-and-focus.md](./events-and-focus.md), [drag-and-drop.md](./drag-and-drop.md)
