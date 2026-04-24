---
name: typescript-to-lua-guidelines
description: Trigger on `.ts` files in game/scripting dirs with TSTL config. Use when transpiling TypeScript to Lua with TSTL 1.24+. Apply for game scripting, clean Lua output, multi-return functions. Keywords: TSTL, TypeScript-to-Lua, namespaces, LuaMultiReturn, decorators, stable tables, coroutines, Lua interop.
---

# TypeScript-to-Lua Coding Guidelines

## Requirements

- TSTL ≥ 1.24; TS strict mode; Lua ≥ 5.3 / LuaJIT 2.1.

## Essentials

- **Code organization** - Favor namespaces/functions over classes for clean Lua output, see [reference/namespaces-vs-classes.md](reference/namespaces-vs-classes.md), [reference/function-patterns.md](reference/function-patterns.md)
- **Performance** - Keep tables stable, design for Lua GC and coroutines, see [reference/stable-tables.md](reference/stable-tables.md), [reference/coroutine-patterns.md](reference/coroutine-patterns.md)
- **TSTL features** - Use LuaMultiReturn, decorators when beneficial, see [reference/multi-return-functions.md](reference/multi-return-functions.md), [reference/tstl-decorators.md](reference/tstl-decorators.md)
- **Optimization** - Avoid heavy TypeScript features, see [reference/avoiding-heavy-features.md](reference/avoiding-heavy-features.md), [reference/performance-tips.md](reference/performance-tips.md)

## Progressive disclosure

- Read [reference/namespaces-vs-classes.md](reference/namespaces-vs-classes.md) - When choosing code organization for clean Lua output
- Read [reference/multi-return-functions.md](reference/multi-return-functions.md) - When implementing Lua-style multiple return values
- Read [reference/function-patterns.md](reference/function-patterns.md) - When defining functions that transpile cleanly
- Read [reference/module-organization.md](reference/module-organization.md) - When structuring TSTL projects or exports
- Read [reference/stable-tables.md](reference/stable-tables.md) - When optimizing for Lua GC or JIT performance
- Read [reference/coroutine-patterns.md](reference/coroutine-patterns.md) - When implementing cooperative multitasking
- Read [reference/lua-interop.md](reference/lua-interop.md) - When calling Lua code from TypeScript
- Read [reference/tstl-decorators.md](reference/tstl-decorators.md) - When using TSTL-specific decorators
- Read [reference/performance-tips.md](reference/performance-tips.md) - When generated Lua code is slow or bloated
- Read [reference/type-safety.md](reference/type-safety.md) - When maintaining types across TS/Lua boundary
- Read [reference/avoiding-heavy-features.md](reference/avoiding-heavy-features.md) - When transpiled output is unexpectedly large
- Read [reference/tsconfig.md](reference/tsconfig.md) - When configuring TSTL compiler options
