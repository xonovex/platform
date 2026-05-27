# evaluation-and-compilation: Compiling a Graph, Topological Evaluation, and Validity-Hash Caching

**Guideline:** Do not walk the editor graph directly every time you run it; lower (compile) it into a flattened, ordered executable form, evaluate nodes in dependency order so every input is resolved before its consumer runs, and cache each output behind a validity hash that folds in the node's own settings plus the validity hashes of all its inputs, so an unchanged subtree is reused instead of recomputed.

**Rationale:** The editor graph is shaped for authoring — it contains subgraph wrappers, interface nodes, comments, and indirection that exist for humans, not for execution. Interpreting that structure on every evaluation pays for all of it repeatedly and forces the evaluator to special-case editor-only constructs. Compiling once into a flat list of executable nodes in topological order separates "what the author drew" from "what runs," so the run loop is a tight pass with no recursion and no editor concepts. Topological order is required for correctness: a node may only run once each input value exists. The validity hash is what makes re-authoring cheap: because each output's hash is derived from the node's settings and its inputs' hashes, an identical subtree produces an identical hash, and a consumer can decide it already has a valid cached result and skip the work entirely. This is decisive when nodes are expensive (filtering or compressing an image, baking a map) — caching trades a little memory for skipping heavy recompute. The right amount of caching is contextual: an offline authoring graph can afford long bakes and aggressive caches, while a graph re-evaluated at runtime must stay lean.

**How to Apply:**

1. Treat compilation as a pre-pass: produce a flattened node/connection list with all editor-only constructs (subgraph wrappers, interface nodes) removed — see subgraphs-and-functions.md for flattening.
2. Build the dependency DAG from the connections and topologically sort it to get evaluation order.
3. Evaluate in order: for each node, gather resolved inputs (wired value or stored default), run its evaluate callback, publish its outputs.
4. For each output, compute `validity_hash = hash(node_settings, input_hashes...)`; carry it on the value traveling the wire.
5. Before a heavy node runs, compare incoming hashes to what produced its cached result; if unchanged, reuse the cache and skip recompute.
6. Cache the compiled/flattened result itself; invalidate and recompile only when the source graph changes, not on every evaluation.
7. Let a node opt into caching its result only when its operation is heavy enough to justify the extra memory.

**Example:**

```c
// Each value on a wire carries a validity hash computed by its producer.
typedef struct value_t {
    uint64_t type_hash;
    uint64_t validity_hash;   // = hash(node settings, all input validity hashes)
    void    *payload;         // image, buffer, draw call, ...
} value_t;

uint64_t compute_validity(const node_t *n, const value_t *inputs, uint32_t n_in) {
    uint64_t h = hash_settings(n->settings);
    for (uint32_t i = 0; i < n_in; ++i)
        h = hash_combine(h, inputs[i].validity_hash); // fold in EVERY consumed input
    return h;
}

void evaluate(const compiled_graph_t *cg) {     // cg is flat + topologically sorted
    for (uint32_t i = 0; i < cg->num_nodes; ++i) {
        node_t *n = &cg->nodes[i];
        value_t *in = gather_inputs(cg, n);     // wired value or stored default
        uint64_t vh = compute_validity(n, in, n->num_inputs);
        if (cache_has(n, vh)) { publish(n, cache_get(n, vh)); continue; } // skip heavy recompute
        value_t out = n->evaluate(n, in);
        out.validity_hash = vh;
        if (n->cacheable) cache_put(n, vh, out); // opt in only when heavy enough
        publish(n, out);
    }
}
```

**Gotchas:**

- The hash must fold in _every_ input that affects the output; an input read but omitted from the hash produces stale reuse with no error — the worst kind of caching bug.
- Topological sort assumes a DAG; a cycle (often introduced by a bad subgraph inline) has no valid order — detect and reject cycles at compile.
- Re-flattening and re-sorting on every evaluation is correct but throws away the win; cache the compiled form and invalidate only on source change.
- Caching everything wastes memory; reserve caches for genuinely heavy nodes and let cheap nodes recompute.
- Evaluation cost is contextual — an offline bake tolerates what a per-frame runtime graph cannot; tune cache aggressiveness to where the graph runs.
- The evaluator publishes values but does not schedule emitted work (a draw call, a compute dispatch); the owning system does that — see gpu-rendering-guide.

**Related:** [references/subgraphs-and-functions.md](./subgraphs-and-functions.md), [references/node-model.md](./node-model.md), [references/typing-and-connections.md](./typing-and-connections.md), **data-oriented-design-guide**, **gpu-rendering-guide**
