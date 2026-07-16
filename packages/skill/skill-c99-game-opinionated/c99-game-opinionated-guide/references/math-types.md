# math-types: Game Math Types

16-byte-aligned vectors/matrices/quaternions as unions exposing both named (`.x`, `.y`) and indexed (`col[]`) access. Pad `vec3f_t` with `_pad` to hit 16B for SIMD.

```c
typedef union {
    struct { float x, y, z, _pad; };
    float col[4];
} vec3f_t;

typedef union {
    float col[16];
    vec4f_t columns[4];
} matrix4f_t;

vec3f_t pos = {.x = 1.0f, .y = 2.0f, .z = 3.0f};
```
