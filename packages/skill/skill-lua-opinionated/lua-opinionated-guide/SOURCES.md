# Sources

## LuaJIT extensions and Xonovex hot-path conventions

- **URLs:**
  - https://luajit.org/extensions.html
  - https://luajit.org/running.html
- **Provenance:** Performance policy and data-layout recommendations are repository-original conventions validated against maintained Lua/LuaJIT workloads
- **Last reviewed:** 2026-07-19
- **Used for:** `SKILL.md` and all `references/`; local lookup caching, stable table shapes, allocation control, trace-friendly loops, and profiling before tuning.
- **Aspects extracted:** LuaJIT runtime/extension boundaries come from upstream; exact hot-path preferences are Xonovex's opinionated overlay.
