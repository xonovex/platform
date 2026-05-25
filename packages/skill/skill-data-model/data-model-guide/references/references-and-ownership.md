# references-and-ownership: Stable Ids, Ownership, and Resolving References

**Guideline:** Link objects to each other by a stable id (local id within a file, GUID across files/sessions), never by raw pointer, and distinguish owning sub-objects (strong) from non-owning references (weak) so deletion has well-defined, dangle-free semantics.

**Rationale:** A data model is relocated constantly: storage grows, objects are saved and reloaded at new addresses, an undo restores a deleted object. Raw pointers break across every one of those events. An id survives serialization and relocation, and is resolved to the current object through one indirection table. Separating ownership from reference makes the object graph a tree of owned data plus a set of cross-links: deleting an object cleanly deletes everything it owns, while cross-links to it simply stop resolving — they become null instead of dangling.

**Techniques:**

- **Refer by id** - Store the target's id, not its address. Dereference through a registry: `resolve(id) -> object*` or null. Cache the resolved pointer only within a scope where no edit can move or delete it.
- **Strong / owning (sub-object)** - A parent owns its sub-objects. Ownership is exclusive and forms a tree: each object has at most one owner. Deleting (or freeing) the parent recursively deletes its owned sub-objects. Serialization writes sub-objects inline under their owner.
- **Weak / reference** - A reference points at an object it does not own. Many references may target one object. On the target's deletion the reference is left pointing at a now-absent id, and `resolve` returns null — a guaranteed safe miss, never a dangling read.
- **Local ids vs GUIDs** - Inside one file, small monotonically-assigned local ids are compact and fast. For links that cross files, span sessions, or merge between users, use a globally-unique id (GUID) so two independently-authored objects never collide.
- **Generation / no-reuse** - If local ids are recycled, pair each with a generation counter (id + generation) so a stale reference to a reused slot is detected; otherwise never reuse an id within a model's lifetime.
- **Resolve at the boundary** - Keep ids in storage and on disk; resolve to pointers only transiently inside an operation. This keeps the model relocatable and serializable end to end.

**How to Apply:**

1. Give every object an id; keep a registry `id -> object*` (an indirection table).
2. Model containment as `sub_object` properties (strong); model cross-links as `reference` properties holding an id (weak).
3. On delete: recursively delete owned sub-objects; leave references untouched — they will resolve to null.
4. On dereference of a `reference`: look up the id; return null if absent or generation mismatches; handle null at every call site.

**Example:**

```c
typedef struct { uint32_t value, generation; } object_id_t; // 0 == null

static object_t *resolve(model_t *m, object_id_t id) {
  if (id.value == 0 || id.value >= m->slot_count) return NULL;
  slot_t *s = &m->slots[id.value];
  if (s->generation != id.generation) return NULL; // stale -> safe null
  return s->object;
}

// Strong: deleting a parent deletes everything it owns (tree).
static void delete_object(model_t *m, object_id_t id) {
  object_t *o = resolve(m, id);
  if (!o) return;
  for (uint32_t i = 0; i < o->sub_count; i++)
    delete_object(m, o->sub_objects[i]); // recurse into owned children
  m->slots[id.value].generation++;       // invalidate outstanding weak refs
  m->slots[id.value].object = NULL;       // weak refs now resolve to NULL
}

// Bad: a stored pointer dangles after delete/relocate/reload.
object_t *cached_target = resolve(m, ref_id); // valid only within this scope
```

**Gotchas:**

- A weak reference whose target was deleted must resolve to null; a model that instead leaves it pointing at whatever object reused the id has a silent aliasing bug.
- An object must have exactly one owner — two strong owners cause double-delete; convert all but one to weak references.
- Cross-file links need GUIDs: local ids are only unique within the file that minted them, so a merge or import will collide them.
- Resolving an id every access has a cost; batch-resolve once per operation, but never cache a pointer across an edit that can delete or relocate.

**Related:** [references/object-model.md](./object-model.md), [references/serialization.md](./serialization.md), [references/undo-redo.md](./undo-redo.md); **data-oriented-design-guide** for the cache-friendly handle/indirection storage mechanics
