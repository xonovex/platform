---
name: lua-opinionated-guidelines
description: Trigger on `.lua` files for LuaJIT performance-critical code. Opinionated patterns: JIT-friendly tables, table pre-allocation, cache lookups. Keywords: LuaJIT, JIT optimization, table pre-allocation, cache lookups, stable table shapes.
---

# Lua Opinionated Guidelines (LuaJIT Performance)

## Requirements

- Lua ≥ 5.4 or LuaJIT 2.1.

## Essentials

- **Module pattern** - Always `local`, one module per file returning table, see [reference/module-pattern.md](reference/module-pattern.md), [reference/local-variables.md](reference/local-variables.md)
- **Code organization** - Prefer table-based modules and simple functions over deep OO, see [reference/module-pattern.md](reference/module-pattern.md), [reference/metatables.md](reference/metatables.md)
- **Performance** - Keep tables stable (JIT-friendly), pre-alloc when size known, see [reference/jit-friendly-tables.md](reference/jit-friendly-tables.md), [reference/cache-lookups.md](reference/cache-lookups.md)
- **Cooperative tasks** - Use coroutines for async patterns, see [reference/coroutines.md](reference/coroutines.md)
- **Validation** - Validate inputs and handle errors, see [reference/input-validation.md](reference/input-validation.md), [reference/error-handling.md](reference/error-handling.md)

## Progressive disclosure

- Read [reference/module-pattern.md](reference/module-pattern.md) - When creating reusable modules or organizing code structure
- Read [reference/local-variables.md](reference/local-variables.md) - When encountering global variable issues or scoping problems
- Read [reference/metatables.md](reference/metatables.md) - When implementing object-oriented patterns or operator overloading
- Read [reference/jit-friendly-tables.md](reference/jit-friendly-tables.md) - When optimizing hot paths or improving JIT performance
- Read [reference/cache-lookups.md](reference/cache-lookups.md) - When reducing table access overhead in performance-critical code
- Read [reference/coroutines.md](reference/coroutines.md) - When implementing cooperative multitasking or async patterns
- Read [reference/input-validation.md](reference/input-validation.md) - When adding type checks or parameter validation
- Read [reference/error-handling.md](reference/error-handling.md) - When handling errors or implementing fallback logic
- Read [reference/string-concatenation.md](reference/string-concatenation.md) - When building strings in loops or formatting output
- Read [reference/idiomatic-patterns.md](reference/idiomatic-patterns.md) - When learning common Lua idioms or patterns
