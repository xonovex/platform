# cache-lookups: Cache Table Lookups

**Guideline:** Cache repeated table field accesses in local variables to improve performance.

**Rationale:** Table lookups are slower than local variable access. Caching reduces redundant table accesses, especially in hot paths and loops.

**Example:**

```lua
-- ❌ Bad - repeated table lookups
function update_position(entity, dt)
    entity.x = entity.x + entity.velocity_x * dt
    entity.y = entity.y + entity.velocity_y * dt
end

-- ✅ Good - cached lookups
function update_position(entity, dt)
    local x = entity.x
    local y = entity.y
    local vx = entity.velocity_x
    local vy = entity.velocity_y

    entity.x = x + vx * dt
    entity.y = y + vy * dt
end

-- ✅ Good - cache module table
local math_sin = math.sin
local math_cos = math.cos

function rotate_point(x, y, angle)
    local cos_a = math_cos(angle)
    local sin_a = math_sin(angle)
    return x * cos_a - y * sin_a,
           x * sin_a + y * cos_a
end
```

**Techniques:**

- Store frequently accessed table fields in locals
- Cache standard library functions (math.sin, etc.)
- Especially important in loops and hot paths
- Balance readability vs performance
