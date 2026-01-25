# coordinate-system: Coordinate System Conventions

**Guideline:** Right-handed Y-up coordinates, CCW winding. Use suffixes `_rh`/`_lh` for explicit handedness. Override with defines.

**Rationale:** Consistent handedness and winding across geometry reduces bugs and ensures interoperability with standard graphics APIs.

**Example:**

```c
// Default (right-handed)
matrix4f_t proj = matrix4f_perspective(fov, aspect, near, far);
matrix4f_t view = matrix4f_look_at(eye, target, up);

// Explicit or Vulkan setup
matrix4f_t view_rh = matrix4f_look_at_rh(eye, target, up);
matrix4f_t view_lh = matrix4f_look_at_lh(eye, target, up);

#define MATH_COORDINATE_SYSTEM_LEFT_HANDED 1
#define MATH_CLIP_SPACE_DEPTH_ZERO_TO_ONE 1
```

**Techniques:**
- Right-handed default: Use +X right, +Y up, +Z forward as base convention
- CCW winding: Define front faces with counter-clockwise vertex order
- Suffix variants: Provide `_rh` and `_lh` function suffixes for explicit systems
- Compile-time overrides: Use defines like `MATH_COORDINATE_SYSTEM_LEFT_HANDED`
- Normal calculation: Cross product respects handedness for correct face normals
