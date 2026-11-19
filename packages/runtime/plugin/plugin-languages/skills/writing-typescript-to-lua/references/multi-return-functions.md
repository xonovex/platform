# multi-return-functions: Lua Multi-Return with $multi

Import `LuaMultiReturn` from `"typescript-to-lua"`, type the return as `LuaMultiReturn<[T1, T2, ...]>`, and return via `$multi(...)`. Destructure at the call site. This emits native Lua multiple returns: cheaper than array/object wrappers.

```typescript
import {LuaMultiReturn} from "typescript-to-lua";

function divmod(a: number, b: number): LuaMultiReturn<[number, number]> {
  return $multi(Math.floor(a / b), a % b);
}
const [quotient, remainder] = divmod(17, 5); // 3, 2

// value-or-error pattern
function parseJSON(
  json: string,
): LuaMultiReturn<[object | null, string | null]> {
  try {
    return $multi(JSON.parse(json), null);
  } catch (e) {
    return $multi(null, (e as Error).message);
  }
}

// variadic: spread into $multi
function unpack<T extends unknown[]>(arr: T): LuaMultiReturn<T> {
  return $multi(...arr);
}
```
