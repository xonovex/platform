# serialization: Save, Load, Versioning, and Migration

**Guideline:** Persist the model through a stable, versioned on-disk schema; serialize references by id, migrate older files forward into the current schema on load, and emit deterministic output so saved files diff cleanly under version control.

**Rationale:** Saved files outlive the code that wrote them. The schema in memory will change — properties get added, renamed, removed — but old files must still open. A versioned format plus per-version migration lets the loader transform an old layout into the current one instead of refusing it. Because the in-memory model already references objects by id (not pointer), serialization is mostly mechanical: write ids, re-resolve on load. Deterministic output (stable ordering, stable formatting) is what makes the file usable with diff/merge tools and code review.

**Techniques:**

- **Stable on-disk schema** - Define the file layout explicitly and independently of the in-memory layout. Tag the file with a format version. Treat the on-disk property names/ids as a contract.
- **Versioning + migration** - On load, read the version; if older, run the chain of migration steps (`v1→v2→v3…`) that each transform the prior layout into the next. New code reads only the current schema; migrations bridge the gap.
- **References by id** - Owned sub-objects serialize inline under their owner; weak references serialize as the target's id/GUID. On load, write objects first, then resolve ids — an id pointing at a missing object loads as null, mirroring runtime delete semantics.
- **Text vs binary** - Text (JSON/structured text) diffs and merges well and is human-inspectable — good for source-controlled project files. Binary is compact and fast — good for caches, large buffers, and runtime assets. Models often save text for authoring and binary for shipping.
- **Deterministic output** - Emit properties and objects in a stable order (by property index, by id), normalize floats/whitespace, and avoid embedding timestamps or hash-ordered maps. Two saves of the same model must be byte-identical so version control shows only real changes.
- **Partial / streaming load** - For large models, write a table of contents / per-object offsets so the loader can map or stream objects on demand instead of reading the whole file. Lazy-resolve references as objects are paged in.
- **Unknown fields** - A forward-and-backward tolerant loader skips properties it doesn't recognize (and ideally preserves them on resave) rather than failing, so a newer file opens in older code without data loss.
- **Changes vs snapshots (save games)** - For mutable runtime state, persist member-level _changes_ (deltas) tracked since creation rather than full snapshots — much smaller, but anything not reported before the save is lost. POD records can use a default (zero-initialized) serializer; non-POD records supply their own serialize/deserialize callbacks. Mark what is persistent declaratively (e.g. in the editor) rather than baking it into every system.
- **Re-spawn from asset references** - When content originates from an asset, save only a reference to the asset plus the local overrides, and re-spawn the hierarchy from the _current_ asset on load. Saves shrink, and content patches apply automatically — but you must keep every asset version that older saves still reference, and structural changes to an asset hierarchy can break the UUID matching that re-maps overridden children.
- **ABI-stable struct evolution** - For versioned API/data structs that must stay binary-compatible: append new fields only at the _end_, never reorder or retype existing ones, and reuse explicitly zero-initialized reserved bytes. A leading `uint32_t size` field lets a callee version-detect a struct passed by pointer. Request an exact major version, accept higher minor/patch; for a breaking change introduce a new struct and support both for a while rather than mutating the old one.

**How to Apply:**

1. Write a header with a magic tag and a format `version`.
2. Serialize each owned object: its type, id, and each set property; sub-objects inline, references as ids.
3. On load, read `version`; if `< current`, apply migration steps in sequence to reach the current schema.
4. Re-resolve all reference ids after all objects exist; emit objects/properties in a stable order for deterministic files.

**Example:**

```c
#define FORMAT_VERSION 3

static bool load(model_t *m, reader_t *r) {
  header_t h = read_header(r);
  if (h.magic != MODEL_MAGIC) return false;
  read_all_objects(m, r);                  // create objects + record ref ids
  for (uint32_t v = h.version; v < FORMAT_VERSION; v++)
    migrate_step(m, v);                    // v1->v2->...->current
  resolve_all_references(m);               // ids -> objects; missing -> null
  return true;
}

// Migration: a property was split in v2; old files have the combined one.
static void migrate_step(model_t *m, uint32_t from) {
  if (from == 1) for_each_object(m, split_size_into_width_height);
}

// Bad: refusing unknown properties means a newer file can never open in older
// code, and a v1 reader rejects v2 files outright. Migrate forward; skip
// (and ideally preserve) properties you don't recognize.
if (prop_unknown) return false; // brittle: loses cross-version compatibility
```

**Gotchas:**

- A loader that hard-fails on an unrecognized property cannot open newer files; skip-and-preserve unknowns instead of rejecting.
- Migrations must be ordered and idempotent per step; a missing or out-of-order step silently mis-reads old data as if it were current.
- Non-deterministic output (map iteration order, embedded timestamps, unnormalized floats) produces spurious diffs and defeats version control review.
- References to objects absent from the file must load as null, not as a fabricated or dangling target — match runtime delete semantics.
- Binary formats are fragile across struct/layout/endianness changes; version them as strictly as text, or restrict binary to regenerable caches.
- A change-tracking save loses any mutation not reported before the save point; ensure every persistent edit flows through the tracking path, or the loaded state silently differs from what the user saw.
- "Reserved, must be zero" struct bytes are unenforceable (Hyrum's Law): callers will pass garbage and break future reuse. A leading `size` field is the more reliable way to evolve a by-pointer struct.
- You cannot freely replace an asset that older saves point to — re-spawning from a structurally-changed asset can leave saved overrides orphaned or mis-mapped.

**Related:** [references/references-and-ownership.md](./references-and-ownership.md), [references/object-model.md](./object-model.md), [references/snapshots-and-threading.md](./snapshots-and-threading.md)
