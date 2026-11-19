# jit-friendly-tables: Numeric Loop Idiom

Iterate dense arrays with the plain numeric for-loop. That simple idiom is the deliverable: do not manually unroll it (`local i = 0; while i < n - 3 do ... i = i + 4 end`) or reach for FFI arrays / SIMD-style tricks unless profiling proves the need.

```lua
for i = 1, #entities do update_entity(entities[i]) end
```
