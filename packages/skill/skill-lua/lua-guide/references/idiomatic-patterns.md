# idiomatic-patterns: Idiomatic Patterns

Lua truthiness idioms (only `nil` and `false` are falsy):

```lua
local value = condition and true_value or false_value  -- ternary; false_value must be truthy
local x = config.x or 0                                 -- default value
local nested = obj and obj.field and obj.field.nested   -- safe navigation
a, b = b, a                                             -- swap via tuple assignment
local x, y, z = table.unpack(position)                  -- 5.4: table.unpack, not global unpack
```

The `and`/`or` ternary is broken when `true_value` can itself be `false`/`nil` — use an `if` then.
