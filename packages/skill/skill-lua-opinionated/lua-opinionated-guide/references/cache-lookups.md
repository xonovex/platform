# cache-lookups: Cache Table Lookups

On hot paths, hoist repeated table-field, global, and stdlib lookups into locals — table/global access is slower than local access. Cache `math.sin`/`math.cos` and other stdlib functions into upvalue locals in hot loops; elsewhere it is noise, so apply where profiling shows it.

```lua
-- ❌ repeated field lookups
function update_position(entity, dt)
    entity.x = entity.x + entity.velocity_x * dt
    entity.y = entity.y + entity.velocity_y * dt
end

-- ✅ cache fields + stdlib funcs into locals
local sin, cos = math.sin, math.cos
function update_position(entity, dt)
    local x, y = entity.x, entity.y
    entity.x = x + entity.velocity_x * dt
    entity.y = y + entity.velocity_y * dt
end
```
