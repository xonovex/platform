# math-types: Game Math Types

**Guideline:** Use 16-byte aligned vectors/matrices/quaternions. Union-based for named and indexed access.

**Rationale:** 16-byte alignment enables SIMD operations and efficient cache access; unions allow both semantic and array-based access.

**Example:**

```c
typedef union {
    struct { float x, y, z, _pad; };
    float col[4];
} vec3f_t;

typedef union {
    float col[16];
    vec4f_t columns[4];
} matrix4f_t;

// Usage
vec3f_t pos = {.x = 1.0f, .y = 2.0f, .z = 3.0f};
matrix4f_t m;
m.columns[3] = (vec4f_t){.x = tx, .y = ty, .z = tz, .w = 1.0f};
```

**Techniques:**

- 16-byte alignment: Pad vec3 with \_pad field for SIMD-friendly 16B size
- Union design: Combine named fields with array access for flexibility
- Column-major matrices: Store matrix4f_t as 4 column vectors
- Quaternion format: Use x, y, z, w order with normalization constraint
- Array access: Support both `.x`, `.y` named and `[0]`, `[1]` indexed access
