# timing: Timing and Frame Calculations

Write every duration as `seconds * fps` (e.g. `0.1 * fps` = 100ms) for frame-accuracy; `interpolate()` ranges are in frames, not seconds. `durationInFrames = totalSeconds * fps`; stagger delays as `index * delaySeconds * fps`. Clamp edges with `extrapolateLeft`/`extrapolateRight`.

```tsx
const startFrame = 1 * fps; // 1 second
const fadeIn = interpolate(frame, [0, 0.5 * fps], [0, 1], {
  extrapolateLeft: "clamp",
});
```
