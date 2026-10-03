# physics-patterns: Physics Engine Patterns

Store `inv_mass` (1/mass); `0` means static and naturally zeroes any applied impulse. Track state in a `uint32_t` bitmask (`|=` set, `&= ~` clear, `&` test) for flags like ACTIVE/SLEEPING. Decouple physics arrays from entity arrays with index sync maps. Split the update loop into integration → broadphase → narrowphase.

```c
typedef struct {
    vec3f_t position, velocity;
    float inv_mass;  // 0 = static (never moves)
} rigidbody_t;

void apply_impulse(rigidbody_t *rb, vec3f_t impulse) {
    rb->velocity.x += impulse.x * rb->inv_mass;
    rb->velocity.y += impulse.y * rb->inv_mass;
    rb->velocity.z += impulse.z * rb->inv_mass;
}
```
