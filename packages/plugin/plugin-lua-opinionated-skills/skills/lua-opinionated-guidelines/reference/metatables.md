# metatables: Metatable Pattern for Simple OO

**Guideline:** Use metatables with `__index` for simple object-oriented programming when needed.

**Rationale:** Provides class-like behavior with inheritance and methods while remaining JIT-friendly. Simpler than complex OO frameworks.

**Example:**

```lua
-- vector.lua
local Vector = {}
Vector.__index = Vector

-- Constructor
function Vector:new(x, y)
    local instance = {
        x = x or 0,
        y = y or 0
    }
    return setmetatable(instance, self)
end

-- Methods
function Vector:length()
    return (self.x * self.x + self.y * self.y) ^ 0.5
end

function Vector:normalize()
    local len = self:length()
    if len == 0 then return self end
    self.x = self.x / len
    self.y = self.y / len
    return self
end

function Vector:add(other)
    return Vector:new(self.x + other.x, self.y + other.y)
end

-- Usage
local v1 = Vector:new(3, 4)
local len = v1:length()  -- 5
local v2 = v1:add(Vector:new(1, 1))

return Vector
```

**Techniques:**

- Create table with `__index` pointing to itself
- Use constructor pattern with `setmetatable`
- Define methods using colon syntax
- Pre-allocate all fields in constructor
