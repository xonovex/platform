---
name: c99-guidelines
description: >-
  Trigger on `.c/.h` files. Use when writing C99 code. Apply for memory management, data structures, low-level operations. Keywords: C99, malloc/free, const-correctness, designated initializers, inline functions, error handling.
---

# C99 Coding Guidelines

## Essentials

- **Memory management** - Free all heap allocations, avoid leaks, see [reference/memory-management.md](reference/memory-management.md)
- **Modern C99** - Use `const`, designated initializers, small functions, see [reference/designated-initializers.md](reference/designated-initializers.md), [reference/const-correctness.md](reference/const-correctness.md)
- **Type safety** - Favor `inline` functions over macros where practical, see [reference/inline-functions.md](reference/inline-functions.md)
- **Quality** - Enable warnings, run static analysis

## Best practices

- **Error handling** - Use return codes, check all fallible operations, see [reference/error-handling.md](reference/error-handling.md)
- **Input validation** - Check bounds, NULL pointers, division by zero
- **Readability** - Small functions, clear naming, comments for non-obvious logic

## Progressive disclosure

- Read [reference/memory-management.md](reference/memory-management.md) - When allocating memory or managing resource lifetimes
- Read [reference/designated-initializers.md](reference/designated-initializers.md) - When initializing structs or arrays with specific values
- Read [reference/inline-functions.md](reference/inline-functions.md) - When replacing macros or writing small utility functions
- Read [reference/compound-literals.md](reference/compound-literals.md) - When creating temporary values without named variables
- Read [reference/const-correctness.md](reference/const-correctness.md) - When marking immutable data or understanding pointer const
- Read [reference/error-handling.md](reference/error-handling.md) - When implementing error codes or handling failures
