# Sources

This guide is an **overlay on c99-guide**: generic C99 idioms remain there, while
**memory-management-guide** owns allocation, ownership, and lifetime theory and
**data-oriented-design-guide** owns stable-handle layout and iteration tradeoffs. This
guide retains only the C99-specific boundary validations and style decisions. The
architecture references below are distilled from the engine blog archive.

## Game-engine development blog (archive)

- **Provenance:** Locally archived game-engine development articles
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Architecture, Gotchas
  - The plain-C plugin/interface model, header-include discipline, hot-reload contract, modular composition, and cross-language binding surface
- **Aspects extracted:**
  - "Physical Design of the engine" — the machine-checkable rule that headers may not include headers (bar a tiny allowlist), one header per system, folders as system boundaries, opaque handle types (`{uint64_t opaque;}`), forward declarations + pointer interfaces, isolating templates/generics, and the resulting acyclic graph + fast incremental builds → `references/physical-design.md`
  - "Little Machines Working Together (Parts 1–2)" — the string-keyed interface/API registry (`add`/`remove`/`first`/`next`), capabilities as plain-C structs of function pointers under a unique id, runtime discovery with zero compile-time coupling, multiple implementations of one interface as the extension-point mechanism, plain C for ABI stability, lean core → `references/plugin-architecture.md`
  - "The Anti-Feature Dream" — decomposing tools into small reusable building blocks assembled by the user instead of monolithic features (the modular/composable design bias) → `references/plugin-architecture.md`, `references/composability.md`
  - "DLL Hot Reloading in Theory and Practice" — reloadable native modules behind a function-pointer API table with all persistent state host-owned, re-fetch the table after each reload, reload at a safe point → `references/hot-reload.md`
  - "Creating Cross-Language APIs" — the C ABI as the universal FFI surface, restricting the API to a portable subset (no untagged unions/variadics/globals), flat data over pointer graphs, call-scoped pointer lifetimes, and generating idiomatic per-language bindings from a machine-readable API spec → `references/cross-language-api.md`

## C ABI / FFI and large-API binding prior art

- **URLs:**
  - System V AMD64 ABI — https://gitlab.com/x86-psABIs/x86-64-ABI
  - Vulkan XML registry (spec-driven header + binding generation) — https://github.com/KhronosGroup/Vulkan-Docs
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Architecture
  - Corroborating the C-ABI-as-lingua-franca and spec-generated-bindings approach
- **Aspects extracted:**
  - Stable C calling convention vs unstable C++ ABI; spec → C header + per-language bindings → `references/cross-language-api.md`

## Modern C / data-oriented C conference talks

- **Provenance:** Locally reviewed recordings and transcripts about modern C and data-oriented implementation
- **Last reviewed:** 2026-06-13
- **Used for:**
  - `SKILL.md` → Architecture (Handles & indices), Safety (Strings), Gotchas
  - The data-oriented decisions that distinguish this style from generic C99
- **Aspects extracted:**
  - "Reference objects via array indexes, not raw pointers" — the general handle model is owned by **data-oriented-design-guide**; the C99-specific bounds-before-generation check and scoped-pointer rule feed `references/safety-validations.md`
  - "Replace the legacy libc string trap" — `strlen`/`strtok` terminator rescans go O(n²) in a loop (the GTA Online JSON-load case); the owning-vs-non-owning split (length-carrying view for reads, bounded caller-owned builder for writes) → `references/string-handling.md`
  - Address/UndefinedBehavior sanitizers as the runtime net for hand-carved arena/caller-owned memory → `references/build-warnings-policy.md`
  - Memory arenas / aggregate "free the whole lifetime at once" allocation reinforce the existing caller-owns-memory direction; the general allocator theory stays in **memory-management-guide**

The C11 `_Generic` overloading and macro-heavy metaprogramming (defer macros, stb_ds-style
meta-header dynamic arrays) from the same talks are intentionally **excluded** from this guide.

## Guide-level synthesis

- **Provenance:** Repository-original integration of the source blocks above; these references combine multiple inputs or maintained conventions rather than one exclusive upstream
- **References:** references/alignment.md, references/file-naming.md, references/implementation-variants.md, references/testing-patterns.md
- **Last reviewed:** 2026-06-13

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
