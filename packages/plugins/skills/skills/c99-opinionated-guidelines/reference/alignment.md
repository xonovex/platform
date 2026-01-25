# alignment: Memory Alignment

**Guideline:** 16 bytes for SIMD, cache line (64/128B) for hot data.

**Rationale:** SIMD loads require alignment. Misalignment causes penalties/crashes. Cache alignment prevents false sharing in concurrent code.

**Example:**

```c
// SIMD-friendly vec3 (16 bytes)
typedef struct { float x, y, z, _pad; } vec3_t;

// Compiler alignment
typedef struct { _Alignas(16) float data[4]; } aligned_vec4_t;

// Cache-line aligned hot data
#define CACHE_LINE_SIZE 64
typedef struct {
    _Alignas(CACHE_LINE_SIZE) uint32_t counter;
    char pad[CACHE_LINE_SIZE - sizeof(uint32_t)];
} atomic_counter_t;

// SoA arrays
typedef struct {
    _Alignas(16) float *x, *y, *z;
    size_t count;
} vec3_soa_t;

// Aligned allocation
void *aligned_alloc_16(size_t size) {
    void *ptr;
    return posix_memalign(&ptr, 16, size) == 0 ? ptr : NULL;
}

// Platform detection
#if defined(__x86_64__) || defined(_M_X64)
    #define CACHE_LINE_SIZE 64
#elif defined(__aarch64__)
    #define CACHE_LINE_SIZE 128
#else
    #define CACHE_LINE_SIZE 64
#endif
```

**Techniques:**
- SIMD alignment: Pad vec3 to 16-byte aligned structures for SIMD instructions
- Compiler directives: Use `_Alignas(16)` for explicit struct field alignment
- SoA layout: Align array-of-structures for batch SIMD operations
- Cache alignment: Align hot data to 64B (x86) or 128B (ARM) cache line size
- Platform detection: Use conditional compilation for architecture-specific alignment
