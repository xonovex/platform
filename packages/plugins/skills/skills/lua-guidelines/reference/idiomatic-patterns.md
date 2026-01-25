# idiomatic-patterns: Idiomatic Patterns

**Guideline:** Use Lua's idiomatic patterns for common operations like ternary expressions, default values, and safe navigation.

**Rationale:** These patterns are widely recognized in Lua community, concise, and leverage Lua's truthiness semantics effectively.

**Example:**

```lua
-- ✅ Ternary operator pattern
local value = condition and true_value or false_value

-- ✅ Default value pattern
local x = config.x or 0
local name = user.name or "Anonymous"

-- ✅ Safe navigation
local value = obj and obj.field and obj.field.nested

-- ✅ Multiple return values
local ok, result = pcall(risky_function)
if ok then
    use_result(result)
else
    handle_error(result)
end

-- ✅ Swap variables
a, b = b, a

-- ✅ Unpack table
local x, y, z = unpack(position)
```

**Techniques:**
- Use `and`/`or` for conditional expressions
- Use `or` for default values
- Use `and` chain for safe navigation
- Leverage multiple return values
- Use tuple assignment for swapping
