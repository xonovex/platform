# animations: Frame-Driven Animation

Drive all motion with `useCurrentFrame()` + `interpolate()` or `spring()`. CSS transitions and Tailwind `animate-*` classes never render — Remotion captures each frame as a still.

```tsx
import {
  Easing,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from "remotion";

const frame = useCurrentFrame();
const {fps} = useVideoConfig();

const opacity = interpolate(frame, [0, 0.5 * fps], [0, 1], {
  extrapolateRight: "clamp",
});
const x = interpolate(frame, [0, 30], [0, 200], {
  easing: Easing.inOut(Easing.ease),
  extrapolateRight: "clamp",
});
const scale = spring({frame, fps, config: {damping: 10, stiffness: 100}});
```

- **Spring presets** (`config`): smooth `{damping: 200}` · snappy `{damping: 20, stiffness: 200}` · bouncy `{damping: 8}` · heavy `{damping: 15, stiffness: 80, mass: 2}`
- **Easings**: `Easing.in`/`out`/`inOut`; curves `Easing.ease`/`quad`/`sin`/`exp`/`circle`; custom `Easing.bezier(0.25, 0.1, 0.25, 1)`
