# jit-friendly-tables: JIT-Friendly Table Operations

Define all fields in the table literal/constructor and never add fields after creation — LuaJIT specializes on table shape, so a later field addition changes the shape and deoptimizes the trace. Pre-allocate arrays to their known size, keep them dense (no `nil` holes, which break `#` and fast array traces), and iterate arrays with numeric `for i = 1, #t` rather than `pairs()`, which the JIT can't compile as tightly.

```lua
-- ✅ stable shape: all fields in the literal
local function create_entity(x, y, type)
    return { x = x, y = y, type = type, velocity_x = 0, velocity_y = 0, health = 100, active = true }
end

-- ❌ unstable shape: fields added after creation
local function create_entity_bad(x, y)
    local entity = { x = x, y = y }
    entity.velocity_x = 0  -- shape change deoptimizes the JIT
    return entity
end

-- ✅ pre-allocate dense array, iterate numerically
local arr = {}
for i = 1, size do arr[i] = 0 end
for i = 1, #entities do update_entity(entities[i]) end  -- not pairs()
```
