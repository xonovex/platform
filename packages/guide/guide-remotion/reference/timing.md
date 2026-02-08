# timing: Timing and Frame Calculations

**Guideline:** Always write durations as seconds multiplied by fps for consistency; use interpolate with frame ranges, not time.

**Rationale:** Seconds \* fps pattern is clear and frame-accurate; consistent approach prevents calculation errors.

**Example:**

```tsx
const frame = useCurrentFrame();
const {fps} = useVideoConfig();
const startFrame = 1 * fps; // 1 second
const fadeIn = interpolate(frame, [0, 0.5 * fps], [0, 1], {
  extrapolateLeft: "clamp",
});
```

**Techniques:**

- Pattern: `seconds * fps` for all durations (0.1 _ fps = 100ms, 1 _ fps = 1s)
- interpolate() ranges in frames: [startFrame, endFrame] not seconds
- Composition: durationInFrames = totalSeconds \* fps
- Stagger delays: index _ delaySeconds _ fps
- Visibility check: frame >= startFrame && frame < endFrame
- Extrapolation: Use extrapolateLeft/extrapolateRight to clamp values at edges
- Common: Flash 0.1s, fade 0.5s, transitions 1-2s
