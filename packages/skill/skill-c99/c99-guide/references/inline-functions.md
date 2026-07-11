# inline-functions: Inline Functions vs Macros

Prefer `static inline` functions over function-like macros: they type-check and evaluate each argument once. A macro re-evaluates its arguments, so `CLAMP(x++, 0, 10)` steps `x` multiple times. C99 has no generics, so write per-type variants (`clampi`, `clampf`). Place `static inline` in headers to avoid multiple-definition errors.

```c
#define CLAMP(v, lo, hi) ((v) < (lo) ? (lo) : ((v) > (hi) ? (hi) : (v)))  // x++ stepped 2-3x

static inline int clampi(int v, int lo, int hi) {
    return v < lo ? lo : (v > hi ? hi : v);
}
```
