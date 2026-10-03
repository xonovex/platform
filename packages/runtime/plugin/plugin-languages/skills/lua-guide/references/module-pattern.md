# module-pattern: Module Pattern

One module per file: a `local` table, functions as `M.name`, single `return M` at the end.

```lua
-- math2d.lua
local M = {}

function M.length(x, y)
    return (x * x + y * y) ^ 0.5
end

function M.distance(x1, y1, x2, y2)
    return M.length(x2 - x1, y2 - y1)  -- reference siblings via M
end

return M
```
