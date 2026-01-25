# data-oriented-design: Data-Oriented Design

**Guideline:** Organize data structures for cache locality using Struct of Arrays (SoA) instead of Array of Structs (AoS) when processing arrays of data.

**Rationale:** Modern CPUs load data in cache lines. Struct of Arrays layout ensures sequential memory access patterns that maximize cache utilization.

**Example:**

```c
// Array of Structs (AoS) - poor cache locality
struct Entity {
    float x, y, z;      // Position
    float vx, vy, vz;   // Velocity
    int health;
    int type;
};

struct Entity entities[1000];
for (size_t i = 0; i < 1000; i++) {
    entities[i].x += entities[i].vx;
}

// Struct of Arrays (SoA) - better cache locality
struct EntitySystem {
    float *positions_x, *positions_y, *positions_z;
    float *velocities_x, *velocities_y, *velocities_z;
    int *health, *type;
    size_t count;
};

// Sequential memory access, better cache usage
for (size_t i = 0; i < system.count; i++) {
    system.positions_x[i] += system.velocities_x[i];
}
```

**Techniques:**
- Separate fields: Move frequently accessed fields into their own arrays
- Cache locality: Sequential memory access maximizes CPU cache utilization
- Array sync: Keep indices synchronized across all component arrays
- Hot loops: Identify performance-critical loops for SoA conversion
- Measurement: Profile before and after refactoring to verify improvements
