# namespaces-vs-classes: Organize for Clean Lua Tables

Use the literal `namespace` keyword: namespaces compile to plain Lua tables with functions, while classes emit metatables + inheritance chains. Numeric enums compile to plain numbers. Nested namespaces map to nested tables: keep depth to 2-3 levels.

```typescript
namespace Vector {
  export interface Vec {
    x: number;
    y: number;
  }
  export function magnitude(v: Vec): number {
    return Math.sqrt(v.x * v.x + v.y * v.y);
  }
}

enum Direction {
  Up = 0,
  Down = 1,
  Left = 2,
  Right = 3,
} // compiles to numbers

// Too deep, verbose Lua: namespace Game.Entities.Physics.Collision.Detection {}
```
