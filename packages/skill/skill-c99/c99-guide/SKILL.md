---
name: c99-guide
description: "Use when editing or reviewing general-purpose C99 — libraries, CLI tools, system code without strong opinions on memory ownership or layout. Triggers on `.c`/`.h` files and prompts about C99 idioms, designated initializers, const-correctness, malloc/free patterns, inline functions, or error returns, even when the user doesn't say 'C'. Skip projects that follow the caller-owns-memory style (use c99-opinionated-guide), game-engine code (use c99-game-opinionated-guide), and C++ work."
---

# C99 Coding Guidelines

## Essentials

- **Memory management** - Free all heap allocations, avoid leaks, see [references/memory-management.md](references/memory-management.md)
- **Modern C99** - Use `const`, designated initializers, small functions, see [references/designated-initializers.md](references/designated-initializers.md), [references/const-correctness.md](references/const-correctness.md)
- **Type safety** - Favor `inline` functions over macros where practical, see [references/inline-functions.md](references/inline-functions.md)
- **Quality** - Enable warnings, run static analysis

## Best practices

- **Error handling** - Use return codes, check all fallible operations, see [references/error-handling.md](references/error-handling.md)
- **Input validation** - Check bounds, NULL pointers, division by zero
- **Readability** - Small functions, clear naming, comments for non-obvious logic

## Gotchas

- Signed integer overflow is undefined behavior — even `INT_MAX + 1` lets the compiler eliminate "impossible" code paths
- Strict aliasing means casting `int*` to `float*` is UB unless via `union` or `memcpy` — silently miscompiled at higher optimization levels
- `malloc` returns memory that's max-aligned; custom arenas must preserve alignment for `_Bool`/`double`/SIMD types
- Designated initializers (`.field = x`) zero-fill unmentioned members — a missing field becomes silent zero, not a compile error

## Progressive disclosure

- Read [references/memory-management.md](references/memory-management.md) - When allocating memory or managing resource lifetimes
- Read [references/designated-initializers.md](references/designated-initializers.md) - When initializing structs or arrays with specific values
- Read [references/inline-functions.md](references/inline-functions.md) - When replacing macros or writing small utility functions
- Read [references/compound-literals.md](references/compound-literals.md) - When creating temporary values without named variables
- Read [references/const-correctness.md](references/const-correctness.md) - When marking immutable data or understanding pointer const
- Read [references/error-handling.md](references/error-handling.md) - When implementing error codes or handling failures
