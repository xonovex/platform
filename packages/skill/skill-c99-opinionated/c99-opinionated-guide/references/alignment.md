# alignment: Memory Alignment (C)

## Guideline

Use C's alignment mechanics — `_Alignas`, over-aligned types, and an aligned allocator — to satisfy the SIMD and cache-line alignment that other concerns require.

## Rationale

This doc is only the C _how_. The _why/when_ lives elsewhere: aligning/padding for vectorization is a layout decision (see **data-oriented-design-guide**, SIMD-friendly layout), and cache-line-aligning/padding hot or atomic fields to avoid false sharing is a concurrency decision (see **lock-free-guide**). Here: how to express those in C.

## How to Apply (C specifics)

1. `_Alignas(16)` (or `alignas`, C11) on a field/type for SIMD; pad a `vec3` to 16 bytes when it must load as a `vec4`.
2. Allocate over-aligned memory with `aligned_alloc` (C11) or `posix_memalign`; `malloc` only guarantees max-align.
3. Pad a struct to a whole multiple of the alignment so an _array_ of it stays aligned per element.
4. Source the cache-line size per target with conditional compilation (64 B x86-64, 128 B some ARM).

## Example

```c
typedef struct { float x, y, z, _pad; } vec3_t;              // 16B, SIMD-loadable
typedef struct { _Alignas(16) float data[4]; } vec4a_t;     // explicit field alignment

#if defined(__aarch64__)
#define CACHE_LINE_SIZE 128
#else
#define CACHE_LINE_SIZE 64
#endif
typedef struct { _Alignas(CACHE_LINE_SIZE) uint32_t v; char pad[CACHE_LINE_SIZE - 4]; } padded_t;

static void *aligned_alloc_16(size_t n) {                   // over-aligned allocation
  void *p; return posix_memalign(&p, 16, n) == 0 ? p : NULL;
}
```

## Gotchas

- `_Alignas` on a struct field aligns the field, but an array of that struct only stays aligned if the struct size is a whole multiple of the alignment — add explicit padding.
- `malloc` returns max-align (fine for scalars), not arbitrary over-alignment; use `aligned_alloc`/`posix_memalign` for SIMD/cache-line needs.

## Related

**data-oriented-design-guide** (SIMD-friendly layout), **lock-free-guide** (false sharing)
