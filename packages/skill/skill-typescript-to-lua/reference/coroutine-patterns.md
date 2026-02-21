# coroutine-patterns: Coroutine-Friendly Patterns

**Guideline:** Design generators for seamless Lua coroutine integration without runtime overhead.

**Rationale:** Lua coroutines provide cooperative multitasking naturally. TypeScript generators compile to coroutines, enabling timers, sequences, and state machines.

**Example:**

```typescript
// Timer using generator/coroutine
function* timer(seconds: number): Generator<void, void> {
  let elapsed = 0;
  while (elapsed < seconds) {
    const delta = yield; // Receives delta-time from Lua
    elapsed += delta;
  }
}

// Animation sequence
function* animatePosition(
  target: {x: number},
  endX: number,
  duration: number,
): Generator<void, boolean> {
  let elapsed = 0;
  const startX = target.x;
  while (elapsed < duration) {
    const delta = yield;
    elapsed += delta;
    const progress = elapsed / duration;
    target.x = startX + (endX - startX) * progress;
  }
  return true;
}

// Chaining coroutines
function* sequence(): Generator<void, void> {
  yield* timer(2); // Wait 2 seconds
  yield* animatePosition(obj, 100, 1); // Animate 1 second
}
```

**Techniques:**

- Use `function*` syntax to define generators that compile to Lua coroutines
- Yield control points with `yield` to pause execution and allow other tasks
- Accept delta-time as yield result to measure elapsed time accurately
- Return final values after completion to signal coroutine success or results
- Chain generators with `yield*` to compose coroutines without overhead
- Type generators as `Generator<YieldType, ReturnType>`
- Use for implementing timers, animations, state machines, and task sequences
