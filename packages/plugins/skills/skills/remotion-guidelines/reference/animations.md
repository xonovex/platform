# animations: Frame-Driven Animation

**Guideline:** Drive all animations with `useCurrentFrame()` and `interpolate()` or `spring()` functions, never CSS transitions.

**Rationale:** Remotion captures each frame as an image; CSS transitions and Tailwind animations won't render in the video output.

**Example:**

```tsx
// Simple fade-in over 1 second (30fps = 30 frames)
const frame = useCurrentFrame();
const opacity = interpolate(frame, [0, 30], [0, 1], {
  extrapolateRight: "clamp",
});
return <div style={{opacity}}>Fades in</div>;

// Spring bounce entrance
const scale = spring({frame, fps: 30, config: {damping: 8}});
return <div style={{transform: `scale(${scale})`}}>Bounces in</div>;

// Bad: Won't animate in Remotion
<div className="animate-bounce">Won't work</div>
```

**Linear Interpolation:**

```tsx
import {interpolate, useCurrentFrame, useVideoConfig} from "remotion";

function FadeIn() {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  // Fade in over 0.5 seconds
  const opacity = interpolate(frame, [0, 0.5 * fps], [0, 1], {
    extrapolateRight: "clamp",
  });

  return <div style={{opacity}}>Content</div>;
}
```

**Spring Animation:**

```tsx
import {spring, useCurrentFrame, useVideoConfig} from "remotion";

function BounceIn() {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();

  const scale = spring({
    frame,
    fps,
    config: {damping: 10, stiffness: 100},
  });

  return <div style={{transform: `scale(${scale})`}}>Content</div>;
}
```

**Spring Presets:**

```tsx
// Smooth, no bounce (subtle reveals)
const smooth = {damping: 200};

// Snappy, minimal bounce (UI elements)
const snappy = {damping: 20, stiffness: 200};

// Bouncy entrance (playful animations)
const bouncy = {damping: 8};

// Heavy, slow motion
const heavy = {damping: 15, stiffness: 80, mass: 2};
```

**Easing Functions:**

```tsx
import {Easing, interpolate, useCurrentFrame} from "remotion";

function EasedMove() {
  const frame = useCurrentFrame();

  const x = interpolate(frame, [0, 30], [0, 200], {
    easing: Easing.inOut(Easing.ease),
    extrapolateRight: "clamp",
  });

  // Available easings: Easing.in, Easing.out, Easing.inOut
  // Curves: Easing.ease, Easing.quad, Easing.sin, Easing.exp, Easing.circle
  // Custom: Easing.bezier(0.25, 0.1, 0.25, 1)

  return <div style={{transform: `translateX(${x}px)`}}>Moving</div>;
}
```

**Bad vs. Good:**

```tsx
// Bad: CSS transitions won't render
<div className="transition-opacity duration-500 opacity-0 hover:opacity-100">Won't animate</div>

// Bad: Tailwind animation
<div className="animate-bounce">Won't work</div>

// Good: Frame-driven
const frame = useCurrentFrame();
const opacity = interpolate(frame, [0, 30], [0, 1]);
<div style={{opacity}}>Renders correctly</div>
```

**Techniques:**
- useCurrentFrame(): Get current frame number from video composition
- interpolate(): Map frame ranges to value ranges with extrapolation control
- spring(): Apply physics-based motion with damping, stiffness, mass
- Spring presets: smooth, snappy, bouncy, heavy for different effects
- Easing functions: Apply easing curves (ease, quad, sin, exp, circle, bezier)
- No CSS transitions: Avoid transition classes, they won't render
- No Tailwind animations: animate-bounce, etc. won't work in Remotion
