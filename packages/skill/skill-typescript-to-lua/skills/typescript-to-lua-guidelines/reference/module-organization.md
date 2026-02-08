# module-organization: Module Organization

**Guideline:** Organize code in flat namespace hierarchies that map cleanly to Lua tables.

**Rationale:** Nested namespaces compile to nested Lua tables, creating efficient module systems aligned with Lua patterns.

**Example:**

```typescript
// Good: 2-level hierarchy, organized by feature
namespace Physics {
  namespace Collision {
    export interface Circle {
      x: number;
      y: number;
      radius: number;
    }
    export function intersect(a: Circle, b: Circle): boolean {
      const dx = a.x - b.x;
      const dy = a.y - b.y;
      const dist = Math.sqrt(dx * dx + dy * dy);
      return dist < a.radius + b.radius;
    }
  }
  export function update(body: Body, dt: number) {
    /* ... */
  }
}

namespace Rendering {
  namespace Color {
    export const RED = {r: 1, g: 0, b: 0};
    export const BLUE = {r: 0, g: 0, b: 1};
  }
  export function draw(mesh: Mesh) {
    /* ... */
  }
}

// Avoid: Too deep (translates to verbose Lua)
namespace Game.Entities.Physics.Collision.Detection {}
```

**Techniques:**

- Use nested namespaces for logical grouping that maps to Lua table structure
- Export functions directly within namespaces (not as properties)
- Keep hierarchy shallow: maximum 2-3 levels deep for clarity
- Group related functionality together in single namespace
- Organize by feature (Physics, Rendering, Input) rather than by file
- Use consistent naming conventions across namespace hierarchy
- Minimize cross-namespace dependencies for cleaner module structure
