# jit-friendly-tables: JIT-Friendly Table Operations

**Guideline:** Keep table shapes stable by pre-allocating all fields and avoid adding fields after creation.

**Rationale:** LuaJIT optimizes based on table shape. Adding fields later causes shape changes that prevent JIT compilation and reduce performance.

**Example:**

```lua
-- ✅ Good - stable tables
local function create_entity(x, y, type)
    -- Pre-allocate all fields
    return {
        x = x,
        y = y,
        type = type,
        velocity_x = 0,
        velocity_y = 0,
        health = 100,
        active = true
    }
end

-- ✅ Good - pre-allocated array
local function create_array(size)
    local arr = {}
    for i = 1, size do
        arr[i] = 0
    end
    return arr
end

-- ❌ Bad - unstable table (fields added after creation)
local function create_entity_bad(x, y)
    local entity = {x = x, y = y}
    -- Adding fields later hurts JIT
    entity.velocity_x = 0
    entity.velocity_y = 0
    return entity
end

-- ❌ Bad - using pairs in hot path
for k, v in pairs(entities) do  -- Slow
    update_entity(v)
end

-- ✅ Good - numeric for loop
for i = 1, #entities do  -- Fast
    update_entity(entities[i])
end
```

**Techniques:**
- Define all fields in table literal or constructor
- Pre-allocate arrays with known size
- Use numeric for loops instead of `pairs` for arrays
- Avoid sparse arrays
