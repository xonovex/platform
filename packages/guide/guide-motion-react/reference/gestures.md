# gestures: Gesture Interactions

**Guideline:** Use whileHover/whileTap/whileFocus with spring physics; never use duration for gestures (springs feel responsive).

**Rationale:** Spring transitions feel instantaneous and natural; duration-based easing delays feedback.

**Example:**

```tsx
<motion.button
  whileHover={{scale: 1.05}}
  whileTap={{scale: 0.95}}
  whileFocus={{boxShadow: "0 0 0 3px rgba(66, 153, 225, 0.6)"}}
  transition={{type: "spring", stiffness: 400, damping: 25}}
/>
```

**Techniques:**

- whileHover/whileTap/whileFocus: Gesture states; always use spring transitions
- Spring presets: Snappy (buttons) 400/25, smooth (cards) 300/20, bouncy 200/10
- Hover effects: y: -8, boxShadow, scale 1.02-1.05
- Magnetic button: useMotionValue + useSpring for cursor tracking
- Dock scaling: Distance-based proximity effect via Math.abs(index - hovered)
- Accessibility: Include whileFocus for keyboard users; test with Tab key
