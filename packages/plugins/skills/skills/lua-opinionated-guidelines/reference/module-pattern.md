# module-pattern: Module Pattern

**Guideline:** Use table-based module pattern with local module table and single return statement.

**Rationale:** Provides clean encapsulation, avoids global namespace pollution, and creates predictable module structure that's easy to reason about.

**Example:**

```lua
-- math2d.lua
local M = {}

-- Simple function in module
function M.length(x, y)
    return (x * x + y * y) ^ 0.5
end

function M.distance(x1, y1, x2, y2)
    local dx = x2 - x1
    local dy = y2 - y1
    return M.length(dx, dy)
end

function M.normalize(x, y)
    local len = M.length(x, y)
    if len == 0 then return 0, 0 end
    return x / len, y / len
end

return M
```

**Techniques:**

- Create local module table `local M = {}`
- Define functions as `M.function_name`
- Return module table at end
- One module per file
