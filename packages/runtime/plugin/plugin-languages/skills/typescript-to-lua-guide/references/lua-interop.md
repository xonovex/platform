# lua-interop: Typing Lua Globals

Declare host-engine globals (LÖVE, Corona, custom engine) with `declare const` and full type signatures for type-safe interop. Use string-literal unions for enum-like params (e.g. `"fill" | "line"`) and nested object types matching the Lua table layout.

```typescript
declare const love: {
  graphics: {
    draw: (drawable: Drawable, x: number, y: number) => void;
    setColor: (r: number, g: number, b: number, a?: number) => void;
  };
  timer: {getDelta: () => number};
};

love.graphics.setColor(1, 0, 0);
const dt = love.timer.getDelta();
```
