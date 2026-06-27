# shader-system: Shaders as Compiled Build Artifacts

## Guideline

Treat shaders as a build artifact: author in a source language, compile to a binary intermediate offline (e.g. SPIR-V, DXIL), reflect the binary to derive binding/layout information, manage feature variants explicitly, key pipeline-state objects on shader-plus-state, and support hot-reload by recompiling and rebuilding affected pipelines.

## Rationale

Compiling shader source at runtime is slow, pulls a full compiler into the shipping binary, and hides syntax errors until the draw call that needs them. Compiling to a binary intermediate offline moves errors to build time, removes the compiler from the runtime, and lets a pipeline cache key on stable bytecode. Reflecting that binary to discover its resource bindings, inline-constant ranges, and vertex inputs means the layout is derived from the shader (single source of truth) instead of hand-duplicated in host code and silently drifting. Variants matter because real shaders branch on features (shadows on/off, skinning, fog): a uniform branch evaluated per fragment wastes the GPU, so compile specialized permutations and pick one at bind time — but each combination multiplies build time and pipeline count, so trade permutations against an ubershader that branches on a constant for cold paths. Hot-reload (recompile the changed module, rebuild only the pipelines that reference it) gives shader iteration in seconds without a restart.

## Techniques

- **Offline compile** - Source → binary intermediate via a shader compiler as a build step; ship bytecode, not source.
- **Reflection** - Parse the compiled module for binding slots, resource types, inline-constant size, and vertex layout; build binding layouts from it automatically.
- **Permutations / variants** - Compile a specialized module per feature-flag combination; select the right one at material/pipeline creation. Best for hot paths and divergent features.
- **Ubershader + compile-time constants** - One module whose branches resolve at pipeline-creation time via specialization constants; fewer source variants, more pipelines only where actually instantiated. Best for many shallow toggles.
- **Pipeline key** - Key the pipeline cache on (shader modules + render state: blend, depth, raster, formats); identical keys share one compiled pipeline.
- **Hot-reload** - Watch shader source mtimes; on change recompile, validate, then rebuild every pipeline whose key references the changed module and swap them in at a frame boundary.
- **Caching** - Persist compiled bytecode and the driver's pipeline cache blob to disk so cold start skips recompilation.
- **Declarative composition** - Author shaders as composable declarations (named resource/constant `imports`, code blocks, state blocks) that merge by rule — code concatenated in order, state last-write-wins — and generate `load_<const>()` / `get_<resource>()` accessors and the interpolator structs automatically, so authoring is decoupled from the binding API. Build a library of stackable declarations rather than monolithic shaders.
- **System-bitmask variant selection** - Give each optional feature/"system" a unique bit; the set of active systems forms a bitmask that selects the precompiled variant in O(1), instead of a string/hash lookup recomputed on every state change.

## Example

```c
// Layout is reflected from the compiled module, not hand-written -> cannot drift.
// (Neutral pseudocode; the concrete compiler/reflector lives in the per-API skill.)
shader_blob vs = compile_offline("mesh.vert", STAGE_VERTEX);   // build step / hot-reload
shader_blob fs = compile_offline("mesh.frag", STAGE_FRAGMENT);
shader_reflection r = reflect(fs);                             // bindings, constant ranges, formats
binding_layout layout = make_layout_from(r);                   // derived, single source of truth

// Variant selection: a feature flag picks a pre-compiled permutation, no per-pixel branch.
uint32_t variant = (material->cast_shadows ? VARIANT_SHADOW : 0)
                 | (mesh->skinned          ? VARIANT_SKIN   : 0);
shader_module fs_module = shader_variant(cache, "mesh.frag", variant);

// Pipeline keyed on shader + state; identical keys return the cached pipeline.
pipeline_key key = { .vs = vs.hash, .fs = fs_module_hash(variant),
                     .blend = BLEND_OPAQUE, .depth = DEPTH_LESS, .color_fmt = RGBA16F };
pipeline pso = pipeline_cache_get_or_build(cache, &key);       // disk-backed cache avoids stalls
```

## Gotchas

- The permutation count is combinatorial: N independent feature flags is 2^N modules and 2^N build/pipeline entries — gate rarely-toggled features behind an ubershader branch instead.
- Reflection reports what the bytecode declares; an unused binding optimized away by the compiler vanishes from reflection, so layouts built purely from reflection can mismatch a hand-written host struct — make the shader the source of truth, not both.
- First use of an uncached pipeline compiles synchronously and stalls the frame; warm the cache at load and persist the driver blob, see [references/binding-model.md](./binding-model.md).
- Hot-reload must rebuild pipelines and swap at a safe frame boundary, never mid-recording into an in-flight command stream.
- Compile-time/specialization constants resolve at pipeline-creation time, not draw time; changing one still requires a new pipeline.

## Related

[references/binding-model.md](./binding-model.md), [references/render-graph.md](./render-graph.md), [references/command-recording-and-frames.md](./command-recording-and-frames.md)
