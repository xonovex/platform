# import-and-compile: Importers, Per-Type Compilers, and Deterministic Cooking

## Guideline

Bring source files in through an importer selected by file format and produce runtime data through a per-asset-type compiler that is a deterministic function of (source bytes, settings, target platform) — same inputs, same output, every time, on every machine.

## Rationale

Splitting import (parse one of many source formats into the intermediate) from compile (cook the intermediate into runtime data per type) lets you add a new source format by writing one importer, and change runtime packing by editing one compiler, without coupling the two. Making the compile step deterministic is what makes everything downstream possible: a content-addressed cache only works if identical inputs reliably produce identical output, and a shareable build cache across machines only works if no hidden state (wall-clock time, hash-map iteration order, absolute paths, environment) leaks into the result. Treating settings and target platform as explicit inputs — not ambient configuration — means a changed compression preset or a different platform deterministically produces a different output and a different cache key, instead of silently reusing a stale cooked artifact. The tradeoff is that compilers must be pure and side-effect-free, which forbids convenient shortcuts like reading global config or embedding the build host name.

## How to Apply

1. Register importers behind a single interface keyed by file extension; the pipeline enumerates them and picks the first capable of the format, then runs the import as an async task into the intermediate.
2. Give each asset type its own compiler that reads the intermediate object plus its settings and emits runtime data — e.g. textures generate and compress mips, meshes optimize vertex cache and quantize attributes, materials lower a shader/material graph to a compiled program.
3. Make settings part of the input signature: mip count, compression format, quantization precision, target platform, and any per-asset overrides all feed the compiler and the cache key.
4. Forbid non-determinism in compiler output: no timestamps, no absolute or machine-specific paths, no unordered-container iteration order, no uninitialized padding in serialized structs.
5. Emit per-target runtime data (texture block format, endianness, alignment, shader bytecode) and fold the target identity into both the output and the cache key so each platform gets its own entry.
6. Run imports and compiles off the main thread so authoring stays responsive; the work is a pure function, so it parallelizes naturally.

## Example

```c
// Importers keyed by extension, behind one interface.
typedef struct asset_io_i {
    bool (*can_import)(const char *ext);
    void (*import)(const char *path, dcc_scene_t *out); // async; lands in intermediate
} asset_io_i;

typedef struct tex_settings_t {
    uint32_t  mip_count;
    enum { BC7, ASTC_6x6, RGBA8 } compression;
    target_t  platform;   // settings + platform are INPUTS, not ambient state
} tex_settings_t;

// Deterministic: out is a pure function of (src, s). No clock, no abs paths, no env.
void texture_compile(const image_t *src, tex_settings_t s, runtime_texture_t *out) {
    generate_mips(src, s.mip_count, out);
    compress(out, s.compression, s.platform); // platform-specific block format
}
```

## Gotchas

- Reading config, locale, or environment variables inside a compiler makes its output depend on the machine; the same source then cooks differently on CI and on a workstation, poisoning a shared cache.
- Hashing a settings struct with uninitialized padding bytes yields a different key on each build even when the meaningful settings are identical — serialize canonical fields, not the raw struct.
- Doing heavy work in the importer instead of the compiler means it cannot be recook per platform or skipped on reimport; importers should land a faithful intermediate and stop.
- A compiler that depends on the order assets are processed (global mutable state, shared counters) is not deterministic even if each input looks pure in isolation.

## Related

[references/raw-vs-runtime-formats.md](./raw-vs-runtime-formats.md), [references/content-hash-and-cache.md](./content-hash-and-cache.md), **data-model-guide**
