# namespaces-vs-classes: Namespace vs Class Design

**Guideline:** Prefer namespaces over classes for cleaner, more efficient Lua output.

**Rationale:** Classes generate metatables and inheritance chains. Namespaces compile to simple Lua tables with functions, resulting in efficient output.

**Example:**

```typescript
// AVOID: Class generates complex Lua with metatables
class Vector {
  constructor(
    public x: number,
    public y: number,
  ) {}
  magnitude(): number {
    return Math.sqrt(this.x * this.x + this.y * this.y);
  }
  normalize(): Vector {
    const m = this.magnitude();
    return new Vector(this.x / m, this.y / m);
  }
}

// PREFER: Namespace generates simple Lua tables
namespace Vector {
  export interface Vec {
    x: number;
    y: number;
  }

  export function magnitude(v: Vec): number {
    return Math.sqrt(v.x * v.x + v.y * v.y);
  }

  export function normalize(v: Vec): Vec {
    const m = magnitude(v);
    return {x: v.x / m, y: v.y / m};
  }

  export function add(a: Vec, b: Vec): Vec {
    return {x: a.x + b.x, y: a.y + b.y};
  }
}

// Usage
const v: Vector.Vec = {x: 3, y: 4};
const mag = Vector.magnitude(v); // 5
const norm = Vector.normalize(v); // { x: 0.6, y: 0.8 }
```

**Techniques:**

- Use `export namespace` for grouping related functions instead of classes
- Export functions directly within namespaces (not static methods)
- Avoid classes for purely functional code, use namespaces for cleaner output
- Keep functions pure and stateless within namespaces
- Use composition patterns to share behavior across namespaces
- Consider inheritance only when truly needed (rare in Lua)
- Prefer namespace-based organization for utilities and helper functions
