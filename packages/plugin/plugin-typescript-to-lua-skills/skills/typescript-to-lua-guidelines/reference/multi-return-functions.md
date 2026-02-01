# multi-return-functions: Lua Multi-Return Functions

**Guideline:** Use `LuaMultiReturn` and `$multi()` for idiomatic Lua multiple-value returns.

**Rationale:** Lua natively supports multiple returns, more efficient than arrays/objects. TSTL provides `LuaMultiReturn` for type-safe multi-returns.

**Example:**

```typescript
import {LuaMultiReturn} from "typescript-to-lua";

// Simple multi-return
function divmod(a: number, b: number): LuaMultiReturn<[number, number]> {
  return $multi(Math.floor(a / b), a % b);
}

// Usage
const [quotient, remainder] = divmod(17, 5);
// quotient = 3, remainder = 2

// Error handling with optional
function parseJSON(
  json: string,
): LuaMultiReturn<[object | null, string | null]> {
  try {
    return $multi(JSON.parse(json), null);
  } catch (e) {
    return $multi(null, (e as Error).message);
  }
}

// Usage
const [data, err] = parseJSON('{"key": "value"}');
if (err) console.error("Parse error:", err);
else console.log("Data:", data);

// Variadic returns
function unpack<T extends any[]>(arr: T): LuaMultiReturn<T> {
  return $multi(...arr);
}
```

**Techniques:**

- Import `LuaMultiReturn` from "typescript-to-lua" package
- Declare return type as `LuaMultiReturn<[type1, type2, ...]>` for multiple values
- Return using `$multi(value1, value2, ...)` function call syntax
- Destructure multi-returns with `const [a, b] = func()` for clarity
- Use for idiomatic Lua patterns like divmod that return quotient and remainder
- Implement error handling by returning optional value with error message
- Support variadic returns with spread operator inside $multi
