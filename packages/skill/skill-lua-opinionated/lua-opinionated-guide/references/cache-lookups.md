# cache-lookups: Cache Global Lookups

On hot paths, hoist repeated global and module-table lookups into locals — global/table access is slower than local access. Apply where profiling shows it.

```lua
local insert = table.insert  -- hoist the module lookup once, reuse in the loop
for i = 1, #items do insert(dst, transform(items[i])) end
```
