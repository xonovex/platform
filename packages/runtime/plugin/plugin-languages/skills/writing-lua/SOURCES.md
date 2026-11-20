# Sources

## Lua 5.4 reference manual

- **URLs:**
  - https://www.lua.org/manual/5.4/
- **Version:** 5.4.0
- **Content SHA256:** 75886da92d130cc217c6f79a02e80223bdc03de61a5823333fa57eafdfd73bdb
- **Last reviewed:** 2026-07-22
- **Used for:** `SKILL.md` and all `references/`; modules, local scope, tables, metatables, coroutines, error handling, iteration, and standard-library behavior.
- **References:** all
- **Aspects extracted:** Lua 5.4 syntax and runtime semantics, including `table.unpack`, metatable behavior, iterator protocols, and coroutine APIs.

## LuaJIT extensions and Xonovex hot-path conventions

- **URLs:**
  - https://luajit.org/extensions.html
  - https://luajit.org/running.html
- **Provenance:** Performance policy and data-layout recommendations are repository-original conventions validated against maintained Lua/LuaJIT workloads
- **Last reviewed:** 2026-07-19
- **Used for:** `references/hot-paths.md`; local lookup caching, stable table shapes, allocation control, trace-friendly loops, and profiling before tuning.
- **References:** `references/hot-paths.md`
- **Aspects extracted:** LuaJIT runtime/extension boundaries come from upstream; exact hot-path preferences are Xonovex conventions.
