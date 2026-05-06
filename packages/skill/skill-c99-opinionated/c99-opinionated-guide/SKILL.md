---
name: c99-opinionated-guide
description: "Use when editing systems or embedded C99 code in projects that follow the opinionated caller-owns-memory, data-oriented style. Triggers on `.c`/`.h` files in systems/embedded/DOD projects and on prompts about SoA layouts, SIMD variants, alignment, strict file naming, even when the user doesn't say 'opinionated'. Skip game-engine code (use c99-game-opinionated-guide), generic C99 without these conventions (use c99-guide), and C++ work."
---

# C99 Opinionated Guidelines (Systems/Game/Embedded)

## Essentials

- **Memory management** - Prefer stack, free all heap allocations, see [reference/memory-management.md](reference/memory-management.md)
- **Modern C99** - Use `const`, designated initializers, small functions, see [reference/designated-initializers.md](reference/designated-initializers.md), [reference/const-correctness.md](reference/const-correctness.md)
- **Type safety** - Favor `inline` functions over macros, see [reference/inline-functions.md](reference/inline-functions.md)
- **Data-oriented design** - Structure for cache locality, see [reference/data-oriented-design.md](reference/data-oriented-design.md)
- **Quality** - Treat warnings as errors, run static analysis

## Architecture

- **Implementation variants** - Scalar → AoS → SoA → SIMD, see [reference/implementation-variants.md](reference/implementation-variants.md)
- **Caller-owns-memory** - Libraries never allocate, caller provides arrays, see [reference/caller-owns-memory.md](reference/caller-owns-memory.md)
- **Alignment** - 16-byte for SIMD, cache line for hot data, see [reference/alignment.md](reference/alignment.md)
- **File naming** - `*_type.h`, `*_impl.h`, `*_aos.h`, `*_soa.h`, `*_simde.h`, see [reference/file-naming.md](reference/file-naming.md)

## Safety

- **Input validation** - Check capacity, bounds, NULL, division, overflow, see [reference/safety-validations.md](reference/safety-validations.md)
- **Work buffers** - Complex functions use caller-provided buffers, see [reference/caller-owns-memory.md](reference/caller-owns-memory.md)
- **SIMD parity** - Test variants against scalar reference, see [reference/testing-patterns.md](reference/testing-patterns.md)

## Progressive disclosure

- Read [reference/memory-management.md](reference/memory-management.md) - When allocating memory or managing resource lifetimes
- Read [reference/designated-initializers.md](reference/designated-initializers.md) - When initializing structs or arrays with specific values
- Read [reference/inline-functions.md](reference/inline-functions.md) - When replacing macros or writing small utility functions
- Read [reference/compound-literals.md](reference/compound-literals.md) - When creating temporary values without named variables
- Read [reference/const-correctness.md](reference/const-correctness.md) - When marking immutable data or understanding pointer const
- Read [reference/data-oriented-design.md](reference/data-oriented-design.md) - When optimizing cache performance or data layouts
- Read [reference/error-handling.md](reference/error-handling.md) - When implementing error codes or handling failures
- Read [reference/implementation-variants.md](reference/implementation-variants.md) - When choosing between scalar, AoS, SoA, or SIMD implementations
- Read [reference/caller-owns-memory.md](reference/caller-owns-memory.md) - When designing APIs where caller provides memory
- Read [reference/alignment.md](reference/alignment.md) - When aligning data for SIMD or cache performance
- Read [reference/file-naming.md](reference/file-naming.md) - When organizing headers by type, implementation, and variant
- Read [reference/testing-patterns.md](reference/testing-patterns.md) - When writing tests with assertions, epsilon comparisons, or parity checks
- Read [reference/safety-validations.md](reference/safety-validations.md) - When validating inputs for capacity, bounds, NULL, or overflow
