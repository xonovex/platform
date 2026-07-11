# namespaces-vs-classes: Organize for Clean Lua Tables

Prefer namespaces over classes: classes emit metatables + inheritance chains, namespaces compile to plain Lua tables with functions. Export functions directly (not static methods) and pass state as the first arg. Numeric enums compile to plain numbers. Nested namespaces map to nested tables — keep depth to 2-3 levels.

```typescript
// AVOID: class → metatables, inheritance runtime
class Vector {
  constructor(
    public x: number,
    public y: number,
  ) {}
  magnitude(): number {
    return Math.sqrt(this.x * this.x + this.y * this.y);
  }
}

// PREFER: namespace → plain table of functions
namespace Vector {
  export interface Vec {
    x: number;
    y: number;
  }
  export function magnitude(v: Vec): number {
    return Math.sqrt(v.x * v.x + v.y * v.y);
  }
  export function add(a: Vec, b: Vec): Vec {
    return {x: a.x + b.x, y: a.y + b.y};
  }
}

enum Direction {
  Up = 0,
  Down = 1,
  Left = 2,
  Right = 3,
} // compiles to numbers

// Too deep — verbose Lua: namespace Game.Entities.Physics.Collision.Detection {}
```
