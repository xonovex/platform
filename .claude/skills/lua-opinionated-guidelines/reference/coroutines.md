# coroutines: Coroutines for Cooperative Tasks

**Guideline:** Use coroutines for cooperative multitasking, timed actions, and state machines.

**Rationale:** Coroutines enable writing sequential code for asynchronous operations without callbacks. Ideal for game timers, animations, and task scheduling.

**Example:**

```lua
-- Coroutine for timed actions
function create_timer(duration)
    return coroutine.create(function()
        local elapsed = 0
        while elapsed < duration do
            local dt = coroutine.yield()
            elapsed = elapsed + dt
        end
        return true  -- Completed
    end)
end

-- Task manager
local TimerManager = {}

function TimerManager:new()
    return setmetatable({timers = {}}, {__index = self})
end

function TimerManager:add_timer(duration, callback)
    local timer = {
        coro = create_timer(duration),
        callback = callback
    }
    table.insert(self.timers, timer)
end

function TimerManager:update(dt)
    local i = 1
    while i <= #self.timers do
        local timer = self.timers[i]
        local ok, done = coroutine.resume(timer.coro, dt)

        if done then
            timer.callback()
            table.remove(self.timers, i)
        else
            i = i + 1
        end
    end
end

-- Usage
local manager = TimerManager:new()
manager:add_timer(2.0, function()
    print("Timer fired after 2 seconds")
end)

-- In game loop
manager:update(dt)
```

**Techniques:**
- Create coroutine with `coroutine.create`
- Use `coroutine.yield()` to pause execution
- Resume with `coroutine.resume(coro, args)`
- Check status with return values or `coroutine.status`
