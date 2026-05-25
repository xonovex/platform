---
name: lua-opinionated-guide
description: "Use when tuning performance-critical Lua hot paths — the tunings especially benefit LuaJIT, and the principles apply to vanilla Lua 5.4 too. An overlay on lua-guide: covers only hot-path performance, not Lua fundamentals. Triggers on `.lua` files in performance-sensitive or LuaJIT projects and on prompts about JIT-friendly tables, table pre-allocation, cache lookups, stable table shapes, or hot-path tuning, even when the user doesn't say 'LuaJIT'. Skip generic Lua scripting — modules, scoping, metatables, coroutines, validation, errors, string building, idioms (use lua-guide) — and TSTL output (use typescript-to-lua-guide)."
---

# Lua Opinionated Guidelines (Performance Tuning)

A performance overlay on **lua-guide**. Apply **lua-guide** for all Lua fundamentals — module pattern, local variables, metatables, coroutines, input validation, error handling, string building, idioms. This skill adds only hot-path tuning: the tunings especially benefit LuaJIT, and the same principles still help vanilla Lua 5.4.

## Requirements

- LuaJIT 2.1 (these tunings especially target JIT compilation) or Lua ≥ 5.4 (principles still apply).

## Essentials

- **Foundation** - All Lua fundamentals live in **lua-guide**; this skill adds hot-path performance tuning on top
- **Stable table shapes** - Pre-allocate all fields, never add fields after creation, so the JIT can specialize, see [references/jit-friendly-tables.md](references/jit-friendly-tables.md)
- **Cache lookups** - Hoist repeated table/global/stdlib lookups into locals on hot paths, see [references/cache-lookups.md](references/cache-lookups.md)

## Gotchas

- Adding a field after table creation changes the table's shape — it deoptimizes the JIT trace even though the code is correct
- `pairs()` in a hot loop can't be JIT-compiled as tightly as a numeric `for i = 1, #t` loop over a dense array
- Sparse arrays (`nil` holes) break both `#` and fast array traces — keep arrays dense
- Caching `math.sin`/`math.cos` into locals matters in hot loops but is noise elsewhere — apply tuning where profiling shows it, not everywhere

## Progressive disclosure

- Read [references/jit-friendly-tables.md](references/jit-friendly-tables.md) - Load when optimizing hot paths or improving JIT performance
- Read [references/cache-lookups.md](references/cache-lookups.md) - Load when reducing table access overhead in performance-critical code
