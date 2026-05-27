---
name: data-model-guide
description: "Use when designing a central in-memory data model / object database for a tool, editor, or engine: typed objects with properties, stable cross-references and sub-object ownership, change notification, undo/redo, and serialization. Triggers on prompts about a runtime object/property schema, referencing objects by stable id/GUID, observing/notifying changes, transactional edits and undo history, save/load with versioning/migration, or copy-on-write snapshots — even when the user doesn't name a specific model. Skip cache/layout optimization (use data-oriented-design-guide), relational/SQL schema design (use sql-postgresql-guide), and runtime ECS component iteration for simulation/rendering."
---

# Data-Model Guidelines

## Essentials

- **One central model** - Route tool/editor state through a single typed object store, not scattered structs, see [references/object-model.md](references/object-model.md)
- **Stable ids, not pointers** - Reference objects by id/GUID so storage can move, save, and reload, see [references/references-and-ownership.md](references/references-and-ownership.md)
- **Make mutation observable** - Edits go through the model so it can notify, undo, and persist, see [references/change-notification.md](references/change-notification.md)

## Object model

- **Typed objects** - Describe each type by a runtime schema, not a hand-written C struct, see [references/object-model.md](references/object-model.md)
- **Property set per type** - bool/int/float/string/reference/sub-object/buffer/array properties, see [references/object-model.md](references/object-model.md)
- **Data-driven types** - Add a type or property at runtime without recompiling, see [references/object-model.md](references/object-model.md)
- **Separate metadata from instances** - Type/schema lives once; instance data is just values, see [references/object-model.md](references/object-model.md)

## References & ownership

- **Strong = owning sub-object** - A parent owns its sub-objects; deleting the parent deletes them, see [references/references-and-ownership.md](references/references-and-ownership.md)
- **Weak = reference** - A reference does not own; on delete it resolves to null, never dangles, see [references/references-and-ownership.md](references/references-and-ownership.md)
- **Local vs global ids** - Local ids inside a file, GUIDs for cross-file/cross-session links, see [references/references-and-ownership.md](references/references-and-ownership.md)
- **Identity vs role** - A GUID names a fixed identity; a name/path names a role that late-binds — default to GUIDs, see [references/references-and-ownership.md](references/references-and-ownership.md)

## Prototypes & overrides

- **Any object can be a prototype** - Instances inherit all properties and store only overrides; resolve a property by walking the prototype chain, see [references/prototypes.md](references/prototypes.md)
- **Per-property override bitmask** - One bit per property marks local-vs-inherited; sub-object sets track inherited / instantiated / removed children, see [references/prototypes.md](references/prototypes.md)
- **Serialize only deltas** - Persist the prototype reference plus overrides; reconstruct inherited values on load, see [references/prototypes.md](references/prototypes.md)

## Mutation

- **Change notification** - Events, dirty flags, dependency recompute, batched per transaction, see [references/change-notification.md](references/change-notification.md)
- **Pull vs push** - Poll a version number, or register callbacks; avoid notification storms, see [references/change-notification.md](references/change-notification.md)
- **Undo/redo** - Record inverse ops or before/after snapshots into an undo journal, see [references/undo-redo.md](references/undo-redo.md)
- **Transactions** - Group many edits into one user-visible undo step, see [references/undo-redo.md](references/undo-redo.md)

## Persistence

- **Stable on-disk schema** - Version the format; migrate old files into the new schema on load, see [references/serialization.md](references/serialization.md)
- **References serialize by id** - Persist ids/GUIDs, re-resolve on load, see [references/serialization.md](references/serialization.md)
- **Deterministic output** - Stable ordering for diff/version control; choose text vs binary, see [references/serialization.md](references/serialization.md)
- **Save changes or re-spawn from assets** - Persist member-level deltas for runtime state, asset references for content; POD serializes by default, see [references/serialization.md](references/serialization.md)
- **ABI-stable struct evolution** - Append fields at the end, add a leading `size`, never reorder/retype; semantic-version the API, see [references/serialization.md](references/serialization.md)
- **Snapshots & threading** - Publish immutable/copy-on-write versions for readers, see [references/snapshots-and-threading.md](references/snapshots-and-threading.md)

## Gotchas

- A weak reference must resolve to null after its target is deleted; never leave it pointing at a recycled id.
- Reusing a freed object id lets a stale reference silently alias a different object — bump a generation or never reuse.
- Firing a notification synchronously inside a setter can re-enter the model mid-edit; queue and flush at transaction end.
- An undo entry that captured a pointer/index breaks once storage moves — capture ids and values, not addresses.
- External side effects (file writes, network) are not undoable; keep them out of the transactional journal.
- A loader that rejects unknown fields cannot open older OR newer files — migrate forward and ignore-or-preserve unknowns.
- Overriding a sub-object in an instance gives it a new id, so references to the prototype's original silently miss — resolve references through the override, don't store the new id everywhere.
- Allowing computed overrides (a property defined as an expression over the prototype's value) drags evaluation order and cycles into the data model — keep overrides concrete values.

## Progressive Disclosure

- Read [references/object-model.md](references/object-model.md) - Load when defining typed objects, property kinds, or data-driven schemas
- Read [references/references-and-ownership.md](references/references-and-ownership.md) - Load when linking objects by id, modeling ownership, or handling deletes
- Read [references/prototypes.md](references/prototypes.md) - Load when adding prefabs/templates/presets: prototype-instance inheritance, per-property overrides, override propagation
- Read [references/change-notification.md](references/change-notification.md) - Load when making edits observable, batching, or recomputing derived data
- Read [references/undo-redo.md](references/undo-redo.md) - Load when adding transactional edits, undo journals, or redo stacks
- Read [references/serialization.md](references/serialization.md) - Load when saving/loading, versioning the format, or migrating old files
- Read [references/snapshots-and-threading.md](references/snapshots-and-threading.md) - Load when giving reader threads a consistent view while writers mutate
