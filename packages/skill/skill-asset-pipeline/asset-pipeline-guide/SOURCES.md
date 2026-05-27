# Sources

## Game-engine development blog (archive)

- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Import and compile, Caching and incrementality, Live iteration, Gotchas
  - A worked instance of a two-representation asset pipeline: faithful intermediate on import, deterministic per-type compile, validity-hash caching, hash-propagation invalidation, and the same transform framework reused offline and at runtime
- **Aspects extracted:**
  - "the engine Asset Pipeline" — importers behind one extension-keyed interface running async into a faithful intermediate; doing minimal massaging on import to stay close to the source; a generic intermediate as least-common-denominator with a typed extension slot for format-specific data; per-type compile steps (textures: mips/compression, meshes: vertex-cache optimization/quantization, materials: shader/material graph) splitting one source into multiple runtime resources; keeping the structural hierarchy immutable and confining edits to overrides on instances so reimport stays unambiguous about authority → `references/raw-vs-runtime-formats.md`, `references/import-and-compile.md`, `references/dependency-tracking.md`
  - "Creation Graphs" — each output computes a validity hash from its local settings plus the validity hashes of its inputs, used to skip redundant computation; selective caching of heavy operations only; the same transform pipeline runs offline (long compiles acceptable) or at runtime (efficiency-critical), enabling hot iteration; hash propagation forms the implicit dependency graph that invalidates exactly the affected outputs → `references/content-hash-and-cache.md`, `references/dependency-tracking.md`, `references/hot-reloading-content.md`, `references/import-and-compile.md`
  - "Referencing Objects: Names vs GUIDs" — stable GUID/object identity as the default for references so they survive rename and move (vs fragile path/name references that break on reorg); content-hash identity where the same content always maps to the same identifier (the Git principle); using stable identity for dependency edges → `references/dependency-tracking.md`, `references/content-hash-and-cache.md`

## General build-cache / incremental-build prior art

- **URLs:**
  - Bazel — content-addressed remote build cache and action determinism — https://bazel.build/remote/caching
  - Git — content-addressed object model (same content → same hash) — https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Caching and incrementality; reference grounding for content-addressed, shareable caches
  - Confirmation that hashing canonical inputs (excluding paths/timestamps/host) to key a content-addressed cache and share it across machines is general build-system practice, not engine-specific
- **Aspects extracted:**
  - Action determinism + hermetic inputs as the precondition for a shareable cache; remote cache hits turning a cold build into a download → `references/import-and-compile.md`, `references/content-hash-and-cache.md`
  - Content-addressed identity (hash of content is the key) and the danger of leaking location/time into the key → `references/content-hash-and-cache.md`

## Refresh Workflow

1. Re-read the upstream source(s) above (the three archived posts and the build-cache references)
2. Diff against the prior pull (or scan for newly added sections / revisions)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
