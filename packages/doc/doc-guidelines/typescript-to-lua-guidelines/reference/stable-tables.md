# stable-tables: Stable Table Patterns

**Guideline:** Define all table fields upfront to enable LuaJIT optimizations.

**Rationale:** LuaJIT performs best with constant table shapes. Adding fields dynamically causes deoptimization. Pre-allocating fields enables optimal JIT compilation.

**Example:**

```typescript
// SLOW: Dynamic property addition (deoptimizes JIT)
const entity = {x: 0, y: 0};
entity.vx = 1; // Added after creation - bad for JIT
entity.vy = 1; // Another dynamic addition
entity.id = 123; // Yet another - very slow

// FAST: All fields defined upfront
interface Entity {
  readonly x: number;
  readonly y: number;
  readonly vx: number;
  readonly vy: number;
  readonly id: number;
}

function createEntity(x: number, y: number, id: number): Entity {
  return {
    x: x,
    y: y,
    vx: 0, // All fields initialized
    vy: 0,
    id: id,
  };
}

// SLOW: Variable-sized arrays
const buffers: number[][] = [];
function addData(id: number, data: number[]) {
  buffers[id] = data; // Variable size - bad for JIT
}

// FAST: Pre-allocated arrays
const maxEntities = 1000;
const posX = new Array<number>(maxEntities); // Fixed size
const posY = new Array<number>(maxEntities);
const velocities: number[][] = new Array(maxEntities);

for (let i = 0; i < maxEntities; i++) {
  posX[i] = 0;
  posY[i] = 0;
  velocities[i] = [0, 0]; // Stable sub-arrays
}
```

**Techniques:**

- Define complete interfaces with all properties and types upfront
- Initialize all fields in factory functions to establish table shape
- Pre-allocate arrays with known sizes to avoid dynamic growth
- Avoid dynamically adding properties after creation (hurts JIT)
- Use `readonly` modifiers to signal immutability and stability
- Create factory functions that set all fields in one operation
- Keep table structure consistent across all instances of a type
