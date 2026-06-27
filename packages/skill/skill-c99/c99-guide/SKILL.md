---
name: c99-guide
description: "Use when editing or reviewing general-purpose C99 — libraries, CLI tools, system code without strong opinions on memory ownership or layout. Triggers on `.c`/`.h` files and prompts about C99 idioms, designated initializers, fixed-width types, const-correctness, malloc/free patterns, inline functions, value-oriented APIs, sanitizers, or error returns, even when the user doesn't say 'C'."
---

# C99 Coding Guidelines

## Essentials

- **Memory management** - Free all heap allocations, avoid leaks, see [references/memory-management.md](references/memory-management.md)
- **Modern C99** - Use `const`, designated initializers (ZII), small functions, see [references/designated-initializers.md](references/designated-initializers.md), [references/const-correctness.md](references/const-correctness.md)
- **Fixed-width types** - Name widths with `<stdint.h>` for sized/serialized data; `size_t` for counts and indices, see [references/fixed-width-types.md](references/fixed-width-types.md)
- **Type safety** - Favor `inline` functions over macros where practical, see [references/inline-functions.md](references/inline-functions.md)
- **Quality** - Pin strict C99, treat warnings as errors, keep an ASan/UBSan debug build, run static analysis, see [references/build-and-warnings.md](references/build-and-warnings.md)

## Best practices

- **Error handling** - Use return codes, check all fallible operations, see [references/error-handling.md](references/error-handling.md)
- **Value-oriented APIs** - Return small results by value or in a `{ok, value}` result struct; reserve out-params for large or multiple results, see [references/value-types.md](references/value-types.md)
- **Strings** - Borrow length-carrying views, write through bounded builders over caller memory — not `strlen`/`strcat`/`strtok` rescans, see [references/string-views.md](references/string-views.md)
- **Input validation** - Check bounds, NULL pointers, division by zero
- **Readability** - Small functions, clear naming, comments for non-obvious logic
- **Paradigm** - Functional style (pure functions, explicit context) → **fp-guide**; object/data modeling → **oop-guide**

## Gotchas

- Signed integer overflow is undefined behavior — even `INT_MAX + 1` lets the compiler eliminate "impossible" code paths
- Strict aliasing means casting `int*` to `float*` is UB unless via `union` or `memcpy` — silently miscompiled at higher optimization levels
- `malloc` returns memory that's max-aligned; custom arenas must preserve alignment for `_Bool`/`double`/SIMD types
- Designated initializers (`.field = x`) zero-fill unmentioned members — a missing field becomes silent zero, not a compile error
- `long` is 32-bit on Windows and 64-bit on most 64-bit Unix targets — a struct of `int`/`long` serializes and hashes differently per platform; use `<stdint.h>` exact-width types for any stored or shared data
- Comparing a signed `int` against an unsigned `size_t` converts the `int` — a negative value becomes a huge `size_t` and the bounds check passes; keep counts/indices in `size_t` end to end
- Return-by-value is for _small_ PODs; returning a large struct just copies it — large or caller-owned results still take a pointer
- `-std=c99` usually still means the GNU dialect — set `C_EXTENSIONS OFF` for true ISO C99, then `#define _XOPEN_SOURCE 700` or POSIX calls (`readlink`, `strnlen`, `ssize_t`) become implicit-declaration errors
- `-Wextra`'s `-Wmissing-field-initializers`/`-Wmissing-braces` fire on intentional zero-init — suppress those two, keep the rest of `-Werror`
- A bare `snprintf` into a fixed buffer trips `-Wformat-truncation`; check its return (`>= size` ⇒ truncated) to handle the case and clear the warning

## Progressive disclosure

- Read [references/build-and-warnings.md](references/build-and-warnings.md) - Load when configuring the C standard, feature-test macros, warning flags, or sanitizers
- Read [references/fixed-width-types.md](references/fixed-width-types.md) - Load when choosing integer types for struct fields, serialized data, counts, or indices
- Read [references/string-views.md](references/string-views.md) - Load when handling strings without repeated terminator scans or hidden allocations
- Read [references/value-types.md](references/value-types.md) - Load when designing function signatures: returning results by value vs. out-parameters
- Read [references/memory-management.md](references/memory-management.md) - Load when allocating memory or managing resource lifetimes
- Read [references/designated-initializers.md](references/designated-initializers.md) - Load when initializing structs or arrays with specific values
- Read [references/inline-functions.md](references/inline-functions.md) - Load when replacing macros or writing small utility functions
- Read [references/compound-literals.md](references/compound-literals.md) - Load when creating temporary values without named variables
- Read [references/const-correctness.md](references/const-correctness.md) - Load when marking immutable data or understanding pointer const
- Read [references/error-handling.md](references/error-handling.md) - Load when implementing error codes or handling failures
