# filename: spring-physics

**Guideline:** Use `type: "spring"` for interactive gestures; adjust `stiffness` (speed), `damping` (bounce), `mass` (weight) for feel.

**Rationale:** Spring animations are interruptible and natural; duration animations are rigid but synchronized; match animation type to use case.

**Example:**

```tsx
// Snappy (buttons, quick interactions)
<motion.div
  whileHover={{scale: 1.05}}
  transition={{type: "spring", stiffness: 400, damping: 25}}
/>

// Bouncy (playful entrance)
<motion.div
  initial={{scale: 0.9}}
  animate={{scale: 1}}
  transition={{type: "spring", stiffness: 200, damping: 8}}
/>

// Smooth (cards, panels)
<motion.div
  whileHover={{y: -8}}
  transition={{type: "spring", stiffness: 300, damping: 20}}
/>
```

**Techniques:**

- `stiffness: 100-500`: Response speed; higher = faster snap to target
- `damping: 10-50`: Bounce control; higher = less bouncy/overshoot
- `mass: 0.5-3`: Perceived weight; higher = heavier, slower response
- Snappy preset: `stiffness: 400, damping: 25` for buttons, quick feedback
- Smooth preset: `stiffness: 300, damping: 20` for regular elements
- Bouncy preset: `stiffness: 200, damping: 8` for attention-grabbing/playful
- Heavy preset: `stiffness: 100, damping: 15, mass: 2` for modals, large elements
- Gentle preset: `stiffness: 150, damping: 15` for subtle, understated movements
- Use `useSpring()` hook for smoothing raw motion values (mouse tracking, scroll)
- Spring for: Gestures, hover, tap, drag interactions, interactive feedback, layout changes
- Duration for: Entrance sequences, scroll reveals, page transitions, synchronized multi-element animations
