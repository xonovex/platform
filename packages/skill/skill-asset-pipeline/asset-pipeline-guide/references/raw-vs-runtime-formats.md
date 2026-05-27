# raw-vs-runtime-formats: Separating Editable Source From Compiled Runtime Data

**Guideline:** Keep two distinct representations of every asset — an editable raw/intermediate form that stays as close as possible to the original authored file, and a derived runtime form produced by a compile step — and never let runtime data be edited directly or treated as the source of truth.

**Rationale:** Source files from authoring tools (FBX, glTF, PSD, WAV) are rich, generic, and built for editing; runtime data is narrow, packed, and built for fast loading on a specific target. Conflating the two forces every editing operation to fight the packed layout and every runtime load to carry editing metadata. By landing imports in an intermediate that does the minimum massaging — just enough to render or audition with decent quality while staying faithful to the original — you keep reimport cheap and unambiguous, and you defer all expensive, lossy, platform-specific transformation to the compile step where it can be cached and redone per target. The tradeoff is two representations to keep in sync, which is exactly what the dependency tracker and content cache exist to automate; the cost of a unified format is far higher, because any change to runtime packing then forces re-authoring and any edit risks corrupting shippable data.

**How to Apply:**

1. On import, produce a typed intermediate object that mirrors the source's structure: scene tree/hierarchy, buffers (vertices, indices, image bits, animation), and objects (meshes, materials, textures, lights, cameras).
2. Do as little transformation as possible at import time — enough for a faithful preview, no more; push mip generation, compression, vertex-cache optimization, and quantization into the per-type compiler.
3. Treat the runtime form as a pure function of the intermediate plus settings; it is regenerable and disposable, so store it in a cache, not in version control as a hand-edited artifact.
4. Confine user edits to the intermediate (or to overrides layered on top of it); when the source file changes, reimport overwrites the intermediate and recompiles downstream.
5. Provide a typed extension slot on the intermediate so a format-specific importer can attach data the generic schema cannot represent, rather than silently dropping it.

**Example:**

```c
// Intermediate: faithful, editable, close to source. Stored in the data model.
typedef struct dcc_scene_t {
    scene_tree_t       tree;     // hierarchy + node transforms, mirrors the source
    buffer_set_t       buffers;  // raw vertex/index/image/anim bits, minimally touched
    object_array_t     objects;  // meshes, materials, textures, lights, cameras
    extension_set_t    ext;      // typed slot for format-specific data the schema lacks
} dcc_scene_t;

// Runtime: narrow, packed, per-target. Derived by compile(), cached, never hand-edited.
typedef struct runtime_mesh_t {
    gpu_buffer_t   vbuf;         // quantized, vertex-cache-optimized for this platform
    gpu_buffer_t   ibuf;
    material_id_t  material;
} runtime_mesh_t;

runtime_mesh_t mesh_compile(const dcc_scene_t *src, mesh_id_t id, target_t platform);
```

**Gotchas:**

- A generic intermediate is a least-common-denominator and will lose valuable source data unless you give importers a typed extension mechanism to carry it forward.
- If users can mutate the intermediate's structural shape (relinking the hierarchy) in ways the source format cannot express, reimport becomes ambiguous about who owns the structure — keep the shape immutable and allow only override-style edits, ideally on instances rather than the prototype.
- Storing checked-in, hand-tweaked runtime data turns the cache into a source of truth and makes platform retargeting impossible; runtime data must always be regenerable from intermediate + settings.
- "Minimal massaging on import" is a discipline: every transformation you sneak into the importer is one you cannot recook per platform or skip on reimport.

**Related:** [references/import-and-compile.md](./import-and-compile.md), [references/dependency-tracking.md](./dependency-tracking.md), **data-model-guide**, **data-oriented-design-guide**
