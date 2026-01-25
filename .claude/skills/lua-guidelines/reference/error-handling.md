# error-handling: Error Handling

**Guideline:** Use explicit error return values (value, error) pattern instead of exceptions for expected failures.

**Rationale:** Follows Lua conventions. Makes error handling explicit and visible. More appropriate for recoverable errors than `assert()`.

**Example:**

```lua
-- ✅ Protected call with error handling
local function load_config(path)
    local f, err = io.open(path, "r")
    if not f then
        return nil, "Failed to open file: " .. err
    end

    local content = f:read("*all")
    f:close()

    local config, err = parse_config(content)
    if not config then
        return nil, "Failed to parse config: " .. err
    end

    return config
end

-- Usage
local config, err = load_config("config.lua")
if not config then
    print("Error loading config:", err)
    -- Use defaults
    config = get_default_config()
end
```

**Techniques:**
- Return `nil, error_message` on failure
- Return value on success
- Check first return value before using
- Use `pcall` for protecting risky operations
