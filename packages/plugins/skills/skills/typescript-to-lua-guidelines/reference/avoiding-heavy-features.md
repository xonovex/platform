# avoiding-heavy-features: Avoiding Heavy TypeScript Features

**Guideline:** Avoid complex inheritance, async/await, and heavy OOP patterns that bloat Lua output.

**Rationale:** Complex features require substantial runtime support and generate verbose Lua. Simpler patterns align with Lua's lightweight nature.

**Example:**

```typescript
// AVOID: Heavy inheritance
class Entity {
  move() {}
}
class GameObject extends Entity {
  render() {}
}
class Enemy extends GameObject {
  attack() {}
}

// PREFER: Composition with namespaces
namespace Physics {
  export interface Body {
    x: number;
    y: number;
  }
  export function move(b: Body, dx: number) {
    b.x += dx;
  }
}

namespace Rendering {
  export function render(b: Body) {
    /* draw */
  }
}

// AVOID: async/await
async function loadData() {
  return await fetch();
}

// PREFER: Generators (become coroutines in Lua)
function* loadData() {
  const data = yield fetchAsync();
  return data;
}
```

**Techniques:**

- Use composition patterns instead of inheritance to keep Lua output lightweight
- Replace async/await with callback functions or coroutines (not supported in Lua)
- Prefer interfaces and namespaces for code organization over class-based designs
- Design data structures as plain objects without methods or properties
- Use function\* generators for cooperative multitasking instead of async APIs
- Keep class hierarchies shallow, prefer interfaces for type contracts
- Avoid decorators and complex OOP patterns that require runtime support
