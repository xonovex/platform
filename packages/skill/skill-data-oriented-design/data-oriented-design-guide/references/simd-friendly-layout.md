# simd-friendly-layout: SIMD-Friendly Data Layout

## Guideline

Lay data out as contiguous, aligned, same-typed columns (SoA or AoSoA) so loops vectorize with cheap aligned loads instead of expensive gathers.

## Rationale

SIMD processes N lanes per instruction, but only when each lane's input is contiguous and the load address is aligned to the vector width. SoA columns give exactly that: lane `k` reads `field[i+k]` as one aligned vector load. AoS forces the compiler to gather one scalar per record (slow, often non-vectorizable). Alignment matters because an unaligned vector load is slower or, on some ISAs, faults; padding to the vector width lets the tail iterate branch-free. Good layout is the difference between auto-vectorization and scalar fallback.

## Techniques

- **SoA columns enable contiguous loads** - Store each field in its own array; the inner loop over the array becomes a sequence of aligned vector loads/ops/stores.
- **Pad count to SIMD width** - Round the element count up to a multiple of the lane count and process full vectors; mask or ignore the padding lanes, avoiding a scalar remainder loop.
- **Alignment (16/32/64B)** - Align column base addresses to the vector width: 16B for SSE/NEON (4 floats), 32B for AVX (8 floats), 64B for AVX-512 (16 floats) — and 64B also aligns to a cache line.
- **AoSoA for cache + SIMD** - Tile width = lane count: each tile is one (or a few) vector(s) and sits in one cache line, giving both locality and vectorizability.
- **Batch of N** - Structure the algorithm to consume N lanes at a time; keep per-lane control flow uniform (no divergent branches) so the vector stays full.

## How to Apply

1. Convert the vectorized field(s) to SoA columns (or AoSoA tiles of width = lane count).
2. Align each column to the vector width with the language's alignment facility; allocate via an aligned allocator.
3. Pad the element count up to a multiple of the lane count; initialize padding to a neutral value.
4. Keep the inner loop branch-free and same-typed; check the disassembly/vectorizer report to confirm vector instructions.

## Example

```c
#define N 1024
#define LANES 8            // AVX: 8 floats
#define NPAD ((N + LANES - 1) & ~(LANES - 1)) // round up to lane multiple

// SoA columns, 32-byte aligned -> aligned vector loads, no gather.
typedef struct {
  _Alignas(32) float x[NPAD];
  _Alignas(32) float y[NPAD];
  _Alignas(32) float out[NPAD];
} cols_t;

// Branch-free, same-typed inner loop -> auto-vectorizes to 8-wide ops.
static void madd(cols_t *c, float k) {
  for (size_t i = 0; i < NPAD; i++) c->out[i] = c->x[i] * k + c->y[i];
}

// Bad: AoS forces a per-record gather; usually falls back to scalar.
typedef struct { float x, y, out; char tag; } rec_t; // stride 16B, gathered
```

## Gotchas

- Padding lanes must hold safe values (e.g. 0) so they cannot produce NaN/inf, div-by-zero, or out-of-range side effects.
- Alignment is necessary but not sufficient: divergent per-lane branches, dependent loads, or aliasing can still block vectorization.
- Aligned allocation is separate from struct alignment — `malloc` only guarantees max-align; use an aligned allocator for over-aligned columns.

## Related

[references/soa-aos-aosoa.md](./soa-aos-aosoa.md), memory-management-guide, [references/cache-behavior.md](./cache-behavior.md), [references/measurement-and-profiling.md](./measurement-and-profiling.md)
