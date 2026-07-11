# string-concatenation: String Concatenation

In loops, collect parts in a table and join with `table.concat(parts, sep)` — repeated `..` allocates a new string each iteration. Use `string.format` for fixed interpolation.

```lua
local parts = {}
for i = 1, 1000 do
    parts[i] = tostring(i)
end
local str = table.concat(parts, ",")
```
