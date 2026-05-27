# hdr-output: HDR Output and Display Color Spaces

**Guideline:** To drive an HDR display, present to a wide-gamut swapchain — PQ-encoded Rec.2020 at ≥10-bit, or extended-linear scRGB at FP16 — keep the entire scene pipeline in linear, scene-referred light, and convert to the display's primaries and transfer function yourself in one final pass scaled to the display's real peak luminance. Never rely on an automatic sRGB backbuffer for HDR.

**Rationale:** SDR bakes in three assumptions — Rec.709 primaries, 8-bit storage, and an sRGB transfer — none of which hold for HDR. HDR needs wider primaries (Rec.2020), at least 10-bit (or FP16) to avoid visible banding across the larger range, and a transfer function (PQ/ST2084 or HLG) that maps code values to _absolute_ luminance in nits. Two things force you to own the final encode rather than delegate to a hardware sRGB backbuffer: blending must happen in linear light to be physically correct, but an auto-transfer backbuffer blends _after_ the curve (wrong, and it mismatches what users expect from OS-native UI text); and you must map the signal to the display's actual peak nits, because almost no panel reaches PQ's 10,000-nit ceiling. Doing linear → primaries → transfer in a single explicit pass keeps blending linear, applies the transfer exactly once, and gives you the paper-white scaling knob.

**How to Apply:**

1. Query the surface's supported color spaces/formats; detect HDR capability rather than assuming it.
2. Choose a swapchain config: PQ + a 10-bit-per-channel packed format (e.g. `A2B10G10R10`) for HDR10, or extended-linear + `RGBA16F` for scRGB. Use 8-bit sRGB only for the SDR path.
3. Render the scene in linear light into an FP16 target; do lighting, post, color grading, and tone mapping there.
4. Composite UI in the same linear space — convert sRGB-authored UI colors to linear exactly once on the way in.
5. Final output pass: linear → CIE-XYZ → output primaries via a 3×3 matrix → apply the display's transfer function (sRGB, or PQ for HDR10).
6. Scale luminance by a paper-white / reference-white level expressed against the display's measured peak nits, not PQ's 10,000-nit maximum.

**Example:**

```glsl
// Final encode for an HDR10 (Rec.2020 + PQ) swapchain. Input is linear scene light.
const mat3 REC709_TO_REC2020 = /* row-major source transposed for column-major */ ...;

vec3 pq_oetf(vec3 L) {                 // L in [0,1] normalized to display peak
    const float m1 = 0.1593017578125, m2 = 78.84375, c1 = 0.8359375,
                c2 = 18.8515625, c3 = 18.6875;
    vec3 Lm = pow(L, vec3(m1));
    return pow((c1 + c2 * Lm) / (1.0 + c3 * Lm), vec3(m2));
}

vec3 encode_hdr10(vec3 linear_rec709, float paper_white_nits, float peak_nits) {
    vec3 rec2020 = REC709_TO_REC2020 * linear_rec709;     // gamut map
    vec3 nits    = rec2020 * paper_white_nits;            // SDR "white" anchored in nits
    return pq_oetf(clamp(nits / 10000.0, 0.0, peak_nits / 10000.0));
}
```

**Gotchas:**

- Color-conversion matrices published online are usually row-major; a column-major renderer must transpose them or every color shifts.
- Applying a transfer function twice (e.g. an sRGB backbuffer on top of your own encode) loses precision — round-tripping sRGB→linear→sRGB in FP visibly errors in the darkest code values. Encode exactly once.
- Feeding PQ the unscaled 10,000-nit range makes everything look washed out and dim; scale to the panel's real peak.
- An automatic sRGB backbuffer cannot be used for correct linear UI blending and does not match OS-native text rendering — own the composite.
- 8-bit storage with Rec.2020 primaries bands badly; 10-bit is the floor for PQ output.
- Calibration needs measured hardware (colorimeter); unit-test the conversion matrices and GPU readbacks for correctness, but trust a meter for final tuning.

**Related:** [references/shader-system.md](./shader-system.md), [references/command-recording-and-frames.md](./command-recording-and-frames.md), **gpu-rendering-vulkan-guide**
