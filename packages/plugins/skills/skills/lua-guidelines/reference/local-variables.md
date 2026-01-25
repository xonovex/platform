# local-variables: Always Use Local Variables

**Guideline:** Declare all variables as `local` to avoid implicit globals.

**Rationale:** Global variables pollute the namespace and create hard-to-find bugs. Locals have clearer scope and are easier to reason about.

**Example:**

```lua
-- ✅ Good - all local
local function calculate(a, b)
    local result = a + b
    local squared = result * result
    return squared
end

local value = calculate(3, 4)

-- ❌ Bad - implicit globals
function calculate(a, b)
    result = a + b        -- Global!
    squared = result * result  -- Global!
    return squared
end

value = calculate(3, 4)  -- Global!
```

**Techniques:**
- Prefix all variable declarations with `local`
- Prefix all function declarations with `local`
- Only export through module return table
- Use linters to catch missing `local` keywords
