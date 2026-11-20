---
name: writing-lua
description: "Use when writing or editing Lua 5.4+ or LuaJIT modules, scripts, or configuration, or making a hot path faster. Triggers on `.lua` files, `require` modules, local scoping, metatables, coroutines, `pcall` error handling, or LuaJIT loop tuning, even when the user doesn't say 'Lua'."
---

# Lua Coding Guidelines

Write Lua modules with local state, explicit error handling, and clear interfaces.

## Requirements

- Lua ≥ 5.4, or LuaJIT 2.1 for the hot-path tunings.

## Essentials

- **Module pattern** - Always `local`, one module per file returning table, see [references/module-pattern.md](references/module-pattern.md), [references/local-variables.md](references/local-variables.md)
- **Code organization** - Prefer table-based modules and simple functions, see [references/module-pattern.md](references/module-pattern.md), [references/metatables.md](references/metatables.md)
- **Cooperative tasks** - Use coroutines for async patterns, see [references/coroutines.md](references/coroutines.md)
- **Validation** - Validate inputs, see [references/input-validation.md](references/input-validation.md)
- **Hot paths** - Plain numeric loops and lookups cached in locals, only where profiling shows the need, see [references/hot-paths.md](references/hot-paths.md)
- **Paradigm** - Functional style → **applying-fp**; class/OO design (metatable-based) → **applying-oop**

## Gotchas

- Arrays are 1-indexed; a `nil` hole breaks the `#` length operator (it stops at the first nil)
- Tables are both array and hash; mixing them is fine but iteration order isn't guaranteed for the hash part
- `/` always returns a float; use `//` for integer floor division: `total // per_page`, not `math.floor(total / per_page)`
- Variables are global by default unless declared `local`: forgetting `local` in a loop counter leaks into the surrounding scope
- Hot-path tuning beyond what profiling shows (manual unrolling, FFI arrays, SIMD-style tricks) adds code and no speed

## Progressive disclosure

- Read [references/module-pattern.md](references/module-pattern.md) - Load when creating reusable modules or organizing code structure
- Read [references/local-variables.md](references/local-variables.md) - Load when encountering global variable issues or scoping problems
- Read [references/metatables.md](references/metatables.md) - Load when implementing object-oriented patterns or operator overloading
- Read [references/coroutines.md](references/coroutines.md) - Load when implementing cooperative multitasking or async patterns
- Read [references/input-validation.md](references/input-validation.md) - Load when adding type checks or parameter validation
- Read [references/idiomatic-patterns.md](references/idiomatic-patterns.md) - Load when learning common Lua idioms or patterns
- Read [references/hot-paths.md](references/hot-paths.md) - Load when a profile shows a hot loop or a LuaJIT trace to tune
