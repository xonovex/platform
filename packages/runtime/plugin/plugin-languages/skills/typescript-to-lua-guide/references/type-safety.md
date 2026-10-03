# type-safety: Types Across the TS/Lua Boundary

Strict typing guides TSTL output and catches errors before transpile. Model results/optionals as discriminated unions paired with a namespace of factories/combinators, and use `readonly` (including `readonly [x, y]` tuples) to signal immutable table shapes.

```typescript
type Result<T, E> = {kind: "ok"; value: T} | {kind: "err"; error: E};

namespace Result {
  export function ok<T, E>(value: T): Result<T, E> {
    return {kind: "ok", value};
  }
  export function err<T, E>(error: E): Result<T, E> {
    return {kind: "err", error};
  }
}

interface Entity {
  readonly id: number;
  readonly velocity: readonly [number, number];
}
```
