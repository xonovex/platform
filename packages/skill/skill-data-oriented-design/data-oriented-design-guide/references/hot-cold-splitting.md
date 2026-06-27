# hot-cold-splitting: Hot/Cold Field Splitting

## Guideline

Separate frequently-accessed (hot) fields from rarely-accessed (cold) fields into distinct structures so hot loops never drag cold bytes through the cache.

## Rationale

A "fat struct" mixes fields touched every frame with fields touched once (names, debug data, config, backreferences). Because the cache loads whole 64-byte lines, every hot iteration also pulls in the cold bytes that share the record, inflating the working set and evicting useful data. Splitting shrinks the hot record so far more useful elements fit per line and per cache level, often turning a capacity-miss-bound loop into a streaming one — frequently the single highest-leverage DOD refactor.

## How to Apply

1. Instrument or reason about which fields the hot loops actually read/write every iteration.
2. Move those into a compact `hot` array; move the rest into a parallel `cold` array indexed the same way (or referenced by a handle).
3. Keep the two in lockstep by index so element `i` in hot corresponds to element `i` in cold.
4. Measure: the hot loop should now load fewer cache lines and show fewer capacity misses.

## Example

```c
// Bad: 80-byte record; the position loop only needs 12 bytes but loads all 80.
typedef struct {
  vec3 pos;          // hot: read every frame
  char name[40];     // cold: UI/debug only
  uint64_t spawn_ms; // cold
  void *script;      // cold
  uint32_t flags;    // cold
} entity_t;
entity_t e[N];
for (size_t i = 0; i < N; i++) e[i].pos = step(e[i].pos); // drags 68 cold bytes

// Good: hot fields packed tight (prefetch streams ~5 entities per line);
//       cold fields parked in a parallel array, untouched by the hot loop.
typedef struct { vec3 pos; } hot_t;            // 12 bytes
typedef struct { char name[40]; uint64_t spawn_ms; void *script; uint32_t flags; } cold_t;
hot_t  hot[N];
cold_t cold[N];                                 // hot[i] <-> cold[i]
for (size_t i = 0; i < N; i++) hot[i].pos = step(hot[i].pos);
```

## Gotchas

- "Hot" is per-loop, not global: a field hot for rendering may be cold for AI. Split by the dominant access pattern, or use multiple views.
- Over-splitting into many tiny arrays adds index bookkeeping and can hurt loops that need several of them together; group co-accessed hot fields.
- This is the field-level companion to SoA: SoA splits every field; hot/cold splits the struct into two by frequency.

## Related

[references/soa-aos-aosoa.md](./soa-aos-aosoa.md), [references/cache-behavior.md](./cache-behavior.md), [references/handles-and-indices.md](./handles-and-indices.md)
