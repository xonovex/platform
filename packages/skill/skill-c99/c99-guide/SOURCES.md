# Sources

The core C99 style here (memory management, const-correctness, inline-over-macros,
compound literals, error handling, string views/builders, build dialect & warnings) is
repo-original / general C knowledge with no single upstream. The Modern-C references below
were distilled from conference-talk material on modern, data-oriented C.

## Modern C / data-oriented C conference talks

- **Last reviewed:** 2026-06-13
- **Used for:**
  - `SKILL.md` → Fixed-width types, Value-oriented APIs, Quality (sanitizers), Gotchas
  - The "Modern C over C89" framing: fixed-width types, ZII, value-oriented APIs, runtime sanitizers
- **Aspects extracted:**
  - Fixed-width type safety via `<stdint.h>` (exact vs least/fast widths, `size_t`/`ptrdiff_t`, `<inttypes.h>` format macros, the portability/determinism rationale) → `references/fixed-width-types.md`
  - Designated initializers as a "killer feature", Zero-Is-Initialization (ZII) as the C analogue of RAII, and declarative config-struct APIs (`_DEFAULT` const + pass-by-value) → `references/designated-initializers.md`
  - Value-oriented APIs: passing/returning small plain-old-data by value to kill pointer aliasing, and `optional`/result structs that bundle a value with its validity (allocation-free, monadic-style) instead of out-parameters → `references/value-types.md`, `references/error-handling.md`
  - Address/UndefinedBehavior sanitizers and running crashes under a debugger rather than `printf`; single- vs multiple-translation-unit (unity build) trade-off → `references/build-and-warnings.md`
  - The legacy-libc trap (`strlen`/`strtok` terminator rescans, the GTA Online JSON-load O(n²) case) and the owning-vs-non-owning string split → already covered in `references/string-views.md`

The C11 `_Generic` overloading and macro-heavy metaprogramming (defer macros, stb_ds-style
meta-header dynamic arrays) from the same talks are intentionally **excluded** — out of scope
for this C99 guide.

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
