---
name: c99-opinionated-guide
description: "Use when editing systems or embedded C99 code in projects that follow the opinionated caller-owns-memory, data-oriented style. Triggers on `.c`/`.h` files in systems/embedded/DOD projects and on prompts about SoA layouts, SIMD variants, alignment, strict file naming, even when the user doesn't say 'opinionated'. Skip game-engine code (use c99-game-opinionated-guide), generic C99 without these conventions (use c99-guide), and C++ work."
---

# C99 Opinionated Guidelines (Systems/Game/Embedded)

## Essentials

- **Memory management** - Prefer stack, free all heap allocations, see [references/memory-management.md](references/memory-management.md)
- **Modern C99** - Use `const`, designated initializers, small functions, see [references/designated-initializers.md](references/designated-initializers.md), [references/const-correctness.md](references/const-correctness.md)
- **Type safety** - Favor `inline` functions over macros, see [references/inline-functions.md](references/inline-functions.md)
- **Data-oriented design** - Layout/cache is a core choice of this style — apply **data-oriented-design-guide**
- **Quality** - Treat warnings as errors, run static analysis

## Architecture

- **Implementation variants** - Ship scalar → AoS → SoA → SIMD with `_*` suffixes + parity tests; layout rationale in **data-oriented-design-guide**, see [references/implementation-variants.md](references/implementation-variants.md)
- **Caller-owns-memory** - This style's default; the general principle (arenas/ownership) lives in **memory-management-guide**, see [references/caller-owns-memory.md](references/caller-owns-memory.md)
- **Alignment** - C `_Alignas`/aligned allocation; the _why_ is in **data-oriented-design-guide** (SIMD) and **lock-free-guide** (false sharing), see [references/alignment.md](references/alignment.md)
- **Composability** - Composable stages/primitives over a uniform currency, explicit caller-wired composition, see [references/composability.md](references/composability.md)
- **Hot reload** - Reloadable native modules via API/function-pointer tables + host-owned state, see [references/hot-reload.md](references/hot-reload.md)
- **File naming** - `*_type.h`, `*_impl.h`, `*_aos.h`, `*_soa.h`, `*_simde.h`, see [references/file-naming.md](references/file-naming.md)

## Safety

- **Input validation** - Check capacity, bounds, NULL, division, overflow, see [references/safety-validations.md](references/safety-validations.md)
- **Work buffers** - Complex functions use caller-provided buffers, see [references/caller-owns-memory.md](references/caller-owns-memory.md)
- **SIMD parity** - Test variants against scalar reference, see [references/testing-patterns.md](references/testing-patterns.md)

## Gotchas

- `static inline` in a header has different linkage than `inline` alone — exactly one TU must emit the definition; mixing causes silent code bloat
- Designated initializers leave unmentioned fields zero-initialized — relying on that for safety means a missing field is silent
- `unsigned` overflow is defined; signed overflow is undefined behavior — never rely on signed wrap
- `alignof`/`alignas` interact subtly with `malloc` (always returns max-align) vs custom allocators

## Progressive disclosure

- Read [references/memory-management.md](references/memory-management.md) - Load when allocating memory or managing resource lifetimes
- Read [references/designated-initializers.md](references/designated-initializers.md) - Load when initializing structs or arrays with specific values
- Read [references/inline-functions.md](references/inline-functions.md) - Load when replacing macros or writing small utility functions
- Read [references/compound-literals.md](references/compound-literals.md) - Load when creating temporary values without named variables
- Read [references/const-correctness.md](references/const-correctness.md) - Load when marking immutable data or understanding pointer const
- Read [references/error-handling.md](references/error-handling.md) - Load when implementing error codes or handling failures
- Read [references/implementation-variants.md](references/implementation-variants.md) - Load when choosing between scalar, AoS, SoA, or SIMD implementations
- Read [references/caller-owns-memory.md](references/caller-owns-memory.md) - Load when designing APIs where caller provides memory
- Read [references/alignment.md](references/alignment.md) - Load when aligning data for SIMD or cache performance
- Read [references/composability.md](references/composability.md) - Load when designing pipelines, multi-stage transforms, or reusable primitive APIs
- Read [references/hot-reload.md](references/hot-reload.md) - Load when making native code reloadable at runtime or designing a plugin/module boundary
- Read [references/file-naming.md](references/file-naming.md) - Load when organizing headers by type, implementation, and variant
- Read [references/testing-patterns.md](references/testing-patterns.md) - Load when writing tests with assertions, epsilon comparisons, or parity checks
- Read [references/safety-validations.md](references/safety-validations.md) - Load when validating inputs for capacity, bounds, NULL, or overflow
