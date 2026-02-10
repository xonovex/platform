# input-validation: Input Validation

**Guideline:** Validate function arguments for type, range, and correctness using assertions.

**Rationale:** Lua is dynamically typed. Explicit validation catches errors early, provides clear error messages, and documents expectations.

**Example:**

```lua
-- ✅ Good - validate inputs
function M.clamp(value, min, max)
    assert(type(value) == "number", "value must be a number")
    assert(type(min) == "number", "min must be a number")
    assert(type(max) == "number", "max must be a number")
    assert(min <= max, "min must be <= max")

    if value < min then return min end
    if value > max then return max end
    return value
end

-- ✅ Good - type checking
function M.divide(a, b)
    assert(type(a) == "number", "a must be a number")
    assert(type(b) == "number", "b must be a number")
    assert(b ~= 0, "division by zero")
    return a / b
end

-- ✅ Good - range checking
function M.set_volume(volume)
    assert(type(volume) == "number", "volume must be a number")
    assert(volume >= 0 and volume <= 1, "volume must be between 0 and 1")
    -- Implementation
end
```

**Techniques:**

- Use `assert()` with descriptive messages
- Check types with `type()` function
- Validate numeric ranges and constraints
- Place validations at function start
