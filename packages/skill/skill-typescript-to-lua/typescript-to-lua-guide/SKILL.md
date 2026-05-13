---
name: typescript-to-lua-guide
description: "Use when editing TypeScript that compiles to Lua via TSTL 1.24+ (game scripting, Defold, embedded engines). Triggers on `.ts` files in game/scripting directories with TSTL config, and on prompts about LuaMultiReturn, namespaces, decorators, stable table layouts, or Lua interop boundaries, even when the user doesn't say 'TSTL'. Skip ordinary Node/browser TypeScript (use typescript-guide) and pure Lua (use lua-guide / lua-opinionated-guide)."
---

# TypeScript-to-Lua Coding Guidelines

## Requirements

- TSTL ≥ 1.24; TS strict mode; Lua ≥ 5.3 / LuaJIT 2.1.

## Essentials

- **Code organization** - Favor namespaces/functions over classes for clean Lua output, see [references/namespaces-vs-classes.md](references/namespaces-vs-classes.md), [references/function-patterns.md](references/function-patterns.md)
- **Performance** - Keep tables stable, design for Lua GC and coroutines, see [references/stable-tables.md](references/stable-tables.md), [references/coroutine-patterns.md](references/coroutine-patterns.md)
- **TSTL features** - Use LuaMultiReturn, decorators when beneficial, see [references/multi-return-functions.md](references/multi-return-functions.md), [references/tstl-decorators.md](references/tstl-decorators.md)
- **Optimization** - Avoid heavy TypeScript features, see [references/avoiding-heavy-features.md](references/avoiding-heavy-features.md), [references/performance-tips.md](references/performance-tips.md)

## Gotchas

- Not all TypeScript features translate — generators, `for-await-of`, dynamic `import()`, and BigInt are unsupported or partially supported
- Translation is source-to-source; runtime semantics follow Lua, not JS — `0` and `""` are truthy in Lua, falsy in JS
- Lua's table indexing (1-based, no `length` for holes) replaces JS arrays — pass-through types abstract this but iteration order changes
- Importing JS-only libraries silently fails at translation — only `@types` packages with Lua-runtime equivalents work
- `tstlc` is the CLI; build configs are different from `tsc` — sharing a `tsconfig.json` with web projects requires careful `compilerOptions` split

## Progressive disclosure

- Read [references/namespaces-vs-classes.md](references/namespaces-vs-classes.md) - Load when choosing code organization for clean Lua output
- Read [references/multi-return-functions.md](references/multi-return-functions.md) - Load when implementing Lua-style multiple return values
- Read [references/function-patterns.md](references/function-patterns.md) - Load when defining functions that transpile cleanly
- Read [references/module-organization.md](references/module-organization.md) - Load when structuring TSTL projects or exports
- Read [references/stable-tables.md](references/stable-tables.md) - Load when optimizing for Lua GC or JIT performance
- Read [references/coroutine-patterns.md](references/coroutine-patterns.md) - Load when implementing cooperative multitasking
- Read [references/lua-interop.md](references/lua-interop.md) - Load when calling Lua code from TypeScript
- Read [references/tstl-decorators.md](references/tstl-decorators.md) - Load when using TSTL-specific decorators
- Read [references/performance-tips.md](references/performance-tips.md) - Load when generated Lua code is slow or bloated
- Read [references/type-safety.md](references/type-safety.md) - Load when maintaining types across TS/Lua boundary
- Read [references/avoiding-heavy-features.md](references/avoiding-heavy-features.md) - Load when transpiled output is unexpectedly large
- Read [references/tsconfig.md](references/tsconfig.md) - Load when configuring TSTL compiler options
