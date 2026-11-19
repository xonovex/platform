# coroutine-patterns: Generators as Lua Coroutines

`function*` generators compile to Lua coroutines with no extra runtime. `yield` pauses and receives the value passed on resume (e.g. delta-time); `yield*` chains generators. Type as `Generator<YieldType, ReturnType>`. Use for timers, tweens, state machines, and task sequences.

```typescript
function* timer(seconds: number): Generator<void, void> {
  let elapsed = 0;
  while (elapsed < seconds) {
    const delta = yield; // resume value from Lua
    elapsed += delta;
  }
}

function* sequence(obj: {x: number}): Generator<void, void> {
  yield* timer(2); // wait 2s
  yield* animate(obj, 1); // then animate 1s
}
```
