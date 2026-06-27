# content-hash-and-cache: Content-Addressed Caching of Cooked Output

## Guideline

Compute a content hash over (source bytes + settings + the content hashes of every dependency) and store each cooked output under that hash in a content-addressed cache, so identical inputs always hit the same entry and the cache is reproducible and shareable across machines.

## Rationale

Once compilation is deterministic, its output is a pure function of its inputs, so the hash of the inputs is a complete identity for the output — the same principle Git uses, where the same content is always referred to by the same identifier. Keying the cache on that hash means a cook is done at most once per distinct (input, settings, platform) combination anywhere: a teammate or CI machine that already produced the artifact can serve it directly, turning a cold build into a download. The hash must fold in the dependencies' hashes (not just the direct source), so that changing an upstream input changes this output's key transitively. The tradeoff is that the hash must cover everything that can affect the output and nothing that cannot: include a path, host name, or timestamp and the cache stops being content-addressed — two machines cooking the same source get different keys and can never share. The reward is reproducibility (the same inputs always rebuild the same bytes) and a cache that is safe to share precisely because the key is derived from content, not location.

## How to Apply

1. For each output, compute a 64-bit (or wider) validity hash from the canonical source content, the canonicalized settings, the target platform, and the validity hashes of each input resource it consumed.
2. Use that hash as the cache key; look it up before cooking — on a hit, skip the compile and serve the stored artifact; on a miss, cook, store, and serve.
3. Hash canonical content, never raw memory: avoid struct padding, normalize floating-point representations, sort unordered collections before hashing.
4. Exclude location and environment from the hash entirely (absolute paths, machine names, wall-clock times) so the key is purely a function of content and settings.
5. Make the cache shareable: a local disk cache backed by a shared remote store lets one machine's cook satisfy another's lookup; because keys are content-derived, no coordination is needed.
6. Apply caching selectively — wrap the expensive cooks (compression, mip filtering, mesh optimization) and skip caching nodes whose compute is cheaper than computing and looking up the hash.

## Example

```c
// validity_hash folds in settings + the hashes of every input the cook consumed.
uint64_t texture_validity_hash(const image_t *src, tex_settings_t s,
                               const uint64_t *input_hashes, size_t n_inputs) {
    hasher_t h = hasher_init();
    hasher_add(&h, image_canonical_bytes(src), image_canonical_size(src));
    hasher_add(&h, &s, settings_canonical_size(&s)); // canonical: no padding
    for (size_t i = 0; i < n_inputs; ++i)
        hasher_add(&h, &input_hashes[i], sizeof input_hashes[i]);
    return hasher_finish(&h);
}

runtime_texture_t cache_get_or_cook(cache_t *c, const image_t *src, tex_settings_t s,
                                    const uint64_t *deps, size_t n) {
    uint64_t key = texture_validity_hash(src, s, deps, n);
    runtime_texture_t out;
    if (cache_lookup(c, key, &out)) return out;        // hit: skip the cook entirely
    texture_compile(src, s, &out);                     // miss: cook deterministically
    cache_store(c, key, &out);                         // shareable: key is content-derived
    return out;
}
```

## Gotchas

- A hash that omits a real input (the settings, a dependency, the target platform) serves stale output after that input changes — under-covering the hash is a correctness bug, not a perf bug.
- A hash that includes irrelevant state (timestamps, absolute paths, host names) over-invalidates and breaks cross-machine sharing — every machine computes a different key for identical content.
- Hashing raw structs drags in uninitialized padding and pointer values; hash canonical serialized content so the same logical input always hashes the same.
- Caching a node whose work is cheaper than the hash-and-lookup makes things slower; reserve the cache for genuinely heavy steps.
- A 64-bit hash can collide in principle; for a shared, long-lived cache prefer a wider hash or verify content on hit if a collision would ship wrong data.

## Related

[references/import-and-compile.md](./import-and-compile.md), [references/dependency-tracking.md](./dependency-tracking.md), **data-model-guide**
