# access-patterns: Linear Access and Avoiding Pointer Chasing

**Guideline:** Iterate data sequentially over contiguous arrays and avoid pointer chasing and random access, so the hardware prefetcher and cache hide memory latency.

**Rationale:** The hardware prefetcher detects constant-stride access and fetches lines ahead of the loop, hiding most miss latency. Pointer chasing (linked lists, trees of heap nodes, graphs of objects) defeats it: each `node = node->next` is a dependent load that cannot start until the previous one returns, so latency is fully exposed and the access is effectively random across the heap. Sequential streaming over a packed array is often an order of magnitude faster than the "same" algorithm over linked nodes, even at equal asymptotic complexity.

**Techniques:**

- **Sequential / streaming / linear access** - Walk arrays front-to-back with a constant stride; this is the prefetcher's ideal and the baseline to design toward.
- **Index-based iteration over linked structures** - Store nodes in an array and link by integer index, not pointer; better, store them in traversal order so iteration is a plain sweep.
- **Batch over one-at-a-time** - Process a run of elements per call so the loaded line, prefetch, and instruction cache are reused before eviction.
- **Avoid random access** - Hash-bucket scatter, indirection tables, and sorting by an unrelated key turn streaming into random fetches; gather/sort into order first, then stream.
- **Hot in front** - When elements have differing liveness, compact live ones to the front so the active range is contiguous.

**How to Apply:**

1. Find loops that follow pointers or jump by data-dependent indices.
2. Replace heap-node graphs with arrays; replace `->next` traversal with `++i` over a packed array in iteration order.
3. If you must index indirectly, pre-sort or gather into a contiguous scratch buffer, then stream that.
4. Verify the prefetcher is engaged: a profiler should show high IPC and low memory-stall on the loop.

**Example:**

```c
// Bad: linked list — each step is a dependent load to a random heap address.
typedef struct node { int v; struct node *next; } node_t;
long sum_list(const node_t *h) {
  long s = 0;
  for (const node_t *n = h; n; n = n->next) s += n->v; // latency fully exposed
  return s;
}

// Good: packed array, constant stride — prefetcher streams ahead.
long sum_array(const int *v, size_t n) {
  long s = 0;
  for (size_t i = 0; i < n; i++) s += v[i];            // latency hidden
  return s;
}
```

**Gotchas:**

- Linked structures can still be fine if nodes are pool-allocated in traversal order and rarely re-linked — it is the random heap addresses, not the "list" concept, that hurts.
- Indexing by a generational handle is sequential-friendly only if the underlying array is itself traversed in order; random handle lookups are still random.

**Related:** [references/handles-and-indices.md](./handles-and-indices.md), [references/cache-behavior.md](./cache-behavior.md), [references/data-as-transforms.md](./data-as-transforms.md), [references/memory-arenas.md](./memory-arenas.md)
