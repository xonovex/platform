# tagged-unions: Tagged Union Pattern

**Guideline:** Type enum + union for polymorphism without virtual functions. Cache-friendly, compiler optimizable.

**Rationale:** Avoids virtual function overhead and pointer chasing; enables compiler exhaustiveness checks and efficient batch processing.

**Example:**

```c
typedef enum {
    SHAPE_SPHERE, SHAPE_BOX, SHAPE_CAPSULE,
} shape_type_t;

typedef struct {
    shape_type_t type;
    union {
        struct { float radius; } sphere;
        struct { vec3f_t half_extents; } box;
        struct { float radius, height; } capsule;
    };
} shape_t;

float shape_volume(const shape_t *s) {
    switch (s->type) {
        case SHAPE_SPHERE:
            return (4.0f/3.0f) * PI * s->sphere.radius * s->sphere.radius * s->sphere.radius;
        case SHAPE_BOX:
            return 8 * s->box.half_extents.x * s->box.half_extents.y * s->box.half_extents.z;
        case SHAPE_CAPSULE:
            return PI * s->capsule.radius * s->capsule.radius * s->capsule.height;
    }
}
```

**Techniques:**
- Type enum: Define all variant types in single enum for exhaustiveness
- Union members: Create named structs within union for each type variant
- Switch dispatch: Use switch on type field with no default clause
- Designated init: Use `.type = SHAPE_BOX, .box = {...}` for clarity
- Small members: Keep union data small; use pointers for large heap data
