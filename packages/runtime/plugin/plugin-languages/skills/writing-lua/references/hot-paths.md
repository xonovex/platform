# hot-paths: Tune Lua Hot Paths

Tune only the code that profiling shows is hot. The tunings help LuaJIT 2.1 most, because they keep its traces simple, and the same principles still help Lua 5.4.

## Numeric loops

Iterate dense arrays with the plain numeric for-loop. That simple idiom is the deliverable: do not manually unroll it (`local i = 0; while i < n - 3 do ... i = i + 4 end`) or reach for FFI arrays / SIMD-style tricks unless profiling proves the need.

```lua
for i = 1, #entities do update_entity(entities[i]) end
```

## Cached lookups

On hot paths, hoist repeated global and module-table lookups into locals: global/table access is slower than local access.

```lua
local insert = table.insert  -- hoist the module lookup once, reuse in the loop
for i = 1, #items do insert(dst, transform(items[i])) end
```
