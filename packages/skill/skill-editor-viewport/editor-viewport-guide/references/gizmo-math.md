# gizmo-math: The Linear Algebra Behind Robust Gizmos

### Guideline

Do gizmo drag geometry in screen space: project the gizmo's axis to 2D, project the 2D cursor onto that 2D line, then lift the result back to world space and intersect with the world axis — never build a 3D "mouse ray" from an arbitrary cursor z and intersect it with the axis. Drive the object by the _delta_ of the projected parameter from drag-start, and guard every division with an epsilon so near-parallel axes and zero-length directions cannot explode.

### How to Apply

1. On drag start, compute the projected parameter once and store it as `axis_start`.
2. Each frame, recompute the projected parameter `s` for the current cursor and apply the _delta_ `ds = s - axis_start` along the axis.
3. To project: build two world points on the gizmo axis, transform both to screen space (`world_to_screen`), project the 2D cursor onto that 2D screen-space line, then `screen_to_world` the projected 2D point into a world ray.
4. Intersect that world ray with the world gizmo axis (line-line) to get the world parameter; multiply the axis direction by `ds` for the world delta.
5. For local-space dragging, rotate the world delta by the inverse of the parent rotation before adding it to the stored local position.
6. Scale handle geometry by distance from the camera so it keeps a fixed pixel size on screen.
7. Guard every helper: bail to a no-op (return the base point, or `{0,0}`) when a denominator is below epsilon.

### Example

```c
// Per-frame axis drag: apply the DELTA from the stored start, not the absolute s.
vec2_t st = gizmo_axis_intersection(world_pos, axis, cursor_px, camera, viewport);
float  ds = st.x - axis_start;                 // start captured on mouse-down
vec3_t world_delta = vec3_mul(axis, ds);
vec3_t local_delta = quat_rotate(parent_rot_inv, world_delta); // for local mode
local->pos = vec3_add(local_start, local_delta);

// 2D point-on-line projection: closest point on A + t*U to P. Epsilon guards |U|~0.
vec2_t point_on_line_2d(vec2_t a, vec2_t u, vec2_t p) {
  const float uu = vec2_dot(u, u);
  if (uu < 1e-5f) return a;                     // degenerate direction -> no move
  return vec2_add(a, vec2_mul(u, vec2_dot(u, vec2_sub(p, a)) / uu));
}

// 3D line-line "intersection": params (s,t) minimizing distance between
// A + s*U and B + t*V (U,V normalized). Bail when nearly parallel.
vec2_t line_line(vec3_t a, vec3_t u, vec3_t b, vec3_t v) {
  const float uv = vec3_dot(u, v);
  if (uv * uv > 1.0f - 1e-5f) return (vec2_t){0, 0}; // parallel -> 1-uv^2 ~ 0
  const float au = vec3_dot(a,u), bu = vec3_dot(b,u);
  const float av = vec3_dot(a,v), bv = vec3_dot(b,v);
  const float denom = 1.0f - uv*uv;
  const float s = (bu + av*uv - bv*uv - au) / denom;
  const float t = (av + bu*uv - au*uv - bv) / denom;
  return (vec2_t){s, t};
}
```

```c
// The helper: do the projection in 2D screen space, then lift back to world.
vec2_t gizmo_axis_intersection(vec3_t world_pos, vec3_t axis, vec2_t cursor_px,
                               const camera_t *cam, const viewport_t *vp) {
  // 1. two world points on the axis -> screen space
  vec2_t a0 = world_to_screen(cam, vp, world_pos);
  vec2_t a1 = world_to_screen(cam, vp, vec3_add(world_pos, axis));
  // 2. project the 2D cursor onto the 2D screen-space axis
  vec2_t p2 = point_on_line_2d(a0, vec2_sub(a1, a0), cursor_px);
  // 3. lift the projected 2D point back to a world-space ray
  vec3_t ray_o, ray_d; screen_to_world_ray(cam, vp, p2, &ray_o, &ray_d);
  // 4. intersect that world ray with the world axis -> param along the axis
  return line_line(world_pos, axis, ray_o, ray_d);
}
```

### Gotchas

- The 3D-mouse-ray approach (set cursor z to 0/1, intersect with the axis) is a skew-line distance minimization, not a projection; it drifts and sends the object backward at steep angles — project in 2D screen space instead.
- Applying the absolute projected parameter as the position snaps the object to the cursor on frame one; always subtract the drag-start parameter and apply only the delta.
- The line-line denominator `1 - (u·v)^2` goes to zero as the axis nears parallel with the view ray; guard with `uv*uv > 1 - 1e-5f` and return a no-op.
- The 2D projection divides by `u·u`; a near-zero-length screen-space axis (axis seen edge-on) makes it blow up — guard with `uu < 1e-5f`.
- Forgetting perspective divide (`x/=w, y/=w, z/=w`) when going to/from clip space puts the screen-space line in the wrong place and the projection silently drifts.
- A handle that is not distance-scaled shrinks to sub-pixel at distance and becomes unclickable; scale geometry by camera distance for constant on-screen size.
- Skipping the inverse-parent-rotation step in local mode applies a world delta to a local position and the object slides along the wrong axes under a rotated parent.

### Related

[references/manipulation-gizmos.md](./manipulation-gizmos.md), **c99-game-opinionated-guide**
