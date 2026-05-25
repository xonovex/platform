# undo-redo: Transactional Edits and the Undo Journal

**Guideline:** Make edits transactional — record either the inverse of each operation or a before/after snapshot of what changed into an undo journal, group the operations of one user action into a single transaction, and maintain a redo stack so undone transactions can be re-applied.

**Rationale:** Undo is the defining feature of an editor's data model, and it only works if every state-changing edit is captured. Recording inverse operations (or before/after values) makes undo a pure data operation: pop the last transaction, apply its inverses, push it onto the redo stack. Grouping into transactions matters because one user action ("move three nodes") is many property writes; the user expects one Ctrl+Z to undo all of them, not three. Tying the transaction boundary to the change-notification commit keeps undo, notification, and persistence consistent.

**Techniques:**

- **Inverse operations** - For each mutation, record an operation that exactly reverses it: `set(id, prop, old)` inverts `set(id, prop, new)`; `delete` inverts `create`. Undo applies inverses in reverse order.
- **Before/after snapshots** - Alternatively store the property's old value (for undo) and new value (for redo). Simpler than authoring inverses, at the cost of storing both values; cheap for scalar properties, expensive for large buffers.
- **Transactions** - Open a transaction at the start of a user action, record every mutation into it, commit at the end. The whole transaction is one undo step. Empty transactions are discarded.
- **Undo / redo stacks** - Commit pushes the transaction onto the undo stack. Undo pops it, applies inverses, pushes it onto the redo stack. Any new edit clears the redo stack (you cannot redo into a branch you've diverged from).
- **Bounding history** - Cap the undo stack by count or by memory; drop the oldest transactions. For large-buffer edits, store deltas or only the changed region rather than whole copies.
- **Undoable vs not** - Pure model mutations are undoable. External side effects — file writes, network calls, spawning a process — are not; keep them out of the journal or make them idempotent and re-issue on redo deliberately.

**How to Apply:**

1. Wrap each user action in `txn_begin` / `txn_commit`; route all mutations through the model so they land in the open transaction.
2. For every mutation, append an entry capturing object id, property, and old (and new) value — ids and values, never pointers or indices.
3. Undo: pop the undo stack, apply inverses in reverse order, push onto redo. Redo: the mirror.
4. On any fresh edit, clear the redo stack; enforce the history bound after commit.

**Example:**

```c
typedef struct { object_id_t id; uint32_t prop; value_t old, new; } edit_t;
typedef struct { edit_t *edits; uint32_t count; } txn_t;

static void txn_commit(model_t *m, txn_t *t) {
  if (t->count == 0) return;          // discard empty transactions
  undo_stack_push(m, t);
  redo_stack_clear(m);                // a new edit invalidates the redo branch
  enforce_history_bound(m);
}

static void undo(model_t *m) {
  txn_t *t = undo_stack_pop(m);
  if (!t) return;
  for (int32_t i = (int32_t)t->count - 1; i >= 0; i--)   // reverse order
    apply_set(m, t->edits[i].id, t->edits[i].prop, t->edits[i].old);
  redo_stack_push(m, t);
}

// Bad: capturing a pointer/index in the journal breaks after storage relocates
// or the object is deleted-then-recreated. Capture { id, value } instead.
edit_t broken = { .target_ptr = obj_ptr }; // dangles on undo of a delete
```

**Gotchas:**

- An edit entry that stored a pointer or array index breaks the moment storage moves or an object is recreated — capture the stable id and the value.
- Forgetting to clear the redo stack on a new edit lets a redo re-apply a stale operation onto diverged state, corrupting the model.
- Side effects (writing a file, sending a request) cannot be undone by replaying inverses; either exclude them or model them explicitly so redo re-issues them on purpose.
- Whole-buffer snapshots for large `buffer`/`array` properties blow up history memory; store region deltas and bound the stack.
- Transaction boundaries should match user intent: too fine and one click takes many undos; too coarse and unrelated edits undo together.

**Related:** [references/change-notification.md](./change-notification.md), [references/references-and-ownership.md](./references-and-ownership.md), [references/snapshots-and-threading.md](./snapshots-and-threading.md)
