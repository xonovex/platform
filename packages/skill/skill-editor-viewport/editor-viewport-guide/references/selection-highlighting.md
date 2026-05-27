# selection-highlighting: Outline Rendering by Id Edge Detection

**Guideline:** Highlight the selection with a crisp constant-width outline by rendering selected objects' ids into a separate single-channel target, then edge-detecting that id target in one fullscreen pass after post-processing — comparing each pixel's id to its neighbors' to find silhouette edges, and comparing selection depth to scene depth so an occluded outline is dimmed rather than dropped.

**Rationale:** A good selection outline must be readable, crisp, anti-aliased, a constant one pixel wide, composited after post-processing, still visible (dimmed) when the selected object is hidden behind unselected geometry, and cheap for any shader to opt into. Writing a selection id into a dedicated render target and doing edge detection on it satisfies all of these at once: edges fall exactly on silhouette boundaries (where the id changes), width is controlled by the sampling kernel rather than the geometry, and the pass runs late so it overlays the final image. The hard part is occlusion. Comparing the selection depth directly against the scene depth would be correct only if the depth buffers were stable — but with temporal anti-aliasing every frame applies a different sub-pixel jitter, so a direct compare makes the outline shimmer. Taking the closest selection depth over a small neighborhood before comparing both smooths jitter and lets the outline bleed slightly over occluders, which reads as a deliberate, soft hidden-outline rather than noise.

**How to Apply:**

1. Maintain a selection set of ids and pass it to renderable objects; in the inner loop, if the owning object's id is in the set, enable a shared `selection` shader feature.
2. That feature renders the object again into a single-channel selection target, writing a unique id (e.g. the low 8 bits of the entity id).
3. In a fullscreen pass, sample a 3x3 neighborhood of the id target with `Gather` (fetch a 2x2 block per instruction in counter-clockwise XYZW order) instead of nine point samples.
4. Compute alpha from how many neighbors differ from the center id (`dot(neighbor != center, 1/8)` accumulated) — the more they differ, the closer to an edge.
5. For occlusion, take the closest selection depth over a small neighborhood (with reverse-Z, the `max` of taps), linearize it, and compare to the resolved scene depth.
6. If the outline pixel is behind scene geometry, multiply its alpha down (e.g. to 0.3) rather than discarding it; early-out when alpha is zero.
7. Composite the outline color after post-processing in sRGB.

**Example:**

```hlsl
// Edge detect on the id target: alpha rises where neighbor ids differ from center.
float4 i0 = sel_id.GatherRed(clamp_pt, uv, int2(-2,-2));
float4 i1 = sel_id.GatherRed(clamp_pt, uv, int2( 0,-2));
float4 i2 = sel_id.GatherRed(clamp_pt, uv, int2(-2, 0));
float4 i3 = sel_id.GatherRed(clamp_pt, uv, int2( 0, 0));
i2.xw = i1.xy; float id_c = i3.w; i3.w = i0.y;          // arrange the 3x3 ring
float a  = dot(float4(i2 != id_c), 1.0/8.0);
      a += dot(float4(i3 != id_c), 1.0/8.0);

// Occlusion: closest selection depth over a neighborhood (reverse-Z -> max),
// linearized, vs resolved scene depth. Dim, don't drop, when hidden.
float4 d0 = sel_depth.GatherRed(clamp_pt, uv, int2(-2,-2));
/* ... gather d1,d2,d3 ... */
float d = max(max(maxc(d0), maxc(d1)), max(maxc(d2), maxc(d3)));
float scene = scn_depth.Sample(clamp_pt, uv).r;
float2 nf = camera_near_far();
bool visible = linearize_depth(d, nf.x, nf.y) <= scene;
a *= visible ? 1.0 : 0.3;                                // hidden outline dimmed
if (a == 0.0) discard;
```

**Gotchas:**

- Direct selection-depth-vs-scene-depth comparison shimmers under TAA because the depth buffers carry a per-frame sub-pixel jitter; take the closest depth over a small neighborhood first.
- Reverse-Z means the closest surface is the maximum depth value, so use `max` over the depth taps, not `min`.
- The selection depth is in the (possibly non-linear, pre-resolve) selection buffer while scene depth is resolved/linear; linearize before comparing or the dim test is wrong.
- Issuing nine individual point samples is wasteful; one `Gather` returns a 2x2 block, so a few gathers cover the kernel — but remember the XYZW tap order (counter-clockwise) when reassembling the ring.
- Discarding instead of dimming occluded outline pixels loses the "object is behind something" cue; multiply alpha down instead.
- Skipping the zero-alpha early-out wastes the whole composite where there is no edge.
- A single-channel id target only distinguishes silhouettes, not colors; per-selection-state colors require writing a color id and doing a weighted neighborhood sum.

**Related:** [references/object-picking.md](./object-picking.md), [references/render-editor-integration.md](./render-editor-integration.md), **gpu-rendering-guide**
