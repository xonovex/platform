# Sources

## Game-engine development blog (archive)

- **URL:** https://archive-host.github.io/blog_archive/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Object model, References & ownership, Mutation, Persistence
  - The central typed object/property "object database" framing for tools and editors
- **Aspects extracted:**
  - "The Story behind The Truth: Designing a Data Model" — typed objects, property kinds, data-driven types, ids → `references/object-model.md`
  - "References, Ownership, and the Object Model" — id/GUID references, strong sub-object ownership vs weak references, delete-to-null → `references/references-and-ownership.md`
  - "Multi-Threading the Truth" — immutable snapshots, copy-on-write versions, atomic publish, per-frame read snapshot → `references/snapshots-and-threading.md`
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
