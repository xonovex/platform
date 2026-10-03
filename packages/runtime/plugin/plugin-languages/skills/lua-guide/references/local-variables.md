# local-variables: Always Use Local Variables

Prefix every variable and function with `local`: Lua defaults to globals, so a missing `local` silently leaks into `_G` and creates cross-file bugs. Export only through the module return table.

```lua
local function calculate(a, b)
    local result = a + b   -- ✅ local; without `local` this is a global
    return result * result
end
```
