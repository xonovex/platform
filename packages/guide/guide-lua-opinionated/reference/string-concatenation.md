# string-concatenation: String Concatenation

**Guideline:** Use `table.concat()` for building strings in loops instead of repeated concatenation.

**Rationale:** String concatenation with `..` creates new string objects each time. For loops, this creates many intermediate strings. Table concatenation is O(n) instead of O(n²).

**Example:**

```lua
-- ❌ Bad - repeated concatenation creates many strings
local str = ""
for i = 1, 1000 do
    str = str .. tostring(i) .. ","
end

-- ✅ Good - table concatenation
local parts = {}
for i = 1, 1000 do
    parts[i] = tostring(i)
end
local str = table.concat(parts, ",")

-- ✅ Good - string.format for simple cases
local message = string.format("Player %s scored %d points", name, score)
```

**Techniques:**

- Build array of string parts
- Use `table.concat(parts, separator)`
- Use `string.format()` for simple interpolation
- Avoid `..` in loops
