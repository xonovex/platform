---
name: asset-pipeline-guide
description: "Use when designing the asset pipeline of a tool or game engine — turning authored sources (FBX, glTF, PNG, WAV, shaders) into runtime-ready data: importers/compilers per asset type, a deterministic compile/cook step, content-addressed caching keyed by a hash of inputs and settings, dependency tracking so an edit reimports only what changed, platform-specific output, and live hot-reloading. Triggers on import, cook/bake, asset/build cache, reimport, dependency invalidation, file watcher, hot-reload of textures/meshes/materials, and 'why does changing one file rebuild everything', even when the user doesn't say 'asset pipeline'. Skip modeling assets as typed objects/references (use data-model-guide), layout of cooked data (use data-oriented-design-guide), and reload-safe C (use c99-opinionated-guide)."
---

# Asset pipeline Guidelines

An asset pipeline turns editable authored sources into runtime-ready data and keeps the two in sync as content changes. The pipeline has two halves: a stable raw/intermediate representation that stays close to the original file, and a derived runtime representation produced by deterministic compile steps whose results are cached by content hash. How assets are modeled as typed objects with stable references is owned by data-model-guide; the memory layout of the cooked output is owned by data-oriented-design-guide; the discipline of writing reloadable C is owned by c99-opinionated-guide.

## Essentials

- **Two representations** - Keep an editable raw/intermediate form separate from the compiled runtime form; never edit the runtime form directly, see [references/raw-vs-runtime-formats.md](references/raw-vs-runtime-formats.md)
- **Importers per format, compilers per type** - Source files enter through a format importer keyed by extension; each asset type owns a deterministic compiler that cooks it to runtime data, see [references/import-and-compile.md](references/import-and-compile.md)
- **Settings are compile inputs** - Import/compile options (mip count, compression, quantization, target platform) are part of the input, not side state, see [references/import-and-compile.md](references/import-and-compile.md)
- **Hash inputs to a cache key** - Compute a content hash over source bytes plus settings plus the input hashes of dependencies; key the cache on it, see [references/content-hash-and-cache.md](references/content-hash-and-cache.md)
- **Track dependencies, reimport the closure** - Record which inputs each output consumed so a source edit invalidates exactly its dependents and nothing else, see [references/dependency-tracking.md](references/dependency-tracking.md)
- **Hot-reload by swapping live data** - Watch sources, recompile only the changed closure, and swap runtime data atomically while the app keeps running, see [references/hot-reloading-content.md](references/hot-reloading-content.md)

## Import and compile

- **Stay close to the source on import** - Do the minimum to land a faithful intermediate; defer all heavy transformation to the per-type compile step, see [references/raw-vs-runtime-formats.md](references/raw-vs-runtime-formats.md)
- **Make compilation deterministic** - Same inputs and settings must always yield byte-identical output; ban wall-clock time, iteration order, and absolute paths from the output, see [references/import-and-compile.md](references/import-and-compile.md)
- **Cook per target platform** - Emit platform-specific runtime data (texture format, endianness, alignment) and fold the target into the cache key, see [references/import-and-compile.md](references/import-and-compile.md)

## Caching and incrementality

- **Content-address the cache** - Store cooked outputs under their content hash so identical inputs hit the same entry and a build cache is shareable across machines, see [references/content-hash-and-cache.md](references/content-hash-and-cache.md)
- **Invalidate by hash propagation** - An output's validity hash folds in the hashes of its inputs, so any upstream change ripples down and only stale outputs recook, see [references/dependency-tracking.md](references/dependency-tracking.md)
- **Reimport the dependents, not the world** - On a source change, recompile the recorded dependency closure of that source, leaving unrelated assets untouched, see [references/dependency-tracking.md](references/dependency-tracking.md)

## Live iteration

- **Watch and recompile in the background** - A file watcher feeds an async compile queue; the app stays responsive while cooks run, see [references/hot-reloading-content.md](references/hot-reloading-content.md)
- **Swap, then retire old data** - Publish the new runtime resource, repoint references, and free the old version only once no in-flight work still reads it, see [references/hot-reloading-content.md](references/hot-reloading-content.md)

## Gotchas

- A "deterministic" compiler that hashes a struct with padding bytes, or iterates a hash map in pointer order, produces a different hash per run and silently defeats the cache — hash canonical content, not raw memory.
- Forgetting that settings are inputs means changing the compression preset does not bust the cache; the user keeps seeing the old cooked texture and blames the importer.
- If the dependency record is incomplete (a compiler reads an included shader header it never reported), an edit to that header reimports nothing and ships stale runtime data — under-reporting dependencies is worse than over-reporting.
- Caching the heavy step is the point; adding a cache to a node whose compute is cheaper than the hash and lookup makes iteration slower, not faster.
- A content hash that includes an absolute path, build host name, or timestamp is no longer content-addressed — two machines cooking the same source get different keys and can never share the cache.
- Hot-reload that frees the old resource immediately can crash a frame already submitted with the old handle; retire old versions behind the in-flight fence, not on swap.
- Letting the editable hierarchy be mutated in ways the source format can't express makes reimport ambiguous about who is authoritative — keep the structural shape immutable and confine edits to overrides on instances.
- A least-common-denominator intermediate silently drops format-specific data (vertex colors, custom attributes); provide a typed extension slot so importers can carry data the generic schema lacks.

## Progressive Disclosure

- Read [references/raw-vs-runtime-formats.md](references/raw-vs-runtime-formats.md) - Load when separating editable source/intermediate data from compiled runtime data, or deciding what to do at import vs compile time
- Read [references/import-and-compile.md](references/import-and-compile.md) - Load when writing importers or per-type compilers, making the cook deterministic, or treating settings and target platform as inputs
- Read [references/content-hash-and-cache.md](references/content-hash-and-cache.md) - Load when hashing inputs+settings into a cache key, building a content-addressed cache, or making the build cache reproducible and shareable
- Read [references/dependency-tracking.md](references/dependency-tracking.md) - Load when recording asset dependencies so a source edit invalidates exactly its dependents, or implementing incremental reimport
- Read [references/hot-reloading-content.md](references/hot-reloading-content.md) - Load when watching sources, recompiling in the background, or swapping live runtime data safely while the app runs
