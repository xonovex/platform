---
name: lua-guide
description: "Use when editing general-purpose Lua 5.4+ — modules, scripts, configuration. Triggers on `.lua` files and prompts about module patterns, local scoping, metatables, coroutines, error handling, even when the user doesn't say 'Lua'."
---

# Lua Coding Guidelines

## Requirements

- Lua ≥ 5.4.

## Essentials

- **Module pattern** - Always `local`, one module per file returning table, see [references/module-pattern.md](references/module-pattern.md), [references/local-variables.md](references/local-variables.md)
- **Code organization** - Prefer table-based modules and simple functions, see [references/module-pattern.md](references/module-pattern.md), [references/metatables.md](references/metatables.md)
- **Cooperative tasks** - Use coroutines for async patterns, see [references/coroutines.md](references/coroutines.md)
- **Validation** - Validate inputs and handle errors, see [references/input-validation.md](references/input-validation.md), [references/error-handling.md](references/error-handling.md)
- **Paradigm** - Functional style → **general-fp-guide**; class/OO design (metatable-based) → **general-oop-guide**

## Gotchas

- Arrays are 1-indexed; a `nil` hole breaks the `#` length operator (it stops at the first nil)
- Tables are both array and hash; mixing them is fine but iteration order isn't guaranteed for the hash part
- 5.3+ distinguishes integers from floats; division `/` always returns float, `//` does integer division — silent bugs in pre-5.3 code
- Variables are global by default unless declared `local` — forgetting `local` in a loop counter leaks into the surrounding scope
- `require` caches modules; reloading needs `package.loaded[name] = nil` before re-requiring

## Progressive disclosure

- Read [references/module-pattern.md](references/module-pattern.md) - Load when creating reusable modules or organizing code structure
- Read [references/local-variables.md](references/local-variables.md) - Load when encountering global variable issues or scoping problems
- Read [references/metatables.md](references/metatables.md) - Load when implementing object-oriented patterns or operator overloading
- Read [references/coroutines.md](references/coroutines.md) - Load when implementing cooperative multitasking or async patterns
- Read [references/input-validation.md](references/input-validation.md) - Load when adding type checks or parameter validation
- Read [references/error-handling.md](references/error-handling.md) - Load when handling errors or implementing fallback logic
- Read [references/string-concatenation.md](references/string-concatenation.md) - Load when building strings in loops or formatting output
- Read [references/idiomatic-patterns.md](references/idiomatic-patterns.md) - Load when learning common Lua idioms or patterns
