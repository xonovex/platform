# function-patterns: Function and Module Patterns

**Guideline:** Use pure functions, module-level constants, and type-safe enums for efficient Lua.

**Rationale:** Pure functions compile to simple Lua without closures. Constants generate efficient tables. Patterns align with Lua's procedural nature.

**Example:**
```typescript
// Pure functions in namespace
namespace Math {
  export function add(a: number, b: number): number { return a + b }
  export function multiply(a: number, b: number): number { return a * b }
}

// Module-level constants
namespace Colors {
  export const RED = 0xFF0000
  export const GREEN = 0x00FF00
  export const BLUE = 0x0000FF
}

// Numeric enums (efficient)
enum Direction {
  Up = 0,
  Down = 1,
  Left = 2,
  Right = 3
}

// Enum utilities
namespace DirectionUtils {
  export function opposite(dir: Direction): Direction {
    if (dir === Direction.Up) return Direction.Down
    if (dir === Direction.Down) return Direction.Up
    if (dir === Direction.Left) return Direction.Right
    return Direction.Left
  }
}

// Generic functions
namespace Array {
  export function map<T, U>(arr: T[], fn: (t: T) => U): U[] {
    const result: U[] = []
    for (let i = 0; i < arr.length; i++) {
      result[i] = fn(arr[i])
    }
    return result
  }
}
```

**Techniques:**
- Export pure functions in namespaces to organize code without closures
- Define module-level constants in dedicated namespaces for reusability
- Use numeric enums for efficient representation that compiles to simple numbers
- Create utility namespaces with static functions for enum operations
- Keep functions stateless to enable inline compilation
- Use type parameters for generic functions to maintain type safety
- Group related utilities in single namespaces to generate efficient tables
