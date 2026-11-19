# input-validation: Input Validation

Validate argument type and range at function start with `assert(cond, message)`: Lua is dynamically typed, so `assert` gives an early, descriptive failure.

```lua
function M.clamp(value, min, max)
    assert(type(value) == "number", "value must be a number")
    assert(min <= max, "min must be <= max")
    if value < min then return min end
    if value > max then return max end
    return value
end
```
