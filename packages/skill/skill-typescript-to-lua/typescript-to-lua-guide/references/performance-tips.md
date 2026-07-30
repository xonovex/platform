# performance-tips: Lua/JIT Performance Shaping

Shape TS source so generated Lua stays JIT-friendly. Define all table fields upfront: adding fields after creation changes table shape and deoptimizes LuaJIT; use `readonly` factory functions. Cache global lookups (`const sqrt = Math.sqrt`) in locals, use numeric `for` loops over `for-of`, and pre-allocate with `new Array<number>(n)` instead of growing via `push`. Reuse scratch allocations in hot paths. The JIT "why" lives in **lua-opinionated-guide**.

```typescript
// FAST: stable shape, all fields set in factory
interface Entity {
  readonly x: number;
  readonly y: number;
  readonly vx: number;
  readonly vy: number;
  readonly id: number;
}
function createEntity(x: number, y: number, id: number): Entity {
  return {x, y, vx: 0, vy: 0, id};
}

// FAST: cached lookup, numeric loop, pre-allocated result
const sqrt = Math.sqrt;
const results = new Array<number>(items.length);
for (let i = 0; i < items.length; i++) results[i] = sqrt(items[i].value);

// FAST: reuse allocation instead of per-iteration table
const delta = {x: 0, y: 0};
function updatePositions(bodies: Body[], dt: number) {
  for (let i = 0; i < bodies.length; i++) {
    const body = bodies[i];
    delta.x = body.vx * dt;
    delta.y = body.vy * dt;
    body.x += delta.x;
    body.y += delta.y;
  }
}
```
