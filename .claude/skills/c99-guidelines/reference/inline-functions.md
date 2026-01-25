# inline-functions: Inline Functions vs Macros

**Guideline:** Prefer static inline functions over macros for type safety and predictable behavior.

**Rationale:** Inline functions provide type checking, proper scoping, and avoid macro side effects like multiple evaluation of arguments.

**Example:**

```c
// Macro - unsafe, no type checking
#define CLAMP(v, lo, hi) ((v) < (lo) ? (lo) : ((v) > (hi) ? (hi) : (v)))

// Inline function - type safe
static inline int clampi(int v, int lo, int hi) {
    if (v < lo) return lo;
    if (v > hi) return hi;
    return v;
}

static inline float clampf(float v, float lo, float hi) {
    if (v < lo) return lo;
    if (v > hi) return hi;
    return v;
}

// Macro side effect: x++ evaluated multiple times
// Inline function: x++ evaluated once safely
int x = 5;
int result = clampi(x++, 0, 10);
```

**Techniques:**
- Replace macros: Replace function-like macros with `static inline` functions
- Type-specific variants: Create `clampi`, `clampf` instead of generic macro
- Header placement: Use inline functions in headers for no multiple definition issues
- Avoid side effects: Inline functions safely evaluate arguments once
- Simple logic: Keep inline functions simple for compiler optimization
