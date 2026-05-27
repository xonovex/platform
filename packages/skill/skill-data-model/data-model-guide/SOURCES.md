# Sources

## Game-engine development blog (archive)

- **URL:** https://archive-host.github.io/blog_archive/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Object model, References & ownership, Mutation, Persistence
  - The central typed object/property "object database" framing for tools and editors
- **Aspects extracted:**
  - "The Story behind The Truth: Designing a Data Model" — typed objects, property kinds, data-driven types, ids, and the strong sub-object ownership vs weak reference (delete-to-null) model → `references/object-model.md`, `references/references-and-ownership.md`
  - "Referencing Objects: Names vs GUIDs" — id/GUID references, GUID-as-identity vs name/path-as-role, default-to-GUID hybrid, resolve-once-at-spawn, rename/auto-patch pitfalls → `references/references-and-ownership.md`
  - "Multi-Threading the Truth" — immutable snapshots, copy-on-write versions, atomic publish, per-frame read snapshot → `references/snapshots-and-threading.md`
  - "One-Click Save Game System (Parts 1–3)" — member-level change tracking (deltas vs snapshots), POD default serializer, asset-reference re-spawn for smaller saves and forward compatibility, declarative persistence opt-in → `references/serialization.md`
  - "API Versioning" — semantic-versioned API structs, ABI-safe evolution (append at end, leading `size` field, reserved-byte reuse), exact-major / higher-minor matching → `references/serialization.md`
  - "The Document Model and the engine" — per-asset vs project-wide save/revert and undo-stack scope trade-offs, runtime-resolved inter-document relationships → `references/undo-redo.md`
  - Change tracking, dirty flags, transaction-batched notification → `references/change-notification.md`
  - Versioned serialization, migration of old files, references serialized by id → `references/serialization.md`

## Editor / DAW / DCC architecture write-ups (command pattern and document models)

- **URLs:**
  - Gang of Four, "Design Patterns" — Command and Memento patterns (undo)
  - https://gameprogrammingpatterns.com/command.html
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Mutation
  - Transactional edits, inverse operations vs before/after snapshots, undo/redo stacks
- **Aspects extracted:**
  - Command/Memento → inverse operations and before/after snapshots → `references/undo-redo.md`
  - Grouping edits into one user-visible transaction, redo-stack invalidation → `references/undo-redo.md`

## Copy-on-write and persistent data structures

- **URLs:**
  - Read-copy-update (RCU) — https://www.kernel.org/doc/html/latest/RCU/whatisRCU.html
  - Persistent data structures (structural sharing) — https://en.wikipedia.org/wiki/Persistent_data_structure
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Persistence, Gotchas
  - Lock-free immutable snapshots, structural sharing, version reclamation
- **Aspects extracted:**
  - Atomic publish + acquire/release, deferred reclamation of old versions → `references/snapshots-and-threading.md`
  - Structural sharing so a small edit copies only touched nodes → `references/snapshots-and-threading.md`

## Serialization, schema versioning, and migration

- **URLs:**
  - https://martinfowler.com/articles/evodb.html (evolutionary schema, migration sequencing)
  - Protocol/format evolution guidance on unknown-field tolerance — https://protobuf.dev/programming-guides/proto3/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Persistence, Gotchas
  - Stable on-disk schema, ordered forward migrations, tolerating unknown fields, deterministic output for diffing
- **Aspects extracted:**
  - Format versioning + ordered migration steps → `references/serialization.md`
  - Skip/preserve unknown fields for forward/backward compatibility → `references/serialization.md`
  - Deterministic, stably-ordered output for version control → `references/serialization.md`

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
