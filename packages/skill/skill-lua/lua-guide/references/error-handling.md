# error-handling: Error Handling

Return `nil, error_message` for expected failures (not `error()`/exceptions); the caller checks the first return before using it. Wrap risky operations that may throw in `pcall`.

```lua
local function load_config(path)
    local f, err = io.open(path, "r")
    if not f then return nil, "open failed: " .. err end
    local content = f:read("*all")
    f:close()
    return parse_config(content)  -- also returns nil, err on failure
end

local config, err = load_config("config.lua")
if not config then config = get_default_config() end
```
