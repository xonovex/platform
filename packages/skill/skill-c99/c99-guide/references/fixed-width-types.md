# Fixed-width integer types

C's built-in `int`/`long`/`unsigned` have implementation-defined width — the standard guarantees only minimums. For any value whose size, range, or byte layout matters — serialized records, hardware/protocol fields, hashes, bitmasks, indices into large arrays, anything that must behave identically across platforms — name the width with `<stdint.h>`.

## Use exact-width types for sized and serialized data

```c
#include <stdint.h>
#include <stdbool.h>

typedef struct {
    uint32_t id;        /* exactly 32 bits on every platform */
    int16_t  offset;    /* wraps and serializes identically everywhere */
    uint8_t  flags;
    bool     active;    /* <stdbool.h>, not an int flag */
} record_t;
```

- `uintN_t` / `intN_t` (`8/16/32/64`) — exact width, two's complement, no padding. The default for struct fields, file/wire formats, bitmasks, and hashes.
- Bare `int` is fine for a _local_ loop counter or small arithmetic where only the C minimum (≥16 bits) matters — don't `uint32_t` a `for` index out of reflex. The rule governs _stored and shared_ data, not every variable.
- `bool` from `<stdbool.h>`, never an `int` flag — it states intent and is one byte.

## size_t, ptrdiff_t, and pointer-sized ints

- `size_t` for sizes, counts, and indices that track an object's element count — it is the type of `sizeof` and the width of the address space. Mixing `int` and `size_t` in a comparison invites a sign-conversion bug: a negative `int` becomes a huge `size_t`.
- `ptrdiff_t` for the signed difference of two pointers.
- `intptr_t` / `uintptr_t` only to round-trip a pointer through an integer (tagging, alignment math) — not as a general "big int".

## Least / fast variants (rarely needed, worth knowing)

- `uint_least32_t` — the smallest type with ≥32 bits; for huge arrays where storage dominates.
- `uint_fast32_t` — the fastest type with ≥32 bits; for a hot scalar where speed dominates.

Reach for exact-width by default; these are for the uncommon case where you are explicitly trading storage against speed.

## Print and scan them portably

`%d` / `%u` do not match fixed-width types on every platform. Use the `<inttypes.h>` macros:

```c
#include <inttypes.h>
printf("id=%" PRIu32 " off=%" PRId16 "\n", r.id, r.offset);
```

## Why it matters

Determinism and portability. A struct of exact-width fields has the same size and layout on every target, so it serializes, hashes, and compares bit-identically. A struct of `int`/`long` does not: `long` is 32-bit on Windows and 64-bit on most 64-bit Unix targets, and that single difference silently breaks save files, network packets, and cross-compiler reproducibility.

**Related:** [references/designated-initializers.md](./designated-initializers.md) (ZII fills the rest), [references/build-and-warnings.md](./build-and-warnings.md) (sign-conversion warnings)
