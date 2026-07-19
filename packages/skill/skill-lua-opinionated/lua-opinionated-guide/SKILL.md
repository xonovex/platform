---
name: lua-opinionated-guide
description: "Use when tuning performance-critical Lua hot paths — the tunings especially benefit LuaJIT, and the principles apply to vanilla Lua 5.4 too. A focused overlay that covers only hot-path performance, not Lua fundamentals. Triggers on `.lua` files in performance-sensitive or LuaJIT projects and on prompts about JIT-friendly tables, table pre-allocation, cache lookups, stable table shapes, or hot-path tuning, even when the user doesn't say 'LuaJIT'."
---

# Lua Opinionated Guidelines (Performance Tuning)

A performance overlay on **lua-guide**. Apply **lua-guide** for all Lua fundamentals — module pattern, local variables, metatables, coroutines, input validation, error handling, string building, idioms. This skill adds only hot-path tuning: the tunings especially benefit LuaJIT, and the same principles still help vanilla Lua 5.4.

## Requirements

- LuaJIT 2.1 (these tunings especially target JIT compilation) or Lua ≥ 5.4 (principles still apply).

## Essentials

- **Foundation** - All Lua fundamentals live in **lua-guide**; this skill adds hot-path performance tuning on top
- **Simple hot loops** - Iterate dense arrays with a plain numeric for-loop; don't over-engineer with unrolling or FFI, see [references/jit-friendly-tables.md](references/jit-friendly-tables.md)
- **Cache lookups** - Hoist repeated global/module-table lookups into locals on hot paths, see [references/cache-lookups.md](references/cache-lookups.md)

## Gotchas

- Iterate dense arrays with the plain numeric `for i = 1, #t do ... end`; the simple idiom is the deliverable — don't manually unroll it (`while i < n - 3 ... i = i + 4`)
- Apply hot-path tuning only where profiling shows a need — don't over-engineer with FFI arrays or SIMD-style tricks beyond what was asked

## Progressive disclosure

- Read [references/jit-friendly-tables.md](references/jit-friendly-tables.md) - Load when optimizing hot paths or improving JIT performance
- Read [references/cache-lookups.md](references/cache-lookups.md) - Load when reducing table access overhead in performance-critical code
