---
name: c99-opinionated-guide
description: "Use when editing systems or embedded C99 code in projects that follow the opinionated caller-owns-memory, data-oriented style. An overlay on c99-guide: covers only the opinionated decisions, not the generic C99 idioms it shares with c99-guide. Triggers on `.c`/`.h` files in systems/embedded/DOD projects and on prompts about caller-owns-memory, SoA/SIMD variants, alignment, index/handle references, caller-owned string views/builders, physical design, plugin architecture, strict file naming, even when the user doesn't say 'opinionated'. Skip generic C99 idioms — const-correctness, designated initializers, inline-vs-macro, compound literals, fixed-width types, value-oriented APIs, baseline error/memory patterns (use c99-guide) — game-engine code (use c99-game-opinionated-guide), and C++ work."
---

# C99 Opinionated Guidelines (Systems/Game/Embedded)

## Essentials

- **Overlay on c99-guide** - This guide carries only the opinionated decisions. For the shared C99 idioms — `const`-correctness, designated initializers (ZII), `inline`-over-macros, compound literals, fixed-width types, value-oriented APIs, and baseline error/return patterns — follow **c99-guide**
- **Memory management** - This style's defining choice: the library doesn't allocate, the caller owns storage — see [references/caller-owns-memory.md](references/caller-owns-memory.md); general allocator theory (arenas/pools/lifetimes) in **memory-management-guide**
- **Data-oriented design** - Layout/cache is a core choice of this style — apply **data-oriented-design-guide**
- **Quality** - Pin strict C99, `-Werror` for correctness but relax the unused-symbol family for libraries, keep an ASan/UBSan debug build, run static analysis, see [references/build-warnings-policy.md](references/build-warnings-policy.md)

## Architecture

- **Implementation variants** - Ship scalar → AoS → SoA → SIMD with `_*` suffixes + parity tests; layout rationale in **data-oriented-design-guide**, see [references/implementation-variants.md](references/implementation-variants.md)
- **Caller-owns-memory** - This style's default; the general principle (arenas/ownership) lives in **memory-management-guide**, see [references/caller-owns-memory.md](references/caller-owns-memory.md)
- **Handles & indices** - Store cross-object references as a stable index or generational handle into a caller-owned array, never a raw pointer — relocatable, serializable, deterministic, see [references/handles-and-indices.md](references/handles-and-indices.md)
- **Alignment** - C `_Alignas`/aligned allocation; the _why_ is in **data-oriented-design-guide** (SIMD) and **lock-free-guide** (false sharing), see [references/alignment.md](references/alignment.md)
- **Composability** - Composable stages/primitives over a uniform currency, explicit caller-wired composition, see [references/composability.md](references/composability.md)
- **Hot reload** - Reloadable native modules via API/function-pointer tables + host-owned state, see [references/hot-reload.md](references/hot-reload.md)
- **Physical design** - Headers don't include headers; one header = one system's interface; opaque handles + forward declarations; acyclic deps and fast incremental builds, see [references/physical-design.md](references/physical-design.md)
- **Plugin architecture** - Small plugins talking through a string-keyed registry of plain-C function-pointer interfaces; runtime discovery, lean core, see [references/plugin-architecture.md](references/plugin-architecture.md)
- **Cross-language APIs** - Plain-C portable-subset surface, flat data over pointers, call-scoped pointers, spec-generated bindings, see [references/cross-language-api.md](references/cross-language-api.md)
- **File naming** - `*_type.h`, `*_impl.h`, `*_aos.h`, `*_soa.h`, `*_simde.h`, see [references/file-naming.md](references/file-naming.md)

## Safety

- **Input validation** - Check capacity, bounds, NULL, division, overflow, see [references/safety-validations.md](references/safety-validations.md)
- **Work buffers** - Complex functions use caller-provided buffers, see [references/caller-owns-memory.md](references/caller-owns-memory.md)
- **Strings** - Length-carrying views for reads, bounded caller-owned builders for writes; no `strlen`/`strtok` terminator rescans, see [references/string-handling.md](references/string-handling.md)
- **SIMD parity** - Test variants against scalar reference, see [references/testing-patterns.md](references/testing-patterns.md)

