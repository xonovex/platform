# dependency-tracking: Recording Dependencies for Exact Incremental Reimport

**Guideline:** Record every input each cook actually consumed — source files, included sub-resources, referenced assets, and settings — so that when a source changes you can recompile exactly its dependents (the transitive closure) and leave everything else untouched.

**Rationale:** Without a dependency record you face two bad options: recook everything on any change (slow), or guess which outputs are affected (wrong). A precise record makes incremental reimport both fast and correct: it tells you the exact set of outputs whose validity hash could change, so an edit to one shared header, texture, or material recompiles only the meshes and materials that referenced it. The mechanism dovetails with content hashing — because an output's validity hash folds in the hashes of its inputs, a changed input propagates a new hash down the dependency edges automatically, and the cache lookup misses for precisely the stale outputs. References between assets should use stable identity (a GUID/object identity that survives renames and moves) rather than fragile paths, so a dependency edge does not break when an asset is reorganized. The dominant failure mode is under-reporting: a dependency the cook read but never recorded is invisible, so editing it reimports nothing and ships stale runtime data. The tradeoff is that compilers must report what they touch, which means routing all input access through a tracked interface rather than reading files directly.

**How to Apply:**

1. While cooking, capture every input the compiler reads — the primary source, any included files (shader headers, sub-meshes), and every referenced asset — into a per-output dependency list.
2. Store dependencies by stable identity (GUID/object id), not by path, so moving or renaming an asset does not silently break the edge; reserve name/path references for cases that genuinely need late binding.
3. On a change, walk the reverse edges to find the dependency closure of the changed input, and queue exactly those outputs for recompile.
4. Let hash propagation do the deciding: recompute validity hashes down the closure; outputs whose hash is unchanged (the edit did not actually affect them) stay cached, so the closure is an upper bound and the hash trims it to the truly stale set.
5. Persist the dependency graph alongside the cache so incremental reimport survives a restart and does not require a full rebuild to rediscover edges.
6. Prefer over-reporting to under-reporting: if unsure whether something is an input, record it — a spurious edge costs a redundant cook, a missing edge ships wrong data.

**Example:**

```c
// Every read goes through the tracker so the cook's true input set is recorded.
typedef struct dep_recorder_t { asset_id_t out; id_array_t *inputs; } dep_recorder_t;

const void *track_read(dep_recorder_t *r, asset_id_t input_id) {
    id_array_push(r->inputs, input_id);     // record by stable identity, not path
    return asset_load(input_id);
}

// material cook records the textures + shader header it actually consumed
void material_compile(dep_recorder_t *r, const material_src_t *m, runtime_material_t *o) {
    const shader_inc_t *hdr = track_read(r, m->shader_header); // edge: header -> material
    for (size_t i = 0; i < m->n_textures; ++i)
        track_read(r, m->textures[i]);                         // edge: texture -> material
    lower_material(m, hdr, o);
}

// On a source change: queue the reverse closure; hash propagation trims the unchanged.
void on_source_changed(graph_t *g, asset_id_t changed) {
    id_array_t closure = reverse_closure(g, changed); // exact set of possible dependents
    for (size_t i = 0; i < closure.len; ++i)
        queue_recook(closure.ids[i]); // cache miss only for those whose validity hash moved
}
```

**Gotchas:**

- An input read outside the tracked interface (a direct `fopen`, an ambient config read) is an invisible dependency; editing it reimports nothing and ships stale data — route all input access through the tracker.
- Path-based dependency edges break on rename/move and create phantom misses or stale hits; use stable identity for edges and accept name references only where late binding is intended.
- Recording only direct inputs misses transitive ones (a header that includes another header); the closure must be transitive or the invalidation is incomplete.
- Treating the reverse closure as the final recook set, without hash propagation, over-rebuilds; let the validity hash trim the closure to the genuinely stale outputs.
- Losing the persisted graph forces a full rebuild and erases incrementality after a crash or cache move — persist it next to the cache.

**Related:** [references/content-hash-and-cache.md](./content-hash-and-cache.md), [references/hot-reloading-content.md](./hot-reloading-content.md), **data-model-guide**
