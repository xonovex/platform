# gestures: Hover, Tap, Focus

Use `whileHover`/`whileTap`/`whileFocus` with spring transitions, never `duration` (springs give instant feedback).

```tsx
<motion.button
  whileHover={{scale: 1.05}}
  whileTap={{scale: 0.95}}
  whileFocus={{boxShadow: "0 0 0 3px rgba(66, 153, 225, 0.6)"}}
  transition={{type: "spring", stiffness: 400, damping: 25}}
/>
```

- Spring presets: snappy (buttons) 400/25, smooth (cards) 300/20, bouncy 200/10
- Hover effects: `y: -8`, `boxShadow`, `scale: 1.02-1.05`
- Magnetic button / cursor tracking: `useMotionValue` + `useSpring`
- Dock scaling: proximity effect via `Math.abs(index - hovered)`
- Always include `whileFocus` for keyboard users
