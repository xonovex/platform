# Sources

The C99 style here (implementation variants, caller-owns-memory, alignment,
handles/indices, string handling, file naming, safety validations, testing patterns) is
repo-original/general C knowledge — those reference files have no single upstream and
are expected to show as "uncovered" in the source audit. This guide is an **overlay on
c99-guide**: the generic C99 idioms it shares — const-correctness, designated
initializers, inline-over-macros, compound literals, and baseline error/memory patterns —
are not duplicated here; c99-guide owns them. The architecture references below are
distilled from the engine blog archive.

## Game-engine development blog (archive)

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

- **Last reviewed:** 2026-06-13
- **Used for:**
  - `SKILL.md` → Architecture (Handles & indices), Safety (Strings), Gotchas
  - The data-oriented decisions that distinguish this style from generic C99
- **Aspects extracted:**
  - "Reference objects via array indexes, not raw pointers" — indices/generational handles into caller-owned arrays survive relocation, serialize position-independently, pack tighter, and stay deterministic; generation counters guard slot reuse → `references/handles-and-indices.md`, sharpened bounded-container/handle-resolution rules in `references/safety-validations.md`
  - "Replace the legacy libc string trap" — `strlen`/`strtok` terminator rescans go O(n²) in a loop (the GTA Online JSON-load case); the owning-vs-non-owning split (length-carrying view for reads, bounded caller-owned builder for writes) → `references/string-handling.md`
  - Address/UndefinedBehavior sanitizers as the runtime net for hand-carved arena/caller-owned memory → `references/build-warnings-policy.md`
  - Memory arenas / aggregate "free the whole lifetime at once" allocation reinforce the existing caller-owns-memory direction; the general allocator theory stays in **memory-management-guide**

The C11 `_Generic` overloading and macro-heavy metaprogramming (defer macros, stb_ds-style
meta-header dynamic arrays) from the same talks are intentionally **excluded** from this guide.

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
