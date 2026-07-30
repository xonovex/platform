# alignment: Memory Alignment (C)

The C mechanics for the alignment other concerns require. The _why/when_ lives elsewhere: SIMD-friendly layout in **data-oriented-design-guide**, cache-line padding against false sharing in **lock-free-guide**.

1. Strict ISO C99 has no portable over-alignment syntax. Put the compiler extension behind one project macro and reject unsupported compilers explicitly. The compact example below targets POSIX with GCC or Clang; other platforms need their own macro and allocator pair.
2. On POSIX, expose `posix_memalign` with `_POSIX_C_SOURCE=200112L` or `_XOPEN_SOURCE=700` before any header; C11 `aligned_alloc` is outside this guide.
3. Pad a struct to a whole multiple of its alignment so an _array_ of it stays aligned per element.
4. Supply cache-line size from target configuration or a measured platform constant; an ISA macro alone does not determine it.

```c
#define _POSIX_C_SOURCE 200112L
#include <stddef.h>
#include <stdlib.h>

#if defined(__GNUC__) || defined(__clang__)
#define PROJECT_ALIGN(N) __attribute__((aligned(N)))
#else
#error "Define PROJECT_ALIGN for this compiler"
#endif

typedef struct { float x, y, z, pad; } vec3_t;
typedef struct { float data[4]; } PROJECT_ALIGN(16) vec4a_t;

void *project_alloc_aligned(size_t alignment, size_t size) {
  void *memory = NULL;
  return posix_memalign(&memory, alignment, size) == 0 ? memory : NULL;
}
```

## Gotchas

- An aligned type in an array only stays aligned when its size is a whole multiple of the requested alignment: add explicit padding where necessary.
- `malloc` guarantees alignment suitable for fundamental types, not arbitrary SIMD or cache-line over-alignment; use the platform allocator paired with the guarded alignment macro.
- `posix_memalign` requires a power-of-two alignment that is also a multiple of `sizeof(void *)`; validate dynamic values before calling it.

## Related

**data-oriented-design-guide** (SIMD-friendly layout), **lock-free-guide** (false sharing)
