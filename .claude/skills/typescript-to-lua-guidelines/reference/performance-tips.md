# performance-tips: Performance Optimization Tips

**Guideline:** Cache lookups, use numeric for-loops, and pre-allocate data structures for optimal performance.

**Rationale:** Lua has different performance characteristics than JavaScript. Lookups, iterators, and dynamic allocation have costs. Pre-allocation aligns with LuaJIT optimization.

**Example:**
```typescript
// SLOW: Repeated lookups, for-of loop
for (const item of items) {
  console.log(item)
  Math.sqrt(item.value)  // Lookup Math each iteration
}

// FAST: Cache lookups, numeric loop
const sqrt = Math.sqrt
const itemsLen = items.length
for (let i = 0; i < itemsLen; i++) {
  const item = items[i]
  console.log(item)
  sqrt(item.value)  // Cached reference
}

// SLOW: Dynamic array growth
const results: number[] = []
for (const item of items) {
  results.push(item * 2)  // Reallocates as it grows
}

// FAST: Pre-allocate
const results = new Array<number>(items.length)
for (let i = 0; i < items.length; i++) {
  results[i] = items[i] * 2  // No reallocation
}

// SLOW: Table allocation in hot path
function updatePositions(bodies: Body[]) {
  for (const body of bodies) {
    const delta = { x: 0, y: 0 }  // Allocates each call
    delta.x = body.vx * dt
    delta.y = body.vy * dt
    body.x += delta.x
    body.y += delta.y
  }
}

// FAST: Reuse allocation
const delta = { x: 0, y: 0 }
function updatePositions(bodies: Body[]) {
  for (let i = 0; i < bodies.length; i++) {
    const body = bodies[i]
    delta.x = body.vx * dt
    delta.y = body.vy * dt
    body.x += delta.x
    body.y += delta.y
  }
}
```

**Techniques:**
- Cache frequently-used global functions in local variables for fast lookup
- Use numeric for-loops (`for (let i = 0; i < arr.length; i++)`) instead of for-of
- Avoid for-of and for-in iterators that generate overhead in Lua
- Pre-allocate arrays when final size is known to avoid dynamic growth
- Minimize table allocations in hot paths (frequently-called functions)
- Store module-level constants in locals within functions that use them
- Profile hot paths and focus optimization on frequently-executed code
- Use LuaJIT-friendly patterns: early allocation, numeric indices, minimal lookups
