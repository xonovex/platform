# tstl-decorators: TSTL JSDoc Annotations

Control Lua code generation with TSTL JSDoc annotations:

- `/** @luaTable */`: interface compiles to a raw Lua table (no wrapper).
- `/** @luaIterator */`: generator usable directly in Lua `for ... in`.
- `/** @noSelf */`: function/method omits the implicit `self` parameter.
- `/** @pureVirtual */`: marks an abstract method with no body.

```typescript
/** @luaTable */
interface Config {
  host: string;
  port: number;
  debug: boolean;
}

/** @luaIterator */
function* enumerate<T>(arr: T[]): Generator<[number, T]> {
  for (let i = 0; i < arr.length; i++) yield [i, arr[i]];
}

class Vector {
  constructor(
    public x: number,
    public y: number,
  ) {}
  /** @noSelf */
  static distance(a: Vector, b: Vector): number {
    return Math.sqrt((a.x - b.x) ** 2 + (a.y - b.y) ** 2);
  }
}
```
