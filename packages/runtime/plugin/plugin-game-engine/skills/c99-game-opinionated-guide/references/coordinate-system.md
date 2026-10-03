# coordinate-system: Coordinate System Conventions

Default to right-handed (+X right, +Y up, +Z forward), CCW front-face winding. Provide `_rh`/`_lh` function suffixes for explicit handedness; cross products respect handedness for normals. Override the defaults at compile time.

```c
matrix4f_t proj = matrix4f_perspective(fov, aspect, near, far);
matrix4f_t view = matrix4f_look_at(eye, target, up);
matrix4f_t view_rh = matrix4f_look_at_rh(eye, target, up);
matrix4f_t view_lh = matrix4f_look_at_lh(eye, target, up);

#define MATH_COORDINATE_SYSTEM_LEFT_HANDED 1
#define MATH_CLIP_SPACE_DEPTH_ZERO_TO_ONE 1
```
