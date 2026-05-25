# shader-system: Shaders as Compiled Build Artifacts

**Guideline:** Treat shaders as a build artifact: author in a source language, compile to an intermediate binary (SPIR-V) offline, reflect the binary to derive descriptor/layout information, manage feature variants explicitly, key pipeline-state objects on shader-plus-state, and support hot-reload by recompiling and rebuilding affected pipelines.

**Rationale:** Compiling shader source at runtime is slow, pulls a full compiler into the shipping binary, and hides syntax errors until the draw call that needs them. Compiling to a binary intermediate offline moves errors to build time, removes the compiler from the runtime, and lets a pipeline cache key on stable bytecode. Reflecting that binary to discover its descriptor bindings, push-constant ranges, and vertex inputs means the layout is derived from the shader (single source of truth) instead of hand-duplicated in C and silently drifting. Variants matter because real shaders branch on features (shadows on/off, skinning, fog): a uniform branch evaluated per fragment wastes the GPU, so compile specialized permutations and pick one at bind time — but each combination multiplies build time and PSO count, so trade permutations against an ubershader that branches on a uniform/spec-constant for cold paths. Hot-reload (recompile the changed module, rebuild only the pipelines that reference it) gives shader iteration in seconds without a restart.

**Techniques:**

- **Offline compile** - Source → SPIR-V via a shader compiler (glslang-class) as a build step; ship bytecode, not source.
- **Reflection** - Parse the compiled module for binding numbers, descriptor types, push-constant size, and vertex layout; build descriptor-set layouts from it automatically.
- **Permutations / variants** - Compile a specialized module per feature-flag combination; select the right one at material/pipeline creation. Best for hot paths and divergent features.
- **Ubershader + specialization constants** - One module whose branches resolve at pipeline-creation time via spec constants; fewer source variants, more PSOs only where actually instantiated. Best for many shallow toggles.
- **PSO key** - Key the pipeline cache on (shader modules + render state: blend, depth, raster, formats); identical keys share one compiled PSO.
- **Hot-reload** - Watch shader source mtimes; on change recompile to SPIR-V, validate, then rebuild every PSO whose key references the changed module and swap them in at a frame boundary.
- **Caching** - Persist compiled SPIR-V and the driver's PSO cache blob to disk so cold start skips recompilation.

**Example:**

```c
// Layout is reflected from the compiled module, not hand-written -> cannot drift.
spirv_blob vs = compile_to_spirv("mesh.vert", stage_vertex);   // build step / hot-reload
spirv_blob fs = compile_to_spirv("mesh.frag", stage_fragment);
shader_reflection r = reflect(fs);                             // bindings, push ranges, formats
VkDescriptorSetLayout set_layout = make_set_layout_from(r);    // derived, single source of truth

// Variant selection: a feature flag picks a pre-compiled permutation, no per-pixel branch.
uint32_t variant = (material->cast_shadows ? VARIANT_SHADOW : 0)
                 | (mesh->skinned        ? VARIANT_SKIN   : 0);
VkShaderModule fs_module = shader_variant(cache, "mesh.frag", variant);

// PSO keyed on shader + state; identical keys return the cached pipeline.
pso_key key = { .vs = vs.hash, .fs = fs_module_hash(variant),
                .blend = BLEND_OPAQUE, .depth = DEPTH_LESS, .color_fmt = RGBA16F };
VkPipeline pso = pso_cache_get_or_build(cache, &key);          // disk-backed cache avoids stalls
```

**Gotchas:**

- The permutation count is combinatorial: N independent feature flags is 2^N modules and 2^N build/PSO entries — gate rarely-toggled features behind an ubershader branch instead.
- Reflection reports what the bytecode declares; an unused binding optimized away by the compiler vanishes from reflection, so layouts built purely from reflection can mismatch a hand-written C struct — make the shader the source of truth, not both.
- First use of an uncached PSO compiles synchronously and stalls the frame; warm the cache at load and persist the driver blob, see [references/pipelines-and-descriptors.md](./pipelines-and-descriptors.md).
- Hot-reload must rebuild PSOs and swap at a safe frame boundary, never mid-recording into an in-flight command buffer.
- Spec constants resolve at pipeline-creation time, not draw time; changing one still requires a new PSO.

**Related:** [references/pipelines-and-descriptors.md](./pipelines-and-descriptors.md), [references/render-graph.md](./render-graph.md), [references/command-buffers-and-frames.md](./command-buffers-and-frames.md)
