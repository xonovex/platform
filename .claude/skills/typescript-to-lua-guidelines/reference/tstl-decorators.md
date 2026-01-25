# tstl-decorators: TSTL Decorators and Annotations

**Guideline:** Use TSTL-specific decorators and JSDoc annotations to control Lua code generation.

**Rationale:** TSTL decorators like `@luaTable` and `@luaIterator` fine-tune generated Lua. Custom decorators bridge TypeScript and Lua idioms.

**Example:**
```typescript
/** @luaTable */
interface Config {
  host: string
  port: number
  debug: boolean
}

/** @luaIterator */
function* enumerate<T>(arr: T[]): Generator<[number, T]> {
  for (let i = 0; i < arr.length; i++) {
    yield [i, arr[i]]
  }
}

class Vector {
  constructor(public x: number, public y: number) {}

  // @noSelf means this function doesn't receive 'self' parameter
  /** @noSelf */
  static distance(a: Vector, b: Vector): number {
    const dx = a.x - b.x
    const dy = a.y - b.y
    return Math.sqrt(dx * dx + dy * dy)
  }

  // Regular method gets self: parameter
  magnitude(): number {
    return Math.sqrt(this.x * this.x + this.y * this.y)
  }
}

// Custom memoization decorator
function memoize(fn: Function) {
  const cache = new Map()
  return (...args: any[]) => {
    const key = JSON.stringify(args)
    if (cache.has(key)) return cache.get(key)
    const result = fn(...args)
    cache.set(key, result)
    return result
  }
}
```

**Techniques:**
- Use `/** @luaTable */` JSDoc annotation for raw Lua table interfaces
- Use `/** @luaIterator */` for generator-based iterators
- Use `/** @noSelf */` for functions that don't use 'self' context
- Use `/** @pureVirtual */` to mark abstract methods
- Implement custom decorators for cross-cutting concerns
- Apply decorators for performance optimizations: memoization, caching
- Use decorators for logging, debugging, and tracing functionality
- Apply decorators at method level for fine-grained code generation control
