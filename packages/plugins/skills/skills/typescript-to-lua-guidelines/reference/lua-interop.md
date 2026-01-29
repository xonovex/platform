# lua-interop: Integration with Lua APIs

**Guideline:** Declare Lua globals with TypeScript types for type-safe integration.

**Rationale:** Lua environments provide global APIs (LÖVE, Corona, custom engines). TypeScript declarations enable autocomplete, type checking, and full interop.

**Example:**

```typescript
// LÖVE 2D game engine declarations
declare const love: {
  graphics: {
    draw: (drawable: Drawable, x: number, y: number) => void;
    print: (text: string, x: number, y: number) => void;
    setColor: (r: number, g: number, b: number, a?: number) => void;
  };
  window: {
    setWidth: (width: number) => void;
    setHeight: (height: number) => void;
    getWidth: () => number;
    getHeight: () => number;
  };
  timer: {
    getDelta: () => number;
    getTime: () => number;
  };
};

// Custom engine example
declare const engine: {
  render: (meshId: number, transform: Matrix4) => void;
  physics: {
    raycast: (origin: Vector3, direction: Vector3) => HitResult | null;
  };
};

// Type-safe usage
love.graphics.setColor(1, 0, 0); // Red
love.graphics.print("Hello", 100, 50);
const width = love.window.getWidth();
```

**Techniques:**

- Use `declare const` to declare Lua global variables and libraries
- Define complete type signatures for all global APIs and functions
- Use string literal types for enum-like parameters (e.g., `"fill" | "line"`)
- Group related functions in nested object types matching Lua structure
- Document expected parameters and return types for each API call
- Use union types for optional or variable-length parameters
- Create type definitions that match exact Lua API signatures
