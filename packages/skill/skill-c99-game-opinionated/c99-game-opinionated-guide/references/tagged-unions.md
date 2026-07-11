# tagged-unions: Tagged Union Pattern

Type enum + union for polymorphism without vtables or pointer chasing. Dispatch with a `switch` on the type field and **no `default` clause** so the compiler flags unhandled variants. Use designated init (`.type = SHAPE_BOX, .box = {...}`); keep union members small, pointing to heap for large data.

```c
typedef enum { SHAPE_SPHERE, SHAPE_BOX, SHAPE_CAPSULE } shape_type_t;

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
        case SHAPE_SPHERE:  return (4.0f/3.0f) * PI * s->sphere.radius * s->sphere.radius * s->sphere.radius;
        case SHAPE_BOX:     return 8 * s->box.half_extents.x * s->box.half_extents.y * s->box.half_extents.z;
        case SHAPE_CAPSULE: return PI * s->capsule.radius * s->capsule.radius * s->capsule.height;
    }
}
```
