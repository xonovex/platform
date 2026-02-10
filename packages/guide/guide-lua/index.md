---
name: lua-guidelines
description: Trigger on `.lua` files. Use when writing Lua 5.4+ code. Apply for modules, scripting, table patterns. Keywords: Lua, module pattern, local variables, metatables, coroutines, error handling.
---

# Lua Coding Guidelines

## Requirements

- Lua ≥ 5.4.

## Essentials

- **Module pattern** - Always `local`, one module per file returning table, see [reference/module-pattern.md](reference/module-pattern.md), [reference/local-variables.md](reference/local-variables.md)
- **Code organization** - Prefer table-based modules and simple functions, see [reference/module-pattern.md](reference/module-pattern.md), [reference/metatables.md](reference/metatables.md)
- **Cooperative tasks** - Use coroutines for async patterns, see [reference/coroutines.md](reference/coroutines.md)
- **Validation** - Validate inputs and handle errors, see [reference/input-validation.md](reference/input-validation.md), [reference/error-handling.md](reference/error-handling.md)

## Progressive disclosure

- Read [reference/module-pattern.md](reference/module-pattern.md) - When creating reusable modules or organizing code structure
- Read [reference/local-variables.md](reference/local-variables.md) - When encountering global variable issues or scoping problems
- Read [reference/metatables.md](reference/metatables.md) - When implementing object-oriented patterns or operator overloading
- Read [reference/coroutines.md](reference/coroutines.md) - When implementing cooperative multitasking or async patterns
- Read [reference/input-validation.md](reference/input-validation.md) - When adding type checks or parameter validation
- Read [reference/error-handling.md](reference/error-handling.md) - When handling errors or implementing fallback logic
- Read [reference/string-concatenation.md](reference/string-concatenation.md) - When building strings in loops or formatting output
- Read [reference/idiomatic-patterns.md](reference/idiomatic-patterns.md) - When learning common Lua idioms or patterns