## Gotchas

- For a library, `-Werror` on the unused-symbol family is wrong: header reflection/mapping tables and interface-mandated parameters are surface a TU may not reference — relax `-Wno-unused-{parameter,variable,but-set-variable,function}`, keep `-Wunused-value` and correctness warnings
- `-std=c99` usually keeps the GNU dialect — set `C_EXTENSIONS OFF`, then `_XOPEN_SOURCE=700` or POSIX calls become implicit-declaration errors
- `static inline` in a header gives each TU its own internal-linkage copy (safe, but bloats if the compiler never inlines it); a plain `inline` definition has external linkage and needs exactly one TU to emit the external definition — don't mix the two storage classes for the same function
- Designated initializers leave unmentioned fields zero-initialized — relying on that for safety means a missing field is silent
- `unsigned` overflow is defined; signed overflow is undefined behavior — never rely on signed wrap
- `alignof`/`alignas` interact subtly with `malloc` (always returns max-align) vs custom allocators
- One header including another silently reintroduces the include cascade and recompilation storms — keep the no-header-includes rule machine-checked in CI
- A cached plugin-interface or cross-module function pointer dangles after reload/unload — re-fetch from the registry, never stash it across that boundary
- A C API that lets a caller keep a borrowed pointer past the call is a lifetime contract a GC language can't honor; default to call-scoped pointers and document the rare exceptions
- A raw pointer stored into a caller-owned array dangles the moment that array is re-bound, compacted, or grown — store an index/handle and resolve it per use; pointer values also vary run to run (ASLR/allocator), breaking determinism and serialization
- A recycled slot makes a bare index alias a different object — add a generation counter to the handle so a stale reference fails its check instead of reading the wrong data
- `strlen`/`strcmp`/`strtok` rescan to the terminator on every call, so a loop over them is silently O(n²) (the GTA Online JSON-load case) — carry length in a view and never rescan

## Progressive disclosure

- Read [references/build-warnings-policy.md](references/build-warnings-policy.md) - Load when configuring the C standard, feature-test macros, the library-vs-app warning policy, or sanitizers
- Read [references/implementation-variants.md](references/implementation-variants.md) - Load when choosing between scalar, AoS, SoA, or SIMD implementations
- Read [references/caller-owns-memory.md](references/caller-owns-memory.md) - Load when designing APIs where caller provides memory
- Read [references/handles-and-indices.md](references/handles-and-indices.md) - Load when storing cross-object references, designing pools, or making structures serializable/relocatable
- Read [references/string-handling.md](references/string-handling.md) - Load when handling strings without terminator rescans or hidden allocations
- Read [references/alignment.md](references/alignment.md) - Load when aligning data for SIMD or cache performance
- Read [references/composability.md](references/composability.md) - Load when designing pipelines, multi-stage transforms, or reusable primitive APIs
- Read [references/hot-reload.md](references/hot-reload.md) - Load when making native code reloadable at runtime or designing a plugin/module boundary
- Read [references/physical-design.md](references/physical-design.md) - Load when organizing headers/translation units, cutting build times, or breaking include/dependency cycles
- Read [references/plugin-architecture.md](references/plugin-architecture.md) - Load when designing a plugin system, an interface/API registry, or runtime discovery between decoupled components
- Read [references/cross-language-api.md](references/cross-language-api.md) - Load when designing a C API to be bound from other languages (Lua/C#/Rust/Python) or generating bindings
- Read [references/file-naming.md](references/file-naming.md) - Load when organizing headers by type, implementation, and variant
- Read [references/testing-patterns.md](references/testing-patterns.md) - Load when writing tests with assertions, epsilon comparisons, or parity checks
- Read [references/safety-validations.md](references/safety-validations.md) - Load when validating inputs for capacity, bounds, NULL, or overflow
