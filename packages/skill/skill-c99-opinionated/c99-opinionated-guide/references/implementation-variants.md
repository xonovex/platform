# implementation-variants: Implementation Variant Strategy (C)

Ship a routine as progressive variants: scalar reference → AoS batch → SoA batch → SIMD, under the `_aos`/`_soa`/`_simde` file/symbol convention. Layout rationale (cache, vectorization) is in **data-oriented-design-guide**; this is the C convention for organizing variants and keeping them honest.

1. Write the scalar single-object reference first; it is the correctness source of truth.
2. Add `_aos`/`_soa`/`_simde` variants as needed, named by suffix (see file-naming).
3. Run the _same_ test suite across all variants. Every batched/SIMD variant must match the scalar output (parity).
4. Only add the variants a workload actually needs; don't pre-ship SIMD for cold code.

## Example

```c
// Scalar reference (correctness source of truth)
float vec3_dot(const vec3_t *a, const vec3_t *b) {
  return a->x * b->x + a->y * b->y + a->z * b->z;
}

// SoA batch (cache-friendly variant; layout rationale -> data-oriented-design-guide)
typedef struct { float *x, *y, *z; size_t count; } vec3_soa_t;
void vec3_dot_soa(float *out, const vec3_soa_t *a, const vec3_soa_t *b) {
  for (size_t i = 0; i < a->count; i++)
    out[i] = a->x[i] * b->x[i] + a->y[i] * b->y[i] + a->z[i] * b->z[i];
}
```

## Related

**data-oriented-design-guide** (layout/SIMD rationale), [references/file-naming.md](./file-naming.md), [references/testing-patterns.md](./testing-patterns.md)
