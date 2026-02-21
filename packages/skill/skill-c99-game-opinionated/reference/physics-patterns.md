# physics-patterns: Physics Engine Patterns

**Guideline:** Use inverse mass (0 = static), bit flags for state, sync maps for decoupled systems.

**Rationale:** Inverse mass simplifies static bodies and impulse application; flags enable efficient state management; sync maps decouple physics from entity systems.

**Example:**

```c
typedef struct {
    vec3f_t position, velocity;
    float inv_mass;  // 0 = static (never moves)
} rigidbody_t;

// Automatically handles static (inv_mass=0 zeroes impulse)
void apply_impulse(rigidbody_t *rb, vec3f_t impulse) {
    rb->velocity.x += impulse.x * rb->inv_mass;
    rb->velocity.y += impulse.y * rb->inv_mass;
    rb->velocity.z += impulse.z * rb->inv_mass;
}
```

**Techniques:**

- Inverse mass: Store `inv_mass` (1/mass); zero for static bodies to disable movement
- Bit flags: Use `uint32_t` with bitmask enums for state like ACTIVE, SLEEPING
- Sync maps: Index maps decouple physics arrays from entity system arrays
- Flag operations: Use `|=` (set), `&= ~` (clear), `&` (test) for bit manipulation
- Update loop: Separate integration, broadphase detection, and narrowphase solving
