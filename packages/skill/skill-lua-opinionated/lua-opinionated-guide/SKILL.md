---
name: lua-opinionated-guide
description: "Use when tuning LuaJIT-targeted Lua for performance-critical paths. Triggers on `.lua` files in LuaJIT projects and on prompts about JIT-friendly tables, table pre-allocation, cache lookups, stable table shapes, or hot-path tuning, even when the user doesn't say 'LuaJIT'. Skip generic Lua scripting (use lua-guide) and TSTL output (use typescript-to-lua-guide)."
---

# Lua Opinionated Guidelines (LuaJIT Performance)

## Requirements

- Lua ≥ 5.4 or LuaJIT 2.1.

## Essentials

- **Module pattern** - Always `local`, one module per file returning table, see [references/module-pattern.md](references/module-pattern.md), [references/local-variables.md](references/local-variables.md)
- **Code organization** - Prefer table-based modules and simple functions over deep OO, see [references/module-pattern.md](references/module-pattern.md), [references/metatables.md](references/metatables.md)
- **Performance** - Keep tables stable (JIT-friendly), pre-alloc when size known, see [references/jit-friendly-tables.md](references/jit-friendly-tables.md), [references/cache-lookups.md](references/cache-lookups.md)
- **Cooperative tasks** - Use coroutines for async patterns, see [references/coroutines.md](references/coroutines.md)
- **Validation** - Validate inputs and handle errors, see [references/input-validation.md](references/input-validation.md), [references/error-handling.md](references/error-handling.md)

## Gotchas

- Module pattern (`local M = {}; return M`) is idiomatic — exposing globals breaks isolation and shadows in unexpected places
- `pcall` is the only way to catch errors — wrapping every entry point makes errors compose; ad-hoc `pcall` leaks failures
- Metatables for OOP work but inheritance is by-hand — prefer composition unless you actually need polymorphic dispatch
- String concatenation with `..` allocates each time — building a string in a loop should use `table.concat` instead

## Progressive disclosure

- Read [references/module-pattern.md](references/module-pattern.md) - When creating reusable modules or organizing code structure
- Read [references/local-variables.md](references/local-variables.md) - When encountering global variable issues or scoping problems
- Read [references/metatables.md](references/metatables.md) - When implementing object-oriented patterns or operator overloading
- Read [references/jit-friendly-tables.md](references/jit-friendly-tables.md) - When optimizing hot paths or improving JIT performance
- Read [references/cache-lookups.md](references/cache-lookups.md) - When reducing table access overhead in performance-critical code
- Read [references/coroutines.md](references/coroutines.md) - When implementing cooperative multitasking or async patterns
- Read [references/input-validation.md](references/input-validation.md) - When adding type checks or parameter validation
- Read [references/error-handling.md](references/error-handling.md) - When handling errors or implementing fallback logic
- Read [references/string-concatenation.md](references/string-concatenation.md) - When building strings in loops or formatting output
- Read [references/idiomatic-patterns.md](references/idiomatic-patterns.md) - When learning common Lua idioms or patterns
