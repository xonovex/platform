# nested-arrays: Variable-Length Nested Data in Bulk Arrays

**Guideline:** Store objects that own variable-length child lists in fixed-size bulk arrays plus an application-specific chunked allocator, not as per-object heap-allocated pointers; size chunks to a cache line, and intern immutable strings so identity is a pointer compare.

**Rationale:** "An object with a list of children" naively becomes a pointer to a heap-allocated growable array per object — one allocation per object, scattered across memory, opaque to the profiler, and a cache miss on every traversal. Holding the child data in a second bulk array and linking fixed-size chunks by index keeps everything contiguous, makes allocation a free-list pop, and lets you tune the layout to the access pattern. The trade-off is explicit: chunking gives up O(1) random indexing for O(n/chunk) traversal and pays internal fragmentation proportional to chunk size, so choose the representation from how the data is actually read.

**How to Apply:**

1. Known small upper bound? Use a **capped fixed array** inline in the struct (`uint32_t children[8]`) — simplest, but wastes space if the cap is loose and must handle overflow. (Safe to raise a cap later in runtime code; not in a file format.)
2. Unbounded growth? Use a **chunked allocator**: hold child indices in fixed-size chunks linked by index, stored in a separate bulk array. Size the chunk to fill a cache line — e.g. 14 × 32-bit indices + 2 link indices = 64 bytes.
3. Many-directional relational queries? Use **linked-list siblings** (`first_child`, `prev/next_sibling`) — good for bidirectional traversal, worse for straight iteration (pointer chasing).
4. Strings as identifiers? **Intern** them into one buffer + hash table so equal strings share one pointer and comparison is pointer equality (ref-count only if you must delete).
5. Prefer bulk arrays + an app-specific allocator over a generic growable container for object pools.

**Example:**

```c
// Bad: one heap allocation per object, scattered; profiler-opaque, miss per traversal
struct object { char *name; struct object **children; uint32_t n, cap; }; // malloc per list

// Good: child indices in cache-line chunks, kept in a bulk array; name interned
enum { CHILD_CHUNK_SIZE = 14 };
struct child_chunk { uint32_t child[CHILD_CHUNK_SIZE]; uint32_t prev_chunk, next_chunk; }; // 64B
struct object { const char *name;  // interned: compare by pointer
                uint32_t num_children, first_chunk; };  // index into the chunk bulk array
```

**Counter-Example:** If children are read by frequent random index (`children[k]`) rather than full traversal, the chunked layout's O(n/chunk) walk is the wrong trade — a capped or contiguous array preserves O(1) indexing. Pick the structure from the dominant access.

**Related:** [handles-and-indices.md](./handles-and-indices.md), [access-patterns.md](./access-patterns.md), [soa-aos-aosoa.md](./soa-aos-aosoa.md)
