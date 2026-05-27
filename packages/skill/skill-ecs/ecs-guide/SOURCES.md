# Sources

## Game-engine development blog (archive)

- **URL:** https://archive-host.github.io/blog_archive/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → all sections
  - The data-oriented ECS storage/iteration model and the change-tracking/external-sync trade-offs
- **Aspects extracted:**
  - "Should entities support multiple instances of the same component?" — single-component-per-type verdict, multiplicity via child entities / list-components, index-into-internal for custom layout → `references/single-vs-multiple-components.md`, `references/storage-and-archetypes.md`
  - "Syncing a data-oriented ECS with a stateful external system" — five change-tracking strategies (brute force + dirty flags, orchestration, callbacks, change lists, "Changing" tag component), out-of-sync window, callbacks breaking parallelism → `references/change-tracking-and-sync.md`
  - "Entity-Component-Systems and Rendering" — `tm_ci_render_i` / `tm_ci_shader_i` plugin interfaces, per-viewer culling into visibility bitmasks, render-graph module injection, extra viewers generated during execution → `references/rendering-integration.md`
  - Entity-type-as-bitmask contiguous storage, filter-and-batch system loops → `references/storage-and-archetypes.md`, `references/systems-and-iteration.md`

## Data-oriented design foundations

- **URLs:**
  - Richard Fabian, "Data-Oriented Design" — https://www.dataorienteddesign.com/dodbook/
  - Mike Acton, "Data-Oriented Design and C++" (CppCon 2014) — https://www.youtube.com/watch?v=rX0ItVEVjHc
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas
  - The cache-locality and existence-based-processing rationale behind contiguous per-type storage
- **Aspects extracted:**
  - Tables/streams, archetype grouping, transforms over bulk data → `references/storage-and-archetypes.md`, `references/systems-and-iteration.md`
  - Detailed cache/layout reasoning deferred to data-oriented-design-guide → `references/storage-and-archetypes.md`

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
