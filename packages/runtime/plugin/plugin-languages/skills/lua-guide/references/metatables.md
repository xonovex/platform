# metatables: Metatable Pattern for Simple OO

Class-like objects via a table whose `__index` points to itself; construct with `setmetatable(instance, self)` and define methods with colon syntax (`self` is implicit). Pre-allocate all fields in the constructor.

```lua
local Vector = {}
Vector.__index = Vector

function Vector:new(x, y)
    return setmetatable({ x = x or 0, y = y or 0 }, self)
end

function Vector:length()
    return (self.x * self.x + self.y * self.y) ^ 0.5
end

return Vector
```
