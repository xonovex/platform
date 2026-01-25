# implementation-variants: Implementation Variant Strategy

**Guideline:** Progressive variants: scalar single → AoS batch → SoA batch → SIMD.

**Rationale:** Different use cases need different trade-offs: debugging (single), existing structs (AoS), cache-friendly (SoA), peak throughput (SIMD).

**Example:**

```c
// Scalar single (reference)
float vec3_dot(const vec3_t *a, const vec3_t *b) {
    return a->x*b->x + a->y*b->y + a->z*b->z;
}

// AoS batch (struct arrays)
void vec3_dot_aos(float *out, const vec3_t *a, const vec3_t *b, size_t n) {
    for (size_t i = 0; i < n; i++)
        out[i] = a[i].x*b[i].x + a[i].y*b[i].y + a[i].z*b[i].z;
}

// SoA batch (component arrays, cache-friendly)
typedef struct { float *x, *y, *z; size_t count; } vec3_soa_t;

void vec3_dot_soa(float *out, const vec3_soa_t *a, const vec3_soa_t *b) {
    for (size_t i = 0; i < a->count; i++)
        out[i] = a->x[i]*b->x[i] + a->y[i]*b->y[i] + a->z[i]*b->z[i];
}
```

**Techniques:**
- Scalar single: Start with single-object reference implementation for debugging
- AoS batch: Add batch processing for struct arrays (1.5x speedup)
- SoA batch: Add cache-friendly variant for large workloads (2-3x speedup)
- SIMD variants: Add vectorized implementations for critical paths (4-10x)
- Parity testing: Test correctness across all variants with same test suite
