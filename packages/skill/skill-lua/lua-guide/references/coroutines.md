# coroutines: Coroutines for Cooperative Tasks

Use coroutines for timers, animations, and state machines: `coroutine.create` a function, `coroutine.yield()` to pause, `coroutine.resume(co, args)` to continue (returns `ok` plus the yielded/returned values), `coroutine.status` to inspect.

```lua
local function create_timer(duration)
    return coroutine.create(function()
        local elapsed = 0
        while elapsed < duration do
            elapsed = elapsed + coroutine.yield()  -- receives dt from resume
        end
        return true
    end)
end

local co = create_timer(2.0)
local ok, done = coroutine.resume(co, dt)  -- done == true when finished
```
